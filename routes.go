package main

import (
	"github.com/gorilla/mux"
	"net/http"
)

func main() {

	protect := mux.NewRouter()
	protect.HandleFunc("/songs", getAll).Methods("GET")

	http.Handle("/songs", authMiddleware(protect))
	http.HandleFunc("/token", generateTokenHandler)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}
}
