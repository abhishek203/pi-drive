package api

import (
	"net/http"

	"github.com/pidrive/pidrive/internal/billing"
)

type BillingHandler struct {
	billingService *billing.BillingService
}

func NewBillingHandler(billingService *billing.BillingService) *BillingHandler {
	return &BillingHandler{billingService: billingService}
}

func (h *BillingHandler) GetPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.billingService.GetPlans()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"plans": plans})
}

func (h *BillingHandler) GetUsage(w http.ResponseWriter, r *http.Request) {
	agent := GetAgent(r)
	if agent == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	usage, err := h.billingService.GetUsage(agent.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, usage)
}

func (h *BillingHandler) Upgrade(w http.ResponseWriter, r *http.Request) {
	agent := GetAgent(r)
	if agent == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	var req struct {
		Plan string `json:"plan"`
	}
	if err := decodeJSON(r, &req); err != nil || req.Plan == "" {
		writeError(w, http.StatusBadRequest, "plan is required")
		return
	}

	// TODO: Integrate Stripe checkout for paid plans
	// For now, allow direct upgrade (useful for free/testing)
	if err := h.billingService.Upgrade(agent.ID, req.Plan); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{
		"message": "upgraded to " + req.Plan,
	})
}

func (h *BillingHandler) GetBilling(w http.ResponseWriter, r *http.Request) {
	agent := GetAgent(r)
	if agent == nil {
		writeError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	info, err := h.billingService.GetBillingInfo(agent.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, info)
}
