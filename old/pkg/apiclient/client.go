package apiclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type ApiClient struct {
	BaseURL    string
	httpClient *http.Client
	Headers    map[string]string
}

// NewClient creates a standard API client with optimized HTTP transport settings
func NewClient(baseURL string, timeout time.Duration, headers map[string]string) *ApiClient {
	// Configure HTTP transport with optimized settings for handling slow APIs
	transport := &http.Transport{
		MaxIdleConns:        100,                      // Increase connection pool size
		MaxIdleConnsPerHost: 20,                       // Allow reasonable idle connections per host
		IdleConnTimeout:     timeout + 30*time.Second, // Keep connections alive at least as long as request timeout
	}

	return &ApiClient{
		BaseURL: baseURL,
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		Headers: headers,
	}
}

// NewClientWithInsecureSkipVerify creates a client that ignores invalid SSL certificates with optimized HTTP transport
func NewClientWithInsecureSkipVerify(baseURL string, timeout time.Duration, headers map[string]string) *ApiClient {
	transport := &http.Transport{
		MaxIdleConns:        100,                      // Increase connection pool size
		MaxIdleConnsPerHost: 20,                       // Allow reasonable idle connections per host
		IdleConnTimeout:     timeout + 30*time.Second, // Keep connections alive at least as long as request timeout
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	return &ApiClient{
		BaseURL: baseURL,
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		Headers: headers,
	}
}

// NewClientWithTLSConfig creates a client with custom TLS configuration and optimized HTTP transport
func NewClientWithTLSConfig(baseURL string, timeout time.Duration, headers map[string]string, tlsConfig *tls.Config) *ApiClient {
	transport := &http.Transport{
		MaxIdleConns:        100,                      // Increase connection pool size
		MaxIdleConnsPerHost: 20,                       // Allow reasonable idle connections per host
		IdleConnTimeout:     timeout + 30*time.Second, // Keep connections alive at least as long as request timeout
		TLSClientConfig:     tlsConfig,
	}

	return &ApiClient{
		BaseURL: baseURL,
		httpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		Headers: headers,
	}
}

func (c *ApiClient) doRequest(method, path string, requestBody any, responseBody any) error {
	return c.doRequestWithErrorBinding(method, path, requestBody, responseBody, nil)
}

func (c *ApiClient) doRequestWithErrorBinding(method, path string, requestBody any, responseBody any, errorBody any) error {
	url := fmt.Sprintf("%s%s", c.BaseURL, path)

	var body io.Reader
	if requestBody != nil {
		jsonData, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("error marshaling JSON: %w", err)
		}
		body = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(context.Background(), method, url, body)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}

	if requestBody != nil {
		if req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", "application/json")
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)

		// If errorBody is provided, try to bind the error response to it
		if errorBody != nil {
			if err := json.Unmarshal(bodyBytes, errorBody); err != nil {
				return err
			}
		}

		return fmt.Errorf("%d: %s", resp.StatusCode, string(bodyBytes))
	}

	if responseBody != nil {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error reading response body: %w", err)
		}

		// Only attempt to decode if there's actual content in the response body
		if len(bodyBytes) > 0 {
			if err := json.Unmarshal(bodyBytes, responseBody); err != nil {
				return fmt.Errorf("error decoding response: %w", err)
			}
		}
	}

	return nil
}

func (c *ApiClient) Get(path string, responseBody any) error {
	return c.doRequest(http.MethodGet, path, nil, responseBody)
}

func (c *ApiClient) GetWithErrorBinding(path string, responseBody any, errorBody any) error {
	return c.doRequestWithErrorBinding(http.MethodGet, path, nil, responseBody, errorBody)
}

func (c *ApiClient) Post(path string, requestBody any, responseBody any) error {
	return c.doRequest(http.MethodPost, path, requestBody, responseBody)
}

func (c *ApiClient) PostWithErrorBinding(path string, requestBody any, responseBody any, errorBody any) error {
	return c.doRequestWithErrorBinding(http.MethodPost, path, requestBody, responseBody, errorBody)
}

func (c *ApiClient) Put(path string, requestBody any, responseBody any) error {
	return c.doRequest(http.MethodPut, path, requestBody, responseBody)
}

func (c *ApiClient) PutWithErrorBinding(path string, requestBody any, responseBody any, errorBody any) error {
	return c.doRequestWithErrorBinding(http.MethodPut, path, requestBody, responseBody, errorBody)
}

func (c *ApiClient) Delete(path string, responseBody any) error {
	return c.doRequest(http.MethodDelete, path, nil, responseBody)
}

func (c *ApiClient) DeleteWithErrorBinding(path string, responseBody any, errorBody any) error {
	return c.doRequestWithErrorBinding(http.MethodDelete, path, nil, responseBody, errorBody)
}

func (c *ApiClient) PostForm(path string, formData map[string]string, responseBody any) error {
	return c.PostFormWithErrorBinding(path, formData, responseBody, nil)
}

func (c *ApiClient) PostFormWithErrorBinding(path string, formData map[string]string, responseBody any, errorBody any) error {
	requestURL := fmt.Sprintf("%s%s", c.BaseURL, path)

	// Convert map to form values
	values := url.Values{}
	for k, v := range formData {
		values.Set(k, v)
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, requestURL, strings.NewReader(values.Encode()))
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)

		// If errorBody is provided, try to bind the error response to it
		if errorBody != nil {
			if err := json.Unmarshal(bodyBytes, errorBody); err == nil {
				return err
			}
		}

		return fmt.Errorf("%d: %s", resp.StatusCode, string(bodyBytes))
	}

	if responseBody != nil {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("error reading response body: %w", err)
		}

		// Only attempt to decode if there's actual content in the response body
		if len(bodyBytes) > 0 {
			if err := json.Unmarshal(bodyBytes, responseBody); err != nil {
				return fmt.Errorf("error decoding response: %w", err)
			}
		}
	}

	return nil
}
