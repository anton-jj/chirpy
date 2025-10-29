package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/anton-jj/chripy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type responeError struct {
	Error string `json:"error"`
}

type apiConfig struct {
	fileServerHits atomic.Int32
	db             *database.Queries
	secret         string
}

type cleanedData struct {
	Cleaned_body string `json:"cleaned_body"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		os.Exit(1)
	}
	dbURL := os.Getenv("DB_URL")
	secret := os.Getenv("SECRET")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		os.Exit(1)
	}
	dbQueries := database.New(db)
	const filePathRoot = "."
	const port = ":8080"

	apiConfig := apiConfig{
		fileServerHits: atomic.Int32{},
		db:             dbQueries,
		secret:         secret,
	}
	log.Print(apiConfig.secret)

	mux := http.NewServeMux()
	fsHandler := apiConfig.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filePathRoot))))
	mux.Handle("/app/", fsHandler)

	mux.HandleFunc("GET /api/healthz", handleHealtz)
	mux.HandleFunc("GET /admin/metrics", apiConfig.handleMetrics)
	mux.HandleFunc("POST /admin/reset", apiConfig.handleReset)
	mux.HandleFunc("POST /api/users", apiConfig.handleUsers)
	mux.HandleFunc("POST /api/login", apiConfig.handleLogin)
	mux.HandleFunc("POST /api/chirps", apiConfig.handleChirpCreate)
	mux.HandleFunc("GET /api/chirps", apiConfig.handleChirpsGetAll)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiConfig.handleGetChrip)

	server := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	log.Printf("Serving files from %s to port: %s", filePathRoot, server.Addr)
	server.ListenAndServe().Error()
}

func handleHealtz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func validateBody(data string) string {
	badWords := map[string]int{
		"kerfuffle": 0,
		"sharbert":  0,
		"fornax":    0,
	}
	var cleanedData []string
	for _, s := range strings.Split(data, " ") {
		log.Println(s)
		if _, ok := badWords[strings.ToLower(s)]; ok {
			cleanedData = append(cleanedData, "****")
		} else {

			cleanedData = append(cleanedData, s)
		}

	}
	return strings.Join(cleanedData, " ")
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write([]byte(message))
}

func respondWithJson(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	data, err := json.Marshal(payload)
	if err != nil {
		log.Printf("failed to marshal payload")
		return
	}
	w.Write([]byte(data))
}
