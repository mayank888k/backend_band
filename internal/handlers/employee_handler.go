package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/modernband/booking/internal/database"
	"github.com/modernband/booking/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	coll := database.GetCollection("employees")
	ctx := context.Background()

	// Check if username already exists
	count, err := coll.CountDocuments(ctx, bson.M{"username": request.Username})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username already exists"})
		return
	}

	// Create new employee
	now := time.Now()
	employee := models.Employee{
		Name:                     request.Name,
		MobileNumber:             request.MobileNumber,
		Email:                    request.Email,
		Address:                  request.Address,
		IsEmployee:               true,
		TotalAmountToBePaid:      request.TotalAmountToBePaid,
		TotalAmountPaidInAdvance: request.TotalAmountPaidInAdvance,
		Username:                 request.Username,
		Password:                 string(hashedPassword),
		CreatedAt:                now,
		UpdatedAt:                now,
	}

	result, err := coll.InsertOne(ctx, employee)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create employee", "details": err.Error()})
		return
	}

	employee.ID = result.InsertedID.(primitive.ObjectID)

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Employee created successfully",
		"employee": employee,
	})
}

// GetAllEmployees retrieves all employees from the database
func GetAllEmployees(c *gin.Context) {
	coll := database.GetCollection("employees")
	ctx := context.Background()

	// Set options for sorting by created_at in descending order
	opts := options.Find().SetSort(bson.M{"created_at": -1})

	cursor, err := coll.Find(ctx, bson.M{}, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve employees", "details": err.Error()})
		return
	}
	defer cursor.Close(ctx)

	// First decode into full Employee structs
	var employees []models.Employee
	if err = cursor.All(ctx, &employees); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse employee data", "details": err.Error()})
		return
	}

	// Convert to EmployeeResponse structs
	var response []models.EmployeeResponse
	for _, emp := range employees {
		response = append(response, models.EmployeeResponse{
			ID:                       emp.ID,
			Name:                     emp.Name,
			MobileNumber:             emp.MobileNumber,
			Email:                    emp.Email,
			Address:                  emp.Address,
			IsEmployee:               emp.IsEmployee,
			TotalAmountToBePaid:      emp.TotalAmountToBePaid,
			TotalAmountPaidInAdvance: emp.TotalAmountPaidInAdvance,
			Username:                 emp.Username,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"employees": response,
		"count":     len(response),
	})
}

// DeleteEmployee deletes an employee by username
func DeleteEmployee(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	coll := database.GetCollection("employees")
	paymentsColl := database.GetCollection("payments")
	ctx := context.Background()

	// Find employee by username
	var employee models.Employee
	err := coll.FindOne(ctx, bson.M{"username": username}).Decode(&employee)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		}
		return
	}

	// Start a session for transaction
	session, err := database.GetDB().Client().StartSession()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start session", "details": err.Error()})
		return
	}
	defer session.EndSession(ctx)

	// Perform transaction
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		// Delete associated payments
		_, err := paymentsColl.DeleteMany(sessCtx, bson.M{"employee_id": employee.ID})
		if err != nil {
			return nil, err
		}

		// Delete employee
		_, err = coll.DeleteOne(sessCtx, bson.M{"_id": employee.ID})
		if err != nil {
			return nil, err
		}

		return nil, nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete employee", "details": err.Error()})
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

	coll := database.GetCollection("employees")
	paymentsColl := database.GetCollection("payments")
	ctx := context.Background()

	// Find employee by username
	var employee models.Employee
	err := coll.FindOne(ctx, bson.M{"username": username}).Decode(&employee)
	if err != nil {
		if err == mongo.ErrNoDocuments {
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
	payment := models.Payment{
		AmountPaid: request.AmountPaid,
		Date:       date,
		EmployeeID: employee.ID,
		CreatedAt:  time.Now(),
	}

	result, err := paymentsColl.InsertOne(ctx, payment)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment", "details": err.Error()})
		return
	}

	payment.ID = result.InsertedID.(primitive.ObjectID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Payment added successfully",
		"payment": payment,
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

	paymentID, err := primitive.ObjectIDFromHex(paymentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid payment ID format"})
		return
	}

	coll := database.GetCollection("employees")
	paymentsColl := database.GetCollection("payments")
	ctx := context.Background()

	// Find employee by username
	var employee models.Employee
	err = coll.FindOne(ctx, bson.M{"username": username}).Decode(&employee)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		}
		return
	}

	// Delete payment
	result, err := paymentsColl.DeleteOne(ctx, bson.M{
		"_id":         paymentID,
		"employee_id": employee.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete payment", "details": err.Error()})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Payment deleted successfully",
	})
}

// GetEmployeeDetails retrieves detailed information about an employee
func GetEmployeeDetails(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	coll := database.GetCollection("employees")
	paymentsColl := database.GetCollection("payments")
	ctx := context.Background()

	// Find employee by username
	var employee models.Employee
	err := coll.FindOne(ctx, bson.M{"username": username}).Decode(&employee)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Employee not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		}
		return
	}

	// Get employee payments
	cursor, err := paymentsColl.Find(ctx, bson.M{"employee_id": employee.ID}, options.Find().SetSort(bson.D{{Key: "date", Value: -1}}))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve payments", "details": err.Error()})
		return
	}
	defer cursor.Close(ctx)

	var payments []models.Payment
	if err = cursor.All(ctx, &payments); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse payment data", "details": err.Error()})
		return
	}

	// Create response
	response := models.EmployeeResponse{
		ID:                       employee.ID,
		Name:                     employee.Name,
		MobileNumber:             employee.MobileNumber,
		Email:                    employee.Email,
		Address:                  employee.Address,
		IsEmployee:               employee.IsEmployee,
		TotalAmountToBePaid:      employee.TotalAmountToBePaid,
		TotalAmountPaidInAdvance: employee.TotalAmountPaidInAdvance,
		Username:                 employee.Username,
		Payments:                 payments,
	}

	c.JSON(http.StatusOK, response)
}
