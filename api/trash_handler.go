package api

import (
	"net/http"

	"github.com/pidrive/pidrive/internal/activity"
	"github.com/pidrive/pidrive/internal/trash"
)

type TrashHandler struct {
	trashService    *trash.TrashService
	activityService *activity.ActivityService
}

func NewTrashHandler(trashService *trash.TrashService, activityService *activity.ActivityService) *TrashHandler {
	return &TrashHandler{trashService: trashService, activityService: activityService}
}

func (h *TrashHandler) List(w http.ResponseWriter, r *http.Request) {
	agent := GetAgent(r)
	if agent == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	items, err := h.trashService.List(agent.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if items == nil {
		items = []trash.TrashItem{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
	})
}

func (h *TrashHandler) Restore(w http.ResponseWriter, r *http.Request) {
	agent := GetAgent(r)
	if agent == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req struct {
		Path string `json:"path"`
	}
	if err := decodeJSON(r, &req); err != nil || req.Path == "" {
		writeError(w, http.StatusBadRequest, "path is required")
		return
	}

	if err := h.trashService.Restore(agent.ID, req.Path); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.activityService.Log(agent.ID, "restore", req.Path, nil)

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *TrashHandler) Empty(w http.ResponseWriter, r *http.Request) {
	agent := GetAgent(r)
	if agent == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	if err := h.trashService.Empty(agent.ID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.activityService.Log(agent.ID, "trash_empty", "", nil)

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
