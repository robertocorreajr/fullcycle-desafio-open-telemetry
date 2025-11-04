package main

import (
    "context"
    "log"
    "net/http"
    "os"

    h "github.com/robertocorreajr/fullcycle-desafio-open-telemetry/internal/http"
    "github.com/robertocorreajr/fullcycle-desafio-open-telemetry/internal/service"
    "github.com/robertocorreajr/fullcycle-desafio-open-telemetry/internal/viacep"
    "github.com/robertocorreajr/fullcycle-desafio-open-telemetry/internal/weather"
    "github.com/robertocorreajr/fullcycle-desafio-open-telemetry/internal/otel"
    "go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

func main() {
    apiKey := os.Getenv("WEATHERAPI_KEY")
    if apiKey == "" {
        log.Fatal("missing WEATHERAPI_KEY")
    }

    zipkinURL := os.Getenv("ZIPKIN_URL")
    if zipkinURL == "" {
        zipkinURL = "http://zipkin:9411/api/v2/spans"
    }

    tp, err := otel.InitTracer(otel.Config{ServiceName: "service-b", ZipkinURL: zipkinURL})
    if err != nil {
        log.Printf("otel init: %v", err)
    }
    defer tp.Shutdown(context.Background())

    cepClient := viacep.New()
    weatherClient := weather.New(apiKey)
    svc := service.New(cepClient, weatherClient)

    router := h.NewRouter(&h.Handler{Svc: svc})
    // wrap router with otelhttp to automatically extract context from incoming requests
    routerHandler := http.Handler(router)
    router = nil
    wrapped := otelhttp.NewHandler(routerHandler, "service-b-router")

    port := os.Getenv("PORT")
    if port == "" {
        port = "8081"
    }
    addr := ":" + port
    log.Printf("listening on %s", addr)
    log.Fatal(http.ListenAndServe(addr, wrapped))
}
