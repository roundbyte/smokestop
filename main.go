package main

import (
	"encoding/json"
	"log"
	"mime"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/roundbyte/smokestopper/store"
)

type smokeStopServer struct {
	store *store.Store
}

func NewSmokerServer() *smokeStopServer {
	store := store.New()
	return &smokeStopServer{store: store}
}

type UserRegistration struct {
	EmailAddr string `json:"emailAddr"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

type UserRegistrationErrors struct {
	EmailAddr string `json:"emailAddr"`
	Username  string `json:"username"`
}

type UserLogin struct {
	EmailAddr string `json:"emailAddr"`
	Password  string `json:"password"`
}

type Response struct {
	Key   string `json:"key"`
	Error bool   `json:"error"`
}

func renderJSON(w http.ResponseWriter, v interface{}) {
	js, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(js)
}

func (server *smokeStopServer) addUserHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling addSmoker at %s\n", req.URL.Path)

	contentType := req.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if mediatype != "application/json" {
		http.Error(w, "expect application/json Content-Type", http.StatusUnsupportedMediaType)
		return
	}
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	var userRegistration UserRegistration
	if err := dec.Decode(&userRegistration); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	server.store.AddUser(userRegistration.EmailAddr, userRegistration.Username, userRegistration.Password)
	renderJSON(w, Response{Key: userRegistration.EmailAddr, Error: false})
}

func (server *smokeStopServer) checkNewUser(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling checkNewUser at %s\n", req.URL.Path)

	contentType := req.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if mediatype != "application/json" {
		http.Error(w, "expect application/json Content-Type", http.StatusUnsupportedMediaType)
		return
	}
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	var userRegistration UserRegistration
	if err := dec.Decode(&userRegistration); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userRegistrationErrors := UserRegistrationErrors{EmailAddr: "", Username: ""}
	server.store.GetAllUsers()
	if value, exists := server.store.Users[userRegistration.EmailAddr]; exists {
		log.Printf("%s exists already for %s\n", userRegistration.EmailAddr, value.Username)
		userRegistrationErrors.EmailAddr = "Email address already exists 👻"
	}
	// key, user
	for _, user := range server.store.Users {
		if user.Username == userRegistration.Username {
			userRegistrationErrors.Username = "Username has already been taken 👻"
		}
	}
	renderJSON(w, userRegistrationErrors)
}

func (server *smokeStopServer) getAllUsersHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling getAllSmokers at %s\n", req.URL.Path)
	server.store.GetAllUsers()
	renderJSON(w, server.store.Users)
}

func (server *smokeStopServer) loginUserHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling loginUser at %s\n", req.URL.Path)

	contentType := req.Header.Get("Content-Type")
	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if mediatype != "application/json" {
		http.Error(w, "expect application/json Content-Type", http.StatusUnsupportedMediaType)
		return
	}
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	var userLogin UserLogin
	if err := dec.Decode(&userLogin); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	passwordsMatch := server.store.CheckUserPassword(userLogin.EmailAddr, userLogin.Password)
	renderJSON(w, Response{Key: userLogin.EmailAddr, Error: !passwordsMatch})
}

func main() {
	router := mux.NewRouter()
	//router.StrictSlash(true)
	server := NewSmokerServer()

	router.HandleFunc("/api/register/", server.addUserHandler).Methods("POST")
	router.HandleFunc("/api/checknewuser/", server.checkNewUser).Methods("GET")
	router.HandleFunc("/api/user/", server.getAllUsersHandler).Methods("GET")
	router.HandleFunc("/api/login/", server.loginUserHandler).Methods("GET")
	log.Fatal(http.ListenAndServe(":3000", router))
}
