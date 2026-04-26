package onec

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const pageSize = 1000

type Client struct {
	baseURL    string
	user       string
	password   string
	httpClient *http.Client
}

func NewClient(baseURL, user, password string) *Client {
	return &Client{
		baseURL:  strings.TrimRight(baseURL, "/"),
		user:     user,
		password: password,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type oneCResponse struct {
	Value []json.RawMessage `json:"value"`
}

func (c *Client) Fetch(ctx context.Context, endpoint string, params url.Values) ([]json.RawMessage, error) {
	var all []json.RawMessage
	skip := 0

	for {
		p := url.Values{}
		for k, v := range params {
			p[k] = v
		}
		p.Set("$format", "json")
		p.Set("$top", strconv.Itoa(pageSize))
		p.Set("$skip", strconv.Itoa(skip))

		rawURL := c.buildURL(endpoint, p)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			return nil, err
		}
		req.SetBasicAuth(c.user, c.password)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("1C OData returned status %d for %s", resp.StatusCode, endpoint)
		}

		var odr oneCResponse
		err = json.NewDecoder(resp.Body).Decode(&odr)
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("decode 1C response for %s: %w", endpoint, err)
		}

		all = append(all, odr.Value...)
		if len(odr.Value) < pageSize {
			break
		}
		skip += pageSize
	}
	return all, nil
}

func (c *Client) buildURL(endpoint string, params url.Values) string {
	segments := strings.Split(endpoint, "/")
	encoded := make([]string, len(segments))
	for i, s := range segments {
		encoded[i] = url.PathEscape(s)
	}
	u := c.baseURL + "/" + strings.Join(encoded, "/")
	if len(params) > 0 {
		// 1C OData does not accept '+' as a space — must use %20
		u += "?" + strings.ReplaceAll(params.Encode(), "+", "%20")
	}
	return u
}
