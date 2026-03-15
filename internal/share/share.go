package share

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"time"

	"github.com/pidrive/pidrive/internal/db"
)

type Share struct {
	ID         string     `json:"id"`
	OwnerID    string     `json:"owner_id"`
	SourcePath string     `json:"source_path"`
	ShareType  string     `json:"share_type"`
	TargetID   *string    `json:"target_id,omitempty"`
	Permission string     `json:"permission"`
	CreatedAt  time.Time  `json:"created_at"`
	ExpiresAt  *time.Time `json:"expires_at,omitempty"`
	Revoked    bool       `json:"revoked"`

	// Populated by joins
	OwnerEmail  string `json:"owner_email,omitempty"`
	TargetEmail string `json:"target_email,omitempty"`
	URL         string `json:"url,omitempty"`
}

type ShareService struct {
	db        *db.DB
	serverURL string
}

func NewShareService(db *db.DB, serverURL string) *ShareService {
	return &ShareService{db: db, serverURL: serverURL}
}

func generateShareID() (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		b[i] = charset[n.Int64()]
	}
	return string(b), nil
}

type CreateShareInput struct {
	OwnerID     string
	SourcePath  string
	ShareType   string // "direct" or "link"
	TargetID    string // for direct shares (registered user)
	TargetEmail string // for pending shares (unregistered user)
	Permission  string
	ExpiresIn   *time.Duration
}

func (s *ShareService) Create(input CreateShareInput) (*Share, error) {
	id, err := generateShareID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate share ID: %w", err)
	}

	var expiresAt *time.Time
	if input.ExpiresIn != nil {
		t := time.Now().Add(*input.ExpiresIn)
		expiresAt = &t
	}

	permission := input.Permission
	if permission == "" {
		permission = "read"
	}

	share := &Share{}
	var targetID *string
	if input.TargetID != "" {
		targetID = &input.TargetID
	}
	var targetEmail *string
	if input.TargetEmail != "" {
		targetEmail = &input.TargetEmail
	}
	status := "active"
	if input.TargetID == "" && input.TargetEmail != "" {
		status = "pending"
	}

	err = s.db.QueryRow(`
		INSERT INTO shares (id, owner_id, source_path, share_type, target_id, target_email, permission, expires_at, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, owner_id, source_path, share_type, target_id, permission, created_at, expires_at, revoked
	`, id, input.OwnerID, input.SourcePath, input.ShareType, targetID, targetEmail, permission, expiresAt, status).Scan(
		&share.ID, &share.OwnerID, &share.SourcePath, &share.ShareType,
		&share.TargetID, &share.Permission, &share.CreatedAt, &share.ExpiresAt, &share.Revoked,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create share: %w", err)
	}

	if share.ShareType == "link" {
		share.URL = fmt.Sprintf("%s/s/%s", s.serverURL, share.ID)
	}

	return share, nil
}

func (s *ShareService) GetByID(id string) (*Share, error) {
	share := &Share{}
	err := s.db.QueryRow(`
		SELECT s.id, s.owner_id, s.source_path, s.share_type, s.target_id,
			s.permission, s.created_at, s.expires_at, s.revoked,
			a.email as owner_email
		FROM shares s
		JOIN agents a ON a.id = s.owner_id
		WHERE s.id = $1
	`, id).Scan(
		&share.ID, &share.OwnerID, &share.SourcePath, &share.ShareType,
		&share.TargetID, &share.Permission, &share.CreatedAt, &share.ExpiresAt,
		&share.Revoked, &share.OwnerEmail,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("share not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get share: %w", err)
	}

	if share.ShareType == "link" {
		share.URL = fmt.Sprintf("%s/s/%s", s.serverURL, share.ID)
	}

	return share, nil
}

// ListByOwner returns shares created by an agent
func (s *ShareService) ListByOwner(ownerID string) ([]Share, error) {
	rows, err := s.db.Query(`
		SELECT s.id, s.owner_id, s.source_path, s.share_type, s.target_id,
			s.permission, s.created_at, s.expires_at, s.revoked,
			COALESCE(t.email, '') as target_email
		FROM shares s
		LEFT JOIN agents t ON t.id = s.target_id
		WHERE s.owner_id = $1 AND NOT s.revoked
		ORDER BY s.created_at DESC
	`, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to list shares: %w", err)
	}
	defer rows.Close()

	return s.scanShares(rows)
}

// ListByTarget returns shares shared with an agent
func (s *ShareService) ListByTarget(targetID string) ([]Share, error) {
	rows, err := s.db.Query(`
		SELECT s.id, s.owner_id, s.source_path, s.share_type, s.target_id,
			s.permission, s.created_at, s.expires_at, s.revoked,
			a.email as owner_email
		FROM shares s
		JOIN agents a ON a.id = s.owner_id
		WHERE s.target_id = $1 AND NOT s.revoked
			AND (s.expires_at IS NULL OR s.expires_at > now())
		ORDER BY s.created_at DESC
	`, targetID)
	if err != nil {
		return nil, fmt.Errorf("failed to list shares: %w", err)
	}
	defer rows.Close()

	var shares []Share
	for rows.Next() {
		var sh Share
		var ownerEmail string
		if err := rows.Scan(
			&sh.ID, &sh.OwnerID, &sh.SourcePath, &sh.ShareType,
			&sh.TargetID, &sh.Permission, &sh.CreatedAt, &sh.ExpiresAt,
			&sh.Revoked, &ownerEmail,
		); err != nil {
			return nil, fmt.Errorf("failed to scan share: %w", err)
		}
		sh.OwnerEmail = ownerEmail
		if sh.ShareType == "link" {
			sh.URL = fmt.Sprintf("%s/s/%s", s.serverURL, sh.ID)
		}
		shares = append(shares, sh)
	}
	return shares, nil
}

func (s *ShareService) scanShares(rows *sql.Rows) ([]Share, error) {
	var shares []Share
	for rows.Next() {
		var sh Share
		var extra string
		if err := rows.Scan(
			&sh.ID, &sh.OwnerID, &sh.SourcePath, &sh.ShareType,
			&sh.TargetID, &sh.Permission, &sh.CreatedAt, &sh.ExpiresAt,
			&sh.Revoked, &extra,
		); err != nil {
			return nil, fmt.Errorf("failed to scan share: %w", err)
		}
		if sh.ShareType == "direct" {
			sh.TargetEmail = extra
		} else {
			sh.OwnerEmail = extra
		}
		if sh.ShareType == "link" {
			sh.URL = fmt.Sprintf("%s/s/%s", s.serverURL, sh.ID)
		}
		shares = append(shares, sh)
	}
	return shares, nil
}

// Revoke marks a share as revoked
func (s *ShareService) Revoke(shareID, ownerID string) error {
	result, err := s.db.Exec(`
		UPDATE shares SET revoked = TRUE, revoked_at = now()
		WHERE id = $1 AND owner_id = $2 AND NOT revoked
	`, shareID, ownerID)
	if err != nil {
		return fmt.Errorf("failed to revoke share: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("share not found or already revoked")
	}
	return nil
}

// RevokeByPathAndTarget revokes a direct share by path and target email
func (s *ShareService) RevokeByPathAndTarget(ownerID, path, targetID string) error {
	result, err := s.db.Exec(`
		UPDATE shares SET revoked = TRUE, revoked_at = now()
		WHERE owner_id = $1 AND source_path = $2 AND target_id = $3
			AND share_type = 'direct' AND NOT revoked
	`, ownerID, path, targetID)
	if err != nil {
		return fmt.Errorf("failed to revoke share: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("share not found or already revoked")
	}
	return nil
}

// ActivatePendingShares converts pending shares to active when a user registers
func (s *ShareService) ActivatePendingShares(email, agentID string) (int64, error) {
	result, err := s.db.Exec(`
		UPDATE shares SET target_id = $1, status = 'active'
		WHERE target_email = $2 AND status = 'pending' AND NOT revoked
	`, agentID, email)
	if err != nil {
		return 0, fmt.Errorf("failed to activate pending shares: %w", err)
	}
	return result.RowsAffected()
}

// CleanupExpired removes expired link shares
func (s *ShareService) CleanupExpired() (int64, error) {
	result, err := s.db.Exec(`
		UPDATE shares SET revoked = TRUE, revoked_at = now()
		WHERE share_type = 'link' AND NOT revoked
			AND expires_at IS NOT NULL AND expires_at < now()
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired shares: %w", err)
	}
	return result.RowsAffected()
}

// FindDirectShare finds an active direct share between owner and target for a path
func (s *ShareService) FindDirectShare(ownerID, path, targetID string) (*Share, error) {
	share := &Share{}
	err := s.db.QueryRow(`
		SELECT id, owner_id, source_path, share_type, target_id, permission,
			created_at, expires_at, revoked
		FROM shares
		WHERE owner_id = $1 AND source_path = $2 AND target_id = $3
			AND share_type = 'direct' AND NOT revoked
	`, ownerID, path, targetID).Scan(
		&share.ID, &share.OwnerID, &share.SourcePath, &share.ShareType,
		&share.TargetID, &share.Permission, &share.CreatedAt, &share.ExpiresAt, &share.Revoked,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find share: %w", err)
	}
	return share, nil
}
