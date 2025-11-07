package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/anton-jj/chripy/internal/auth"
	"github.com/anton-jj/chripy/internal/database"
	"github.com/google/uuid"
)

type parameters struct {
	Password string `json:"password"`
	Email    string `json:"email"`
}

type userStruct struct {
	ID           uuid.UUID `json:"id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Email        string    `json:"email"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
}

func (aCfg *apiConfig) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var params parameters

	if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
		respondWithError(w, 400, "invalid json format")
		return
	}

	tok, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "header missing or malformed")
	}

	if strings.Count(tok, ".") != 2 {
		respondWithError(w, 401, "malformed token")
		return
	}

	_, err = auth.ValidateJWT(tok, aCfg.secret)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}


	hashedPass, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, 401, "Unauthorized")
		return
	}
	updateUserParams := database.UpdateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPass,
	}
	err = aCfg.db.UpdateUser(r.Context(), updateUserParams)
	if err != nil {
		respondWithError(w, 500, "Failed to update database")
		return
	}
	
	respondWithJson(w, 200, struct{Email string `json:"email"`}{Email: params.Email})

}
func (aCfg *apiConfig) handleLogin(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	var params parameters

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

	checked, err := auth.CheckPasswordHash(params.Password, user.HashedPassword)
	if err != nil || !checked {
		respondWithError(w, 401, "Incorrect email or password")
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, 401, "Could not create a token")
		return
	}
	log.Println(refreshToken)

	token, err := auth.MakeJWT(user.ID, aCfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, 401, "Could not create a token")
		return
	}
	log.Println(token)

	tokenParams := database.CreateRefreshTokenParams{
		Token: sql.NullString{
			Valid:  true,
			String: refreshToken,
		},
		UserID:    user.ID,
		ExpiresAt: time.Now().AddDate(0, 0, 60),
		RevokedAt: sql.NullTime{},
	}
	_, err = aCfg.db.CreateRefreshToken(r.Context(), tokenParams)
	if err != nil {
		respondWithError(w, 401, "Failed to create refresh token")
		return
	}

	resp := userStruct{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.UpdatedAt,
		Email:        user.Email,
		Token:        token,
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

	token, err := auth.MakeJWT(user.ID, aCfg.secret, time.Hour)
	if err != nil {
		respondWithError(w, 401, "Could not create a token")
		return
	}
	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, 401, "Failed to create refresh token")
		return
	}
	resp := userStruct{
		ID:           user.ID,
		CreatedAt:    user.CreatedAt,
		UpdatedAt:    user.CreatedAt,
		Email:        user.Email,
		Token:        token,
		RefreshToken: refreshToken,
	}
	respondWithJson(w, 201, resp)

}
