package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"

	"github.com/Instamojo/sample-sdk-server/lib"
	"github.com/gorilla/mux"
	"github.com/instamojo/sample-sdk-server/model"
)

func main() {
	log.SetFlags(log.Lshortfile)

	router := mux.NewRouter()
	router.HandleFunc("/order/", createOrder).Methods("POST")
	router.HandleFunc("/order", createOrder).Methods("POST")
	router.HandleFunc("/status", statusHandler).Methods("GET")
	router.HandleFunc("/refund/", refundHandler).Methods("POST")
	router.HandleFunc("/ping", pingHandler).Methods("GET")

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	serverAddr := fmt.Sprintf(":%s", port)
	fmt.Printf("Starting server on port %s\n", port)
	log.Fatal(http.ListenAndServe(serverAddr, LoggingHandler(router)))
}

func createOrder(w http.ResponseWriter, r *http.Request) {
	printRequest(r)

	if r.Body == nil {
		log.Println("no body")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var getOrderIDRequest model.GetOrderIDRequest
	goErr := json.NewDecoder(r.Body).Decode(&getOrderIDRequest)
	if goErr != nil {
		log.Printf("decoder error %v", goErr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	createdOrder, err := lib.CreateOrder(getOrderIDRequest)
	if err != nil {
		log.Fatalf("Order creation failed. Error : %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("Created order: %+v", createdOrder)

	w.Header().Set("Content-Type", "application/json")
	bytes, err := json.Marshal(createdOrder)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(bytes)
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	env := r.FormValue("env")
	orderID := r.FormValue("order_id")
	transactionID := r.FormValue("transaction_id")

	gatewayOrderStatus, err := lib.GetOrderStatus(env, orderID, transactionID)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	bytes, err := json.Marshal(gatewayOrderStatus)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(bytes)
}

func refundHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	env := r.FormValue("env")
	transactionID := r.FormValue("transaction_id")
	amount := r.FormValue("amount")
	refundType := r.FormValue("type")
	body := r.FormValue("body")

	statusCode, err := lib.InitiateRefund(env, transactionID, amount, refundType, body)
	if err != nil {
		log.Println(err)
	}
	w.WriteHeader(statusCode)
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func printRequest(request *http.Request) {
	requestDump, err := httputil.DumpRequest(request, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(requestDump))
}

func printResponse(response *http.Response) {
	responseDump, err := httputil.DumpResponse(response, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(responseDump))
}
