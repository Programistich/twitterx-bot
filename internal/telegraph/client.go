package telegraph

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const DefaultBaseURL = "https://api.telegra.ph"

// Client - HTTP клієнт для Telegraph API
type Client struct {
	httpClient *http.Client
	baseURL    string
}

// NewClient створює новий клієнт
// baseURL - опціональний, для тестування. Якщо порожній - використовується DefaultBaseURL
func NewClient(httpClient *http.Client, baseURL string) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	return &Client{
		httpClient: httpClient,
		baseURL:    baseURL,
	}
}

// CreateAccount створює новий Telegraph акаунт
func (c *Client) CreateAccount(ctx context.Context, req CreateAccountRequest) (*Account, error) {
	form := url.Values{}
	form.Set("short_name", req.ShortName)
	if req.AuthorName != "" {
		form.Set("author_name", req.AuthorName)
	}
	if req.AuthorURL != "" {
		form.Set("author_url", req.AuthorURL)
	}

	var resp Response[Account]
	if err := c.post(ctx, "/createAccount", form, &resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, fmt.Errorf("%w: %s", ErrAPIError, resp.Error)
	}

	return &resp.Result, nil
}

// CreatePage створює нову сторінку
func (c *Client) CreatePage(ctx context.Context, req CreatePageRequest) (*Page, error) {
	// Серіалізуємо content в JSON
	contentJSON, err := json.Marshal(req.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal content: %w", err)
	}

	form := url.Values{}
	form.Set("access_token", req.AccessToken)
	form.Set("title", req.Title)
	form.Set("content", string(contentJSON))
	if req.AuthorName != "" {
		form.Set("author_name", req.AuthorName)
	}
	if req.AuthorURL != "" {
		form.Set("author_url", req.AuthorURL)
	}
	form.Set("return_content", fmt.Sprintf("%t", req.ReturnContent))

	var resp Response[Page]
	if err := c.post(ctx, "/createPage", form, &resp); err != nil {
		return nil, err
	}

	if !resp.OK {
		return nil, fmt.Errorf("%w: %s", ErrAPIError, resp.Error)
	}

	return &resp.Result, nil
}

// post виконує POST запит
func (c *Client) post(ctx context.Context, path string, form url.Values, result any) error {
	reqURL := c.baseURL + path

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}
