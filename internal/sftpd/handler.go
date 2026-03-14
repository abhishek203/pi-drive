package sftpd

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/sftp"
)

// chrootHandler implements sftp.Handlers rooted at a specific directory
type chrootHandler struct {
	root string
}

func (h *chrootHandler) resolve(path string) string {
	// Clean and make relative
	clean := filepath.Clean("/" + path)
	// Prevent escaping root
	full := filepath.Join(h.root, clean)
	if !strings.HasPrefix(full, h.root) {
		return h.root
	}
	return full
}

// FileReader - implements sftp.FileReader
func (h *chrootHandler) Fileread(r *sftp.Request) (io.ReaderAt, error) {
	path := h.resolve(r.Filepath)
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// FileWriter - implements sftp.FileWriter
func (h *chrootHandler) Filewrite(r *sftp.Request) (io.WriterAt, error) {
	path := h.resolve(r.Filepath)
	os.MkdirAll(filepath.Dir(path), 0755)

	flags := os.O_WRONLY | os.O_CREATE
	if r.Pflags().Trunc {
		flags |= os.O_TRUNC
	}
	if r.Pflags().Append {
		flags |= os.O_APPEND
	}

	f, err := os.OpenFile(path, flags, 0644)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// FileCmder - implements sftp.FileCmder
func (h *chrootHandler) Filecmd(r *sftp.Request) error {
	switch r.Method {
	case "Setstat":
		return nil
	case "Rename":
		src := h.resolve(r.Filepath)
		dst := h.resolve(r.Target)
		return os.Rename(src, dst)
	case "Rmdir":
		return os.RemoveAll(h.resolve(r.Filepath))
	case "Remove":
		return os.Remove(h.resolve(r.Filepath))
	case "Mkdir":
		return os.MkdirAll(h.resolve(r.Filepath), 0755)
	case "Symlink":
		return nil // disallow symlinks
	}
	return sftp.ErrSSHFxOpUnsupported
}

// FileLister - implements sftp.FileLister
func (h *chrootHandler) Filelist(r *sftp.Request) (sftp.ListerAt, error) {
	path := h.resolve(r.Filepath)

	switch r.Method {
	case "List":
		entries, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}
		infos := make([]os.FileInfo, 0, len(entries))
		for _, e := range entries {
			info, err := e.Info()
			if err != nil {
				continue
			}
			infos = append(infos, info)
		}
		return listerat(infos), nil

	case "Stat":
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		return listerat([]os.FileInfo{info}), nil

	case "Readlink":
		target, err := os.Readlink(path)
		if err != nil {
			return nil, err
		}
		return listerat([]os.FileInfo{fakeFileInfo(target)}), nil
	}

	return nil, sftp.ErrSSHFxOpUnsupported
}

// listerat implements sftp.ListerAt
type listerat []os.FileInfo

func (l listerat) ListAt(ls []os.FileInfo, offset int64) (int, error) {
	if offset >= int64(len(l)) {
		return 0, io.EOF
	}
	n := copy(ls, l[offset:])
	if n < len(ls) {
		return n, io.EOF
	}
	return n, nil
}

// fakeFileInfo for readlink results
type fakeInfo struct {
	name string
}

func fakeFileInfo(name string) os.FileInfo    { return &fakeInfo{name: name} }
func (f *fakeInfo) Name() string              { return f.name }
func (f *fakeInfo) Size() int64               { return 0 }
func (f *fakeInfo) Mode() os.FileMode         { return 0644 }
func (f *fakeInfo) ModTime() time.Time        { return time.Time{} }
func (f *fakeInfo) IsDir() bool               { return false }
func (f *fakeInfo) Sys() interface{}          { return nil }
