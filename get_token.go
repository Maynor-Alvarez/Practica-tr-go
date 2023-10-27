package main

import (
	"github.com/golang-jwt/jwt/v5"
	"log"
	"net/http"
	"strings"
	"time"
)

var secret = []byte("tribal_secret")

func generateTokenHandler(w http.ResponseWriter, r *http.Request) {

	user := r.URL.Query().Get("user")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  1,
		"username": user,
		"exp":      time.Now().Add(time.Hour * 1).Unix(),
		"iat":      time.Now().Unix(),
		"nbf":      time.Now().Unix(),
	})

	tokenString, err := token.SignedString(secret)
	if err != nil {
		log.Printf("Error al generar token %v", err)
		http.Error(w, "Error al generar token", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(tokenString))

}

func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Acceso no autorizado", http.StatusUnauthorized)
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return secret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Acceso no autorizado", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
