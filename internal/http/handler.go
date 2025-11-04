package http

import (
    "encoding/json"
    "errors"
    "fmt"
    stdhttp "net/http"

    "github.com/gorilla/mux"

    "github.com/robertocorreajr/fullcycle-desafio-open-telemetry/internal/service"
)

type Handler struct {
    Svc *service.Service
}

func NewRouter(h *Handler) *mux.Router {
    r := mux.NewRouter()
    r.HandleFunc("/weather/{zipcode}", h.getWeather).Methods(stdhttp.MethodGet)

    r.HandleFunc("/health", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
        w.WriteHeader(200)
        w.Write([]byte("ok"))
    }).Methods(stdhttp.MethodGet)
    r.HandleFunc("/ready", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
        w.WriteHeader(200)
        w.Write([]byte("ok"))
    }).Methods(stdhttp.MethodGet)
    return r
}

func (h *Handler) getWeather(w stdhttp.ResponseWriter, r *stdhttp.Request) {
    zip := mux.Vars(r)["zipcode"]

    resp, err := h.Svc.GetWeatherByCEP(r.Context(), zip)
    if err != nil {
        // log server error
        fmt.Printf("Erro ao processar CEP %s: %v\n", zip, err)

        switch {
        case errors.Is(err, service.ErrInvalidZip):
            httpJSON(w, stdhttp.StatusUnprocessableEntity, map[string]string{"message": "invalid zipcode"})
            return
        case errors.Is(err, service.ErrNotFound):
            httpJSON(w, stdhttp.StatusNotFound, map[string]string{"message": "can not find zipcode"})
            return
        default:
            httpJSON(w, stdhttp.StatusInternalServerError, map[string]string{"message": "internal error"})
            return
        }
    }

    httpJSON(w, stdhttp.StatusOK, resp)
}

func httpJSON(w stdhttp.ResponseWriter, code int, v any) {
    w.Header().Set("Content-Type", "application/json; charset=utf-8")
    w.WriteHeader(code)
    _ = json.NewEncoder(w).Encode(v)
}
