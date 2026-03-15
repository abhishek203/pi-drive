package api

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/pidrive/pidrive/internal/activity"
	"github.com/pidrive/pidrive/internal/auth"
	"github.com/pidrive/pidrive/internal/billing"
	shareService "github.com/pidrive/pidrive/internal/share"
)

type ShareHandler struct {
	shareService    *shareService.ShareService
	fileManager     *shareService.FileManager
	authService     *auth.AuthService
	emailService    *auth.EmailService
	activityService *activity.ActivityService
	billingService  *billing.BillingService
}

func NewShareHandler(
	ss *shareService.ShareService,
	fm *shareService.FileManager,
	as *auth.AuthService,
	es *auth.EmailService,
	act *activity.ActivityService,
	bs *billing.BillingService,
) *ShareHandler {
	return &ShareHandler{
		shareService:    ss,
		fileManager:     fm,
		authService:     as,
		emailService:    es,
		activityService: act,
		billingService:  bs,
	}
}

func (h *ShareHandler) Share(w http.ResponseWriter, r *http.Request) {
	agent := GetAgent(r)
	if agent == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req struct {
		Path        string `json:"path"`
		ToEmail     string `json:"to_email,omitempty"`
		TargetEmail string `json:"target_email,omitempty"`
		Link        bool   `json:"link,omitempty"`
		Permission  string `json:"permission,omitempty"`
		Expires     string `json:"expires,omitempty"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Path == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	// Validate file exists
	if !h.fileManager.FileExists(agent.ID, req.Path) {
		writeError(w, http.StatusNotFound, "file not found: "+req.Path)
		return
	}

	input := shareService.CreateShareInput{
		OwnerID:    agent.ID,
		SourcePath: req.Path,
		Permission: req.Permission,
	}

	// Parse expiry
	if req.Expires != "" {
		dur, err := parseDuration(req.Expires)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid expires format: "+req.Expires)
			return
		}
		input.ExpiresIn = &dur
	}

	if req.Link {
		// Link share
		input.ShareType = "link"
		sh, err := h.shareService.Create(input)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// No file copy — link serves directly from owner's file

		h.activityService.Log(agent.ID, "share", req.Path, map[string]interface{}{
			"type": "link", "share_id": sh.ID, "url": sh.URL,
		})

		writeJSON(w, http.StatusCreated, sh)
	} else if req.ToEmail != "" || req.TargetEmail != "" {
		email := req.ToEmail
		if email == "" {
			email = req.TargetEmail
		}
		// Direct share
		target, err := h.authService.GetAgentByEmail(email)
		if err != nil {
			writeError(w, http.StatusNotFound, "agent not found: "+email)
			return
		}

		input.ShareType = "direct"
		input.TargetID = target.ID

		sh, err := h.shareService.Create(input)
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}

		// No file copy — shared files are served from owner's directory via WebDAV
		h.activityService.Log(agent.ID, "share", req.Path, map[string]interface{}{
			"type": "direct", "to": email,
		})
		h.activityService.Log(target.ID, "received", req.Path, map[string]interface{}{
			"from": agent.Email,
		})

		// Notify recipient
		go h.emailService.SendShareNotification(email, agent.Email, filepath.Base(req.Path))

		writeJSON(w, http.StatusCreated, sh)
	} else {
		writeError(w, http.StatusBadRequest, "either to_email or link=true is required")
	}
}

// ShareLink handles POST /api/share/link — create a link share
func (h *ShareHandler) ShareLink(w http.ResponseWriter, r *http.Request) {
	agent := GetAgent(r)
	if agent == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req struct {
		Path         string `json:"path"`
		ExpiresHours int    `json:"expires_hours,omitempty"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Path == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	if !h.fileManager.FileExists(agent.ID, req.Path) {
		writeError(w, http.StatusNotFound, "file not found: "+req.Path)
		return
	}

	input := shareService.CreateShareInput{
		OwnerID:    agent.ID,
		SourcePath: req.Path,
		ShareType:  "link",
	}
	if req.ExpiresHours > 0 {
		dur := time.Duration(req.ExpiresHours) * time.Hour
		input.ExpiresIn = &dur
	}

	sh, err := h.shareService.Create(input)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.activityService.Log(agent.ID, "share", req.Path, map[string]interface{}{
		"type": "link", "share_id": sh.ID, "url": sh.URL,
	})

	writeJSON(w, http.StatusCreated, sh)
}

func (h *ShareHandler) ListShared(w http.ResponseWriter, r *http.Request) {
	agent := GetAgent(r)
	if agent == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	byMe, err := h.shareService.ListByOwner(agent.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	withMe, err := h.shareService.ListByTarget(agent.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"shared_by_me":   byMe,
		"shared_with_me": withMe,
	})
}

func (h *ShareHandler) Revoke(w http.ResponseWriter, r *http.Request) {
	agent := GetAgent(r)
	if agent == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	shareID := r.PathValue("id")
	if shareID == "" {
		writeError(w, http.StatusBadRequest, "share ID is required")
		return
	}

	// Get the share to clean up files
	sh, err := h.shareService.GetByID(shareID)
	if err != nil {
		writeError(w, http.StatusNotFound, "share not found")
		return
	}

	if err := h.shareService.Revoke(shareID, agent.ID); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// No file cleanup needed — shares are references, not copies

	h.activityService.Log(agent.ID, "revoke", sh.SourcePath, map[string]interface{}{
		"share_id": shareID,
	})

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// ServeShareLink handles GET /s/:id — public endpoint for link shares
func (h *ShareHandler) ServeShareLink(w http.ResponseWriter, r *http.Request) {
	shareID := r.PathValue("id")
	if shareID == "" {
		writeError(w, http.StatusBadRequest, "share ID is required")
		return
	}

	sh, err := h.shareService.GetByID(shareID)
	if err != nil {
		writeError(w, http.StatusNotFound, "share not found")
		return
	}

	if sh.Revoked {
		writeError(w, http.StatusGone, "this share has been revoked")
		return
	}

	if sh.ExpiresAt != nil && time.Now().After(*sh.ExpiresAt) {
		writeError(w, http.StatusGone, "this share has expired")
		return
	}

	if sh.ShareType != "link" {
		writeError(w, http.StatusBadRequest, "not a link share")
		return
	}

	// Serve directly from owner's file — no copy
	filePath := filepath.Join(
		h.fileManager.AgentFilesPath(sh.OwnerID),
		sh.SourcePath,
	)
	if _, err := os.Stat(filePath); err != nil {
		writeError(w, http.StatusNotFound, "file not found")
		return
	}

	// Track bandwidth
	h.billingService.TrackBandwidth(sh.OwnerID, 0, fileSize(filePath))

	// Serve the file
	filename := filepath.Base(filePath)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	http.ServeFile(w, r, filePath)
}

func fileSize(path string) int64 {
	info, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return info.Size()
}

func parseDuration(s string) (time.Duration, error) {
	// Support formats like "7d", "24h", "30m"
	if len(s) < 2 {
		return 0, fmt.Errorf("invalid duration: %s", s)
	}

	unit := s[len(s)-1]
	val := s[:len(s)-1]

	var multiplier time.Duration
	switch unit {
	case 'd':
		multiplier = 24 * time.Hour
	case 'h':
		multiplier = time.Hour
	case 'm':
		multiplier = time.Minute
	default:
		return time.ParseDuration(s)
	}

	var n int
	_, err := fmt.Sscanf(val, "%d", &n)
	if err != nil {
		return 0, err
	}

	return time.Duration(n) * multiplier, nil
}
