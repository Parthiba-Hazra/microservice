package models

type Order struct {
	ID        int         `json:"id"`
	UserID    int         `json:"user_id"`
	Status    string      `json:"status"`
	Total     float64     `json:"total"`
	CreatedAt string      `json:"created_at"`
	Items     []OrderItem `json:"items"`
}

type OrderItem struct {
	ID        int     `json:"id"`
	OrderID   int     `json:"order_id"`
	ProductID int     `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type OrderInput struct {
	Items []OrderItemInput `json:"items"`
}

type OrderItemInput struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type OrderPlacedEvent struct {
	OrderID int             `json:"order_id"`
	UserID  int             `json:"user_id"`
	Items   []OrderItemInfo `json:"items"`
}

type OrderItemInfo struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type OrderShippedEvent struct {
	OrderID int `json:"order_id"`
}

type Product struct {
	ID    int     `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
}
