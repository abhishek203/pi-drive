package share

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// FileManager handles file copy operations for sharing
type FileManager struct {
	masterMountPath string
}

func NewFileManager(masterMountPath string) *FileManager {
	return &FileManager{masterMountPath: masterMountPath}
}

// AgentFilesPath returns the absolute path to an agent's files directory
func (f *FileManager) AgentFilesPath(agentID string) string {
	return filepath.Join(f.masterMountPath, "agents", agentID, "files")
}

// SharedDirPath returns the absolute path to an agent's incoming shared directory
func (f *FileManager) SharedDirPath(agentID string) string {
	return filepath.Join(f.masterMountPath, "shared", "direct", "to-"+agentID)
}

// LinkDirPath returns the absolute path to a link share directory
func (f *FileManager) LinkDirPath(shareID string) string {
	return filepath.Join(f.masterMountPath, "shared", "links", shareID)
}

// EnsureAgentDirs creates the agent's directories if they don't exist
func (f *FileManager) EnsureAgentDirs(agentID string) error {
	dirs := []string{
		f.AgentFilesPath(agentID),
		f.SharedDirPath(agentID),
	}
	for _, d := range dirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", d, err)
		}
	}
	return nil
}

// FileExists checks if a file exists in an agent's files directory
func (f *FileManager) FileExists(agentID, relPath string) bool {
	absPath := filepath.Join(f.AgentFilesPath(agentID), relPath)
	_, err := os.Stat(absPath)
	return err == nil
}

// CopyToShared copies a file from an agent's files to another agent's shared directory
func (f *FileManager) CopyToShared(ownerID, relPath, targetID string) (string, error) {
	srcPath := filepath.Join(f.AgentFilesPath(ownerID), relPath)
	filename := filepath.Base(relPath)

	destDir := f.SharedDirPath(targetID)
	destPath := filepath.Join(destDir, filename)

	// Handle filename conflicts
	if _, err := os.Stat(destPath); err == nil {
		ext := filepath.Ext(filename)
		name := strings.TrimSuffix(filename, ext)
		destPath = filepath.Join(destDir, fmt.Sprintf("%s (from %s)%s", name, ownerID[:8], ext))
	}

	if err := copyFileOrDir(srcPath, destPath); err != nil {
		return "", fmt.Errorf("failed to copy to shared: %w", err)
	}

	return filepath.Base(destPath), nil
}

// CopyToLink copies a file to the links directory
func (f *FileManager) CopyToLink(ownerID, relPath, shareID string) error {
	srcPath := filepath.Join(f.AgentFilesPath(ownerID), relPath)
	destDir := f.LinkDirPath(shareID)

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create link dir: %w", err)
	}

	destPath := filepath.Join(destDir, filepath.Base(relPath))
	return copyFileOrDir(srcPath, destPath)
}

// RemoveFromShared removes a shared file from a target agent's shared directory
func (f *FileManager) RemoveFromShared(targetID, filename string) error {
	path := filepath.Join(f.SharedDirPath(targetID), filename)
	return os.RemoveAll(path)
}

// RemoveLink removes a link share directory
func (f *FileManager) RemoveLink(shareID string) error {
	return os.RemoveAll(f.LinkDirPath(shareID))
}

// GetLinkFilePath returns the file path for a link share
func (f *FileManager) GetLinkFilePath(shareID string) (string, error) {
	linkDir := f.LinkDirPath(shareID)
	entries, err := os.ReadDir(linkDir)
	if err != nil {
		return "", fmt.Errorf("link share directory not found: %w", err)
	}
	if len(entries) == 0 {
		return "", fmt.Errorf("link share directory is empty")
	}
	return filepath.Join(linkDir, entries[0].Name()), nil
}

// GetUsedBytes returns the total size of files in an agent's directory
func (f *FileManager) GetUsedBytes(agentID string) (int64, error) {
	var total int64
	err := filepath.Walk(f.AgentFilesPath(agentID), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	return total, err
}

func copyFileOrDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if info.IsDir() {
		return copyDir(src, dst)
	}
	return copyFile(src, dst)
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}

func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())
		if err := copyFileOrDir(srcPath, dstPath); err != nil {
			return err
		}
	}
	return nil
}
