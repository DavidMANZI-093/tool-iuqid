package liquid

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/DavidMANZI-093/tool-iquid/pkg/utils"
)

type Client struct {
	BaseURL    string
	Username   string
	Password   string
	HTTPClient *http.Client
}

func NewClient(baseURL, username, password string, timeout time.Duration) *Client {
	jar, _ := cookiejar.New(nil)
	return &Client{
		BaseURL:  baseURL,
		Username: username,
		Password: password,
		HTTPClient: &http.Client{
			Jar:     jar,
			Timeout: timeout,
			Transport: &http.Transport{
				ForceAttemptHTTP2: false,
				DisableKeepAlives: true,
			},
		},
	}
}

func (c *Client) Login() error {
	utils.LogDebug("Requesting %s/", c.BaseURL)
	req, err := http.NewRequest("GET", c.BaseURL+"/", nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/145.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	req.Header.Set("Cache-Control", "max-age=0")
	req.Header.Set("Connection", "keep-alive")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to get index: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read index.html: %w", err)
	}
	body := string(bodyBytes)

	nonce := extractVar(body, `var nonce\s*=\s*["'](.*?)["'];`)
	token := extractVar(body, `var token\s*=\s*["'](.*?)["'];`)
	pubkeyPEM := extractVar(body, `var pubkey\s*=\s*["']([\s\S]*?)["'];`)

	if nonce == "" && token == "" && pubkeyPEM == "" {
		utils.LogDebug("Failed to extract ANY variables. Page might be different")
	} else {
		utils.LogDebug("Extracted nonce=%v, token=%v, pubkeyFound=%v", nonce != "", token != "", pubkeyPEM != "")
	}

	if nonce == "" || token == "" || pubkeyPEM == "" {
		snippet := body
		if len(snippet) > 1000 {
			snippet = snippet[:1000]
		}
		return fmt.Errorf("failed to extract login variables. Body snippet: %s", snippet)
	}

	pubkeyPEM = strings.ReplaceAll(pubkeyPEM, "\\\n", "")
	pubkeyPEM = strings.ReplaceAll(pubkeyPEM, "\\", "")

	pubkeyPEM = strings.TrimSpace(pubkeyPEM)
	if !strings.Contains(pubkeyPEM, "\n") {
		pubkeyPEM = strings.Replace(pubkeyPEM, "-----BEGIN PUBLIC KEY-----", "-----BEGIN PUBLIC KEY-----\n", 1)
		pubkeyPEM = strings.Replace(pubkeyPEM, "-----END PUBLIC KEY-----", "\n-----END PUBLIC KEY-----", 1)
	}

	utils.LogDebug("Extracted PubKey: %s", pubkeyPEM)

	pubKey, err := ParsePublicKey(pubkeyPEM)
	if err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}

	aesKey, err := GenerateRandomBytes(16)
	if err != nil {
		return err
	}
	iv, err := GenerateRandomBytes(16)
	if err != nil {
		return err
	}

	encodedPassword := url.QueryEscape(c.Password)

	postDataStr := fmt.Sprintf("&username=%s&password=%s&csrf_token=%s&nonce=%s&enckey=%s&enciv=%s",
		c.Username,
		encodedPassword,
		token,
		nonce,
		Base64UrlEscape(base64.StdEncoding.EncodeToString(aesKey)),
		Base64UrlEscape(base64.StdEncoding.EncodeToString(iv)),
	)

	ct, err := EncryptAES(aesKey, iv, []byte(postDataStr))
	if err != nil {
		return fmt.Errorf("aes encrypt failed: %w", err)
	}

	aesKeyInfo := base64.StdEncoding.EncodeToString(aesKey) + " " + base64.StdEncoding.EncodeToString(iv)
	ck, err := EncryptRSA(pubKey, []byte(aesKeyInfo))
	if err != nil {
		return fmt.Errorf("rsa encrypt failed: %w", err)
	}

	finalData := fmt.Sprintf("encrypted=1&ct=%s&ck=%s", ct, ck)

	loginURL := c.BaseURL + "/login.cgi"
	req, err = http.NewRequest("POST", loginURL, strings.NewReader(finalData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Referer", c.BaseURL+"/")
	req.Header.Set("Origin", c.BaseURL)

	resp, err = c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("login request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 299 {
		return nil
	}

	if resp.StatusCode == 200 {
		if resp.Header.Get("X-SID") != "" {
			return nil
		}

		u, _ := url.Parse(c.BaseURL)
		cookies := c.HTTPClient.Jar.Cookies(u)
		for _, cookie := range cookies {
			if cookie.Name == "sid" {
				return nil
			}
		}
	}

	respBody, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(respBody))
}

func extractVar(body, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(body)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

func (c *Client) Reboot() error {
	rebootURL := c.BaseURL + "/reboot.cgi"
	req, err := http.NewRequest("GET", rebootURL, nil)
	if err != nil {
		return err
	}
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to fetch reboot page: %w", err)
	}
	defer resp.Body.Close()
	bodyBytes, _ := io.ReadAll(resp.Body)
	body := string(bodyBytes)

	token := ""
	tokenRe := regexp.MustCompile(`csrf_token=([a-zA-Z0-9]+)`)
	if match := tokenRe.FindStringSubmatch(body); len(match) > 1 {
		token = match[1]
	} else {
		inputRe := regexp.MustCompile(`<input[^>]+name=["']csrf_token["'][^>]+value=["'](.*?)["']`)
		if match := inputRe.FindStringSubmatch(body); len(match) > 1 {
			token = match[1]
		}
	}

	if token == "" {
		utils.LogWarn("Could not extract csrf_token from reboot page. Trying without or using session cookie?")
	}

	pubkeyPEM := ""
	pubkeyRe := regexp.MustCompile(`var pubkey = '([\s\S]*?)';`)
	if match := pubkeyRe.FindStringSubmatch(body); len(match) > 1 {
		pubkeyPEM = match[1]
		pubkeyPEM = strings.ReplaceAll(pubkeyPEM, "\\\n", "")
		pubkeyPEM = strings.ReplaceAll(pubkeyPEM, "\\", "")
	}

	if pubkeyPEM == "" {
		return fmt.Errorf("could not extract public key from reboot page")
	}

	pubkeyPEM = strings.TrimSpace(pubkeyPEM)
	if !strings.Contains(pubkeyPEM, "\n") {
		pubkeyPEM = strings.Replace(pubkeyPEM, "-----BEGIN PUBLIC KEY-----", "-----BEGIN PUBLIC KEY-----\n", 1)
		pubkeyPEM = strings.Replace(pubkeyPEM, "-----END PUBLIC KEY-----", "\n-----END PUBLIC KEY-----", 1)
	}

	plainData := "data"
	if token != "" {
		plainData += "&csrf_token=" + token
	}

	aesKey, err := GenerateRandomBytes(16)
	if err != nil {
		return err
	}
	iv, err := GenerateRandomBytes(16)
	if err != nil {
		return err
	}

	ct, err := EncryptAES(aesKey, iv, []byte(plainData))
	if err != nil {
		return fmt.Errorf("AES encryption failed: %w", err)
	}

	pubKey, err := ParsePublicKey(pubkeyPEM)
	if err != nil {
		return fmt.Errorf("invalid public key: %w", err)
	}

	aesKeyInfo := base64.StdEncoding.EncodeToString(aesKey) + " " + base64.StdEncoding.EncodeToString(iv)
	ck, err := EncryptRSA(pubKey, []byte(aesKeyInfo))
	if err != nil {
		return fmt.Errorf("RSA encryption failed: %w", err)
	}

	finalData := fmt.Sprintf("encrypted=1&ct=%s&ck=%s", ct, ck)

	targetURL := rebootURL + "?reboot"
	utils.LogDebug("Sending ENCRYPTED reboot request to %s (Token: %s)", targetURL, token)

	req, err = http.NewRequest("POST", targetURL, strings.NewReader(finalData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")
	req.Header.Set("Referer", rebootURL)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, err = c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send reboot request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	utils.LogDebug("Reboot Response: %s", string(respBody))

	if resp.StatusCode >= 400 {
		return fmt.Errorf("reboot request returned status: %d", resp.StatusCode)
	}

	return nil
}
