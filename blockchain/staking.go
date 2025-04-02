package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"
)

// StakeStatus represents the status of a stake
type StakeStatus string

const (
	StakeActive    StakeStatus = "ACTIVE"
	StakeUnstaking StakeStatus = "UNSTAKING"
	StakeWithdrawn StakeStatus = "WITHDRAWN"
	StakePenalized StakeStatus = "PENALIZED"
)

// Stake represents a user's staked tokens
type Stake struct {
	ID           string      `json:"id"`
	Owner        string      `json:"owner"`
	Amount       float64     `json:"amount"`
	StartTime    time.Time   `json:"startTime"`
	EndTime      time.Time   `json:"endTime"`
	Duration     int64       `json:"duration"` // in seconds
	Status       StakeStatus `json:"status"`
	RewardRate   float64     `json:"rewardRate"`
	TotalReward  float64     `json:"totalReward"`
	LastClaim    time.Time   `json:"lastClaim"`
	UnstakeTime  time.Time   `json:"unstakeTime"`
	WithdrawTime time.Time   `json:"withdrawTime"`
	StakingPower float64     `json:"stakingPower"`
	Votes        []string    `json:"votes"` // IDs of proposals voted on
}

// Validator represents a validator node in the network
type Validator struct {
	Address            string    `json:"address"`
	PublicKey          string    `json:"publicKey"`
	StakedAmount       float64   `json:"stakedAmount"`
	TotalStaked        float64   `json:"totalStaked"`
	Commission         float64   `json:"commission"` // e.g., 0.05 for 5%
	Uptime             float64   `json:"uptime"`     // e.g., 0.998 for 99.8%
	StartTime          time.Time `json:"startTime"`
	ValidatorSince     time.Time `json:"validatorSince"`
	LastBlockValidated time.Time `json:"lastBlockValidated"`
	BlocksValidated    int       `json:"blocksValidated"`
	Delegators         []string  `json:"delegators"` // Addresses of delegators
	Active             bool      `json:"active"`
	Jailed             bool      `json:"jailed"`
	JailReason         string    `json:"jailReason"`
	JailTime           time.Time `json:"jailTime"`
	UnjailTime         time.Time `json:"unjailTime"`
}

// StakingPool manages all staking operations
type StakingPool struct {
	Stakes           map[string]*Stake     `json:"stakes"`
	Validators       map[string]*Validator `json:"validators"`
	TotalStaked      float64               `json:"totalStaked"`
	StakingAPY       float64               `json:"stakingAPY"`
	ValidatorAPY     float64               `json:"validatorAPY"`
	MinStakeAmount   float64               `json:"minStakeAmount"`
	MinStakeDuration int64                 `json:"minStakeDuration"` // in seconds
	MaxStakeDuration int64                 `json:"maxStakeDuration"` // in seconds
	UnstakingPeriod  int64                 `json:"unstakingPeriod"`  // in seconds
	LastRewardTime   time.Time             `json:"lastRewardTime"`
	mutex            sync.RWMutex
}

// NewStakingPool creates a new staking pool
func NewStakingPool() *StakingPool {
	return &StakingPool{
		Stakes:           make(map[string]*Stake),
		Validators:       make(map[string]*Validator),
		StakingAPY:       0.07,               // 7% APY for normal staking
		ValidatorAPY:     0.12,               // 12% APY for validators
		MinStakeAmount:   10.0,               // Minimum 10 tokens to stake
		MinStakeDuration: 7 * 24 * 60 * 60,   // 7 days
		MaxStakeDuration: 365 * 24 * 60 * 60, // 1 year
		UnstakingPeriod:  3 * 24 * 60 * 60,   // 3 days cooldown
		LastRewardTime:   time.Now(),
	}
}

// CreateStake creates a new stake for a user
func (sp *StakingPool) CreateStake(owner string, amount float64, duration int64) (*Stake, error) {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	// Validate parameters
	if amount < sp.MinStakeAmount {
		return nil, fmt.Errorf("stake amount must be at least %.2f", sp.MinStakeAmount)
	}

	if duration < sp.MinStakeDuration {
		return nil, fmt.Errorf("stake duration must be at least %d seconds", sp.MinStakeDuration)
	}

	if duration > sp.MaxStakeDuration {
		return nil, fmt.Errorf("stake duration cannot exceed %d seconds", sp.MaxStakeDuration)
	}

	// Calculate staking power and reward rate
	// Longer stakes get better rates
	durationFactor := float64(duration) / float64(sp.MaxStakeDuration)
	rewardRate := sp.StakingAPY * (1 + durationFactor*0.5) // Up to 50% bonus for max duration

	// Create a new stake
	now := time.Now()
	stake := &Stake{
		Owner:        owner,
		Amount:       amount,
		StartTime:    now,
		EndTime:      now.Add(time.Duration(duration) * time.Second),
		Duration:     duration,
		Status:       StakeActive,
		RewardRate:   rewardRate,
		LastClaim:    now,
		StakingPower: amount * (1 + durationFactor), // More power for longer stakes
		Votes:        []string{},
	}

	// Generate a unique ID
	stakingData, _ := json.Marshal(map[string]interface{}{
		"owner":     owner,
		"amount":    amount,
		"startTime": now,
		"duration":  duration,
	})
	hash := sha256.Sum256(stakingData)
	stake.ID = hex.EncodeToString(hash[:])

	// Save the stake
	sp.Stakes[stake.ID] = stake

	// Update total staked amount
	sp.TotalStaked += amount

	return stake, nil
}

// GetStakesByOwner returns all stakes owned by a specific address
func (sp *StakingPool) GetStakesByOwner(owner string) []*Stake {
	sp.mutex.RLock()
	defer sp.mutex.RUnlock()

	var ownerStakes []*Stake
	for _, stake := range sp.Stakes {
		if stake.Owner == owner {
			ownerStakes = append(ownerStakes, stake)
		}
	}

	return ownerStakes
}

// GetStake returns a stake by its ID
func (sp *StakingPool) GetStake(stakeID string) (*Stake, error) {
	sp.mutex.RLock()
	defer sp.mutex.RUnlock()

	stake, exists := sp.Stakes[stakeID]
	if !exists {
		return nil, fmt.Errorf("stake not found: %s", stakeID)
	}

	return stake, nil
}

// CalculateRewards calculates the rewards earned by a stake since the last claim
func (sp *StakingPool) CalculateRewards(stakeID string) (float64, error) {
	sp.mutex.RLock()
	defer sp.mutex.RUnlock()

	stake, exists := sp.Stakes[stakeID]
	if !exists {
		return 0, fmt.Errorf("stake not found: %s", stakeID)
	}

	if stake.Status != StakeActive {
		return 0, fmt.Errorf("stake is not active: %s", stakeID)
	}

	// Calculate the time elapsed since the last claim
	now := time.Now()
	elapsed := now.Sub(stake.LastClaim).Seconds()

	// Calculate the reward based on the APY
	annualRate := stake.RewardRate
	periodRate := annualRate * (elapsed / (365 * 24 * 60 * 60)) // Convert to period rate
	reward := stake.Amount * periodRate

	return reward, nil
}

// ClaimRewards allows a user to claim their staking rewards
func (sp *StakingPool) ClaimRewards(stakeID string) (float64, error) {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	stake, exists := sp.Stakes[stakeID]
	if !exists {
		return 0, fmt.Errorf("stake not found: %s", stakeID)
	}

	if stake.Status != StakeActive {
		return 0, fmt.Errorf("stake is not active: %s", stakeID)
	}

	// Calculate rewards
	reward, err := sp.CalculateRewards(stakeID)
	if err != nil {
		return 0, err
	}

	// Update the last claim time
	now := time.Now()
	stake.LastClaim = now

	// Update total reward
	stake.TotalReward += reward

	return reward, nil
}

// InitiateUnstake begins the unstaking process
func (sp *StakingPool) InitiateUnstake(stakeID string) error {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	stake, exists := sp.Stakes[stakeID]
	if !exists {
		return fmt.Errorf("stake not found: %s", stakeID)
	}

	if stake.Status != StakeActive {
		return fmt.Errorf("stake is not active: %s", stakeID)
	}

	// Check if the minimum staking period has passed
	now := time.Now()
	if now.Before(stake.StartTime.Add(time.Duration(sp.MinStakeDuration) * time.Second)) {
		return fmt.Errorf("minimum staking period (%d seconds) has not passed", sp.MinStakeDuration)
	}

	// Calculate final rewards before unstaking
	reward, _ := sp.CalculateRewards(stakeID)
	stake.TotalReward += reward

	// Update stake status and set unstaking time
	stake.Status = StakeUnstaking
	stake.UnstakeTime = now
	stake.LastClaim = now

	return nil
}

// WithdrawStake completes the unstaking process and returns the funds
func (sp *StakingPool) WithdrawStake(stakeID string) (float64, error) {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	stake, exists := sp.Stakes[stakeID]
	if !exists {
		return 0, fmt.Errorf("stake not found: %s", stakeID)
	}

	if stake.Status != StakeUnstaking {
		return 0, fmt.Errorf("stake is not in unstaking status: %s", stakeID)
	}

	// Check if the unstaking period has passed
	now := time.Now()
	cooldownEndTime := stake.UnstakeTime.Add(time.Duration(sp.UnstakingPeriod) * time.Second)
	if now.Before(cooldownEndTime) {
		timeRemaining := cooldownEndTime.Sub(now)
		return 0, fmt.Errorf("unstaking period not completed, %s remaining", timeRemaining)
	}

	// Calculate any remaining rewards
	totalAmount := stake.Amount + stake.TotalReward

	// Update the stake status
	stake.Status = StakeWithdrawn
	stake.WithdrawTime = now

	// Update total staked amount
	sp.TotalStaked -= stake.Amount

	return totalAmount, nil
}

// RegisterValidator registers a new validator
func (sp *StakingPool) RegisterValidator(address, publicKey string, stakedAmount float64) (*Validator, error) {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	// Check if validator already exists
	if _, exists := sp.Validators[address]; exists {
		return nil, fmt.Errorf("validator already registered: %s", address)
	}

	// Minimum stake for validators is higher
	minValidatorStake := sp.MinStakeAmount * 10 // 10x minimum stake
	if stakedAmount < minValidatorStake {
		return nil, fmt.Errorf("validator stake must be at least %.2f", minValidatorStake)
	}

	// Create a new validator
	now := time.Now()
	validator := &Validator{
		Address:         address,
		PublicKey:       publicKey,
		StakedAmount:    stakedAmount,
		TotalStaked:     stakedAmount,
		Commission:      0.10, // 10% default commission
		Uptime:          1.0,  // Start with perfect uptime
		StartTime:       now,
		ValidatorSince:  now,
		BlocksValidated: 0,
		Delegators:      []string{},
		Active:          true,
		Jailed:          false,
	}

	// Create a stake for the validator
	validatorStake := &Stake{
		Owner:        address,
		Amount:       stakedAmount,
		StartTime:    now,
		EndTime:      now.Add(365 * 24 * time.Hour), // 1 year by default
		Duration:     365 * 24 * 60 * 60,
		Status:       StakeActive,
		RewardRate:   sp.ValidatorAPY,
		LastClaim:    now,
		StakingPower: stakedAmount * 2, // Validators get 2x staking power
		Votes:        []string{},
	}

	// Generate stake ID
	stakingData, _ := json.Marshal(map[string]interface{}{
		"owner":     address,
		"amount":    stakedAmount,
		"startTime": now,
		"validator": true,
	})
	hash := sha256.Sum256(stakingData)
	validatorStake.ID = hex.EncodeToString(hash[:])

	// Save the validator and stake
	sp.Validators[address] = validator
	sp.Stakes[validatorStake.ID] = validatorStake

	// Update total staked amount
	sp.TotalStaked += stakedAmount

	return validator, nil
}

// DelegateToValidator allows a user to delegate tokens to a validator
func (sp *StakingPool) DelegateToValidator(delegator, validatorAddress string, amount float64) error {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	// Check if validator exists
	validator, exists := sp.Validators[validatorAddress]
	if !exists {
		return fmt.Errorf("validator not found: %s", validatorAddress)
	}

	// Check if validator is active
	if !validator.Active || validator.Jailed {
		return fmt.Errorf("validator is not active: %s", validatorAddress)
	}

	// Minimum delegation amount
	if amount < sp.MinStakeAmount {
		return fmt.Errorf("delegation amount must be at least %.2f", sp.MinStakeAmount)
	}

	// Create a new stake for delegation
	now := time.Now()
	delegationStake := &Stake{
		Owner:        delegator,
		Amount:       amount,
		StartTime:    now,
		EndTime:      now.Add(30 * 24 * time.Hour), // 30 days by default
		Duration:     30 * 24 * 60 * 60,
		Status:       StakeActive,
		RewardRate:   sp.StakingAPY * 1.5, // 50% bonus for delegation
		LastClaim:    now,
		StakingPower: amount * 1.5, // 1.5x staking power for delegated stakes
		Votes:        []string{},
	}

	// Generate stake ID
	stakingData, _ := json.Marshal(map[string]interface{}{
		"owner":     delegator,
		"validator": validatorAddress,
		"amount":    amount,
		"startTime": now,
	})
	hash := sha256.Sum256(stakingData)
	delegationStake.ID = hex.EncodeToString(hash[:])

	// Save the stake
	sp.Stakes[delegationStake.ID] = delegationStake

	// Update validator
	validator.TotalStaked += amount
	validator.Delegators = append(validator.Delegators, delegator)

	// Update total staked amount
	sp.TotalStaked += amount

	return nil
}

// UndelegateFromValidator removes delegation from a validator
func (sp *StakingPool) UndelegateFromValidator(delegator, validatorAddress string, stakeID string) error {
	// Initialize unstaking process like with normal stakes
	err := sp.InitiateUnstake(stakeID)
	if err != nil {
		return err
	}

	// The rest is handled by the normal unstaking process
	return nil
}

// UpdateValidator updates a validator's information
func (sp *StakingPool) UpdateValidator(address string, updates map[string]interface{}) error {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	validator, exists := sp.Validators[address]
	if !exists {
		return fmt.Errorf("validator not found: %s", address)
	}

	// Apply updates
	if commission, ok := updates["commission"].(float64); ok {
		if commission < 0 || commission > 0.5 {
			return errors.New("commission must be between 0% and 50%")
		}
		validator.Commission = commission
	}

	if publicKey, ok := updates["publicKey"].(string); ok {
		validator.PublicKey = publicKey
	}

	return nil
}

// ValidateBlock records that a validator has validated a block
func (sp *StakingPool) ValidateBlock(validatorAddress string) error {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	validator, exists := sp.Validators[validatorAddress]
	if !exists {
		return fmt.Errorf("validator not found: %s", validatorAddress)
	}

	if !validator.Active || validator.Jailed {
		return fmt.Errorf("validator is not active: %s", validatorAddress)
	}

	// Update validator stats
	validator.BlocksValidated++
	validator.LastBlockValidated = time.Now()

	return nil
}

// DistributeRewards distributes rewards to all active stakes
func (sp *StakingPool) DistributeRewards() {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	now := time.Now()

	// Calculate the time elapsed since the last reward distribution
	elapsed := now.Sub(sp.LastRewardTime).Seconds()

	// Calculate the total rewards to distribute
	// In a real system, this would come from block rewards, transaction fees, etc.
	// For this demo, we'll use a simple formula
	dailyRewardRate := 0.0002                                  // 0.02% daily reward rate
	periodRate := dailyRewardRate * (elapsed / (24 * 60 * 60)) // Convert to period rate

	totalRewards := sp.TotalStaked * periodRate

	// Calculate total staking power to distribute rewards proportionally
	var totalStakingPower float64
	for _, stake := range sp.Stakes {
		if stake.Status == StakeActive {
			totalStakingPower += stake.StakingPower
		}
	}

	// Distribute rewards based on staking power
	for _, stake := range sp.Stakes {
		if stake.Status == StakeActive {
			// Calculate stake's share of rewards
			stakeShare := stake.StakingPower / totalStakingPower
			stakeReward := totalRewards * stakeShare

			// Apply validator commission if applicable
			for _, validator := range sp.Validators {
				if validator.Address == stake.Owner {
					// Validator's own stake doesn't pay commission
					break
				}

				// Check if this stake is delegated to this validator
				for _, delegator := range validator.Delegators {
					if delegator == stake.Owner {
						// Apply commission
						commissionAmount := stakeReward * validator.Commission
						stakeReward -= commissionAmount
						break
					}
				}
			}

			// Update stake rewards
			stake.TotalReward += stakeReward
		}
	}

	// Update last reward time
	sp.LastRewardTime = now
}

// GetStakingStats returns statistics about the staking pool
func (sp *StakingPool) GetStakingStats() map[string]interface{} {
	sp.mutex.RLock()
	defer sp.mutex.RUnlock()

	stats := map[string]interface{}{
		"totalStaked":      sp.TotalStaked,
		"stakingAPY":       sp.StakingAPY,
		"validatorAPY":     sp.ValidatorAPY,
		"activeStakes":     0,
		"activeValidators": 0,
		"totalDelegated":   0.0,
		"averageStake":     0.0,
	}

	// Count active stakes and validators
	var activeStakes int
	var totalDelegated float64

	for _, stake := range sp.Stakes {
		if stake.Status == StakeActive {
			activeStakes++
		}
	}

	for _, validator := range sp.Validators {
		if validator.Active && !validator.Jailed {
			stats["activeValidators"] = stats["activeValidators"].(int) + 1
			totalDelegated += validator.TotalStaked - validator.StakedAmount
		}
	}

	stats["activeStakes"] = activeStakes
	stats["totalDelegated"] = totalDelegated

	if activeStakes > 0 {
		stats["averageStake"] = sp.TotalStaked / float64(activeStakes)
	}

	return stats
}

// JailValidator puts a validator in jail
func (sp *StakingPool) JailValidator(address, reason string, jailDuration int64) error {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	validator, exists := sp.Validators[address]
	if !exists {
		return fmt.Errorf("validator not found: %s", address)
	}

	// Set jail status
	validator.Jailed = true
	validator.JailReason = reason
	validator.JailTime = time.Now()
	validator.UnjailTime = time.Now().Add(time.Duration(jailDuration) * time.Second)
	validator.Active = false

	return nil
}

// UnjailValidator releases a validator from jail
func (sp *StakingPool) UnjailValidator(address string) error {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	validator, exists := sp.Validators[address]
	if !exists {
		return fmt.Errorf("validator not found: %s", address)
	}

	if !validator.Jailed {
		return fmt.Errorf("validator is not jailed: %s", address)
	}

	// Check if jail time is over
	now := time.Now()
	if now.Before(validator.UnjailTime) {
		timeRemaining := validator.UnjailTime.Sub(now)
		return fmt.Errorf("validator still jailed for %s", timeRemaining)
	}

	// Release from jail
	validator.Jailed = false
	validator.Active = true

	return nil
}

// CalculateAPY recalculates APY based on network conditions
func (sp *StakingPool) CalculateAPY() {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	// Base APY
	baseAPY := 0.05 // 5%

	// Adjust based on total staked amount
	// Lower APY when more tokens are staked (scarcity)
	stakingFactor := math.Max(0.5, math.Min(1.5, 1.0-(sp.TotalStaked/1000000)*0.5))

	// New staking APY
	sp.StakingAPY = baseAPY * stakingFactor

	// Validator APY is higher
	sp.ValidatorAPY = sp.StakingAPY * 1.5
}

// CompoundRewards compounds rewards into the stake
func (sp *StakingPool) CompoundRewards(stakeID string) error {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	stake, exists := sp.Stakes[stakeID]
	if !exists {
		return fmt.Errorf("stake not found: %s", stakeID)
	}

	if stake.Status != StakeActive {
		return fmt.Errorf("stake is not active: %s", stakeID)
	}

	// Calculate rewards
	reward, err := sp.CalculateRewards(stakeID)
	if err != nil {
		return err
	}

	// Compound rewards into the stake
	stake.Amount += reward
	stake.LastClaim = time.Now()

	// Update staking power
	durationFactor := float64(stake.Duration) / float64(sp.MaxStakeDuration)
	stake.StakingPower = stake.Amount * (1 + durationFactor)

	// Update total staked amount
	sp.TotalStaked += reward

	return nil
}

// ProcessExpiredStakes processes all expired stakes
func (sp *StakingPool) ProcessExpiredStakes() {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	now := time.Now()

	for _, stake := range sp.Stakes {
		// Only process active stakes
		if stake.Status != StakeActive {
			continue
		}

		// Check if stake has expired
		if now.After(stake.EndTime) {
			// Auto-initiate unstaking
			stake.Status = StakeUnstaking
			stake.UnstakeTime = now

			// Calculate final rewards
			elapsed := now.Sub(stake.LastClaim).Seconds()
			periodRate := stake.RewardRate * (elapsed / (365 * 24 * 60 * 60))
			reward := stake.Amount * periodRate
			stake.TotalReward += reward
			stake.LastClaim = now
		}
	}
}

// ExtendStake extends the duration of an active stake
func (sp *StakingPool) ExtendStake(stakeID string, additionalDuration int64) error {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	stake, exists := sp.Stakes[stakeID]
	if !exists {
		return fmt.Errorf("stake not found: %s", stakeID)
	}

	if stake.Status != StakeActive {
		return fmt.Errorf("stake is not active: %s", stakeID)
	}

	// Calculate new duration
	newDuration := stake.Duration + additionalDuration

	// Validate new duration
	if newDuration > sp.MaxStakeDuration {
		return fmt.Errorf("total stake duration cannot exceed %d seconds", sp.MaxStakeDuration)
	}

	// Update stake
	stake.Duration = newDuration
	stake.EndTime = stake.StartTime.Add(time.Duration(newDuration) * time.Second)

	// Recalculate reward rate and staking power
	durationFactor := float64(newDuration) / float64(sp.MaxStakeDuration)
	stake.RewardRate = sp.StakingAPY * (1 + durationFactor*0.5)
	stake.StakingPower = stake.Amount * (1 + durationFactor)

	return nil
}
