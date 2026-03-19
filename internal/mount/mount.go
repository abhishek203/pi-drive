package mount

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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
