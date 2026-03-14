package trash

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pidrive/pidrive/internal/db"
)

type TrashItem struct {
	Path             string    `json:"path"`
	DeletedAt        time.Time `json:"deleted_at"`
	RecoverableUntil time.Time `json:"recoverable_until"`
	SizeBytes        int64     `json:"size_bytes"`
}

type TrashService struct {
	db              *db.DB
	masterMountPath string
	trashDays       int
}

func NewTrashService(db *db.DB, masterMountPath string, trashDays int) *TrashService {
	return &TrashService{
		db:              db,
		masterMountPath: masterMountPath,
		trashDays:       trashDays,
	}
}

// trashDir returns the JuiceFS trash directory for an agent
// JuiceFS stores trash in .trash/ at the filesystem root with timestamps
func (s *TrashService) trashDir(agentID string) string {
	return filepath.Join(s.masterMountPath, "agents", agentID, "files", ".trash")
}

// List returns all items in an agent's trash
func (s *TrashService) List(agentID string) ([]TrashItem, error) {
	trashPath := s.trashDir(agentID)

	// JuiceFS trash may not exist if nothing was deleted
	if _, err := os.Stat(trashPath); os.IsNotExist(err) {
		return nil, nil
	}

	var items []TrashItem
	err := filepath.Walk(trashPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if path == trashPath {
			return nil
		}

		relPath, _ := filepath.Rel(trashPath, path)
		items = append(items, TrashItem{
			Path:             relPath,
			DeletedAt:        info.ModTime(),
			RecoverableUntil: info.ModTime().Add(time.Duration(s.trashDays) * 24 * time.Hour),
			SizeBytes:        info.Size(),
		})

		if info.IsDir() {
			return filepath.SkipDir // don't recurse into subdirs for listing
		}
		return nil
	})

	return items, err
}

// Restore moves a file from trash back to the agent's files directory
func (s *TrashService) Restore(agentID, trashRelPath string) error {
	trashPath := filepath.Join(s.trashDir(agentID), trashRelPath)
	if _, err := os.Stat(trashPath); os.IsNotExist(err) {
		return fmt.Errorf("file not found in trash: %s", trashRelPath)
	}

	// Restore to agent's files directory
	destPath := filepath.Join(s.masterMountPath, "agents", agentID, "files", filepath.Base(trashRelPath))

	// Handle conflict
	if _, err := os.Stat(destPath); err == nil {
		ext := filepath.Ext(destPath)
		base := destPath[:len(destPath)-len(ext)]
		destPath = fmt.Sprintf("%s (restored)%s", base, ext)
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return fmt.Errorf("failed to create restore dir: %w", err)
	}

	return os.Rename(trashPath, destPath)
}

// Empty permanently deletes all trash for an agent
func (s *TrashService) Empty(agentID string) error {
	trashPath := s.trashDir(agentID)
	if _, err := os.Stat(trashPath); os.IsNotExist(err) {
		return nil // nothing to empty
	}
	return os.RemoveAll(trashPath)
}
