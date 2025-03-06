package utils

import (
	"encoding/json"
	"github.com/JMURv/sso/internal/auth"
	"github.com/JMURv/sso/internal/dto"
	"net/http"
	"time"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func StatusResponse(w http.ResponseWriter, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
}

func SuccessResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
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

func SetAuthCookies(w http.ResponseWriter, access, refresh string) {
	http.SetCookie(
		w, &http.Cookie{
			Name:     "access",
			Value:    access,
			Expires:  time.Now().Add(auth.AccessTokenDuration),
			HttpOnly: true,
			Secure:   true,
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
		},
	)

	http.SetCookie(
		w, &http.Cookie{
			Name:     "refresh",
			Value:    refresh,
			Expires:  time.Now().Add(auth.RefreshTokenDuration),
			HttpOnly: true,
			Secure:   true,
			Path:     "/",
			SameSite: http.SameSiteStrictMode,
		},
	)
}

func ParseDeviceByRequest(r *http.Request) (dto.DeviceRequest, bool) {
	ip, ok := r.Context().Value("ip").(string)
	if !ok {
		return dto.DeviceRequest{}, false
	}

	ua, ok := r.Context().Value("ua").(string)
	if !ok {
		return dto.DeviceRequest{}, false
	}

	return dto.DeviceRequest{
		IP: ip,
		UA: ua,
	}, true
}
