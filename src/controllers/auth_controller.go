package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/LeonardoMuller13/digital-bank-api/src/config"
	"github.com/LeonardoMuller13/digital-bank-api/src/database"
	"github.com/LeonardoMuller13/digital-bank-api/src/models"
	"github.com/bitly/go-simplejson"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Credentials struct {
	Password string `json:"password"`
	Cpf      string `json:"cpf"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	creds := &Credentials{}
	err := json.NewDecoder(r.Body).Decode(creds)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Internal error.")
		return
	}

	account := &models.Account{}
	result := database.DB.First(&account, "cpf = ?", creds.Cpf)
	if result.Error != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Username not found.")
		return
	}
	if err = bcrypt.CompareHashAndPassword([]byte(account.Secret), []byte(creds.Password)); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Not authorized.")
		return
	}
	tokenString, err := createToken(account.ID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal error.")
	}
	w.WriteHeader(http.StatusOK)

	response := simplejson.New()
	response.Set("token", tokenString)

	payload, err := response.MarshalJSON()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprint(w, "Internal error.")
	}
	w.Write(payload)
	return
}

func createToken(user uint) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"user": user,
			"exp":  time.Now().Add(time.Hour * 24).Unix(),
		})

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load configurations: %v", err)
	}
	jwtKey := []byte(cfg.JwtSecretKey)

	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
