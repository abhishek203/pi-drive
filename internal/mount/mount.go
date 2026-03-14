package mount

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pidrive/pidrive/internal/config"
)

type MountService struct {
	cfg *config.Config
}

func NewMountService(cfg *config.Config) *MountService {
	return &MountService{cfg: cfg}
}

// FormatJuiceFS formats a new JuiceFS filesystem (run once on server init)
func (m *MountService) FormatJuiceFS() error {
	// Check if already formatted by trying to mount
	cmd := exec.Command("juicefs", "status", m.cfg.RedisURL)
	if err := cmd.Run(); err == nil {
		return nil // already formatted
	}

	args := []string{
		"format",
		"--storage", "s3",
		"--bucket", fmt.Sprintf("%s/%s", m.cfg.S3Endpoint, m.cfg.S3Bucket),
		"--access-key", m.cfg.S3AccessKey,
		"--secret-key", m.cfg.S3SecretKey,
		"--trash-days", fmt.Sprintf("%d", m.cfg.JuiceFSTrashDays),
	}

	args = append(args, m.cfg.RedisURL, "pidrive")

	cmd = exec.Command("juicefs", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("juicefs format failed: %w", err)
	}

	return nil
}

// MountMaster mounts the master JuiceFS filesystem (server-side)
func (m *MountService) MountMaster() error {
	mountPath := m.cfg.JuiceFSMountPath

	// Check if already mounted
	if isMounted(mountPath) {
		return nil
	}

	if err := os.MkdirAll(mountPath, 0755); err != nil {
		return fmt.Errorf("failed to create mount path: %w", err)
	}

	args := []string{
		"mount",
		m.cfg.RedisURL,
		mountPath,
		"--cache-dir", m.cfg.JuiceFSCacheDir,
		"--cache-size", fmt.Sprintf("%d", m.cfg.JuiceFSCacheMB),
		"--background",
	}

	cmd := exec.Command("juicefs", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("juicefs mount failed: %w", err)
	}

	// Create base directories
	dirs := []string{
		filepath.Join(mountPath, "agents"),
		filepath.Join(mountPath, "shared", "direct"),
		filepath.Join(mountPath, "shared", "links"),
	}
	for _, d := range dirs {
		os.MkdirAll(d, 0755)
	}

	return nil
}

// UnmountMaster unmounts the master JuiceFS filesystem
func (m *MountService) UnmountMaster() error {
	cmd := exec.Command("juicefs", "umount", m.cfg.JuiceFSMountPath)
	return cmd.Run()
}

// MountConfig contains info needed by the CLI to set up agent mounts
type MountConfig struct {
	AgentID         string `json:"agent_id"`
	RedisURL        string `json:"redis_url"`
	S3Endpoint      string `json:"s3_endpoint"`
	S3Bucket        string `json:"s3_bucket"`
	S3AccessKey     string `json:"s3_access_key"`
	S3SecretKey     string `json:"s3_secret_key"`
	CacheDir        string `json:"cache_dir"`
	CacheSizeMB     int    `json:"cache_size_mb"`
	MasterMountPath string `json:"master_mount_path"`
}

// GetMountConfig returns the config an agent CLI needs to mount
func (m *MountService) GetMountConfig(agentID string) *MountConfig {
	return &MountConfig{
		AgentID:         agentID,
		RedisURL:        m.cfg.RedisURL,
		S3Endpoint:      m.cfg.S3Endpoint,
		S3Bucket:        m.cfg.S3Bucket,
		S3AccessKey:     m.cfg.S3AccessKey,
		S3SecretKey:     m.cfg.S3SecretKey,
		CacheDir:        m.cfg.JuiceFSCacheDir,
		CacheSizeMB:     m.cfg.JuiceFSCacheMB,
		MasterMountPath: m.cfg.JuiceFSMountPath,
	}
}

// SetupAgentMount sets up bind mounts for an agent (called on the agent's machine)
func SetupAgentMount(mountCfg *MountConfig, drivePath string) error {
	// First, mount JuiceFS if not already mounted
	masterPath := mountCfg.MasterMountPath
	if !isMounted(masterPath) {
		os.MkdirAll(masterPath, 0755)

		args := []string{
			"mount",
			mountCfg.RedisURL,
			masterPath,
			"--cache-dir", mountCfg.CacheDir,
			"--cache-size", fmt.Sprintf("%d", mountCfg.CacheSizeMB),
			"--background",
		}

		cmd := exec.Command("juicefs", args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("juicefs mount failed: %w", err)
		}
	}

	myPath := filepath.Join(drivePath, "my")
	sharedPath := filepath.Join(drivePath, "shared")

	os.MkdirAll(myPath, 0755)
	os.MkdirAll(sharedPath, 0755)

	agentFilesPath := filepath.Join(masterPath, "agents", mountCfg.AgentID, "files")
	agentSharedPath := filepath.Join(masterPath, "shared", "direct", "to-"+mountCfg.AgentID)

	os.MkdirAll(agentFilesPath, 0755)
	os.MkdirAll(agentSharedPath, 0755)

	// Bind mount
	if !isMounted(myPath) {
		if err := bindMount(agentFilesPath, myPath); err != nil {
			return fmt.Errorf("failed to bind mount files: %w", err)
		}
	}

	if !isMounted(sharedPath) {
		if err := bindMount(agentSharedPath, sharedPath); err != nil {
			return fmt.Errorf("failed to bind mount shared: %w", err)
		}
	}

	// Write agent ID file (read-only)
	idFile := filepath.Join(drivePath, ".agent-id")
	os.WriteFile(idFile, []byte(mountCfg.AgentID), 0444)

	return nil
}

// TeardownAgentMount unmounts an agent's drive
func TeardownAgentMount(drivePath string) error {
	myPath := filepath.Join(drivePath, "my")
	sharedPath := filepath.Join(drivePath, "shared")

	unmount(sharedPath)
	unmount(myPath)

	return nil
}

func bindMount(src, dst string) error {
	if runtime.GOOS == "linux" {
		cmd := exec.Command("mount", "--bind", src, dst)
		return cmd.Run()
	}
	// macOS: use bindfs or symlink as fallback
	if _, err := exec.LookPath("bindfs"); err == nil {
		cmd := exec.Command("bindfs", src, dst)
		return cmd.Run()
	}
	// Fallback: symlink (less isolation, but works)
	os.Remove(dst)
	return os.Symlink(src, dst)
}

func unmount(path string) error {
	if !isMounted(path) {
		return nil
	}
	if runtime.GOOS == "linux" {
		return exec.Command("umount", path).Run()
	}
	return exec.Command("umount", path).Run()
}

func isMounted(path string) bool {
	if runtime.GOOS == "linux" {
		out, err := exec.Command("mountpoint", "-q", path).CombinedOutput()
		_ = out
		return err == nil
	}
	// macOS: check mount output
	out, err := exec.Command("mount").Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(out), path)
}
