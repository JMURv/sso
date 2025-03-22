package utils

import (
	"encoding/json"
	"fmt"
	"github.com/JMURv/sso/internal/auth"
	"github.com/JMURv/sso/internal/dto"
	"github.com/go-playground/validator/v10"
	"net/http"
	"time"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

type ErrorsResponse struct {
	Errors []string `json:"errors"`
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

	msgs := make([]string, 0, 1)
	if errs, ok := err.(validator.ValidationErrors); ok {
		msgs = make([]string, 0, len(errs))
		for _, fe := range errs {
			msgs = append(msgs, fmt.Sprintf("%s failed on the %s rule", fe.Field(), fe.Tag()))
		}
	} else {
		msgs = append(msgs, err.Error())
	}

	json.NewEncoder(w).Encode(
		&ErrorsResponse{
			Errors: msgs,
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
