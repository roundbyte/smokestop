package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/roundbyte/smokestop/middleware"
	"github.com/roundbyte/smokestop/server"
)

func main() {
	godotenv.Load()
	router := mux.NewRouter()
	server := server.New()

	router.HandleFunc("/api/users/", server.GetAllUsers).Methods("GET")
	router.HandleFunc("/api/users/register/", server.RegisterUser).Methods("POST")
	router.HandleFunc("/api/users/login/", server.LoginUserHandler).Methods("POST")
	router.HandleFunc("/api/users/logout/", server.LogoutUserHandler).Methods("POST")
	router.HandleFunc("/api/users/check/", server.CheckNewUserHandler).Methods("POST")
	router.HandleFunc("/api/secret/", server.SecretHandler).Methods("GET")

	handler := middleware.Logging(router)

	log.Fatal(http.ListenAndServe(":"+os.Getenv("SERVERPORT"), handler))
}
