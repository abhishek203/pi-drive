package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/pidrive/pidrive/internal/db"
)

type Agent struct {
	ID                   string    `json:"id"`
	Email                string    `json:"email"`
	Name                 string    `json:"name"`
	Plan                 string    `json:"plan"`
	QuotaBytes           int64     `json:"quota_bytes"`
	UsedBytes            int64     `json:"used_bytes"`
	CreatedAt            time.Time `json:"created_at"`
	Verified             bool      `json:"verified"`
	VerificationCode     *string   `json:"-"`
	VerificationExpires  *time.Time `json:"-"`
}

type AuthService struct {
	db *db.DB
}

func NewAuthService(db *db.DB) *AuthService {
	return &AuthService{db: db}
}

// GenerateAPIKey creates a new API key with pk_ prefix
func GenerateAPIKey() (key string, hash string, prefix string, err error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 32)
	for i := range b {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", "", "", err
		}
		b[i] = charset[n.Int64()]
	}
	key = "pk_" + string(b)
	prefix = key[:10]

	h := sha256.Sum256([]byte(key))
	hash = hex.EncodeToString(h[:])

	return key, hash, prefix, nil
}

// GenerateVerificationCode creates a 6-digit code
func GenerateVerificationCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(999999))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// HashAPIKey hashes an API key for lookup
func HashAPIKey(key string) string {
	h := sha256.Sum256([]byte(key))
	return hex.EncodeToString(h[:])
}

// Register creates a new agent account
func (s *AuthService) Register(email, name string) (apiKey string, agent *Agent, err error) {
	// Check if already exists
	var existingID string
	err = s.db.QueryRow("SELECT id FROM agents WHERE email = $1", email).Scan(&existingID)
	if err == nil {
		return "", nil, fmt.Errorf("agent with email %s already exists", email)
	}
	if err != sql.ErrNoRows {
		return "", nil, fmt.Errorf("failed to check existing agent: %w", err)
	}

	key, hash, prefix, err := GenerateAPIKey()
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	code, err := GenerateVerificationCode()
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate verification code: %w", err)
	}

	agent = &Agent{}
	err = s.db.QueryRow(`
		INSERT INTO agents (email, name, api_key_hash, api_key_prefix, verification_code, verification_expires, verified)
		VALUES ($1, $2, $3, $4, $5, $6, FALSE)
		RETURNING id, email, name, plan, quota_bytes, used_bytes, created_at, verified
	`, email, name, hash, prefix, code, time.Now().Add(24*time.Hour)).Scan(
		&agent.ID, &agent.Email, &agent.Name, &agent.Plan,
		&agent.QuotaBytes, &agent.UsedBytes, &agent.CreatedAt, &agent.Verified,
	)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create agent: %w", err)
	}

	return key, agent, nil
}

// Login sends a verification code to the agent's email
func (s *AuthService) Login(email string) (code string, err error) {
	code, err = GenerateVerificationCode()
	if err != nil {
		return "", fmt.Errorf("failed to generate code: %w", err)
	}

	result, err := s.db.Exec(`
		UPDATE agents SET verification_code = $1, verification_expires = $2
		WHERE email = $3
	`, code, time.Now().Add(15*time.Minute), email)
	if err != nil {
		return "", fmt.Errorf("failed to update verification: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return "", fmt.Errorf("no agent found with email %s", email)
	}

	return code, nil
}

// Verify checks the verification code and returns a new API key
func (s *AuthService) Verify(email, code string) (apiKey string, agent *Agent, err error) {
	var storedCode sql.NullString
	var expiresAt sql.NullTime
	var agentID string

	err = s.db.QueryRow(`
		SELECT id, verification_code, verification_expires FROM agents WHERE email = $1
	`, email).Scan(&agentID, &storedCode, &expiresAt)
	if err == sql.ErrNoRows {
		return "", nil, fmt.Errorf("no agent found with email %s", email)
	}
	if err != nil {
		return "", nil, fmt.Errorf("failed to lookup agent: %w", err)
	}

	if !storedCode.Valid || storedCode.String != code {
		return "", nil, fmt.Errorf("invalid verification code")
	}

	if expiresAt.Valid && time.Now().After(expiresAt.Time) {
		return "", nil, fmt.Errorf("verification code expired")
	}

	// Generate new API key
	key, hash, prefix, err := GenerateAPIKey()
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate API key: %w", err)
	}

	agent = &Agent{}
	err = s.db.QueryRow(`
		UPDATE agents SET api_key_hash = $1, api_key_prefix = $2, verified = TRUE,
			verification_code = NULL, verification_expires = NULL
		WHERE id = $3
		RETURNING id, email, name, plan, quota_bytes, used_bytes, created_at, verified
	`, hash, prefix, agentID).Scan(
		&agent.ID, &agent.Email, &agent.Name, &agent.Plan,
		&agent.QuotaBytes, &agent.UsedBytes, &agent.CreatedAt, &agent.Verified,
	)
	if err != nil {
		return "", nil, fmt.Errorf("failed to update agent: %w", err)
	}

	return key, agent, nil
}

// Authenticate looks up an agent by API key
func (s *AuthService) Authenticate(apiKey string) (*Agent, error) {
	hash := HashAPIKey(apiKey)

	agent := &Agent{}
	err := s.db.QueryRow(`
		SELECT id, email, name, plan, quota_bytes, used_bytes, created_at, verified
		FROM agents WHERE api_key_hash = $1
	`, hash).Scan(
		&agent.ID, &agent.Email, &agent.Name, &agent.Plan,
		&agent.QuotaBytes, &agent.UsedBytes, &agent.CreatedAt, &agent.Verified,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invalid API key")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %w", err)
	}

	return agent, nil
}

// GetAgentByEmail looks up an agent by email
func (s *AuthService) GetAgentByEmail(email string) (*Agent, error) {
	agent := &Agent{}
	err := s.db.QueryRow(`
		SELECT id, email, name, plan, quota_bytes, used_bytes, created_at, verified
		FROM agents WHERE email = $1
	`, email).Scan(
		&agent.ID, &agent.Email, &agent.Name, &agent.Plan,
		&agent.QuotaBytes, &agent.UsedBytes, &agent.CreatedAt, &agent.Verified,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no agent found with email %s", email)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to lookup agent: %w", err)
	}
	return agent, nil
}

// GetAgentByID looks up an agent by ID
func (s *AuthService) GetAgentByID(id string) (*Agent, error) {
	agent := &Agent{}
	err := s.db.QueryRow(`
		SELECT id, email, name, plan, quota_bytes, used_bytes, created_at, verified
		FROM agents WHERE id = $1
	`, id).Scan(
		&agent.ID, &agent.Email, &agent.Name, &agent.Plan,
		&agent.QuotaBytes, &agent.UsedBytes, &agent.CreatedAt, &agent.Verified,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no agent found with id %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to lookup agent: %w", err)
	}
	return agent, nil
}

// UpdateUsedBytes updates the storage usage for an agent
func (s *AuthService) UpdateUsedBytes(agentID string, usedBytes int64) error {
	_, err := s.db.Exec("UPDATE agents SET used_bytes = $1 WHERE id = $2", usedBytes, agentID)
	return err
}
