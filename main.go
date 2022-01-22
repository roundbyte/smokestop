package main

import (
	"fmt"
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

	router.HandleFunc("/api/register/", server.RegisterUserHandler).Methods("POST")
	router.HandleFunc("/api/checknewuser/", server.CheckNewUserHandler).Methods("POST")
	router.HandleFunc("/api/user/", server.GetAllUsersHandler).Methods("GET")
	router.HandleFunc("/api/login/", server.LoginUserHandler).Methods("POST")
	router.HandleFunc("/api/secret/", server.SecretHandler).Methods("GET")
	router.HandleFunc("/api/logout/", server.LogoutUserHandler).Methods("POST")

	handler := middleware.Logging(router)

	portString := fmt.Sprintf(":%s", os.Getenv("SERVERPORT"))
	log.Fatal(http.ListenAndServe(portString, handler))
}
