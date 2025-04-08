package utils

import (
	"encoding/json"
	"fmt"
	"github.com/JMURv/sso/internal/auth"
	"github.com/JMURv/sso/internal/config"
	"github.com/JMURv/sso/internal/dto"
	"github.com/JMURv/sso/internal/hdl"
	"github.com/JMURv/sso/internal/hdl/validation"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"strings"
	"time"
)

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

func ParsePaginationValues(r *http.Request) (int, int) {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = config.DefaultPage
	}

	size, err := strconv.Atoi(r.URL.Query().Get("size"))
	if err != nil || size < 1 {
		size = config.DefaultSize
	}

	return page, size
}

func ParseAndValidate(w http.ResponseWriter, r *http.Request, dst any) bool {
	var err error
	if err = json.NewDecoder(r.Body).Decode(dst); err != nil {
		zap.L().Debug(
			hdl.ErrDecodeRequest.Error(),
			zap.Error(err),
		)
		ErrResponse(w, http.StatusBadRequest, hdl.ErrDecodeRequest)
		return false
	}

	if err = validation.V.Struct(dst); err != nil {
		ErrResponse(w, http.StatusBadRequest, err)
		return false
	}

	return true
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
				filters[key] = strings.Split(values[0], ",")
			}
		}
	}
	return filters
}
