package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Credentials struct {
	APIKey string `toml:"api_key"`
	Server string `toml:"server"`
	Mount  string `toml:"mount_path"`
}

type Client struct {
	creds  *Credentials
	http   *http.Client
}

func credentialsPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".pidrive", "credentials")
}

func LoadCredentials() (*Credentials, error) {
	path := credentialsPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("not logged in. Run 'pidrive register' or 'pidrive login' first")
	}

	var creds Credentials
	if err := toml.Unmarshal(data, &creds); err != nil {
		return nil, fmt.Errorf("invalid credentials file: %w", err)
	}
	return &creds, nil
}

func SaveCredentials(creds *Credentials) error {
	path := credentialsPath()
	os.MkdirAll(filepath.Dir(path), 0700)

	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	return toml.NewEncoder(f).Encode(creds)
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
