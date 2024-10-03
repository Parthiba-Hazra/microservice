package models

type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Inventory   int     `json:"inventory"`
	CreatedAt   string  `json:"created_at"`
}

type InventoryUpdateEvent struct {
	ProductID    int `json:"product_id"`
	NewInventory int `json:"new_inventory"`
}

type OrderPlacedEvent struct {
	OrderID int         `json:"order_id"`
	Items   []OrderItem `json:"items"`
}

type OrderItem struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}
