package lib

import (
	"bytes"
	"encoding/json"
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

var refundTypes = map[string]string{
	"RFD": "Duplicate/delayed payment.",
	"TNR": "Product/service no longer available.",
	"QFL": "Customer not satisfied.",
	"QNR": "Product lost/damaged.",
	"EWN": "Digital download issue.",
	"TAN": "Event was canceled/changed.",
	"PTH": "Problem not described above.",
}

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
		return nil, prErr
	}

	// Create Order
	order, oErr := createOrderForGWOrder(gatewayOrderResponse.Order.ID)
	if oErr != nil {
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

	jsonPaymentRequest, _ := json.Marshal(gatewayOrder)
	httpRequest, _ := http.NewRequest("POST", imojoURL+"/v2/gateway/orders/", bytes.NewBuffer(jsonPaymentRequest))
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

	var gatewayOrderResponse model.GatewayOrderResponse
	decodeErr := json.NewDecoder(httpResponse.Body).Decode(&gatewayOrderResponse)
	if decodeErr != nil {
		return nil, decodeErr
	}

	return &gatewayOrderResponse, nil
}

func createOrderForGWOrder(gatewayOrderID string) (*model.Order, error) {
	log.Println("Creating order for payment request")
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

//GetOrderStatus return the status of the order referencing either orderID or transactionID. Preference will be given to
//orderID
func GetOrderStatus(env, authorizationHeader, orderID, transactionID string) ([]byte, error) {
	// statusRequest, err := http.NewRequest("GET", imojoURL+"/v2/gateway/orders/", nil)
	// if err != nil {
	// 	return []byte(""), err
	// }
	//
	// statusRequest.Header.Set("Authorization", authorizationHeader)
	// statusRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	//
	// client := &http.Client{}
	// resp, err := client.Do(statusRequest)
	// if err != nil {
	// 	return []byte(""), err
	// }
	//
	// data, err := ioutil.ReadAll(resp.Body)
	// defer resp.Body.Close()
	//
	// if err != nil {
	// 	return []byte(""), err
	// }
	//
	// return data, nil
	return []byte(""), nil
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
func InitiateRefund(env, authorizationHeader, transactionID, amount, refundType, body string) (int, error) {
	// refundURL := PROD_URL
	// if env == "test" {
	// 	refundURL = TEST_URL
	// }
	//
	// if _, exist := refundTypes[refundType]; !exist {
	// 	return http.StatusBadRequest, errors.New("Invalid refund type " + refundType)
	// }
	//
	// data, err := GetOrderStatus(env, authorizationHeader, "", transactionID)
	// if err != nil {
	// 	return http.StatusInternalServerError, err
	// }
	//
	// var jsonResponse struct {
	// 	ID            string `json:"id"`
	// 	TransactionID string `json:"transaction_id"`
	// 	Payments      []struct {
	// 		ID     string `json:"id"`
	// 		Status string `json:"status"`
	// 	} `json:"payments"`
	// 	Success bool   `json:"success"`
	// 	Message string `json:"message"`
	// }
	//
	// if err := json.Unmarshal(data, &jsonResponse); err != nil {
	// 	return http.StatusInternalServerError, err
	// }
	//
	// if jsonResponse.Success || len(jsonResponse.Payments) < 1 {
	// 	return http.StatusBadRequest, errors.New(jsonResponse.Message)
	// }
	//
	// status := jsonResponse.Payments[0].Status
	// paymentID := jsonResponse.Payments[0].ID
	//
	// if status != "successful" {
	// 	return http.StatusBadRequest, errors.New("Cannot initiate refund for an Unsuccessful transaction")
	// }
	//
	// refundURL += fmt.Sprintf("/v2/payments/%s/refund/", paymentID)
	// params := url.Values{}
	// params.Set("type", refundType)
	// params.Set("refund_amount", amount)
	// params.Set("body", body)
	//
	// refundRequest, err := http.NewRequest("POST", refundURL, bytes.NewBufferString(params.Encode()))
	// if err != nil {
	// 	return http.StatusInternalServerError, err
	// }
	//
	// refundRequest.Header.Set("Authorization", authorizationHeader)
	// refundRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	//
	// client := &http.Client{}
	// resp, err := client.Do(refundRequest)
	// if err != nil {
	// 	return http.StatusInternalServerError, err
	// }
	// _, err = ioutil.ReadAll(resp.Body)
	// if err != nil {
	// 	return http.StatusInternalServerError, err
	// }
	//
	// return resp.StatusCode, nil
	return 0, nil
}
