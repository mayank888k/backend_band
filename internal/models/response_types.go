package models

// LoginResponse represents the response for login endpoint
type LoginResponse struct {
	Username string `json:"username"`
	IsAdmin  bool   `json:"isAdmin"`
	Message  string `json:"message"`
}

// EmployeeResponse represents the detailed employee info including payments
type EmployeeResponse struct {
	ID                       uint      `json:"id"`
	Name                     string    `json:"name"`
	MobileNumber             string    `json:"mobileNumber"`
	Email                    string    `json:"email"`
	Address                  string    `json:"address"`
	TotalAmountToBePaid      float64   `json:"totalAmountToBePaid"`
	TotalAmountPaidInAdvance float64   `json:"totalAmountPaidInAdvance"`
	Username                 string    `json:"username"`
	IsEmployee               bool      `json:"isEmployee"`
	Payments                 []Payment `json:"payments,omitempty"`
}

// ErrorResponse represents error message in API responses
type ErrorResponse struct {
	Error string `json:"error"`
}
