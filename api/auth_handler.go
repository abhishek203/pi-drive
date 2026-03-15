package api

import (
	"log"
	"net/http"

	"github.com/pidrive/pidrive/internal/auth"
	"github.com/pidrive/pidrive/internal/share"
)

type AuthHandler struct {
	authService  *auth.AuthService
	emailService *auth.EmailService
	shareService *share.ShareService
}

func NewAuthHandler(authService *auth.AuthService, emailService *auth.EmailService, shareService *share.ShareService) *AuthHandler {
	return &AuthHandler{authService: authService, emailService: emailService, shareService: shareService}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if !registerLimiter.allow(getClientIP(r)) {
		writeError(w, http.StatusTooManyRequests, "too many registration attempts, try again later")
		return
	}

	var req struct {
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "email and name are required")
		return
	}

	apiKey, agent, err := h.authService.Register(req.Email, req.Name)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error())
		return
	}

	// Send verification email
	code, _ := h.authService.Login(req.Email)
	h.emailService.SendVerificationCode(req.Email, code)

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"api_key": apiKey,
		"agent":   agent,
		"message": "Check your email for the verification code",
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" {
		writeError(w, http.StatusBadRequest, "email is required")
		return
	}

	if !loginLimiter.allow(req.Email) {
		writeError(w, http.StatusTooManyRequests, "too many login attempts, try again later")
		return
	}

	code, err := h.authService.Login(req.Email)
	if err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}

	h.emailService.SendVerificationCode(req.Email, code)

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "Verification code sent to " + req.Email,
	})
}

func (h *AuthHandler) Verify(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Email == "" || req.Code == "" {
		writeError(w, http.StatusBadRequest, "email and code are required")
		return
	}

	if !verifyLimiter.allow(req.Email) {
		writeError(w, http.StatusTooManyRequests, "too many verification attempts, try again later")
		return
	}

	apiKey, agent, err := h.authService.Verify(req.Email, req.Code)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	// Activate any pending shares for this email
	if activated, err := h.shareService.ActivatePendingShares(req.Email, agent.ID); err == nil && activated > 0 {
		log.Printf("[shares] Activated %d pending shares for %s", activated, req.Email)
	}

	// Notify admin of new signup
	go h.emailService.SendAdminNotification("abhishek@ressl.ai", agent.Email, agent.Name)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"api_key": apiKey,
		"agent":   agent,
	})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	agent := GetAgent(r)
	if agent == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}
	writeJSON(w, http.StatusOK, agent)
}
