package weather

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "net/url"
    "time"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
)

type Client interface {
    CurrentTempC(ctx context.Context, query string) (float64, error)
}

type HTTPClient struct {
    APIKey string
    HTTP   *http.Client
}

func New(apiKey string) *HTTPClient {
    return &HTTPClient{
        APIKey: apiKey,
        HTTP: &http.Client{
            Timeout: 5 * time.Second,
        },
    }
}

type weatherAPIResponse struct {
    Current struct {
        TempC float64 `json:"temp_c"`
    } `json:"current"`
}

func (c *HTTPClient) CurrentTempC(ctx context.Context, query string) (float64, error) {
    tracer := otel.Tracer("weatherapi")
    ctx, span := tracer.Start(ctx, "weather.CurrentTempC")
    defer span.End()
    span.SetAttributes(attribute.String("query", query))

    u := url.URL{
        Scheme: "https",
        Host:   "api.weatherapi.com",
        Path:   "/v1/current.json",
    }
    q := u.Query()
    q.Set("key", c.APIKey)
    q.Set("q", query) // example: "Rio de Janeiro,RJ,Brazil"
    u.RawQuery = q.Encode()

    req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
    if err != nil {
        return 0, err
    }

    res, err := c.HTTP.Do(req)
    if err != nil {
        return 0, err
    }
    defer res.Body.Close()

    if res.StatusCode >= 400 {
        return 0, fmt.Errorf("weatherapi http error: %d", res.StatusCode)
    }

    var out weatherAPIResponse
    if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
        return 0, err
    }
    return out.Current.TempC, nil
}
