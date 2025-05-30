# 🧮 Golang Order Matching System

A simplified order matching engine built using GO and MySQL. Inspired by stock exchange matching engines, it supports market and limit orders, partial fills, and price-time priority matching.

---

## ✅ Features

- Submit buy/sell orders via REST API (market or limit type)
- Match orders based on price-time priority
- Support partial fills
- Maintain order book and trade history
- Cancel pending orders
- MySQL-backed persistence using raw SQL (no ORM)

---

## ⚙️ Tech Stack

- Language: Go 
- Database: MySQL 
- Framework: Gin 
- SQL: Raw SQL only 

---

## 📦 Project Structure
- `api/` # API route handlers
- `db/` # DB connection and schema
- `models/` # Order and Trade structs
- `services/` # Matching engine logic
- `main.go` # Entry point
- `go.mod` # Module definition
- `schema.sql` # SQL schema for MySQL
- `README.md` # Documentation



---

## 🛠️ Dependencies and Setup

###  Install Go (v1.18+ recommended)
- [click to download](https://go.dev/dl/go1.22.0.windows-amd64.msi)



###  Clone and Set Up
```bash
git clone https://github.com/00SnowFlake/GOLANG_ORDER_MATCHING_SYSTEM.git
cd GOLANG_ORDER_MATCHING_SYSTEM
go mod tidy
```
###  Install MySQL 
- [click to download](https://dev.mysql.com/downloads/file/?id=541636)

- Create a database: `order_matching`
```sql
CREATE DATABASE IF NOT EXISTS order_matching;
```
- Optionally use MySQL Workbench
-  Configure MySQL in main.go
- Update your DSN:
```go
dsn := "root:password@tcp(127.0.0.1:3306)/order_matching?parseTime=true"
```
🗄️ Database Initialization

Run the schema to create tables:

```bash
mysql -u root -p order_matching < db/schema.sql
```
Or manually in Workbench:
```mysql
CREATE TABLE IF NOT EXISTS orders (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL,
    side ENUM('buy', 'sell') NOT NULL,
    type ENUM('limit', 'market') NOT NULL,
    price DECIMAL(10,2),
    initial_quantity INT NOT NULL,
    remaining_quantity INT NOT NULL,
    status ENUM('open', 'filled', 'canceled') NOT NULL DEFAULT 'open',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS trades (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    symbol VARCHAR(20) NOT NULL,
    buy_order_id BIGINT NOT NULL,
    sell_order_id BIGINT NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    quantity INT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (buy_order_id) REFERENCES orders(id),
    FOREIGN KEY (sell_order_id) REFERENCES orders(id)
);
```
🚀 How to Start the Server
```bash
go run main.go
```
Server will run at `http://localhost:6080`

## 📬 API Usage Examples (via curl)

-  `Place a Buy Limit Order`
```bash
curl -X POST http://localhost:6080/orders \
-H "Content-Type: application/json" \
-d '{
  "symbol": "ETHUSD",
  "side": "buy",
  "type": "limit",
  "price": 2500,
  "initial_quantity": 10
}'
```
 - `Place a Sell Limit Order`
```bash
curl -X POST http://localhost:6080/orders \
-H "Content-Type: application/json" \
-d '{
  "symbol": "ETHUSD",
  "side": "sell",
  "type": "limit",
  "price": 2500,
  "initial_quantity": 5
}'
```
 - `Place a Market Order`
```bash
curl -X POST http://localhost:6080/orders \
-H "Content-Type: application/json" \
-d '{
  "symbol": "ETHUSD",
  "side": "buy",
  "type": "market",
  "initial_quantity": 4
}'
```
 - `Cancel an Order`
```bash
curl -X DELETE http://localhost:6080/orders/3
```
 - `Get Order Book`
```bash
curl http://localhost:6080/orderbook?symbol=ETHUSD
```
-  `Get Order Status`
```bash
curl http://localhost:6080/orders/1
```
 - `View Trades`
```bash
curl http://localhost:6080/trades?symbol=ETHUSD
```


## 📌 Assumptions & Design Decisions

- Orders are matched using price-time priority.  
- Market orders must be filled immediately or are canceled.  
- Limit orders stay open if not fully matched.  
- Trades always use the price of the resting order (like real exchanges).  
- Only raw SQL is used (no ORM or GORM).  
- Database design ensures foreign key integrity between orders and trades.   
- Matching and all DB updates are done inside a single transaction for atomicity.




## 📞 Contact

Feel free to reach out for improvements or questions:

Author: Anjali Gupta

GitHub: https://github.com/00SnowFlake