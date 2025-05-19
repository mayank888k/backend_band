package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/modernband/booking/internal/database"
	"github.com/modernband/booking/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// Login handles user authentication (no JWT)
func Login(c *gin.Context) {
	var request models.LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	coll := database.GetCollection("employees")
	ctx := context.Background()

	// Find employee by username
	var employee models.Employee
	err := coll.FindOne(ctx, bson.M{"username": request.Username}).Decode(&employee)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		}
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(employee.Password), []byte(request.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Return success response (no token)
	c.JSON(http.StatusOK, gin.H{
		"employee": gin.H{
			"id":                       employee.ID,
			"name":                     employee.Name,
			"mobileNumber":             employee.MobileNumber,
			"email":                    employee.Email,
			"address":                  employee.Address,
			"isEmployee":               employee.IsEmployee,
			"totalAmountToBePaid":      employee.TotalAmountToBePaid,
			"totalAmountPaidInAdvance": employee.TotalAmountPaidInAdvance,
			"username":                 employee.Username,
		},
	})
}

// AdminLogin handles admin user authentication (no JWT)
func AdminLogin(c *gin.Context) {
	var request models.LoginRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	coll := database.GetCollection("admin_users")
	ctx := context.Background()

	// Find admin user by username
	var adminUser models.AdminUser
	err := coll.FindOne(ctx, bson.M{"username": request.Username}).Decode(&adminUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		}
		return
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(adminUser.Password), []byte(request.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
		return
	}

	// Return success response (no token)
	c.JSON(http.StatusOK, gin.H{
		"admin": gin.H{
			"id":           adminUser.ID,
			"name":         adminUser.Name,
			"mobileNumber": adminUser.MobileNumber,
			"email":        adminUser.Email,
			"username":     adminUser.Username,
			"isAdmin":      adminUser.IsAdminUser,
		},
	})
}
