package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"

	"context"
	"time"

	"github.com/robertocorreajr/fullcycle-desafio-open-telemetry/internal/otel"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type Req struct {
	Cep string `json:"cep"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	zipkinURL := os.Getenv("ZIPKIN_URL")
	if zipkinURL == "" {
		zipkinURL = "http://zipkin:9411/api/v2/spans"
	}

	tp, err := otel.InitTracer(otel.Config{ServiceName: "service-a", ZipkinURL: zipkinURL})
	if err != nil {
		log.Printf("otel init: %v", err)
	}
	defer tp.Shutdown(context.Background())

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req Req
		b, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(b, &req); err != nil {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte(`{"message":"invalid zipcode"}`))
			return
		}

		// validate: must be a string of exactly 8 digits
		cepRe := regexp.MustCompile(`^\d{8}$`)
		if !cepRe.MatchString(req.Cep) {
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte(`{"message":"invalid zipcode"}`))
			return
		}

		// forward to service B
		url := os.Getenv("SERVICE_B_URL")
		if url == "" {
			url = "http://service-b:8081/weather/" + req.Cep
		} else {
			url = url + "/weather/" + req.Cep
		}

		// create instrumented client to propagate trace context
		client := &http.Client{Transport: otelhttp.NewTransport(http.DefaultTransport)}
		reqOut, _ := http.NewRequestWithContext(r.Context(), http.MethodGet, url, nil)
		resp, err := client.Do(reqOut)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"message":"internal error"}`))
			return
		}
		defer resp.Body.Close()

		w.WriteHeader(resp.StatusCode)
		io.Copy(w, resp.Body)
	})

	http.Handle("/submit", otelhttp.NewHandler(handler, "submit"))

	srv := ":" + port
	log.Printf("service-a listening on %s", srv)
	server := &http.Server{Addr: srv, ReadTimeout: 5 * time.Second, WriteTimeout: 10 * time.Second}
	log.Fatal(server.ListenAndServe())
}
