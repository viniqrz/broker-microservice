package entity

type Order struct {
	ID string
	Investor *Investor
	Asset *Asset
	Shares int
	PendingShares int
	Price float64
	OrderType string
	Status string
	Transactions []*Transaction
}

func NewOrder(orderID string, investor *Investor, asset *Asset, shares int, price float64, orderType string) *Order {
	return &Order{
		ID: orderID,
		Investor: investor,
		Asset: asset,
		Shares: shares,
		PendingShares: shares,
		Price: price,
		OrderType: orderType,
		Status: "OPEN",
		Transactions: []*Transaction{},
	}
}

func (o *Order) decreasePendingShares(shares int) {
	o.PendingShares -= shares
}

func (o *Order) close() {
	o.Status = "CLOSED"
}

func (o *Order) appendTransaction(transaction *Transaction) {
	o.Transactions = append(o.Transactions, transaction)
}