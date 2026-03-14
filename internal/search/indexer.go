package search

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pidrive/pidrive/internal/db"
)

var indexableExtensions = map[string]bool{
	".txt": true, ".md": true, ".csv": true, ".json": true,
	".yaml": true, ".yml": true, ".toml": true, ".log": true,
	".py": true, ".js": true, ".ts": true, ".go": true,
	".rs": true, ".java": true, ".rb": true, ".sh": true,
	".sql": true, ".html": true, ".css": true, ".xml": true,
	".env": true, ".cfg": true, ".ini": true, ".conf": true,
}

const maxFileSize = 50 * 1024 * 1024 // 50 MB

type Indexer struct {
	db              *db.DB
	masterMountPath string
	interval        time.Duration
	stopCh          chan struct{}
}

func NewIndexer(db *db.DB, masterMountPath string, interval time.Duration) *Indexer {
	return &Indexer{
		db:              db,
		masterMountPath: masterMountPath,
		interval:        interval,
		stopCh:          make(chan struct{}),
	}
}

func (idx *Indexer) Start() {
	go func() {
		// Initial index after short delay
		time.Sleep(5 * time.Second)
		idx.indexAll()

		ticker := time.NewTicker(idx.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				idx.indexAll()
			case <-idx.stopCh:
				return
			}
		}
	}()
}

func (idx *Indexer) Stop() {
	close(idx.stopCh)
}

func (idx *Indexer) indexAll() {
	agentsDir := filepath.Join(idx.masterMountPath, "agents")
	entries, err := os.ReadDir(agentsDir)
	if err != nil {
		// Not an error on first run
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		agentID := entry.Name()
		// Skip non-UUID directory names
		if !isValidUUID(agentID) {
			continue
		}
		filesDir := filepath.Join(agentsDir, agentID, "files")
		if err := idx.indexAgent(agentID, filesDir); err != nil {
			log.Printf("[indexer] failed to index agent %s: %v", agentID[:8], err)
		}
	}
}

func isValidUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
		} else {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				return false
			}
		}
	}
	return true
}

// IndexAgent indexes a single agent's files
func (idx *Indexer) IndexAgent(agentID string) error {
	filesDir := filepath.Join(idx.masterMountPath, "agents", agentID, "files")
	return idx.indexAgent(agentID, filesDir)
}

func (idx *Indexer) indexAgent(agentID, filesDir string) error {
	seen := make(map[string]bool)

	err := filepath.WalkDir(filesDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if d.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(filesDir, path)
		seen[relPath] = true

		info, err := d.Info()
		if err != nil {
			return nil
		}

		// Skip large files
		if info.Size() > maxFileSize {
			return nil
		}

		ext := strings.ToLower(filepath.Ext(path))
		if !indexableExtensions[ext] {
			return nil
		}

		// Check if already indexed with same hash
		contentHash, err := hashFile(path)
		if err != nil {
			return nil
		}

		var existingHash string
		err = idx.db.QueryRow(
			"SELECT content_hash FROM search_index WHERE agent_id = $1 AND path = $2",
			agentID, relPath,
		).Scan(&existingHash)

		if err == nil && existingHash == contentHash {
			return nil // unchanged
		}

		// Read content
		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		// Truncate content for indexing (keep first 1MB)
		text := string(content)
		if len(text) > 1024*1024 {
			text = text[:1024*1024]
		}

		_, err = idx.db.Exec(`
			INSERT INTO search_index (agent_id, path, filename, content, size_bytes, modified_at, content_hash)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			ON CONFLICT (agent_id, path) DO UPDATE SET
				filename = EXCLUDED.filename,
				content = EXCLUDED.content,
				size_bytes = EXCLUDED.size_bytes,
				modified_at = EXCLUDED.modified_at,
				content_hash = EXCLUDED.content_hash,
				indexed_at = now()
		`, agentID, relPath, filepath.Base(path), text, info.Size(), info.ModTime(), contentHash)

		return nil
	})
	if err != nil {
		return err
	}

	// Remove stale entries
	rows, err := idx.db.Query("SELECT path FROM search_index WHERE agent_id = $1", agentID)
	if err != nil {
		return err
	}
	defer rows.Close()

	var stale []string
	for rows.Next() {
		var p string
		rows.Scan(&p)
		if !seen[p] {
			stale = append(stale, p)
		}
	}

	for _, p := range stale {
		idx.db.Exec("DELETE FROM search_index WHERE agent_id = $1 AND path = $2", agentID, p)
	}

	return nil
}

func hashFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:]), nil
}
