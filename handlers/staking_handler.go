package handlers

import (
	"BkC/blockchain"
	"BkC/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// StakeRequest represents a request to stake tokens
type StakeRequest struct {
	Amount   float64 `json:"amount"`
	Duration int64   `json:"duration"`
}

// StakeResponse represents the response to a stake request
type StakeResponse struct {
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	StakeID    string `json:"stakeId,omitempty"`
	Amount     float64 `json:"amount,omitempty"`
	RewardRate float64 `json:"rewardRate,omitempty"`
	EndDate    string `json:"endDate,omitempty"`
}

// StakingHandler handles the staking UI page
func StakingHandler(bc *blockchain.Blockchain) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		utils.LogRequest(r)
		
		// Check authentication
		cookie, err := r.Cookie("session")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		
		mu.Lock()
		session, exists := sessions[cookie.Value]
		mu.Unlock()
		if !exists || session == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		
		// Log access
		utils.LogAuditEvent(
			utils.EventTypeUIAccess,
			session.Username,
			r.RemoteAddr,
			"Staking interface access",
			utils.RiskLow,
			nil,
		)
		
		// Get user's stakes
		stakes := bc.StakingPool.GetStakesByOwner(session.Username)
		
		// Get staking statistics
		stakingStats := bc.StakingPool.GetStakingStats()
		
		// Calculate total staked and rewards for this user
		var userTotalStaked float64
		var userTotalRewards float64
		
		for _, stake := range stakes {
			userTotalStaked += stake.Amount
			userTotalRewards += stake.TotalReward
			
			// Add calculated but unclaimed rewards
			if stake.Status == blockchain.StakeActive {
				reward, _ := bc.StakingPool.CalculateRewards(stake.ID)
				userTotalRewards += reward
			}
		}
		
		// Get user balance
		userBalance := bc.GetBalance(session.Username)
		
		// Prepare template data
		templateData := map[string]interface{}{
			"Stakes":          stakes,
			"StakingStats":    stakingStats,
			"Username":        session.Username,