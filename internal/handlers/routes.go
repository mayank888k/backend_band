package handlers

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// SetupRoutes configures all the routes for the API
func SetupRoutes(router *gin.Engine) {
	// Enable CORS
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})

	// API routes
	api := router.Group("/api")
	{
		// Booking endpoints
		api.POST("/book", CreateBooking)
		api.GET("/booking", GetBooking)
		api.GET("/bookings", GetAllBookings)
		api.DELETE("/bookings/past", DeletePastBookings)
		api.DELETE("/bookings/:id", DeleteBooking)

		// Authentication endpoint
		api.POST("/login", Login)
		api.POST("/signin", AdminLogin)

		// Employee endpoints
		api.POST("/employees", CreateEmployee)
		api.GET("/employees", GetAllEmployees)
		api.DELETE("/employees/:username", DeleteEmployee)
		api.POST("/employees/:username/payments", AddPayment)
		api.DELETE("/employees/:username/payments/:paymentID", DeletePayment)
		api.GET("/employees/:username", GetEmployeeDetails)

		// Admin endpoints
		api.POST("/admin", CreateAdminUser)

		// Health check
		api.GET("/health", func(c *gin.Context) {
			port := "8081"  // Default port
			if p := os.Getenv("PORT"); p != "" {
				port = p
			}
			c.JSON(http.StatusOK, gin.H{
				"status": "UP",
				"message": "Server is running on port " + port,
			})
		})
	}

	// Default route for unknown paths
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
	})
}
