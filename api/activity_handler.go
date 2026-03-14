package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/pidrive/pidrive/internal/activity"
)

type ActivityHandler struct {
	activityService *activity.ActivityService
}

func NewActivityHandler(activityService *activity.ActivityService) *ActivityHandler {
	return &ActivityHandler{activityService: activityService}
}

func (h *ActivityHandler) List(w http.ResponseWriter, r *http.Request) {
	agent := GetAgent(r)
	if agent == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	opts := activity.ListOptions{}

	if since := r.URL.Query().Get("since"); since != "" {
		dur, err := parseDuration(since)
		if err == nil {
			t := time.Now().Add(-dur)
			opts.Since = &t
		}
	}

	if action := r.URL.Query().Get("type"); action != "" {
		opts.Action = action
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if n, err := strconv.Atoi(limit); err == nil {
			opts.Limit = n
		}
	}

	events, err := h.activityService.List(agent.ID, opts)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if events == nil {
		events = []activity.Event{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"events": events,
	})
}
