package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/anton-jj/chripy/internal/auth"
	"github.com/anton-jj/chripy/internal/database"

	"github.com/google/uuid"
)

type Chirp struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Body      string    `json:"body"`
	User_id   uuid.UUID `json:"user_id"`
}

func (aCfg *apiConfig) handleDeleteChirp(w http.ResponseWriter, r *http.Request) {
	tok, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "header missing or malformed")
		return
	}
	if strings.Count(tok, ".") != 2 {
		respondWithError(w, 401, "header malformed")
		return
	}

	userId, err := auth.ValidateJWT(tok, aCfg.secret)
	if err != nil {
		respondWithError(w, 403, "Unauthorized")
		return
	}

	id, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 500, "Failed to get pathvariable")
		return
	}

	chirp, err := aCfg.db.GetChirpById(r.Context(), id)
	if err != nil {
		respondWithError(w, 403, "Unauthorized")
		return
	}

	if chirp.UserID != userId {
		respondWithError(w, 403, "Unauthorized")
		return
	}
	err = aCfg.db.DeleteChirpById(r.Context(), id)
	if err != nil {
		respondWithError(w, 404, "Chirp is not found")
		return
	}
	respondWithJson(w, 204, nil)

}
func (aCfg *apiConfig) handleGetChirp(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("chirpID"))
	if err != nil {
		respondWithError(w, 500, "error while parsing the pathvariable")
		return
	}
	chirp, err := aCfg.db.GetChirpById(r.Context(), id)
	if err != nil {
		respondWithError(w, 404, "chirp not found")
		return
	}

	resp := Chirp{
		Id:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		User_id:   chirp.UserID,
	}

	respondWithJson(w, 200, resp)
}
func (aCfg *apiConfig) handleChirpsGetAll(w http.ResponseWriter, r *http.Request) {
	chirps, err := aCfg.db.GetAllChirps(r.Context())
	if err != nil {
		respondWithError(w, 500, "failed to get chirps")
	}

	var jsonChirps []Chirp
	for _, chirp := range chirps {
		jsonChirps = append(jsonChirps, Chirp{
			Id:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			User_id:   chirp.UserID,
		})
	}
	respondWithJson(w, 200, jsonChirps)

}
func (aCfg *apiConfig) handleChirpCreate(w http.ResponseWriter, r *http.Request) {

	type parameters struct {
		Body string `json:"body"`
	}

	var params parameters

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, 400, "invalid json format")
		return
	}

	if len(params.Body) > 140 {
		respondWithError(w, 400, "chirp to long")
		return
	}

	var cleanedBody string = validateBody(params.Body)

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	if strings.Count(token, ".") != 2 {
		respondWithError(w, 401, "Unauthorized")
		return

	}
	validToken, err := auth.ValidateJWT(token, aCfg.secret)
	log.Println(validToken)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}

	dbParams := database.CreateChirpParams{
		Body:   cleanedBody,
		UserID: validToken,
	}
	log.Println(dbParams)
	chirp, err := aCfg.db.CreateChirp(r.Context(), dbParams)

	if err != nil {
		respondWithError(w, 500, "database failed to create chirp")
		return
	}

	resp := Chirp{
		Id:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		User_id:   chirp.UserID,
	}
	respondWithJson(w, 201, resp)

}
