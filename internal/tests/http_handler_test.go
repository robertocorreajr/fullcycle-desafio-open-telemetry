package tests

import (
    "context"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "testing"

    apphttp "github.com/robertocorreajr/fullcycle-desafio-open-telemetry/internal/http"
    "github.com/robertocorreajr/fullcycle-desafio-open-telemetry/internal/service"
    "github.com/robertocorreajr/fullcycle-desafio-open-telemetry/internal/types"
)

type fakeCEP struct{}

func (f *fakeCEP) Lookup(ctx context.Context, cep string) (*types.ViaCEPResult, error) {
    if cep == "00000000" {
        return &types.ViaCEPResult{Erro: true}, nil
    }
    return &types.ViaCEPResult{Localidade: "São Paulo", UF: "SP"}, nil
}

type fakeWeather struct{}

func (f *fakeWeather) CurrentTempC(ctx context.Context, q string) (float64, error) {
    return 28.5, nil
}

func TestGetWeather_OK(t *testing.T) {
    svc := service.New(&fakeCEP{}, &fakeWeather{})
    h := &apphttp.Handler{Svc: svc}
    router := apphttp.NewRouter(h)

    req := httptest.NewRequest(http.MethodGet, "/weather/01001000", nil)
    rr := httptest.NewRecorder()
    router.ServeHTTP(rr, req)

    if rr.Code != http.StatusOK {
        t.Fatalf("expected 200 got %d", rr.Code)
    }

    var resp types.WeatherResponse
    if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
        t.Fatalf("decode: %v", err)
    }
}

func TestGetWeather_InvalidZip(t *testing.T) {
    svc := service.New(&fakeCEP{}, &fakeWeather{})
    h := &apphttp.Handler{Svc: svc}
    router := apphttp.NewRouter(h)

    req := httptest.NewRequest(http.MethodGet, "/weather/01001", nil) // invalid length
    rr := httptest.NewRecorder()
    router.ServeHTTP(rr, req)

    if rr.Code != http.StatusUnprocessableEntity {
        t.Fatalf("expected 422 got %d", rr.Code)
    }
}

func TestGetWeather_NotFound(t *testing.T) {
    svc := service.New(&fakeCEP{}, &fakeWeather{})
    h := &apphttp.Handler{Svc: svc}
    router := apphttp.NewRouter(h)

    req := httptest.NewRequest(http.MethodGet, "/weather/00000000", nil) // fakeCEP returns Erro=true
    rr := httptest.NewRecorder()
    router.ServeHTTP(rr, req)

    if rr.Code != http.StatusNotFound {
        t.Fatalf("expected 404 got %d", rr.Code)
    }
}
