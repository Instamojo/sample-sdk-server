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

// GatewayOrder is the request to create a gateway order
type GatewayOrder struct {
	ID string `json:"id"`

	Name string `json:"name"`

	Email string `json:"email"`

	Phone string `json:"phone"`

	Amount string `json:"amount"`

	Description string `json:"description"`

	Currency string `json:"currency"`

	TransactionID string `json:"transaction_id"`

	Payments []Payment `json:"payments"`

	Status string `json:"status"`

	RedirectURL string `json:"redirect_url"`
}

// GatewayOrderResponse is the response of Create GatewayOrder call
type GatewayOrderResponse struct {
	Order GatewayOrder `json:"order"`
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

// GatewayOrderStatus returns the status of the payment order
type GatewayOrderStatus struct {
	Amount string `json:"amount"`

	Status string `json:"status"`

	PaymentID string `json:"payment_id"`
}

// Payment has details of order payment
type Payment struct {
	ID string `json:"id"`

	Status string `json:"status"`

	InstrumentType string `json:"instrument_type"`

	BillingInstrument string `json:"billing_instrument"`

	Failure string `json:"failure"`
}
