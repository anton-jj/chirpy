package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/anton-jj/chripy/internal/auth"
	"github.com/anton-jj/chripy/internal/database"
	"github.com/google/uuid"
)

type parameters struct {
	Password         string `json:"password"`
	Email            string `json:"email"`
}

type userStruct struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
	Token     string    `json:"token"`
	RefreshToken     string    `json:"refresh_token"`

}

func (aCfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var params parameters

	var expiresIn int64 = 3600
	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, 400, "invalid json format")
		return
	}

	user, err := aCfg.db.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, 400, "cant find user")
		return
	}

	if params.Email == "" {
		respondWithError(w, 400, "email cant be empty")
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, 401, "Could not create a token")
		return
	}

	token, err := auth.MakeJWT(user.ID, aCfg.secret, time.Duration(expiresIn)*time.Second)
	if err != nil {
		respondWithError(w, 401, "Could not create a token")
		return
	}

	checked, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil || !checked {
		respondWithError(w, 401, "Incorrect email or password")
		return
	}

	tokenParams := database.CreateRefreshTokenParams {
		Token: sql.NullString{
			Valid: true,
			String: refreshToken,
		},
		UserID: user.ID,
		ExpiresAt: time.Now().AddDate(0,0,60),
		RevokedAt: sql.NullTime{},



	}
	_, err = aCfg.db.CreateRefreshToken(r.Context(), tokenParams)
	if err != nil {
		respondWithError(w, 401, "Failed to create refresh token")
		return
	}


	resp := userStruct{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Email:     user.Email,
		Token:     token,
		RefreshToken: refreshToken,
	}

	respondWithJson(w, 200, resp)

}
func (aCfg *apiConfig) handleUsers(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

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
		Email:          params.Email,
		HashedPassword: hashed,
	}
	user, err := aCfg.db.CreateUser(r.Context(), userParams)
	if err != nil {
		log.Println("creating user", user)
		respondWithError(w, 500, "failed to create a user")
		return
	}
	resp := userStruct{
		ID:        user.ID,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.CreatedAt,
		Email:     user.Email,
	}
	respondWithJson(w, 201, resp)

}
