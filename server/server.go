package server

import (
	"encoding/json"
	"errors"
	"log"
	"mime"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/roundbyte/smokestop/store"
)

type Server struct {
	store       *store.Store
	cookieStore *sessions.CookieStore
}

func New() *Server {
	store := store.New()
	cookieStore := sessions.NewCookieStore([]byte(os.Getenv("SECRETKEY")))
	return &Server{store: store, cookieStore: cookieStore}
}

type Response struct {
	Data string `json:"data"`
	Err  bool   `json:"err"`
}

// User Resistration

type UserRegistration struct {
	EmailAddr string `json:"emailAddr"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

func (server *Server) RegisterUser(w http.ResponseWriter, req *http.Request) {
	if err := parseBodyJSON(w, req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	var ur UserRegistration
	if err := dec.Decode(&ur); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err := server.checkUserRegistration(ur)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	id, err := server.store.RegisterUser(ur.EmailAddr, ur.Username, ur.Password)
	var response Response
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	response = Response{
		Data: id,
		Err:  false,
	}
	renderJSON(w, response)
}

type UserRegistrationErrors struct {
	EmailAddrErr string `json:"emailAddrErr"`
	UsernameErr  string `json:"usernameErr"`
}

// Check new user credentials

func (server *Server) CheckNewUserHandler(w http.ResponseWriter, req *http.Request) {
	if err := parseBodyJSON(w, req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()
	var ur UserRegistration
	if err := dec.Decode(&ur); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ure, _ := server.checkUserRegistration(ur)
	renderJSON(w, ure)
}

func (server *Server) checkUserRegistration(ur UserRegistration) (UserRegistrationErrors, error) {
	ure := UserRegistrationErrors{EmailAddrErr: "", UsernameErr: ""}
	var err error = nil
	if err := server.store.GetAllUsers(); err != nil {
		log.Printf("Unable to store.GetAllUsers %s\n", err.Error())
	}
	for _, user := range server.store.Users {
		if user.EmailAddr == ur.EmailAddr {
			ure.EmailAddrErr = "Email address has already in use 👻"
			err = errors.New("Email address already in use 🚫")
		}
		if user.Username == ur.Username {
			ure.UsernameErr = "Username has already in use 👻"
			err = errors.New("Username has already in use 🚫")
		}
	}
	return ure, err
}

func (server *Server) GetAllUsers(w http.ResponseWriter, req *http.Request) {
	if err := server.store.GetAllUsers(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	renderJSON(w, server.store.Users)
}

type UserLogin struct {
	EmailAddr string `json:"emailAddr"`
	Password  string `json:"password"`
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
	if err := server.store.GetAllUsers(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	var userId string = ""
	var response Response
	for key, user := range server.store.Users {
		log.Println(key)
		if user.EmailAddr == ul.EmailAddr {
			userId = key
		}
	}
	if userId == "" {
		response = Response{Data: "Non existent user", Err: true}
		renderJSON(w, response)
		return
	}

	err := server.store.DoesPasswordMatch(userId, ul.Password)
	if err != nil {
		response = Response{Data: err.Error(), Err: true}
	} else {
		session, _ := server.cookieStore.Get(req, "session-name")
		session.Values["userId"] = userId
		session.Values["authenticated"] = true
		session.Save(req, w)
		response = Response{Data: "Logged in, set cookie", Err: false}
	}
	renderJSON(w, response)
}

func (server *Server) LogoutUserHandler(w http.ResponseWriter, req *http.Request) {
	session, _ := server.cookieStore.Get(req, "session-name")
	session.Options.MaxAge = -1
	session.Save(req, w)
	response := Response{Data: "Logged out", Err: false}
	renderJSON(w, response)
}

func (server *Server) SecretHandler(w http.ResponseWriter, req *http.Request) {
	server.store.GetAllUsers()
	session, err := server.cookieStore.Get(req, "session-name")
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}
	log.Printf("id: %s, authenticated: %d", session.Values["userId"], session.Values["authenticated"])
	renderJSON(w, server.store.Users)
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
		return errors.New("StatusUnsupportedMediaType")
	}
	return nil
}
