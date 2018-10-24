package lib

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/instamojo/sample-sdk-server/config"
	"github.com/instamojo/sample-sdk-server/model"
)

const testENV = "test"
const prodENV = "production"

var env string
var clientID string
var clientSecret string
var imojoURL string
var client http.Client

func init() {
	client = http.Client{}
	setDefaultEnvironment()
}

func setDefaultEnvironment() {
	// Defaults to test config
	setEnviroment(testENV)
}

func setEnviroment(newEnv string) {
	// Don't change if already set
	if env == newEnv {
		return
	}

	env = newEnv
	if env == prodENV {
		imojoURL = config.Config.ProdURL
		clientID = config.Config.ProdClientID
		clientSecret = config.Config.ProdClientSecret

	} else {
		imojoURL = config.Config.TestURL
		clientID = config.Config.TestClientID
		clientSecret = config.Config.TestClientSecret
	}
}

func fetchToken() (*model.OAuth2Token, error) {
	log.Println("Fetching new access token")
	values := url.Values{}
	values.Set("client_id", clientID)
	values.Set("client_secret", clientSecret)
	values.Set("grant_type", "client_credentials")
	httpRequest, err := http.NewRequest("POST", imojoURL+"/oauth2/token/", bytes.NewBufferString(values.Encode()))
	if err != nil {
		return nil, err
	}

	httpRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	httpResponse, err := client.Do(httpRequest)
	if err != nil {
		return nil, err
	}

	token := &model.OAuth2Token{}
	decodeErr := json.NewDecoder(httpResponse.Body).Decode(token)
	if decodeErr != nil {
		return nil, decodeErr
	}

	return token, nil
}

// CreateOrder will create a new payment order and returns the same
func CreateOrder(request model.GetOrderIDRequest) (*model.Order, error) {
	setEnviroment(strings.ToLower(request.Env))

	// Create GatewayOrder
	gatewayOrderResponse, prErr := createGatewayOrder(request)
	if prErr != nil {
		log.Printf("Error %v", prErr)
		return nil, prErr
	}

	// Create Order
	order, oErr := createOrderForGWOrder(gatewayOrderResponse.Order.ID)
	if oErr != nil {
		log.Printf("Error %v", oErr)
		return nil, oErr
	}

	log.Printf("Created order with ID %s", order.OrderID)
	return order, nil
}

func createGatewayOrder(getOrderIDRequest model.GetOrderIDRequest) (*model.GatewayOrderResponse, error) {
	log.Println("Creating payment request")
	gatewayOrder := model.GatewayOrder{}
	gatewayOrder.Name = getOrderIDRequest.BuyerName
	gatewayOrder.Email = getOrderIDRequest.BuyerEmail
	gatewayOrder.Phone = getOrderIDRequest.BuyerPhone
	gatewayOrder.Amount = getOrderIDRequest.Amount
	gatewayOrder.Description = getOrderIDRequest.Description
	gatewayOrder.Currency = "INR"
	gatewayOrder.TransactionID = uuid.New().String()
	gatewayOrder.RedirectURL = imojoURL + "/integrations/android/redirect/"

	jsonPaymentRequest, _ := json.Marshal(gatewayOrder)
	httpRequest, _ := http.NewRequest("POST", imojoURL+"/v2/gateway/orders/", bytes.NewBuffer(jsonPaymentRequest))
	token, tErr := fetchToken()
	if tErr != nil {
		log.Printf("Error %v", tErr)
		return nil, tErr
	}

	httpRequest.Header.Set("Authorization", "Bearer "+token.AccessToken)
	httpRequest.Header.Set("Content-Type", "application/json")

	httpResponse, err := client.Do(httpRequest)
	if err != nil {
		log.Printf("Error %v", err)
		return nil, err
	}

	var gatewayOrderResponse model.GatewayOrderResponse
	decodeErr := json.NewDecoder(httpResponse.Body).Decode(&gatewayOrderResponse)
	if decodeErr != nil {
		log.Printf("Decode Error %v", decodeErr)
		return nil, decodeErr
	}

	return &gatewayOrderResponse, nil
}

func createOrderForGWOrder(gatewayOrderID string) (*model.Order, error) {
	log.Printf("Creating order for payment request ID %s", gatewayOrderID)
	orderRequest := model.OrderRequest{}
	orderRequest.PaymentRequestID = gatewayOrderID

	jsonOrderRequest, _ := json.Marshal(orderRequest)
	httpRequest, _ := http.NewRequest("POST", imojoURL+"/v2/gateway/orders/payment-request/", bytes.NewBuffer(jsonOrderRequest))
	token, tErr := fetchToken()
	if tErr != nil {
		return nil, tErr
	}

	httpRequest.Header.Set("Authorization", "Bearer "+token.AccessToken)
	httpRequest.Header.Set("Content-Type", "application/json")

	httpResponse, err := client.Do(httpRequest)
	if err != nil {
		return nil, err
	}

	var createdOrder model.Order
	decodeErr := json.NewDecoder(httpResponse.Body).Decode(&createdOrder)
	if decodeErr != nil {
		return nil, decodeErr
	}

	return &createdOrder, nil
}

// GetOrderStatus return the status of the order referencing either orderID or transactionID.
// Preference will be given to orderID
func GetOrderStatus(env, orderID, transactionID string) (*model.GatewayOrderStatus, error) {
	setEnviroment(strings.ToLower(env))

	gatewayOrder, err := getGatewayOrder(orderID, transactionID)
	if err != nil {
		log.Printf("Error %v", err)
		return nil, err
	}

	var gatewayOrderStatus model.GatewayOrderStatus
	gatewayOrderStatus.Amount = gatewayOrder.Amount
	gatewayOrderStatus.Status = gatewayOrder.Status

	if len(gatewayOrder.Payments) > 0 {
		gatewayOrderStatus.PaymentID = gatewayOrder.Payments[0].ID
	}

	return &gatewayOrderStatus, nil
}

func getGatewayOrder(orderID, transactionID string) (*model.GatewayOrder, error) {
	orderURL := imojoURL + "/v2/gateway/orders/"
	if orderID == "" {
		orderURL += "transaction_id:" + transactionID + "/"

	} else {
		orderURL += "id:" + orderID + "/"
	}

	orderRequest, err := http.NewRequest("GET", orderURL, nil)
	if err != nil {
		log.Printf("Error %v", err)
		return nil, err
	}

	token, tErr := fetchToken()
	if tErr != nil {
		log.Printf("Error %v", tErr)
		return nil, tErr
	}

	orderRequest.Header.Set("Authorization", "Bearer "+token.AccessToken)
	orderRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	httpResponse, err := client.Do(orderRequest)
	if err != nil {
		log.Printf("Error %v", err)
		return nil, err
	}

	var gatewayOrder model.GatewayOrder
	decodeErr := json.NewDecoder(httpResponse.Body).Decode(&gatewayOrder)
	if decodeErr != nil {
		log.Printf("Decode Error %v", decodeErr)
		return nil, decodeErr
	}

	return &gatewayOrder, nil
}

//InitiateRefund wil initiate refund for the paymentID for the given with given refund reason
//refundType should be within the following types
//RFD: Duplicate/delayed payment.
//TNR: Product/service no longer available.
//QFL: Customer not satisfied.
//QNR: Product lost/damaged.
//EWN: Digital download issue.
//TAN: Event was canceled/changed.
//PTH: Problem not described above.
func InitiateRefund(env, transactionID, amount string) (int, error) {
	setEnviroment(env)

	gatewayOrder, err := getGatewayOrder("", transactionID)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	if gatewayOrder.Success || len(gatewayOrder.Payments) < 1 {
		return http.StatusBadRequest, errors.New(gatewayOrder.Message)
	}

	payment := gatewayOrder.Payments[0]

	if payment.Status != "successful" {
		return http.StatusBadRequest, errors.New("Cannot initiate refund for an Unsuccessful transaction")
	}

	refundURL := imojoURL + "/v2/payments/" + payment.ID + "/refund/"
	params := url.Values{}
	params.Set("type", "PTH")
	params.Set("refund_amount", amount)
	params.Set("body", "Refund the Amount")

	refundRequest, err := http.NewRequest("POST", refundURL, bytes.NewBufferString(params.Encode()))
	if err != nil {
		return http.StatusInternalServerError, err
	}

	token, tErr := fetchToken()
	if tErr != nil {
		return 0, tErr
	}

	refundRequest.Header.Set("Authorization", "Bearer "+token.AccessToken)
	refundRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	httpResponse, err := client.Do(refundRequest)
	if err != nil {
		return http.StatusInternalServerError, err
	}

	return httpResponse.StatusCode, nil
}
