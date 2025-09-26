package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"

	"github.com/anton-jj/chripy/internal/database"
	_ "github.com/lib/pq"
)

type responeError struct {
	Error string `json:"error"`
}

type apiConfig struct {
	fileServerHits atomic.Int32
	db             *database.Queries
}

type cleanedData struct {
	Cleaned_body string `json:"cleaned_body"`
}

func (af *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		af.fileServerHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (aCfg *apiConfig) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	bodyString := fmt.Sprintf("<html><body><h1>Welcome, Chirpy Admin</h1><p>Chirpy has been visited %d times!</p></body></html>", aCfg.fileServerHits.Load())
	w.Write([]byte(bodyString))
}

func (aCfg *apiConfig) handleReset(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	aCfg.fileServerHits.Store(0)
	w.Write([]byte("hits reset to 0"))
}

func main() {
	dbURL := os.Getenv("DB_URL")
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
	}

	mux := http.NewServeMux()
	fsHandler := apiConfig.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filePathRoot))))
	mux.Handle("/app/", fsHandler)

	mux.HandleFunc("GET /api/healthz", handleHealtz)
	mux.HandleFunc("GET /admin/metrics", apiConfig.handleMetrics)
	mux.HandleFunc("POST /admin/reset", apiConfig.handleReset)
	mux.HandleFunc("POST /api/validate_chirp", handleValidate)
	mux.HandleFunc("POST /api/users", apiConfig.handleUsers)

	server := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	log.Printf("Serving files from %s to port: %s", filePathRoot, server.Addr)
	server.ListenAndServe().Error()
}
func (aCfg *apiConfig) handleUsers(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()
	type parameters struct {
		Email string `json:"email"`
	}

	var params parameters

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, 400, "ivanlid json format")
		return
	}

	if params.Email == "" {
		respondWithError(w, 400, "email cant be empty")
	}
	fmt.Println(params.Email)
	user, err := aCfg.db.CreateUser(r.Context(), params.Email)
	if err != nil {
		log.Println("creating user", user)
		log.Println(err)
		respondWithError(w, 500, "failed to create a user")
		return
	}
	respondWithJson(w, 201, user)

}
func handleHealtz(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func handleValidate(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()
	type parameters struct {
		Body string `json:"body"`
	}

	var params parameters
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, 500, "invalid json body")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "chirp to long")
		return
	}

	var cleanedBody string = removeProfane(params.Body)

	respondWithJson(w, 200, cleanedData{Cleaned_body: cleanedBody})
}

func removeProfane(data string) string {
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
