package api

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/net/webdav"

	"github.com/pidrive/pidrive/internal/share"
)

// agentFS is a virtual filesystem that shows:
//
//	/my/       → agent's own files (read/write)
//	/shared/   → files shared with this agent (read-only, from other agents' dirs)
type agentFS struct {
	agentID      string
	mountPath    string
	shareService *share.ShareService
	fileManager  *share.FileManager
}

func (fs *agentFS) myRoot() string {
	return filepath.Join(fs.mountPath, "agents", fs.agentID, "files")
}

func (fs *agentFS) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	if isSharedPath(name) {
		return os.ErrPermission
	}
	realPath := fs.resolveMy(name)
	if realPath == "" {
		return os.ErrPermission
	}
	return os.MkdirAll(realPath, perm)
}

func (fs *agentFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	name = cleanPath(name)

	// Root directory
	if name == "/" || name == "" {
		return &virtualDir{name: "/", entries: []os.FileInfo{
			&dirInfo{name: "my"},
			&dirInfo{name: "shared"},
		}}, nil
	}

	// /my or /my/...
	if name == "/my" {
		return os.Open(fs.myRoot())
	}
	if strings.HasPrefix(name, "/my/") {
		realPath := filepath.Join(fs.myRoot(), name[4:])
		os.MkdirAll(filepath.Dir(realPath), 0755)
		return os.OpenFile(realPath, flag, perm)
	}

	// /shared
	if name == "/shared" {
		return fs.openSharedRoot()
	}

	// /shared/<email>/...
	if strings.HasPrefix(name, "/shared/") {
		return fs.openSharedFile(name[8:], flag)
	}

	return nil, os.ErrNotExist
}

func (fs *agentFS) RemoveAll(ctx context.Context, name string) error {
	if isSharedPath(name) {
		return os.ErrPermission
	}
	realPath := fs.resolveMy(name)
	if realPath == "" {
		return os.ErrPermission
	}
	return os.RemoveAll(realPath)
}

func (fs *agentFS) Rename(ctx context.Context, oldName, newName string) error {
	if isSharedPath(oldName) || isSharedPath(newName) {
		return os.ErrPermission
	}
	oldReal := fs.resolveMy(oldName)
	newReal := fs.resolveMy(newName)
	if oldReal == "" || newReal == "" {
		return os.ErrPermission
	}
	return os.Rename(oldReal, newReal)
}

func (fs *agentFS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	name = cleanPath(name)

	if name == "/" || name == "" {
		return &dirInfo{name: "/"}, nil
	}
	if name == "/my" {
		return &dirInfo{name: "my"}, nil
	}
	if name == "/shared" {
		return &dirInfo{name: "shared"}, nil
	}
	if strings.HasPrefix(name, "/my/") {
		return os.Stat(filepath.Join(fs.myRoot(), name[4:]))
	}
	if strings.HasPrefix(name, "/shared/") {
		return fs.statShared(name[8:])
	}
	return nil, os.ErrNotExist
}

// --- shared directory helpers ---

func (fs *agentFS) openSharedRoot() (webdav.File, error) {
	shares, err := fs.shareService.ListByTarget(fs.agentID)
	if err != nil {
		return nil, err
	}

	// Group by owner email
	seen := make(map[string]bool)
	var entries []os.FileInfo
	for _, sh := range shares {
		if sh.OwnerEmail != "" && !seen[sh.OwnerEmail] {
			seen[sh.OwnerEmail] = true
			entries = append(entries, &dirInfo{name: sh.OwnerEmail})
		}
	}
	return &virtualDir{name: "shared", entries: entries}, nil
}

func (fs *agentFS) openSharedFile(relPath string, flag int) (webdav.File, error) {
	// relPath is like "alice@company.com/report.txt" or "alice@company.com"
	parts := strings.SplitN(relPath, "/", 2)
	email := parts[0]

	shares, err := fs.shareService.ListByTarget(fs.agentID)
	if err != nil {
		return nil, err
	}

	if len(parts) == 1 {
		// List files shared by this email
		var entries []os.FileInfo
		for _, sh := range shares {
			if sh.OwnerEmail == email {
				// Stat the actual file from owner's dir
				ownerFilePath := filepath.Join(fs.mountPath, "agents", sh.OwnerID, "files", sh.SourcePath)
				info, err := os.Stat(ownerFilePath)
				if err == nil {
					entries = append(entries, info)
				}
			}
		}
		return &virtualDir{name: email, entries: entries}, nil
	}

	// Open specific file: check if share exists
	filename := parts[1]
	for _, sh := range shares {
		if sh.OwnerEmail == email && (sh.SourcePath == filename || filepath.Base(sh.SourcePath) == filename) {
			ownerFilePath := filepath.Join(fs.mountPath, "agents", sh.OwnerID, "files", sh.SourcePath)
			return os.OpenFile(ownerFilePath, os.O_RDONLY, 0)
		}
	}
	return nil, os.ErrNotExist
}

func (fs *agentFS) statShared(relPath string) (os.FileInfo, error) {
	parts := strings.SplitN(relPath, "/", 2)
	email := parts[0]

	shares, err := fs.shareService.ListByTarget(fs.agentID)
	if err != nil {
		return nil, err
	}

	if len(parts) == 1 {
		// Check if this email has shared anything with us
		for _, sh := range shares {
			if sh.OwnerEmail == email {
				return &dirInfo{name: email}, nil
			}
		}
		return nil, os.ErrNotExist
	}

	filename := parts[1]
	for _, sh := range shares {
		if sh.OwnerEmail == email && (sh.SourcePath == filename || filepath.Base(sh.SourcePath) == filename) {
			ownerFilePath := filepath.Join(fs.mountPath, "agents", sh.OwnerID, "files", sh.SourcePath)
			return os.Stat(ownerFilePath)
		}
	}
	return nil, os.ErrNotExist
}

// --- helpers ---

func cleanPath(name string) string {
	name = filepath.Clean("/" + name)
	if name == "." {
		return "/"
	}
	return name
}

func isSharedPath(name string) bool {
	name = cleanPath(name)
	return name == "/shared" || strings.HasPrefix(name, "/shared/")
}

func (fs *agentFS) resolveMy(name string) string {
	name = cleanPath(name)
	if strings.HasPrefix(name, "/my/") {
		return filepath.Join(fs.myRoot(), name[4:])
	}
	if name == "/my" {
		return fs.myRoot()
	}
	return ""
}

// --- virtual directory ---

type virtualDir struct {
	name    string
	entries []os.FileInfo
	pos     int
}

func (d *virtualDir) Close() error                   { return nil }
func (d *virtualDir) Read([]byte) (int, error)       { return 0, fmt.Errorf("is a directory") }
func (d *virtualDir) Write([]byte) (int, error)      { return 0, fmt.Errorf("is a directory") }
func (d *virtualDir) Seek(int64, int) (int64, error) { return 0, fmt.Errorf("is a directory") }

func (d *virtualDir) Readdir(count int) ([]os.FileInfo, error) {
	if d.pos >= len(d.entries) {
		if count <= 0 {
			return nil, nil
		}
		return nil, io.EOF
	}
	if count <= 0 || count > len(d.entries)-d.pos {
		count = len(d.entries) - d.pos
	}
	entries := d.entries[d.pos : d.pos+count]
	d.pos += count
	if d.pos >= len(d.entries) {
		return entries, io.EOF
	}
	return entries, nil
}

func (d *virtualDir) Stat() (os.FileInfo, error) {
	return &dirInfo{name: d.name}, nil
}

// --- dirInfo for virtual directories ---

type dirInfo struct {
	name string
}

func (d *dirInfo) Name() string       { return d.name }
func (d *dirInfo) Size() int64        { return 0 }
func (d *dirInfo) Mode() os.FileMode  { return os.ModeDir | 0755 }
func (d *dirInfo) ModTime() time.Time { return time.Now() }
func (d *dirInfo) IsDir() bool        { return true }
func (d *dirInfo) Sys() interface{}   { return nil }
