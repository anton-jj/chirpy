package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/anton-jj/chripy/internal/auth"
	"github.com/anton-jj/chripy/internal/database"
	"github.com/google/uuid"
)

func (aCfg *apiConfig) handleUsers(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()
	type parameters struct {
		Password string `json:"password"`
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
	hashed, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, 500, "error hashing password")
		return 
	}
	userParams := database.CreateUserParams{
		Email: params.Email,	
		HashedPassword: hashed,
	}
	user, err := aCfg.db.CreateUser(r.Context(), userParams)
	if err != nil {
		log.Println("creating user", user)
		log.Println(err)
		respondWithError(w, 500, "failed to create a user")
		return
	}
	type userStruct struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}
	resp := userStruct{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.CreatedAt,
		Email:     user.Email,
	}
	respondWithJson(w, 201, resp)

}
