package server

import (
	"encoding/json"
	"errors"
	"log"
	"mime"
	"net/http"

	"github.com/roundbyte/smokestop/store"
)

type Server struct {
	store *store.Store
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

func NewServer() *Server {
	store := store.New()
	return &Server{store: store}
}

func (server *Server) RegisterUserHandler(w http.ResponseWriter, req *http.Request) {
	if err := parseBodyJSON(w, req); err != nil {
		return
	}
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	var ur UserRegistration
	if err := dec.Decode(&ur); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	server.store.AddUser(ur.EmailAddr, ur.Username, ur.Password)
	response := Response{Key: ur.EmailAddr, Error: false}
	renderJSON(w, response)
}

func (server *Server) CheckNewUserHandler(w http.ResponseWriter, req *http.Request) {
	if err := parseBodyJSON(w, req); err != nil {
		return
	}
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	var ur UserRegistration
	if err := dec.Decode(&ur); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ure := UserRegistrationErrors{EmailAddr: "", Username: ""}
	server.store.GetAllUsers()
	if value, exists := server.store.Users[ur.EmailAddr]; exists {
		log.Printf("%s exists already for %s\n", ur.EmailAddr, value.Username)
		ure.EmailAddr = "Email address already exists 👻"
	}
	for _, user := range server.store.Users {
		if user.Username == ur.Username {
			ure.Username = "Username has already been taken 👻"
		}
	}
	renderJSON(w, ure)
}

func (server *Server) GetAllUsersHandler(w http.ResponseWriter, req *http.Request) {
	server.store.GetAllUsers()
	renderJSON(w, server.store.Users)
}

func (server *Server) LoginUserHandler(w http.ResponseWriter, req *http.Request) {
	if err := parseBodyJSON(w, req); err != nil {
		return
	}
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()

	var ul UserLogin
	if err := dec.Decode(&ul); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	passwordsMatch := server.store.CheckUserPassword(ul.EmailAddr, ul.Password)
	response := Response{Key: ul.EmailAddr, Error: !passwordsMatch}
	renderJSON(w, response)
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

func parseBodyJSON(w http.ResponseWriter, req *http.Request) error {
	contentType := req.Header.Get("Content-Type")
	if mediatype, _, err := mime.ParseMediaType(contentType); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return errors.New("StatusBadRequest")
	} else if mediatype != "application/json" {
		http.Error(w, "expect application/json Content-Type", http.StatusUnsupportedMediaType)
		return errors.New("expect application/json Content-Type")
	}
	return nil
}
