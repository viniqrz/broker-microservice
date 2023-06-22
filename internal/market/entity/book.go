package entity

import (
	"container/heap"
	"sync"
)

type Book struct {
	Order []*Order
	Transactions []*Transaction
	OrdersChan chan *Order
	OrdersChanOut chan *Order
	Wg *sync.WaitGroup
}

func NewBook(orderChan chan *Order, orderChanOut chan *Order, wg *sync.WaitGroup) *Book {
	return &Book{
		Order: []*Order{},
		Transactions: []*Transaction{},
		OrdersChan: orderChan,
		OrdersChanOut: orderChanOut,
		Wg: wg,
	}
}

func (b *Book) Trade(){
	buyOrders := make(map[string]*OrderQueue) 
	sellOrders := make(map[string]*OrderQueue) 

	for order := range b.OrdersChan {

		asset := order.Asset.ID

		if (buyOrders[asset] == nil) {
			buyOrders[asset] = NewOrderQueue()
			heap.Init(buyOrders[asset])
		}

		if (sellOrders[asset] == nil) {
			sellOrders[asset] = NewOrderQueue()
			heap.Init(sellOrders[asset])
		}

		assetBuyOrders := buyOrders[asset]
		assetSellOrders := sellOrders[asset]

		if order.OrderType == "BUY" {
			assetBuyOrders.Push(order)
			if assetSellOrders.Len() > 0 && assetSellOrders.Orders[0].Price <= order.Price {
				sellOrder := assetSellOrders.Pop().(*Order)

				transaction := NewTransaction(sellOrder, order, order.Shares, sellOrder.Price)
				b.AddTransaction(transaction, b.Wg)
				sellOrder.appendTransaction(transaction)
				order.appendTransaction(transaction)

				b.OrdersChanOut <- sellOrder
				b.OrdersChanOut <- order

				if sellOrder.PendingShares > 0 {
					assetSellOrders.Push(sellOrder)
				}
			}
		}
		
		if order.OrderType == "SELL" {
			assetSellOrders.Push(order)
			if assetBuyOrders.Len() > 0 && assetBuyOrders.Orders[0].Price >= order.Price {
				buyOrder := assetBuyOrders.Pop().(*Order)

				transaction := NewTransaction(order, buyOrder, order.Shares, buyOrder.Price)
				b.AddTransaction(transaction, b.Wg)
				buyOrder.appendTransaction(transaction)
				order.appendTransaction(transaction)

				b.OrdersChanOut <- buyOrder
				b.OrdersChanOut <- order

				if buyOrder.PendingShares > 0 {
					assetBuyOrders.Push(buyOrder)
				}
			}
		}

	}
}

func (b *Book) AddTransaction(transaction *Transaction, wg *sync.WaitGroup) {
	defer wg.Done()

	b.Transactions = append(b.Transactions, transaction)

	sellingShares := transaction.SellingOrder.PendingShares
	buyingShares := transaction.BuyingOrder.PendingShares

	var mutualShares int;

	if buyingShares < sellingShares {
		mutualShares = buyingShares
	} else {
		mutualShares = sellingShares
	}

	transaction.SellingOrder.Investor.DecreaseAssetPositionByAmount(transaction.SellingOrder.Asset.ID, mutualShares)
	transaction.SellingOrder.decreasePendingShares(mutualShares)
	
	transaction.BuyingOrder.Investor.IncreaseAssetPositionByAmount(transaction.BuyingOrder.Asset.ID, mutualShares)
	transaction.BuyingOrder.decreasePendingShares(mutualShares)

	transaction.calculateTotal()

	if transaction.BuyingOrder.PendingShares == 0 {
		transaction.BuyingOrder.close()
	}

	if transaction.SellingOrder.PendingShares == 0 {
		transaction.SellingOrder.close()
	}
}

