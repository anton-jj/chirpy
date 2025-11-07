package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/anton-jj/chripy/internal/auth"
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
	mux.HandleFunc("POST /api/refresh", apiConfig.handleRefresh)
	mux.HandleFunc("POST /api/revoke", apiConfig.handleRevoke)
	mux.HandleFunc("PUT /api/users", apiConfig.handleUpdateUser)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiConfig.handleGetChirp)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiConfig.handleDeleteChirp)

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

func (aCfg *apiConfig) handleRevoke(w http.ResponseWriter, r *http.Request) {

	auth, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Something wrong with headers")
		return
	}

	token, err := aCfg.db.GetRefreshToken(r.Context(), sql.NullString{Valid: true, String: auth})
	if err != nil {
		respondWithError(w, 401, "Could not find token in database")
		return
	}
	params := database.UpdateRevokedAtParams {
		RevokedAt: sql.NullTime{
		Valid: true, 
		Time: time.Now().UTC() },
		UpdatedAt: time.Now(),
		Token: token.Token,
	}
	err = aCfg.db.UpdateRevokedAt(r.Context(), params)
	if err != nil {
		respondWithError(w, 401, "Failed to update token in database")
		return
	}

	respondWithJson(w, 204, nil)

}
func (aCfg *apiConfig) handleRefresh(w http.ResponseWriter, r *http.Request) {

	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	row, err := aCfg.db.GetRefreshToken(r.Context(), sql.NullString{Valid: true, String: refreshToken})
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	if row.RevokedAt.Valid {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	if row.ExpiresAt.Before(time.Now().UTC()) {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	user, err := aCfg.db.GetUserFromRefreshToken(r.Context(), sql.NullString{Valid: true, String:row.Token.String})
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	
	}

	newJWT, err := auth.MakeJWT(user.ID, aCfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, 500, "failed to create new jwt")
		return
	}


	respondWithJson(w, 200, struct{ Token string `json:"token"`} {Token: newJWT})

}

func validateBody(data string) string {
	badWords := map[string]int{
		"kerfuffle": 0,
		"sharbert":  0,
		"fornax":    0,
	}
	var cleanedData []string
	for _, s := range strings.Split(data, " ") {
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
