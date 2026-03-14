package search

import (
	"fmt"
	"strings"
	"time"

	"github.com/pidrive/pidrive/internal/db"
)

type Result struct {
	Path       string    `json:"path"`
	Filename   string    `json:"filename"`
	Snippet    string    `json:"snippet"`
	SizeBytes  int64     `json:"size_bytes"`
	ModifiedAt time.Time `json:"modified_at"`
}

type SearchOptions struct {
	Query      string
	AgentID    string
	Types      []string // file extensions to filter
	ModifiedIn *time.Duration
	MyOnly     bool
	SharedOnly bool
	Limit      int
}

type SearchService struct {
	db *db.DB
}

func NewSearchService(db *db.DB) *SearchService {
	return &SearchService{db: db}
}

func (s *SearchService) Search(opts SearchOptions) ([]Result, error) {
	// Build the agent filter: own files + files shared with this agent
	agentFilter := ""
	args := []interface{}{}
	argIdx := 1

	if opts.MyOnly {
		agentFilter = fmt.Sprintf("si.agent_id = $%d", argIdx)
		args = append(args, opts.AgentID)
		argIdx++
	} else if opts.SharedOnly {
		agentFilter = fmt.Sprintf(`si.agent_id IN (
			SELECT s.owner_id FROM shares s
			WHERE s.target_id = $%d AND NOT s.revoked
				AND (s.expires_at IS NULL OR s.expires_at > now())
		)`, argIdx)
		args = append(args, opts.AgentID)
		argIdx++
	} else {
		// Own files + shared with me
		agentFilter = fmt.Sprintf(`(si.agent_id = $%d OR si.agent_id IN (
			SELECT s.owner_id FROM shares s
			WHERE s.target_id = $%d AND NOT s.revoked
				AND (s.expires_at IS NULL OR s.expires_at > now())
		))`, argIdx, argIdx+1)
		args = append(args, opts.AgentID, opts.AgentID)
		argIdx += 2
	}

	// FTS query
	queryFilter := fmt.Sprintf("si.search_vector @@ plainto_tsquery('english', $%d)", argIdx)
	args = append(args, opts.Query)
	argIdx++

	// Snippet
	snippetExpr := fmt.Sprintf(
		"ts_headline('english', coalesce(si.content, ''), plainto_tsquery('english', $%d), 'MaxWords=30, MinWords=10, MaxFragments=1')",
		argIdx,
	)
	args = append(args, opts.Query)
	argIdx++

	// Optional filters
	var extraFilters []string

	if len(opts.Types) > 0 {
		placeholders := make([]string, len(opts.Types))
		for i, t := range opts.Types {
			if !strings.HasPrefix(t, ".") {
				t = "." + t
			}
			placeholders[i] = fmt.Sprintf("$%d", argIdx)
			args = append(args, t)
			argIdx++
		}
		extraFilters = append(extraFilters,
			fmt.Sprintf("lower(substring(si.filename from '\\.\\w+$')) IN (%s)", strings.Join(placeholders, ",")))
	}

	if opts.ModifiedIn != nil {
		extraFilters = append(extraFilters,
			fmt.Sprintf("si.modified_at >= $%d", argIdx))
		args = append(args, time.Now().Add(-*opts.ModifiedIn))
		argIdx++
	}

	where := fmt.Sprintf("WHERE %s AND %s", agentFilter, queryFilter)
	for _, f := range extraFilters {
		where += " AND " + f
	}

	limit := opts.Limit
	if limit <= 0 {
		limit = 20
	}

	query := fmt.Sprintf(`
		SELECT si.path, si.filename, %s as snippet, si.size_bytes, si.modified_at
		FROM search_index si
		%s
		ORDER BY ts_rank(si.search_vector, plainto_tsquery('english', $%d)) DESC
		LIMIT $%d
	`, snippetExpr, where, argIdx, argIdx+1)
	args = append(args, opts.Query, limit)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	defer rows.Close()

	var results []Result
	for rows.Next() {
		var r Result
		if err := rows.Scan(&r.Path, &r.Filename, &r.Snippet, &r.SizeBytes, &r.ModifiedAt); err != nil {
			return nil, fmt.Errorf("failed to scan result: %w", err)
		}
		results = append(results, r)
	}
	return results, nil
}
