package handlers

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/modernband/booking/internal/database"
	"github.com/modernband/booking/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// CreateEmployee handles the creation of a new employee
func CreateEmployee(c *gin.Context) {
	var request models.CreateEmployeeRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash password"})
		return
	}

	db := database.GetDB()

	// Check if username already exists
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM employees WHERE username = ?", request.Username).Scan(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username already exists"})
		return
	}

	// Create new employee
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := db.Exec(`
		INSERT INTO employees (
			name, mobile_number, email, address, is_employee, 
			total_amount_to_be_paid, total_amount_paid_in_advance, 
			username, password, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		request.Name, request.MobileNumber, request.Email, request.Address, true,
		request.TotalAmountToBePaid, request.TotalAmountPaidInAdvance,
		request.Username, string(hashedPassword), now, now,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create employee", "details": err.Error()})
		return
	}

	// Get the ID of the newly created employee
	id, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get employee ID", "details": err.Error()})
		return
	}

	// Return success response without exposing the password
	c.JSON(http.StatusCreated, gin.H{
		"message": "Employee created successfully",
		"employee": models.EmployeeResponse{
			ID:                       uint(id),
			Name:                     request.Name,
			MobileNumber:             request.MobileNumber,
			Email:                    request.Email,
			Address:                  request.Address,
			TotalAmountToBePaid:      request.TotalAmountToBePaid,
			TotalAmountPaidInAdvance: request.TotalAmountPaidInAdvance,
			Username:                 request.Username,
			IsEmployee:               true,
			Payments:                 []models.Payment{},
		},
	})
}

// GetAllEmployees returns a list of all employees
func GetAllEmployees(c *gin.Context) {
	db := database.GetDB()

	rows, err := db.Query(`
		SELECT id, name, mobile_number, email, address, 
		is_employee, total_amount_to_be_paid, total_amount_paid_in_advance, 
		username, created_at, updated_at 
		FROM employees
	`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve employees", "details": err.Error()})
		return
	}
	defer rows.Close()

	var employees []models.EmployeeResponse
	for rows.Next() {
		var employee models.EmployeeResponse
		var id int
		var createdAt, updatedAt string

		err := rows.Scan(
			&id, &employee.Name, &employee.MobileNumber, &employee.Email, &employee.Address,
			&employee.IsEmployee, &employee.TotalAmountToBePaid, &employee.TotalAmountPaidInAdvance,
			&employee.Username, &createdAt, &updatedAt,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse employee data", "details": err.Error()})
			return
		}

		employee.ID = uint(id)
		employees = append(employees, employee)
	}

	c.JSON(http.StatusOK, gin.H{
		"employees": employees,
		"count":     len(employees),
	})
}

// DeleteEmployee deletes an employee by username
func DeleteEmployee(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	db := database.GetDB()

	// Check if employee exists
	var id int
	err := db.QueryRow("SELECT id FROM employees WHERE username = ?", username).Scan(&id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		}
		return
	}

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to begin transaction", "details": err.Error()})
		return
	}

	// Delete associated payments
	_, err = tx.Exec("DELETE FROM payments WHERE employee_id = ?", id)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete employee payments", "details": err.Error()})
		return
	}

	// Delete employee
	_, err = tx.Exec("DELETE FROM employees WHERE id = ?", id)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete employee", "details": err.Error()})
		return
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Employee deleted successfully",
	})
}

// AddPayment adds a payment to an employee
func AddPayment(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	var request models.CreatePaymentRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	db := database.GetDB()

	// Find employee by username
	var employeeID int
	err := db.QueryRow("SELECT id FROM employees WHERE username = ?", username).Scan(&employeeID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		}
		return
	}

	// Parse date
	date, err := time.Parse("2006-01-02", request.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid date format. Use YYYY-MM-DD"})
		return
	}

	// Create payment
	now := time.Now().UTC().Format(time.RFC3339)
	result, err := db.Exec(
		"INSERT INTO payments (amount_paid, date, employee_id, created_at) VALUES (?, ?, ?, ?)",
		request.AmountPaid, date.Format(time.RFC3339), employeeID, now,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment", "details": err.Error()})
		return
	}

	// Get the ID of the newly created payment
	paymentID, err := result.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get payment ID", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Payment added successfully",
		"payment": models.Payment{
			ID:         uint(paymentID),
			AmountPaid: request.AmountPaid,
			Date:       date,
			EmployeeID: uint(employeeID),
			CreatedAt:  time.Now(),
		},
	})
}

// DeletePayment deletes a payment record
func DeletePayment(c *gin.Context) {
	username := c.Param("username")
	paymentIDStr := c.Param("paymentID")
	if username == "" || paymentIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and payment ID are required"})
		return
	}

	// Parse payment ID
	paymentID, err := strconv.Atoi(paymentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment ID"})
		return
	}

	db := database.GetDB()

	// Find employee by username
	var employeeID int
	err = db.QueryRow("SELECT id FROM employees WHERE username = ?", username).Scan(&employeeID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		}
		return
	}

	// Find and delete payment
	result, err := db.Exec("DELETE FROM payments WHERE id = ? AND employee_id = ?", paymentID, employeeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete payment", "details": err.Error()})
		return
	}

	// Check if payment was found and deleted
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get affected rows", "details": err.Error()})
		return
	}

	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found for this employee"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Payment deleted successfully",
	})
}

// GetEmployeeDetails gets detailed employee info including payments
func GetEmployeeDetails(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	db := database.GetDB()

	// Find employee by username
	var employee models.EmployeeResponse
	var id int
	var createdAt, updatedAt string

	err := db.QueryRow(`
		SELECT id, name, mobile_number, email, address, 
		is_employee, total_amount_to_be_paid, total_amount_paid_in_advance, 
		username, created_at, updated_at 
		FROM employees WHERE username = ?
	`, username).Scan(
		&id, &employee.Name, &employee.MobileNumber, &employee.Email, &employee.Address,
		&employee.IsEmployee, &employee.TotalAmountToBePaid, &employee.TotalAmountPaidInAdvance,
		&employee.Username, &createdAt, &updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		}
		return
	}

	employee.ID = uint(id)

	// Get employee payments
	rows, err := db.Query(`
		SELECT id, amount_paid, date, created_at 
		FROM payments 
		WHERE employee_id = ?
		ORDER BY date DESC
	`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve payments", "details": err.Error()})
		return
	}
	defer rows.Close()

	var payments []models.Payment
	for rows.Next() {
		var payment models.Payment
		var paymentID int
		var dateStr, createdAtStr string

		err := rows.Scan(&paymentID, &payment.AmountPaid, &dateStr, &createdAtStr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse payment data", "details": err.Error()})
			return
		}

		payment.ID = uint(paymentID)
		payment.EmployeeID = uint(id)
		payment.Date, _ = time.Parse(time.RFC3339, dateStr)
		payment.CreatedAt, _ = time.Parse(time.RFC3339, createdAtStr)

		payments = append(payments, payment)
	}

	employee.Payments = payments

	// Return employee details with payments
	c.JSON(http.StatusOK, gin.H{
		"employee": employee,
	})
}
