package billing

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pidrive/pidrive/internal/db"
)

type Plan struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	StorageBytes   int64  `json:"storage_bytes"`
	BandwidthBytes int64  `json:"bandwidth_bytes"`
	PriceCents     int    `json:"price_cents"`
	StripePriceID  string `json:"stripe_price_id,omitempty"`
}

type Usage struct {
	UsedBytes          int64  `json:"used_bytes"`
	QuotaBytes         int64  `json:"quota_bytes"`
	Plan               string `json:"plan"`
	BandwidthToday     int64  `json:"bandwidth_today"`
	BandwidthThisMonth int64  `json:"bandwidth_this_month"`
}

type BillingInfo struct {
	AgentID              string     `json:"agent_id"`
	Plan                 string     `json:"plan"`
	StripeCustomerID     *string    `json:"stripe_customer_id,omitempty"`
	StripeSubscriptionID *string    `json:"stripe_subscription_id,omitempty"`
	CurrentPeriodStart   *time.Time `json:"current_period_start,omitempty"`
	CurrentPeriodEnd     *time.Time `json:"current_period_end,omitempty"`
	Status               string     `json:"status"`
}

type BillingService struct {
	db *db.DB
}

func NewBillingService(db *db.DB) *BillingService {
	return &BillingService{db: db}
}

// GetPlans returns all available plans
func (s *BillingService) GetPlans() ([]Plan, error) {
	rows, err := s.db.Query("SELECT id, name, storage_bytes, bandwidth_bytes, price_cents, COALESCE(stripe_price_id, '') FROM plans ORDER BY price_cents")
	if err != nil {
		return nil, fmt.Errorf("failed to get plans: %w", err)
	}
	defer rows.Close()

	var plans []Plan
	for rows.Next() {
		var p Plan
		if err := rows.Scan(&p.ID, &p.Name, &p.StorageBytes, &p.BandwidthBytes, &p.PriceCents, &p.StripePriceID); err != nil {
			return nil, err
		}
		plans = append(plans, p)
	}
	return plans, nil
}

// GetUsage returns storage and bandwidth usage for an agent
func (s *BillingService) GetUsage(agentID string) (*Usage, error) {
	usage := &Usage{}

	err := s.db.QueryRow(`
		SELECT used_bytes, quota_bytes, plan FROM agents WHERE id = $1
	`, agentID).Scan(&usage.UsedBytes, &usage.QuotaBytes, &usage.Plan)
	if err != nil {
		return nil, fmt.Errorf("failed to get usage: %w", err)
	}

	// Today's bandwidth
	s.db.QueryRow(`
		SELECT COALESCE(bytes_in + bytes_out, 0) FROM bandwidth_usage
		WHERE agent_id = $1 AND date = CURRENT_DATE
	`, agentID).Scan(&usage.BandwidthToday)

	// This month's bandwidth
	s.db.QueryRow(`
		SELECT COALESCE(SUM(bytes_in + bytes_out), 0) FROM bandwidth_usage
		WHERE agent_id = $1 AND date >= date_trunc('month', CURRENT_DATE)
	`, agentID).Scan(&usage.BandwidthThisMonth)

	return usage, nil
}

// TrackBandwidth records bandwidth usage
func (s *BillingService) TrackBandwidth(agentID string, bytesIn, bytesOut int64) error {
	_, err := s.db.Exec(`
		INSERT INTO bandwidth_usage (agent_id, date, bytes_in, bytes_out)
		VALUES ($1, CURRENT_DATE, $2, $3)
		ON CONFLICT (agent_id, date) DO UPDATE SET
			bytes_in = bandwidth_usage.bytes_in + EXCLUDED.bytes_in,
			bytes_out = bandwidth_usage.bytes_out + EXCLUDED.bytes_out
	`, agentID, bytesIn, bytesOut)
	return err
}

// CheckQuota returns true if the agent is within quota
func (s *BillingService) CheckQuota(agentID string, additionalBytes int64) (bool, error) {
	var usedBytes, quotaBytes int64
	err := s.db.QueryRow("SELECT used_bytes, quota_bytes FROM agents WHERE id = $1", agentID).
		Scan(&usedBytes, &quotaBytes)
	if err != nil {
		return false, err
	}
	return usedBytes+additionalBytes <= quotaBytes, nil
}

// CheckBandwidthQuota returns true if the agent is within daily bandwidth quota
func (s *BillingService) CheckBandwidthQuota(agentID string) (bool, error) {
	var plan string
	err := s.db.QueryRow("SELECT plan FROM agents WHERE id = $1", agentID).Scan(&plan)
	if err != nil {
		return false, err
	}

	var bandwidthLimit int64
	err = s.db.QueryRow("SELECT bandwidth_bytes FROM plans WHERE id = $1", plan).Scan(&bandwidthLimit)
	if err != nil {
		return false, err
	}

	if bandwidthLimit < 0 {
		return true, nil // unlimited
	}

	var todayUsage int64
	s.db.QueryRow(`
		SELECT COALESCE(bytes_in + bytes_out, 0) FROM bandwidth_usage
		WHERE agent_id = $1 AND date = CURRENT_DATE
	`, agentID).Scan(&todayUsage)

	return todayUsage < bandwidthLimit, nil
}

// Upgrade changes an agent's plan and quota
func (s *BillingService) Upgrade(agentID, planID string) error {
	var storageBytes int64
	err := s.db.QueryRow("SELECT storage_bytes FROM plans WHERE id = $1", planID).Scan(&storageBytes)
	if err != nil {
		return fmt.Errorf("plan not found: %w", err)
	}

	_, err = s.db.Exec("UPDATE agents SET plan = $1, quota_bytes = $2 WHERE id = $3",
		planID, storageBytes, agentID)
	return err
}

// GetBillingInfo returns billing details for an agent
func (s *BillingService) GetBillingInfo(agentID string) (*BillingInfo, error) {
	info := &BillingInfo{AgentID: agentID}

	err := s.db.QueryRow("SELECT plan FROM agents WHERE id = $1", agentID).Scan(&info.Plan)
	if err != nil {
		return nil, err
	}

	err = s.db.QueryRow(`
		SELECT COALESCE(stripe_customer_id, ''), COALESCE(stripe_subscription_id, ''),
			current_period_start, current_period_end, status
		FROM billing WHERE agent_id = $1
	`, agentID).Scan(
		&info.StripeCustomerID, &info.StripeSubscriptionID,
		&info.CurrentPeriodStart, &info.CurrentPeriodEnd, &info.Status,
	)
	if err == sql.ErrNoRows {
		info.Status = "none"
		return info, nil
	}

	return info, err
}

// EnsureBillingRecord creates a billing record if it doesn't exist
func (s *BillingService) EnsureBillingRecord(agentID string) error {
	_, err := s.db.Exec(`
		INSERT INTO billing (agent_id) VALUES ($1) ON CONFLICT (agent_id) DO NOTHING
	`, agentID)
	return err
}

// UpdateStripeInfo updates Stripe details for an agent
func (s *BillingService) UpdateStripeInfo(agentID, customerID, subscriptionID string, periodStart, periodEnd time.Time) error {
	_, err := s.db.Exec(`
		UPDATE billing SET
			stripe_customer_id = $1,
			stripe_subscription_id = $2,
			current_period_start = $3,
			current_period_end = $4,
			status = 'active'
		WHERE agent_id = $5
	`, customerID, subscriptionID, periodStart, periodEnd, agentID)
	return err
}
