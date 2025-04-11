package http

import (
	"errors"
	"github.com/JMURv/sso/internal/ctrl"
	"github.com/JMURv/sso/internal/hdl"
	mid "github.com/JMURv/sso/internal/hdl/http/middleware"
	"github.com/JMURv/sso/internal/hdl/http/utils"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (h *Handler) RegisterOAuth2Routes() {
	h.router.Get("/api/auth/oauth2/{provider}/start", h.startOAuth2)
	h.router.With(mid.Device).Get("/api/auth/oauth2/{provider}/callback", h.handleOAuth2Callback)
}

func (h *Handler) startOAuth2(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	if provider == "" {
		utils.ErrResponse(w, http.StatusBadRequest, ErrInvalidURL)
		return
	}

	res, err := h.ctrl.GetOAuth2AuthURL(r.Context(), provider)
	if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, res.URL, http.StatusTemporaryRedirect)
}

func (h *Handler) handleOAuth2Callback(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	if provider == "" {
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

	res, err := h.ctrl.HandleOAuth2Callback(r.Context(), &d, provider, code, state)
	if err != nil {
		if errors.Is(err, ctrl.ErrNotFound) {
			utils.ErrResponse(w, http.StatusNotFound, err)
			return
		}
		utils.ErrResponse(w, http.StatusInternalServerError, hdl.ErrInternal)
		return
	}

	utils.SetAuthCookies(w, res.Access, res.Refresh)
	http.Redirect(w, r, res.SuccessURL, http.StatusTemporaryRedirect)
}
