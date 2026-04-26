package trading

// ItemPrice represents pricing for an item
type ItemPrice struct {
	BuyPrice  int
	SellPrice int
	Stock     int // -1 = unlimited
}

// Merchant represents an NPC trader
type Merchant struct {
	ID          string
	Name        string
	Description string
	Inventory   map[string]*ItemPrice
	Currency    int // Merchant's gold
}

// NewMerchant creates a merchant with default prices
func NewMerchant(id, name, desc string) *Merchant {
	m := &Merchant{
		ID:          id,
		Name:        name,
		Description: desc,
		Inventory:   make(map[string]*ItemPrice),
		Currency:    1000,
	}

	// Set default prices
	m.SetPrice("dirt", 1, 0)
	m.SetPrice("stone", 3, 1)
	m.SetPrice("coal", 10, 5)
	m.SetPrice("iron_ore", 25, 12)
	m.SetPrice("gold_ore", 50, 25)
	m.SetPrice("diamond", 100, 50)
	m.SetPrice("wooden_pickaxe", 15, 7)
	m.SetPrice("stone_pickaxe", 40, 20)
	m.SetPrice("iron_pickaxe", 100, 50)
	m.SetPrice("food", 5, 2)
	m.SetPrice("potion", 30, 15)

	return m
}

// SetPrice sets buy/sell price for an item
func (m *Merchant) SetPrice(itemID string, buy, sell int) {
	m.Inventory[itemID] = &ItemPrice{
		BuyPrice:  buy,
		SellPrice: sell,
		Stock:     -1, // Unlimited by default
	}
}

// GetPrice returns item price info
func (m *Merchant) GetPrice(itemID string) *ItemPrice {
	return m.Inventory[itemID]
}

// CanBuy checks if merchant can sell to player
func (m *Merchant) CanBuy(itemID string, quantity int, playerGold int) bool {
	price, exists := m.Inventory[itemID]
	if !exists {
		return false
	}
	if price.Stock >= 0 && price.Stock < quantity {
		return false
	}
	return playerGold >= price.BuyPrice*quantity
}

// CanSell checks if merchant can buy from player
func (m *Merchant) CanSell(itemID string, quantity int) bool {
	price, exists := m.Inventory[itemID]
	if !exists {
		return false
	}
	return m.Currency >= price.SellPrice*quantity
}

// Buy executes a buy transaction (player buys from merchant)
func (m *Merchant) Buy(itemID string, quantity int) (cost int, ok bool) {
	if !m.CanBuy(itemID, quantity, 999999) {
		return 0, false
	}

	price := m.Inventory[itemID]
	cost = price.BuyPrice * quantity

	if price.Stock >= 0 {
		price.Stock -= quantity
	}
	m.Currency += cost

	return cost, true
}

// Sell executes a sell transaction (player sells to merchant)
func (m *Merchant) Sell(itemID string, quantity int) (payout int, ok bool) {
	if !m.CanSell(itemID, quantity) {
		return 0, false
	}

	price := m.Inventory[itemID]
	payout = price.SellPrice * quantity
	m.Currency -= payout

	// Restock merchant inventory
	if price.Stock >= 0 {
		price.Stock += quantity
	}

	return payout, true
}
