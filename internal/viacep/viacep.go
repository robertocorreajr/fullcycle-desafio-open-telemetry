package viacep

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "github.com/robertocorreajr/fullcycle-desafio-open-telemetry/internal/types"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/attribute"
)

type Client interface {
    Lookup(ctx context.Context, cep string) (*types.ViaCEPResult, error)
}

type HTTPClient struct {
    BaseURL string
    HTTP    *http.Client
}

func New() *HTTPClient {
    return &HTTPClient{
        BaseURL: "https://viacep.com.br/ws",
        HTTP: &http.Client{
            Timeout: 5 * time.Second,
        },
    }
}

func (c *HTTPClient) Lookup(ctx context.Context, cep string) (*types.ViaCEPResult, error) {
    tracer := otel.Tracer("viacep")
    ctx, span := tracer.Start(ctx, "viacep.Lookup")
    defer span.End()
    span.SetAttributes(attribute.String("cep", cep))
    req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/%s/json/", c.BaseURL, cep), nil)
    if err != nil {
        return nil, err
    }

    res, err := c.HTTP.Do(req)
    if err != nil {
        return nil, err
    }
    defer res.Body.Close()

    if res.StatusCode == 404 || res.StatusCode == 400 {
        return &types.ViaCEPResult{Erro: true}, nil
    }

    if res.StatusCode >= 400 {
        return nil, fmt.Errorf("viacep http error: %d", res.StatusCode)
    }

    var out types.ViaCEPResult
    if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
        return nil, err
    }

    return &out, nil
}
