package liquid

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
)

func Base64UrlEscape(b64 string) string {
	b64 = strings.ReplaceAll(b64, "+", "-")
	b64 = strings.ReplaceAll(b64, "/", "_")
	b64 = strings.ReplaceAll(b64, "=", ".")
	return b64
}

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func ParsePublicKey(pemStr string) (*rsa.PublicKey, error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.New("failed to parse PEM block containing the public key")
	}
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		if pub, err = x509.ParsePKCS1PublicKey(block.Bytes); err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}
	}
	switch pub := pub.(type) {
	case *rsa.PublicKey:
		return pub, nil
	default:
		return nil, errors.New("key is not of type RSA pointer")
	}
}

func EncryptRSA(pub *rsa.PublicKey, data []byte) (string, error) {
	enc, err := rsa.EncryptPKCS1v15(rand.Reader, pub, data)
	if err != nil {
		return "", err
	}
	return Base64UrlEscape(base64.StdEncoding.EncodeToString(enc)), nil
}

func PKCS7Padding(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(data, padtext...)
}

func EncryptAES(key, iv, data []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	data = PKCS7Padding(data, block.BlockSize())
	mode := cipher.NewCBCEncrypter(block, iv)
	ciphertext := make([]byte, len(data))
	mode.CryptBlocks(ciphertext, data)

	return base64.RawStdEncoding.EncodeToString(ciphertext), nil
}

func SHA256Str(val1, val2 string) string {
	hash := sha256.Sum256([]byte(val1 + ":" + val2))
	return base64.StdEncoding.EncodeToString(hash[:])
}

func SHA256Url(val1, val2 string) string {
	return Base64UrlEscape(SHA256Str(val1, val2))
}
