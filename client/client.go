package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tomo/searxng-cli/model"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func New(baseURL string, timeout int) *Client {
	t := time.Duration(timeout) * time.Second
	return &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: t,
		},
	}
}

type SearchParams struct {
	Query      string
	Categories string
	Engines    string
	Language   string
	TimeRange  string
	Pageno     int
}

type searxngAPIResponse struct {
	Results []struct {
		Title    string `json:"title"`
		URL      string `json:"url"`
		Content  string `json:"content"`
		Category string `json:"category"`
		Engine   string `json:"engine"`
	} `json:"results"`
}

func (c *Client) Search(params SearchParams) ([]model.Result, error) {
	path := strings.TrimRight(c.BaseURL, "/") + "/search"
	u, err := url.Parse(path)
	if err != nil {
		return nil, fmt.Errorf("invalid base URL: %w", err)
	}
	q := u.Query()
	q.Set("q", params.Query)
	q.Set("format", "json")
	if params.Categories != "" {
		q.Set("categories", params.Categories)
	}
	if params.Engines != "" {
		q.Set("engines", params.Engines)
	}
	if params.Language != "" {
		q.Set("language", params.Language)
	}
	if params.TimeRange != "" {
		q.Set("time_range", params.TimeRange)
	}
	if params.Pageno > 0 {
		q.Set("pageno", fmt.Sprintf("%d", params.Pageno))
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "searxng-cli/1.0")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		truncated := body
		if len(truncated) > 200 {
			truncated = truncated[:200]
		}
		err := fmt.Errorf("API returned %d: %s", resp.StatusCode, string(truncated))
		if resp.StatusCode == http.StatusTooManyRequests {
			err = fmt.Errorf("API returned %d: %w\n  hint: check SearXNG limiter settings (limiter: false in settings.yml)", resp.StatusCode, err)
		}
		return nil, err
	}

	var apiResp searxngAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	results := make([]model.Result, 0, len(apiResp.Results))
	for _, r := range apiResp.Results {
		results = append(results, model.Result{
			Title:    r.Title,
			URL:      r.URL,
			Snippet:  r.Content,
			Category: r.Category,
			Engine:   r.Engine,
		})
	}
	return results, nil
}
