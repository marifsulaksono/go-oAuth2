package main

import (
	"g-oAuth2/controller"
	"g-oAuth2/domain"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

const PORT = ":8080"

func main() {
	// load .env files
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error load .env files : %v", err)
	}

	domain.InitGoogleConfig()

	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Welcome, please login first"))
	}).Methods(http.MethodGet)

	r.HandleFunc("/auth/google-auth", controller.LoginGoogle).Methods(http.MethodGet)
	r.HandleFunc("/auth/google-callback", controller.CallbackGoogle).Methods(http.MethodGet)

	log.Printf("Server start at %s", PORT)
	err := http.ListenAndServe(PORT, r)
	if err != nil {
		log.Fatalf("Error start server : %v", err)
	}
}
