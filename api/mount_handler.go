package api

import (
	"net/http"

	"github.com/pidrive/pidrive/internal/activity"
	"github.com/pidrive/pidrive/internal/share"
)

type MountHandler struct {
	fileManager     *share.FileManager
	activityService *activity.ActivityService
}

func NewMountHandler(fileManager *share.FileManager, activityService *activity.ActivityService) *MountHandler {
	return &MountHandler{
		fileManager:     fileManager,
		activityService: activityService,
	}
}

func (h *MountHandler) Mount(w http.ResponseWriter, r *http.Request) {
	agent := GetAgent(r)
	if agent == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	// Ensure agent directories exist
	if err := h.fileManager.EnsureAgentDirs(agent.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create agent directories: "+err.Error())
		return
	}

	// Log activity
	h.activityService.Log(agent.ID, "mount", "", nil)

	// Return the agent's mount path (server-side mount, no raw creds exposed)
	mountPath := h.fileManager.AgentFilesPath(agent.ID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"mount_path": mountPath,
		"agent_id":   agent.ID,
		"status":     "mounted",
	})
}

func (h *MountHandler) Unmount(w http.ResponseWriter, r *http.Request) {
	agent := GetAgent(r)
	if agent == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	h.activityService.Log(agent.ID, "unmount", "", nil)

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
