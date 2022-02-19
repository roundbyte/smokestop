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

	router.HandleFunc("/api/users/", server.GetUsers).Methods("GET")
	router.HandleFunc("/api/register/", server.RegisterUser).Methods("POST")
	router.HandleFunc("/api/login/", server.LoginUserHandler).Methods("POST")
	router.HandleFunc("/api/logout/", server.LogoutUserHandler).Methods("POST")
	router.HandleFunc("/api/verify/", server.VerifyUserHandler).Methods("POST")

	handler := middleware.Logging(router)

	log.Fatal(http.ListenAndServe(":"+os.Getenv("SERVERPORT"), handler))
}
