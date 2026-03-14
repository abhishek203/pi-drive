package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pidrive/pidrive/internal/search"
)

type SearchHandler struct {
	searchService *search.SearchService
	indexer       *search.Indexer
}

func NewSearchHandler(searchService *search.SearchService, indexer *search.Indexer) *SearchHandler {
	return &SearchHandler{searchService: searchService, indexer: indexer}
}

func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	agent := GetAgent(r)
	if agent == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	q := r.URL.Query().Get("q")
	if q == "" {
		writeError(w, http.StatusBadRequest, "q parameter is required")
		return
	}

	opts := search.SearchOptions{
		Query:   q,
		AgentID: agent.ID,
	}

	if types := r.URL.Query().Get("type"); types != "" {
		opts.Types = strings.Split(types, ",")
	}

	if modified := r.URL.Query().Get("modified"); modified != "" {
		dur, err := parseDuration(modified)
		if err == nil {
			opts.ModifiedIn = &dur
		}
	}

	if r.URL.Query().Get("my_only") == "true" {
		opts.MyOnly = true
	}
	if r.URL.Query().Get("shared_only") == "true" {
		opts.SharedOnly = true
	}

	if limit := r.URL.Query().Get("limit"); limit != "" {
		if n, err := strconv.Atoi(limit); err == nil {
			opts.Limit = n
		}
	}

	results, err := h.searchService.Search(opts)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "search failed: "+err.Error())
		return
	}

	if results == nil {
		results = []search.Result{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"query":   q,
		"count":   len(results),
		"results": results,
	})
}

func (h *SearchHandler) TriggerIndex(w http.ResponseWriter, r *http.Request) {
	agent := GetAgent(r)
	if agent == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	go func() {
		h.indexer.IndexAgent(agent.ID)
	}()

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "indexing started",
	})
}

// parseDurationSearch is a helper that matches the parseDuration in share_handler
// but we can reuse the one from share_handler since they're in the same package
func init() {
	// parseDuration is defined in share_handler.go and available here
	_ = time.Now // just to use the time import
}
