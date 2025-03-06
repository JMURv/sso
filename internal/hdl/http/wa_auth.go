package http

import (
	"github.com/JMURv/sso/internal/hdl"
	mid "github.com/JMURv/sso/internal/hdl/http/middleware"
	"github.com/JMURv/sso/internal/hdl/http/utils"
	"github.com/google/uuid"
	"net/http"
)

func RegisterWebAuthnRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("/api/auth/webauthn/register/start", mid.Apply(h.webauthnRegistrationStart, mid.Auth))
	mux.HandleFunc("/api/auth/webauthn/register/finish", mid.Apply(h.webauthnRegistrationFinish, mid.Auth))
	mux.HandleFunc("/api/auth/webauthn/login/start", h.webauthnLoginStart)
	mux.HandleFunc("/api/auth/webauthn/login/finish", h.webauthnLoginFinish)
}

func (h *Handler) webauthnRegistrationStart(w http.ResponseWriter, r *http.Request) {
	uid, ok := r.Context().Value("uid").(uuid.UUID)
	if !ok {
		utils.ErrResponse(w, http.StatusUnauthorized, hdl.ErrInternal)
		return
	}

	user, err := h.ctrl.GetUserForWebAuthn(r.Context(), uid)
	if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, err)
		return
	}

	options, sessionData, err := webAuthn.BeginRegistration(user)
	if err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, err)
		return
	}

	if err := h.ctrl.StoreWebAuthnSession(r.Context(), "registration", uid.String(), sessionData); err != nil {
		utils.ErrResponse(w, http.StatusInternalServerError, err)
		return
	}

	utils.SuccessResponse(w, http.StatusOK, options)
}

func (h *Handler) webauthnRegistrationFinish(w http.ResponseWriter, r *http.Request) {
	// Similar flow for parsing and verifying response
	// Use webAuthn.FinishRegistration()
}

func (h *Handler) webauthnLoginStart(w http.ResponseWriter, r *http.Request) {
	// Get username/email from request
	// Retrieve user credentials
	// Generate assertion options
}

func (h *Handler) webauthnLoginFinish(w http.ResponseWriter, r *http.Request) {
	// Verify assertion response
	// Generate JWT tokens on success
}
