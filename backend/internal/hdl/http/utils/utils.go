package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/JMURv/sso/internal/config"
	"github.com/JMURv/sso/internal/dto"
	"github.com/JMURv/sso/internal/hdl"
	"github.com/JMURv/sso/internal/hdl/validation"
	"github.com/JMURv/sso/internal/repo/s3"
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
		zap.L().Error(
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

var ErrInvalidFileUpload = errors.New("invalid file upload")
var ErrFileTooLarge = errors.New("file too large")
var ErrInvalidFileType = errors.New("invalid file type")

func ParseFileField(r *http.Request, fieldName string, fileReq *s3.UploadFileRequest) error {
	file, header, err := r.FormFile(fieldName)
	if err != nil && err != http.ErrMissingFile {
		return ErrInvalidFileUpload
	}

	if header != nil {
		defer func(file multipart.File) {
			if err = file.Close(); err != nil {
				zap.L().Error("failed to close file", zap.Error(err))
			}
		}(file)

		if header.Size > 10<<20 {
			zap.L().Debug("file too large", zap.String("field", fieldName), zap.Int64("size", header.Size))
			return ErrFileTooLarge
		}

		fileReq.File, err = io.ReadAll(file)
		if err != nil {
			zap.L().Error("failed to read file", zap.Error(err))
			return hdl.ErrInternal
		}

		fileReq.ContentType = http.DetectContentType(fileReq.File)
		if !strings.HasPrefix(fileReq.ContentType, "image/") {
			zap.L().Debug("invalid file type", zap.String("field", fieldName), zap.String("type", fileReq.ContentType))
			return ErrInvalidFileType
		}

		fileReq.Filename = header.Filename
	}

	return nil
}

func GetAuthCookies(accessStr string, refreshStr string) (*http.Cookie, *http.Cookie) {
	access := &http.Cookie{
		Name:     config.AccessCookieName,
		Value:    accessStr,
		Expires:  time.Now().Add(config.AccessTokenDuration),
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	}

	refresh := &http.Cookie{
		Name:     config.RefreshCookieName,
		Value:    refreshStr,
		Expires:  time.Now().Add(config.RefreshTokenDuration),
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		SameSite: http.SameSiteStrictMode,
	}

	return access, refresh
}

func SetAuthCookies(w http.ResponseWriter, access, refresh string) {
	accessCookie, refreshCookie := GetAuthCookies(access, refresh)
	http.SetCookie(w, accessCookie)
	http.SetCookie(w, refreshCookie)
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
	filters := make(map[string]any, len(r.URL.Query()))
	for key, values := range r.URL.Query() {
		switch key {
		case "page", "size":
			continue
		case "roles":
			filters[key] = strings.Split(values[0], ",")
		case "sort":
			filters[key] = values[0]
		default:
			filters[key] = values[0]
		}
	}
	return filters
}
