package onec

import (
	"bytes"
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

const (
	OrderOrgKey      = "c32d0d64-fbe9-11ed-827c-fa163e38e500"
	OrderCurrencyKey = "c26a4d87-c6e2-4aca-ab05-1b02be6ecaec"
)

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

type OrderItem struct {
	NomenclatureKey string
	Name            string
	Quantity        int
	Price           float64
}

type CreateOrderInput struct {
	CounterpartKey string
	Comment        string
	TotalAmount    float64
	Items          []OrderItem
}

type CreatedOrder struct {
	RefKey string
	Number string
}

func (c *Client) CreateOrder(ctx context.Context, in CreateOrderInput) (*CreatedOrder, error) {
	type itemBody struct {
		LineNumber      int     `json:"LineNumber"`
		NomenclatureKey string  `json:"Номенклатура_Key"`
		Quantity        float64 `json:"Количество"`
		Price           float64 `json:"Цена"`
		Sum             float64 `json:"Сумма"`
		Total           float64 `json:"Всего"`
	}

	items := make([]itemBody, len(in.Items))
	for i, item := range in.Items {
		sum := item.Price * float64(item.Quantity)
		items[i] = itemBody{
			LineNumber:      i + 1,
			NomenclatureKey: item.NomenclatureKey,
			Quantity:        float64(item.Quantity),
			Price:           item.Price,
			Sum:             sum,
			Total:           sum,
		}
	}

	body := map[string]any{
		"Date":                time.Now().Format("2006-01-02T15:04:05"),
		"Организация_Key":     OrderOrgKey,
		"Контрагент_Key":      in.CounterpartKey,
		"ВалютаДокумента_Key": OrderCurrencyKey,
		"Комментарий":         in.Comment,
		"СуммаДокумента":      in.TotalAmount,
		"Запасы":              items,
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal 1C order: %w", err)
	}

	endpoint := "Document_" + url.PathEscape("ЗаказПокупателя")
	rawURL := c.baseURL + "/" + endpoint + "?%24format=json"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rawURL, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(c.user, c.password)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		var errBody struct {
			Error struct {
				Message struct {
					Value string `json:"value"`
				} `json:"message"`
			} `json:"odata.error"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&errBody)
		return nil, fmt.Errorf("1C returned %d: %s", resp.StatusCode, errBody.Error.Message.Value)
	}

	var result struct {
		RefKey string `json:"Ref_Key"`
		Number string `json:"Number"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode 1C create order response: %w", err)
	}

	return &CreatedOrder{RefKey: result.RefKey, Number: result.Number}, nil
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
