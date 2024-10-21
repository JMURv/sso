package utils

import (
	"github.com/goccy/go-json"
	"net/http"
	"strings"
)

type Response struct {
	Data any `json:"data"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func SuccessPaginatedResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

func SuccessResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(
		&Response{
			Data: data,
		},
	)
}

func ErrResponse(w http.ResponseWriter, statusCode int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(
		&ErrorResponse{
			Error: err.Error(),
		},
	)
}

func ParseFiltersByURL(r *http.Request) map[string]any {
	filters := make(map[string]any)
	for key, values := range r.URL.Query() {
		switch {
		case key == "page":
			continue
		case key == "size":
			continue
		case key == "sort":
			continue
		case len(values) > 0:
			if strings.HasSuffix(key, "[min]") || strings.HasSuffix(key, "[max]") {
				baseKey := strings.TrimSuffix(key, "[min]")
				baseKey = strings.TrimSuffix(baseKey, "[max]")

				if filters[baseKey] == nil {
					filters[baseKey] = make(map[string]any)
				}

				if strings.HasSuffix(key, "[min]") {
					filters[baseKey].(map[string]any)["min"] = values[0]
				} else {
					filters[baseKey].(map[string]any)["max"] = values[0]
				}
			} else {
				filters[key] = values
			}
		}
	}
	return filters
}
