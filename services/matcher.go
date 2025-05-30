package services

import (
	"github.com/himsapphire/GOLANG_ORDER_MATCHING_SYSTEM/db"
	"github.com/himsapphire/GOLANG_ORDER_MATCHING_SYSTEM/models"
)

func MatchOrder(order models.Order) error {
	tx, err := db.DB.Begin()
	if err != nil {
		return err
	}

	oppositeSide := "buy"
	orderBy := "price DESC"
	if order.Side == "buy" {
		oppositeSide = "sell"
		orderBy = "price ASC"
	}

	query := `
        SELECT id, symbol, side, type, price, initial_quantity, remaining_quantity, status, created_at
        FROM orders
        WHERE symbol = ? AND side = ? AND status = 'open'
        AND type = 'limit'
        ORDER BY ` + orderBy + `, created_at ASC
    `
	rows, err := tx.Query(query, order.Symbol, oppositeSide)
	if err != nil {
		tx.Rollback()
		return err
	}

	var matches []models.Order
	for rows.Next() {
		var match models.Order
		err := rows.Scan(&match.ID, &match.Symbol, &match.Side, &match.Type,
			&match.Price, &match.InitialQuantity, &match.RemainingQuantity,
			&match.Status, &match.CreatedAt)
		if err != nil {
			rows.Close()
			tx.Rollback()
			return err
		}
		matches = append(matches, match)
	}
	rows.Close()

	remaining := order.RemainingQuantity

	for _, match := range matches {
		if remaining <= 0 {
			break
		}

		// Price check for limit orders
		if order.Type == "limit" {
			if (order.Side == "buy" && order.Price < match.Price) ||
				(order.Side == "sell" && order.Price > match.Price) {
				break
			}
		}

		tradeQty := min(remaining, match.RemainingQuantity)
		priceUsed := match.Price

		_, err = tx.Exec(`
			INSERT INTO trades (symbol, buy_order_id, sell_order_id, price, quantity)
			VALUES (?, ?, ?, ?, ?)`,
			order.Symbol, chooseBuy(order, match), chooseSell(order, match), priceUsed, tradeQty)
		if err != nil {
			tx.Rollback()
			return err
		}

		newRem := match.RemainingQuantity - tradeQty
		matchStatus := "open"
		if newRem == 0 {
			matchStatus = "filled"
		}

		_, err = tx.Exec(`UPDATE orders SET remaining_quantity = ?, status = ? WHERE id = ?`,
			newRem, matchStatus, match.ID)
		if err != nil {
			tx.Rollback()
			return err
		}

		remaining -= tradeQty
	}

	// Final status logic
	var newStatus string
	if remaining == 0 {
		newStatus = "filled"
	} else if order.Type == "market" {
		newStatus = "canceled" // market order cannot wait
	} else {
		newStatus = "open" // limit order can wait
	}

	_, err = tx.Exec(`UPDATE orders SET remaining_quantity = ?, status = ? WHERE id = ?`,
		remaining, newStatus, order.ID)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func chooseBuy(a, b models.Order) int64 {
	if a.Side == "buy" {
		return a.ID
	}
	return b.ID
}

func chooseSell(a, b models.Order) int64 {
	if a.Side == "sell" {
		return a.ID
	}
	return b.ID
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
