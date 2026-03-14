package activity

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pidrive/pidrive/internal/db"
)

type Event struct {
	ID        int64           `json:"id"`
	AgentID   string          `json:"agent_id"`
	Action    string          `json:"action"`
	Path      string          `json:"path,omitempty"`
	Details   json.RawMessage `json:"details,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

type ActivityService struct {
	db *db.DB
}

func NewActivityService(db *db.DB) *ActivityService {
	return &ActivityService{db: db}
}

func (s *ActivityService) Log(agentID, action, path string, details map[string]interface{}) error {
	var detailsJSON []byte
	var err error
	if details != nil {
		detailsJSON, err = json.Marshal(details)
		if err != nil {
			return fmt.Errorf("failed to marshal details: %w", err)
		}
	}

	_, err = s.db.Exec(`
		INSERT INTO activity (agent_id, action, path, details) VALUES ($1, $2, $3, $4)
	`, agentID, action, path, detailsJSON)
	return err
}

type ListOptions struct {
	Since  *time.Time
	Action string
	Limit  int
}

func (s *ActivityService) List(agentID string, opts ListOptions) ([]Event, error) {
	query := `SELECT id, agent_id, action, COALESCE(path, ''), details, created_at
		FROM activity WHERE agent_id = $1`
	args := []interface{}{agentID}
	argIdx := 2

	if opts.Since != nil {
		query += fmt.Sprintf(" AND created_at >= $%d", argIdx)
		args = append(args, *opts.Since)
		argIdx++
	}
	if opts.Action != "" {
		query += fmt.Sprintf(" AND action = $%d", argIdx)
		args = append(args, opts.Action)
		argIdx++
	}

	query += " ORDER BY created_at DESC"

	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}
	query += fmt.Sprintf(" LIMIT $%d", argIdx)
	args = append(args, limit)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query activity: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var e Event
		var details []byte
		if err := rows.Scan(&e.ID, &e.AgentID, &e.Action, &e.Path, &details, &e.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan activity: %w", err)
		}
		if details != nil {
			e.Details = json.RawMessage(details)
		}
		events = append(events, e)
	}
	return events, nil
}
