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
	ExpireDate string `json:"expire-date,omitempty"`
}

type Request struct {
	Receipt string `json:"receipt"`
}

type responseWriter struct {
	http.ResponseWriter
	status int
	body   []byte
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body = append(rw.body, b...)
	return rw.ResponseWriter.Write(b)
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func validateReceipt(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	rw := &responseWriter{
		ResponseWriter: w,
		status:         200,
	}
	w = rw

	log.Printf("Request: Method=%s Path=%s RemoteAddr=%s UserAgent=%s",
		r.Method,
		r.URL.Path,
		r.RemoteAddr,
		r.UserAgent(),
	)

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		logResponse(rw, startTime)
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		logResponse(rw, startTime)
		return
	}

	if len(req.Receipt) == 0 {
		http.Error(w, "Receipt Required", http.StatusBadRequest)
		logResponse(rw, startTime)
		return
	}

	lastDigit, err := strconv.Atoi(string(req.Receipt[len(req.Receipt)-1]))
	if err != nil {
		http.Error(w, "Invalid Receipt Format", http.StatusBadRequest)
		logResponse(rw, startTime)
		return
	}

	location, _ := time.LoadLocation("America/Chicago")
	isValid := lastDigit%2 == 1
	response := Response{
		Status: isValid,
	}

	if isValid {
		response.ExpireDate = time.Now().In(location).Add(24 * time.Hour).Format("2006-01-02 15:04:05")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	logResponse(rw, startTime)
}

func logResponse(rw *responseWriter, startTime time.Time) {
	duration := time.Since(startTime)
	log.Printf("Response: Status=%d Duration=%v Body=%s",
		rw.status,
		duration,
		string(rw.body),
	)
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)

	http.HandleFunc("/receipt/validate", validateReceipt)

	log.Printf("Server starting on :80")
	log.Fatal(http.ListenAndServe(":80", nil))
}
