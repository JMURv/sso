package http

import (
	"github.com/JMURv/sso/internal/hdl"
	mid "github.com/JMURv/sso/internal/hdl/http/middleware"
	"github.com/JMURv/sso/internal/hdl/http/utils"
	"net/http"
	"strings"
)

func RegisterOAuth2Routes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc(
		"/api/auth/oauth2/{provider}/start", mid.Apply(
			h.startOAuth2,
			mid.AllowedMethods(http.MethodGet),
		),
	)

	mux.HandleFunc(
		"/api/auth/oauth2/{provider}/callback", mid.Apply(
			h.handleOAuth2Callback,
			mid.AllowedMethods(http.MethodGet),
			mid.Device,
		),
	)
}

func (h *Handler) startOAuth2(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 6 {
		utils.ErrResponse(w, http.StatusBadRequest, ErrInvalidURL)
		return
	}
	provider := parts[4]

	res, err := h.ctrl.GetOAuth2AuthURL(r.Context(), provider)
	if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, res.URL, http.StatusTemporaryRedirect)
}

func (h *Handler) handleOAuth2Callback(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 6 {
		utils.ErrResponse(w, http.StatusBadRequest, ErrInvalidURL)
		return
	}
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	d, ok := utils.ParseDeviceByRequest(r)
	if !ok {
		utils.ErrResponse(w, http.StatusBadRequest, hdl.ErrNoDeviceInfo)
		return
	}

	res, err := h.ctrl.HandleOAuth2Callback(r.Context(), &d, parts[4], code, state)
	if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.SetAuthCookies(w, res.Access, res.Refresh)

	// TODO: get redirect from config
	http.Redirect(w, r, "http://localhost:3000/", http.StatusTemporaryRedirect)
	//utils.SuccessResponse(w, http.StatusOK, res)
}
