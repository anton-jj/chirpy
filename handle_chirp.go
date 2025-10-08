package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

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

func (aCfg *apiConfig) handleGetChrip(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("chirpID"))
	fmt.Println(id)
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
		User_id:   chirp.UserID.UUID,
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
			User_id:   chirp.UserID.UUID,
		})
	}
	respondWithJson(w, 200, jsonChirps)

}
func (aCfg *apiConfig) handleChirpCreate(w http.ResponseWriter, r *http.Request) {

	type parameters struct {
		Body    string    `json:"body"`
		User_id uuid.UUID `json:""`
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

	dbParams := database.CreateChirpParams{
		Body: cleanedBody,
		UserID: uuid.NullUUID{
			Valid: true,
			UUID:  params.User_id,
		},
	}
	chirp, err := aCfg.db.CreateChirp(r.Context(), dbParams)
	if err != nil {
		respondWithError(w, 500, "database failed to create chirp")
	}

	resp := Chirp{
		Id:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		User_id:   chirp.UserID.UUID,
	}
	respondWithJson(w, 201, resp)

}
