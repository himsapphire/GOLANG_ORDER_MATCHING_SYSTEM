package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/himsapphire/GOLANG_ORDER_MATCHING_SYSTEM/db"
	"github.com/himsapphire/GOLANG_ORDER_MATCHING_SYSTEM/models"
	"github.com/himsapphire/GOLANG_ORDER_MATCHING_SYSTEM/services"
)


func CreateOrder(c *gin.Context) {
	var order models.Order

	// Parse JSON input
	if err := c.BindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// Validate input fields
	if (order.Side != "buy" && order.Side != "sell") ||
		(order.Type != "limit" && order.Type != "market") ||
		order.InitialQuantity <= 0 ||
		(order.Type == "limit" && order.Price <= 0) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order parameters"})
		return
	}

	// Set initial fields
	order.Status = "open"
	order.RemainingQuantity = order.InitialQuantity

	// Insert into database
	res, err := db.DB.Exec(`
        INSERT INTO orders (symbol, side, type, price, initial_quantity, remaining_quantity, status)
        VALUES (?, ?, ?, ?, ?, ?, ?)`,
		order.Symbol, order.Side, order.Type, order.Price,
		order.InitialQuantity, order.RemainingQuantity, order.Status,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert order into DB"})
		return
	}

	// Get inserted order ID
	order.ID, _ = res.LastInsertId()

	// Run matching logic
	if err := services.MatchOrder(order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Matching engine error"})
		return
	}

	//  Re-fetch the updated order from DB (to get updated status, remaining_quantity, created_at)
	row := db.DB.QueryRow(`
        SELECT id, symbol, side, type, price, initial_quantity, remaining_quantity, status, created_at
        FROM orders WHERE id = ?`, order.ID,
	)
	err = row.Scan(&order.ID, &order.Symbol, &order.Side, &order.Type, &order.Price,
		&order.InitialQuantity, &order.RemainingQuantity, &order.Status, &order.CreatedAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated order"})
		return
	}

	//  Return final order state
	c.JSON(http.StatusOK, order)
}


func CancelOrder(c *gin.Context) {
	orderID := c.Param("id")

	// Only allow canceling open orders
	res, err := db.DB.Exec(`UPDATE orders SET status = 'canceled' WHERE id = ? AND status = 'open'`, orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel order"})
		return
	}

	affected, _ := res.RowsAffected()
	if affected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found or already filled"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order canceled", "order_id": orderID})
}

func GetOrderBook(c *gin.Context) {
	symbol := c.Query("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Symbol is required"})
		return
	}

	// Top 5 buy orders (highest price first)
	buyRows, err := db.DB.Query(`
        SELECT price, remaining_quantity FROM orders
        WHERE symbol = ? AND side = 'buy' AND status = 'open'
        ORDER BY price DESC, created_at ASC LIMIT 5
    `, symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}
	defer buyRows.Close()

	buyOrders := []gin.H{}
	for buyRows.Next() {
		var price float64
		var qty int
		buyRows.Scan(&price, &qty)
		buyOrders = append(buyOrders, gin.H{"price": price, "quantity": qty})
	}

	// Top 5 sell orders (lowest price first)
	sellRows, err := db.DB.Query(`
        SELECT price, remaining_quantity FROM orders
        WHERE symbol = ? AND side = 'sell' AND status = 'open'
        ORDER BY price ASC, created_at ASC LIMIT 5
    `, symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}
	defer sellRows.Close()

	sellOrders := []gin.H{}
	for sellRows.Next() {
		var price float64
		var qty int
		sellRows.Scan(&price, &qty)
		sellOrders = append(sellOrders, gin.H{"price": price, "quantity": qty})
	}

	c.JSON(http.StatusOK, gin.H{
		"symbol": symbol,
		"bids":   buyOrders,
		"asks":   sellOrders,
	})
}

func ListTrades(c *gin.Context) {
	symbol := c.Query("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Symbol is required"})
		return
	}

	rows, err := db.DB.Query(`
        SELECT id, symbol, buy_order_id, sell_order_id, price, quantity, created_at
        FROM trades WHERE symbol = ? ORDER BY created_at DESC LIMIT 50
    `, symbol)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to query trades"})
		return
	}
	defer rows.Close()

	trades := []models.Trade{}
	for rows.Next() {
		var t models.Trade
		rows.Scan(&t.ID, &t.Symbol, &t.BuyOrderID, &t.SellOrderID,
			&t.Price, &t.Quantity, &t.CreatedAt)
		trades = append(trades, t)
	}

	c.JSON(http.StatusOK, trades)
}

func GetOrderStatus(c *gin.Context) {
	orderID := c.Param("id")

	row := db.DB.QueryRow(`SELECT id, symbol, side, type, price, initial_quantity, remaining_quantity, status, created_at
        FROM orders WHERE id = ?`, orderID)

	var order models.Order
	err := row.Scan(&order.ID, &order.Symbol, &order.Side, &order.Type,
		&order.Price, &order.InitialQuantity, &order.RemainingQuantity,
		&order.Status, &order.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		}
		return
	}

	c.JSON(http.StatusOK, order)
}
