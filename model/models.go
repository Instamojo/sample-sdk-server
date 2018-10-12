package model

// GetOrderIDRequest is the request from the android app
// to create and retrieve the Instamojo OrderID
// OrderID can used to complete the payment with in the app using instamojo-android-sdk
type GetOrderIDRequest struct {
	Env string `json:"env"`

	BuyerName string `json:"buyer_name"`

	BuyerEmail string `json:"buyer_email"`

	BuyerPhone string `json:"buyer_phone"`

	Amount string `json:"amount"`

	Description string `json:"description"`
}

// OAuth2Token Token response from token endpoint
type OAuth2Token struct {
	AccessToken string `json:"access_token"`

	ExpiresIn int `json:"expires_in"`

	TokenType string `json:"token_type"`

	Scope string `json:"scope"`
}

// PaymentRequest is request for a new payment for a buyer
type PaymentRequest struct {
	ID string `json:"id"`

	Purpose string `json:"purpose"`

	BuyerName string `json:"buyer_name"`

	Email string `json:"email"`

	Phone string `json:"phone"`

	Amount string `json:"amount"`
}

// OrderRequest is request to create an order for a payment request
type OrderRequest struct {
	PaymentRequestID string `json:"id"`
}

// Order for a payment request
type Order struct {
	OrderID string `json:"order_id"`

	Name string `json:"name"`

	Email string `json:"email"`

	Phone string `json:"phone"`

	Amount string `json:"amount"`
}
