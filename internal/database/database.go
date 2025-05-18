package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"

	_ "modernc.org/sqlite" // Pure Go SQLite driver
)

var (
	db   *sql.DB
	once sync.Once
)

// GetDB returns a singleton database connection
func GetDB() *sql.DB {
	once.Do(func() {
		var err error
		// Create data directory if it doesn't exist
		if _, err := os.Stat("./data"); os.IsNotExist(err) {
			if err := os.Mkdir("./data", 0755); err != nil {
				log.Fatalf("Failed to create data directory: %v", err)
			}
		}

		// Open SQLite database
		db, err = sql.Open("sqlite", "./data/app.db")
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}

		// Test the connection
		if err = db.Ping(); err != nil {
			log.Fatalf("Failed to ping database: %v", err)
		}

		// Initialize the database schema
		if err = initDB(db); err != nil {
			log.Fatalf("Failed to initialize database schema: %v", err)
		}
	})

	return db
}

// InitDB creates the necessary tables if they don't exist
func initDB(db *sql.DB) error {
	// Enable foreign keys
	_, err := db.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		return fmt.Errorf("error enabling foreign keys: %w", err)
	}

	// Create bookings table
	bookingsTable := `
	CREATE TABLE IF NOT EXISTS bookings (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT NOT NULL,
		phone TEXT NOT NULL,
		additional_phone TEXT,
		package_type TEXT NOT NULL,
		event_date DATETIME NOT NULL,
		venue TEXT NOT NULL,
		city TEXT NOT NULL,
		customization TEXT,
		band_time TEXT,
		custom_time_slot TEXT,
		number_of_people INTEGER,
		number_of_lights INTEGER,
		number_of_dhols INTEGER,
		ghoda_baggi INTEGER,
		ghodi_for_baraat BOOLEAN,
		fireworks BOOLEAN,
		fireworks_amount INTEGER,
		flower_canon BOOLEAN,
		doli_for_vidai BOOLEAN,
		amount INTEGER NOT NULL,
		advance_payment INTEGER NOT NULL,
		phone_verified BOOLEAN DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// Create employees table
	employeesTable := `
	CREATE TABLE IF NOT EXISTS employees (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		mobile_number TEXT NOT NULL,
		email TEXT NOT NULL,
		address TEXT NOT NULL,
		is_employee BOOLEAN DEFAULT 1,
		total_amount_to_be_paid REAL NOT NULL,
		total_amount_paid_in_advance REAL NOT NULL,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// Create payments table
	paymentsTable := `
	CREATE TABLE IF NOT EXISTS payments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		amount_paid REAL NOT NULL,
		date DATETIME NOT NULL,
		employee_id INTEGER NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (employee_id) REFERENCES employees(id) ON DELETE CASCADE
	);`

	// Create admin_users table
	adminUsersTable := `
	CREATE TABLE IF NOT EXISTS admin_users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		mobile_number TEXT NOT NULL,
		email TEXT NOT NULL,
		username TEXT NOT NULL UNIQUE,
		password TEXT NOT NULL,
		is_admin_user BOOLEAN DEFAULT 1,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	// Execute SQL to create tables
	_, err = db.Exec(bookingsTable)
	if err != nil {
		return fmt.Errorf("error creating bookings table: %w", err)
	}

	_, err = db.Exec(employeesTable)
	if err != nil {
		return fmt.Errorf("error creating employees table: %w", err)
	}

	_, err = db.Exec(paymentsTable)
	if err != nil {
		return fmt.Errorf("error creating payments table: %w", err)
	}

	_, err = db.Exec(adminUsersTable)
	if err != nil {
		return fmt.Errorf("error creating admin_users table: %w", err)
	}

	log.Println("Database schema initialized successfully")
	return nil
}
