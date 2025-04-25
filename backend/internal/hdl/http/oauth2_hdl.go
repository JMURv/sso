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
	h.router.Get("/auth/oauth2/{provider}/start", h.startOAuth2)
	h.router.With(mid.Device).Get("/auth/oauth2/{provider}/callback", h.handleOAuth2Callback)
}

// startOAuth2 godoc
//
//	@Summary		Start OAuth2 authentication flow
//	@Description	Redirects user to the OAuth2 provider's authorization URL
//	@Tags			OAuth2
//	@Accept			json
//	@Produce		json
//	@Param			provider	path		string					true	"OAuth2 provider identifier"
//	@Success		307			{object}	nil						"Redirect to provider auth URL"
//	@Failure		400			{object}	utils.ErrorsResponse	"invalid provider"
//	@Failure		500			{object}	utils.ErrorsResponse	"internal error"
//	@Router			/auth/oauth2/{provider}/start [get]
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

// handleOAuth2Callback godoc
//
//	@Summary		Handle OAuth2 provider callback
//	@Description	Processes provider callback, exchanges code, sets authentication cookies, and redirects to success URL
//	@Tags			OAuth2
//	@Accept			json
//	@Produce		json
//	@Param			provider	path		string					true	"OAuth2 provider identifier"
//	@Param			code		query		string					true	"Authorization code returned by provider"
//	@Param			state		query		string					true	"State parameter for CSRF mitigation"
//	@Param			X-Real-IP	header		string					true	"Client real IP address"
//	@Param			User-Agent	header		string					true	"Client User-Agent"
//	@Success		307			{object}	nil						"Redirect to success URL"
//	@Failure		400			{object}	utils.ErrorsResponse	"invalid request or missing device info"
//	@Failure		404			{object}	utils.ErrorsResponse	"provider not supported or resource not found"
//	@Failure		500			{object}	utils.ErrorsResponse	"internal error"
//	@Router			/auth/oauth2/{provider}/callback [get]
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
