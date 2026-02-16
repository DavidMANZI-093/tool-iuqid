package liquid

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
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
		util.LogDebug("Extracted nonce=%v, token=%v, pubkeyFound=%v", nonce != "", token != "", pubkeyPEM != "")
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

	// TODO: Continue
}

func extractVar(body, pattern string) string {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(body)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
