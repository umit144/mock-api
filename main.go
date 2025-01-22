package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"
)

type Response struct {
	Status     bool   `json:"status"`
	ExpireDate string `json:"expire-date"`
}

type Request struct {
	Receipt string `json:"receipt"`
}

func validateReceipt(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if len(req.Receipt) == 0 {
		http.Error(w, "Receipt Required", http.StatusBadRequest)
		return
	}

	lastDigit, err := strconv.Atoi(string(req.Receipt[len(req.Receipt)-1]))
	if err != nil {
		http.Error(w, "Invalid Receipt Format", http.StatusBadRequest)
		return
	}

	location, _ := time.LoadLocation("America/Chicago")
	response := Response{
		Status:     lastDigit%2 == 1,
		ExpireDate: time.Now().In(location).Add(24 * time.Hour).Format("2006-01-02 15:04:05"),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/receipt/validate", validateReceipt)

	log.Printf("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
