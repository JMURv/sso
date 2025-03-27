package http

import (
	"github.com/JMURv/sso/internal/hdl"
	mid "github.com/JMURv/sso/internal/hdl/http/middleware"
	"github.com/JMURv/sso/internal/hdl/http/utils"
	"net/http"
	"strings"
)

func RegisterOIDCRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc(
		"/api/auth/oidc/{provider}/start", mid.Apply(
			h.startOIDC,
			mid.AllowedMethods(http.MethodGet),
		),
	)

	mux.HandleFunc(
		"/api/auth/oidc/{provider}/callback", mid.Apply(
			h.handleOIDCCallback,
			mid.AllowedMethods(http.MethodGet),
			mid.Device,
		),
	)
}

func (h *Handler) startOIDC(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 6 {
		utils.ErrResponse(w, http.StatusBadRequest, ErrInvalidURL)
		return
	}

	res, err := h.ctrl.GetOIDCAuthURL(r.Context(), parts[4])
	if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, res.URL, http.StatusTemporaryRedirect)
}

func (h *Handler) handleOIDCCallback(w http.ResponseWriter, r *http.Request) {
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

	res, err := h.ctrl.HandleOIDCCallback(r.Context(), &d, parts[4], code, state)
	if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, err)
		return
	}

	utils.SetAuthCookies(w, res.Access, res.Refresh)
	http.Redirect(w, r, res.SuccessURL, http.StatusTemporaryRedirect)
}
