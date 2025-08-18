package apiclient

import (
	"bytes"
	"context"
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

func NewClient(baseURL string, timeout time.Duration, headers map[string]string) *ApiClient {
	return &ApiClient{
		BaseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
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
		if err := json.NewDecoder(resp.Body).Decode(responseBody); err != nil {
			return err
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
		if err := json.NewDecoder(resp.Body).Decode(responseBody); err != nil {
			return err
		}
	}

	return nil
}
