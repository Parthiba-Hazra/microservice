package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Init() {
	var err error
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("POSTGRES_HOST", "localhost"),
		getEnv("POSTGRES_PORT", "5432"),
		getEnv("POSTGRES_USER", "user"),
		getEnv("POSTGRES_PASSWORD", "password"),
		getEnv("POSTGRES_DB", "mydb"),
	)

	maxRetries := 12
	for retries := 0; retries < maxRetries; retries++ {
		DB, err = sql.Open("postgres", dsn)
		if err != nil {
			log.Printf("Failed to open database connection: %v", err)
		} else {
			err = DB.Ping()
			if err == nil {
				log.Println("Successfully connected to the database")
				break
			}
			log.Printf("Failed to ping database: %v", err)
		}
		time.Sleep(5 * time.Second)
		log.Println("Retrying database connection...")
	}

	if err != nil {
		log.Fatalf("Could not connect to the database after %d attempts", maxRetries)
	}

	// Create orders table if not exists
	createTables()
}

func createTables() {
	query := `
    CREATE TABLE IF NOT EXISTS orders (
        id SERIAL PRIMARY KEY,
        user_id INT NOT NULL,
        status VARCHAR(50) NOT NULL,
        total DECIMAL(10,2) NOT NULL,
        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
    );
    CREATE TABLE IF NOT EXISTS order_items (
        id SERIAL PRIMARY KEY,
        order_id INT REFERENCES orders(id),
        product_id INT NOT NULL,
        quantity INT NOT NULL,
        price DECIMAL(10,2) NOT NULL
    );
    `
	_, err := DB.Exec(query)
	if err != nil {
		log.Fatalf("Failed to create orders and order_items tables: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
