package trading

import (
	"sync"
)

// TradingManager handles player trading
type TradingManager struct {
	mu        sync.RWMutex
	merchants map[string]*Merchant
	playerGold int
	priceMultiplier float64 // Market fluctuation
}

// NewTradingManager creates new trading manager
func NewTradingManager() *TradingManager {
	tm := &TradingManager{
		merchants:       make(map[string]*Merchant),
		playerGold:      100, // Starting gold
		priceMultiplier: 1.0,
	}
	
	// Create default merchant
	tm.AddMerchant(NewMerchant("general", "General Store", "Sells basic goods"))
	
	return tm
}

// AddMerchant adds a merchant
func (tm *TradingManager) AddMerchant(m *Merchant) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.merchants[m.ID] = m
}

// GetMerchant returns a merchant
func (tm *TradingManager) GetMerchant(id string) *Merchant {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.merchants[id]
}

// GetAllMerchants returns all merchants
func (tm *TradingManager) GetAllMerchants() []*Merchant {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	
	var list []*Merchant
	for _, m := range tm.merchants {
		list = append(list, m)
	}
	return list
}

// GetPlayerGold returns player gold
func (tm *TradingManager) GetPlayerGold() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.playerGold
}

// AddGold adds gold to player
func (tm *TradingManager) AddGold(amount int) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.playerGold += amount
}

// RemoveGold removes gold from player
func (tm *TradingManager) RemoveGold(amount int) bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	if tm.playerGold < amount {
		return false
	}
	tm.playerGold -= amount
	return true
}

// BuyItem buys item from merchant
func (tm *TradingManager) BuyItem(merchantID, itemID string, quantity int) bool {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	merchant, exists := tm.merchants[merchantID]
	if !exists {
		return false
	}
	
	price := merchant.GetPrice(itemID)
	if price == nil {
		return false
	}
	
	cost := int(float64(price.BuyPrice) * tm.priceMultiplier * float64(quantity))
	
	if tm.playerGold < cost {
		return false
	}
	
	if _, ok := merchant.Buy(itemID, quantity); !ok {
		return false
	}
	
	tm.playerGold -= cost
	return true
}

// SellItem sells item to merchant
func (tm *TradingManager) SellItem(merchantID, itemID string, quantity int) int {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	merchant, exists := tm.merchants[merchantID]
	if !exists {
		return 0
	}
	
	price := merchant.GetPrice(itemID)
	if price == nil {
		return 0
	}
	
	payout := int(float64(price.SellPrice) * tm.priceMultiplier * float64(quantity))
	
	if _, ok := merchant.Sell(itemID, quantity); !ok {
		return 0
	}
	
	tm.playerGold += payout
	return payout
}

// SetPriceMultiplier sets market price multiplier
func (tm *TradingManager) SetPriceMultiplier(multiplier float64) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	if multiplier < 0.1 {
		multiplier = 0.1
	}
	if multiplier > 3.0 {
		multiplier = 3.0
	}
	tm.priceMultiplier = multiplier
}

// GetPriceModifier returns current price modifier
func (tm *TradingManager) GetPriceModifier() float64 {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.priceMultiplier
}
