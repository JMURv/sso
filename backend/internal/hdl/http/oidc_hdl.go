package http

import (
	"net/http"

	"github.com/JMURv/sso/internal/hdl"
	mid "github.com/JMURv/sso/internal/hdl/http/middleware"
	"github.com/JMURv/sso/internal/hdl/http/utils"
	"github.com/go-chi/chi/v5"
)

func (h *Handler) RegisterOIDCRoutes() {
	h.router.Get("/auth/oidc/{provider}/start", h.startOIDC)
	h.router.With(mid.Device).Get("/auth/oidc/{provider}/callback", h.handleOIDCCallback)
}

// startOIDC godoc
//
//	@Summary		Start OIDC authentication flow
//	@Description	Redirects user to the OIDC provider's authorization URL
//	@Tags			OIDC
//	@Accept			json
//	@Produce		json
//	@Param			provider	path		string					true	"OIDC provider identifier"
//	@Success		307			{object}	nil						"Redirect to provider auth URL"
//	@Failure		400			{object}	utils.ErrorsResponse	"invalid provider"
//	@Failure		500			{object}	utils.ErrorsResponse	"internal error"
//	@Router			/auth/oidc/{provider}/start [get]
func (h *Handler) startOIDC(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	if provider == "" {
		utils.ErrResponse(w, http.StatusBadRequest, ErrInvalidURL)
		return
	}

	res, err := h.ctrl.GetOIDCAuthURL(r.Context(), provider)
	if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, err)
		return
	}

	http.Redirect(w, r, res.URL, http.StatusTemporaryRedirect)
}

// handleOIDCCallback godoc
//
//	@Summary		Handle OIDC provider callback
//	@Description	Processes provider callback, exchanges code, sets authentication cookies, and redirects to success URL
//	@Tags			OIDC
//	@Accept			json
//	@Produce		json
//	@Param			provider	path		string					true	"OIDC provider identifier"
//	@Param			code		query		string					true	"Authorization code returned by provider"
//	@Param			state		query		string					false	"State parameter for CSRF mitigation"
//	@Param			X-Real-IP	header		string					true	"Client real IP address"
//	@Param			User-Agent	header		string					true	"Client User-Agent"
//	@Success		307			{object}	nil						"Redirect to success URL"
//	@Failure		400			{object}	utils.ErrorsResponse	"invalid request or missing device info"
//	@Failure		404			{object}	utils.ErrorsResponse	"provider not supported or resource not found"
//	@Failure		500			{object}	utils.ErrorsResponse	"internal error"
//	@Router			/auth/oidc/{provider}/callback [get]
func (h *Handler) handleOIDCCallback(w http.ResponseWriter, r *http.Request) {
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

	res, err := h.ctrl.HandleOIDCCallback(r.Context(), &d, provider, code, state)
	if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, err)
		return
	}

	utils.SetAuthCookies(w, res.Access, res.Refresh)
	http.Redirect(w, r, res.SuccessURL, http.StatusTemporaryRedirect)
}
