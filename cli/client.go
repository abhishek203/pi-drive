package cli

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

type Credentials struct {
	APIKey string
	Server string
	Mount  string
}

type Client struct {
	creds *Credentials
	http  *http.Client
}

func credentialsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".pidrive", "credentials")
}

func parseCredentials(data []byte) (*Credentials, error) {
	creds := &Credentials{}
	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return nil, fmt.Errorf("invalid line %q", line)
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		unquoted, err := strconv.Unquote(value)
		if err != nil {
			return nil, fmt.Errorf("invalid value for %s: %w", key, err)
		}

		switch key {
		case "api_key":
			creds.APIKey = unquoted
		case "server":
			creds.Server = unquoted
		case "mount_path":
			creds.Mount = unquoted
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return creds, nil
}

func encodeCredentials(creds *Credentials) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "api_key = %s\n", strconv.Quote(creds.APIKey))
	fmt.Fprintf(&b, "server = %s\n", strconv.Quote(creds.Server))
	fmt.Fprintf(&b, "mount_path = %s\n", strconv.Quote(creds.Mount))
	return []byte(b.String())
}

func LoadCredentials() (*Credentials, error) {
	path := credentialsPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("not logged in. Run 'pidrive register' or 'pidrive login' first")
	}

	creds, err := parseCredentials(data)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials file: %w", err)
	}
	return creds, nil
}

func SaveCredentials(creds *Credentials) error {
	path := credentialsPath()
	os.MkdirAll(filepath.Dir(path), 0700)
	return os.WriteFile(path, encodeCredentials(creds), 0600)
}

func NewClient() (*Client, error) {
	creds, err := LoadCredentials()
	if err != nil {
		return nil, err
	}
	return &Client{
		creds: creds,
		http:  &http.Client{},
	}, nil
}

func NewClientWithServer(server string) *Client {
	return &Client{
		creds: &Credentials{Server: server},
		http:  &http.Client{},
	}
}

func (c *Client) Server() string {
	return c.creds.Server
}

func (c *Client) MountPath() string {
	if c.creds.Mount != "" {
		return c.creds.Mount
	}
	if runtime.GOOS == "darwin" {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "drive")
	}
	return "/drive"
}

func (c *Client) doRequest(method, path string, body interface{}) (*http.Response, error) {
	url := c.creds.Server + path

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.creds.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.creds.APIKey)
	}

	return c.http.Do(req)
}

func (c *Client) Get(path string) (map[string]interface{}, error) {
	resp, err := c.doRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return c.parseResponse(resp)
}

func (c *Client) Post(path string, body interface{}) (map[string]interface{}, error) {
	resp, err := c.doRequest("POST", path, body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return c.parseResponse(resp)
}

func (c *Client) Delete(path string) (map[string]interface{}, error) {
	resp, err := c.doRequest("DELETE", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return c.parseResponse(resp)
}

func (c *Client) parseResponse(resp *http.Response) (map[string]interface{}, error) {
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if resp.StatusCode >= 400 {
		msg := "request failed"
		if e, ok := result["error"].(string); ok {
			msg = e
		}
		return nil, fmt.Errorf("%s (HTTP %d)", msg, resp.StatusCode)
	}

	return result, nil
}

func (c *Client) DownloadFile(url, destPath string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	if c.creds.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.creds.APIKey)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("download failed (HTTP %d)", resp.StatusCode)
	}

	os.MkdirAll(filepath.Dir(destPath), 0755)
	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}
