package api

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"golang.org/x/net/webdav"
)

// Per-agent WebDAV handlers with shared lock state
var (
	agentHandlers   = make(map[string]*webdav.Handler)
	agentHandlersMu sync.Mutex
)

// InvalidateAgentHandler removes the cached handler so it gets recreated with fresh share data
func InvalidateAgentHandler(agentID string) {
	agentHandlersMu.Lock()
	defer agentHandlersMu.Unlock()
	delete(agentHandlers, agentID)
}

func (s *Server) getAgentWebDAVHandler(agentID string) *webdav.Handler {
	agentHandlersMu.Lock()
	defer agentHandlersMu.Unlock()

	if h, ok := agentHandlers[agentID]; ok {
		return h
	}

	// Ensure agent dirs exist
	agentRoot := filepath.Join(s.cfg.JuiceFSMountPath, "agents", agentID, "files")
	os.MkdirAll(agentRoot, 0755)

	h := &webdav.Handler{
		Prefix: "/webdav",
		FileSystem: &agentFS{
			agentID:      agentID,
			mountPath:    s.cfg.JuiceFSMountPath,
			shareService: s.shareService,
			fileManager:  s.fileManager,
		},
		LockSystem: webdav.NewMemLS(),
	}
	agentHandlers[agentID] = h
	return h
}

func (s *Server) serveWebDAV(w http.ResponseWriter, r *http.Request) {
	// OPTIONS must respond without auth for DAV discovery
	if r.Method == "OPTIONS" {
		w.Header().Set("DAV", "1, 2")
		w.Header().Set("Allow", "OPTIONS, GET, HEAD, PUT, DELETE, MKCOL, COPY, MOVE, PROPFIND, PROPPATCH, LOCK, UNLOCK")
		w.Header().Set("MS-Author-Via", "DAV")
		w.WriteHeader(http.StatusOK)
		return
	}

	// Authenticate via Basic Auth (username=anything, password=API key)
	_, apiKey, ok := r.BasicAuth()
	if !ok || apiKey == "" {
		w.Header().Set("DAV", "1, 2")
		w.Header().Set("WWW-Authenticate", `Basic realm="pidrive"`)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	agent, err := s.authService.Authenticate(apiKey)
	if err != nil || !agent.Verified {
		w.Header().Set("WWW-Authenticate", `Basic realm="pidrive"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Set DAV headers on all responses
	w.Header().Set("DAV", "1, 2")

	handler := s.getAgentWebDAVHandler(agent.ID)

	log.Printf("[webdav] %s %s (agent: %s)", r.Method, r.URL.Path, agent.Email)
	handler.ServeHTTP(w, r)
}
