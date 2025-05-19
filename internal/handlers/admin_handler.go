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

// CreateAdminUser creates a new admin user
func CreateAdminUser(c *gin.Context) {
	var request models.CreateAdminUserRequest
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

	adminColl := database.GetCollection("admin_users")
	employeeColl := database.GetCollection("employees")
	ctx := context.Background()

	// Check if username already exists in admin users
	count, err := adminColl.CountDocuments(ctx, bson.M{"username": request.Username})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Admin username already exists"})
		return
	}

	// Check if username already exists in employees
	count, err = employeeColl.CountDocuments(ctx, bson.M{"username": request.Username})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		return
	}

	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username already exists as an employee"})
		return
	}

	// Create new admin user
	now := time.Now()
	adminUser := models.AdminUser{
		ID:           primitive.NewObjectID(),
		Name:         request.Name,
		MobileNumber: request.MobileNumber,
		Email:        request.Email,
		Username:     request.Username,
		Password:     string(hashedPassword),
		IsAdminUser:  true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	_, err = adminColl.InsertOne(ctx, adminUser)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create admin user", "details": err.Error()})
		return
	}

	// Return success response without exposing the password
	c.JSON(http.StatusCreated, gin.H{
		"message": "Admin user created successfully",
		"admin": gin.H{
			"id":           adminUser.ID,
			"name":         adminUser.Name,
			"mobileNumber": adminUser.MobileNumber,
			"email":        adminUser.Email,
			"username":     adminUser.Username,
			"isAdminUser":  adminUser.IsAdminUser,
			"createdAt":    adminUser.CreatedAt,
		},
	})
}

// GetAllAdminUsers retrieves all admin users
func GetAllAdminUsers(c *gin.Context) {
	coll := database.GetCollection("admin_users")
	ctx := context.Background()

	cursor, err := coll.Find(ctx, bson.M{}, options.Find().SetSort(bson.M{"created_at": -1}))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve admin users", "details": err.Error()})
		return
	}
	defer cursor.Close(ctx)

	var adminUsers []models.AdminUser
	if err = cursor.All(ctx, &adminUsers); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse admin user data", "details": err.Error()})
		return
	}

	// Create response without passwords
	var response []gin.H
	for _, admin := range adminUsers {
		response = append(response, gin.H{
			"id":           admin.ID,
			"name":         admin.Name,
			"mobileNumber": admin.MobileNumber,
			"email":        admin.Email,
			"username":     admin.Username,
			"isAdminUser":  admin.IsAdminUser,
			"createdAt":    admin.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"admins": response,
		"count":  len(response),
	})
}

// DeleteAdminUser deletes an admin user by username
func DeleteAdminUser(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	coll := database.GetCollection("admin_users")
	ctx := context.Background()

	// Find admin user by username
	var adminUser models.AdminUser
	err := coll.FindOne(ctx, bson.M{"username": username}).Decode(&adminUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Admin user not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		}
		return
	}

	// Delete admin user
	result, err := coll.DeleteOne(ctx, bson.M{"_id": adminUser.ID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete admin user", "details": err.Error()})
		return
	}

	if result.DeletedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Admin user not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Admin user deleted successfully",
	})
}

// UpdateAdminUser updates an admin user's information
func UpdateAdminUser(c *gin.Context) {
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username is required"})
		return
	}

	var request struct {
		Name         string `json:"name"`
		MobileNumber string `json:"mobileNumber"`
		Email        string `json:"email"`
		Password     string `json:"password,omitempty"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data", "details": err.Error()})
		return
	}

	coll := database.GetCollection("admin_users")
	ctx := context.Background()

	// Find admin user by username
	var adminUser models.AdminUser
	err := coll.FindOne(ctx, bson.M{"username": username}).Decode(&adminUser)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "Admin user not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error", "details": err.Error()})
		}
		return
	}

	// Prepare update fields
	update := bson.M{
		"name":          request.Name,
		"mobile_number": request.MobileNumber,
		"email":         request.Email,
		"updated_at":    time.Now(),
	}

	// Update password if provided
	if request.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not hash password"})
			return
		}
		update["password"] = string(hashedPassword)
	}

	// Update admin user
	result, err := coll.UpdateOne(
		ctx,
		bson.M{"_id": adminUser.ID},
		bson.M{"$set": update},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update admin user", "details": err.Error()})
		return
	}

	if result.ModifiedCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Admin user not found or no changes made"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Admin user updated successfully",
	})
}
