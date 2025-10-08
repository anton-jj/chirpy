package main

import (
	"fmt"
	"net/http"
)

func (aCfg *apiConfig) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	bodyString := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", aCfg.fileServerHits.Load())
	w.Write([]byte(bodyString))
}
