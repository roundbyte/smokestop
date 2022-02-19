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

// Step 1. User Registration

func (server *Server) RegisterUser(w http.ResponseWriter, req *http.Request) {
	// Decode the JSON body
	if err := parseBodyJSON(w, req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()

	// Decode to the User Registration Form to a go struct
	var userRegistrationForm store.UserRegistrationForm
	if err := dec.Decode(&userRegistrationForm); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update the store
	if err := server.updateStore(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Check the email address
	if err := server.checkEmailAddressAvailable(userRegistrationForm.EmailAddr); err != nil {
		response := Response{Data: err.Error(), Err: true}
		renderJSON(w, response)
		return
	}

	// Check the username
	if err := server.checkUsernameAvailable(userRegistrationForm.Username); err != nil {
		response := Response{Data: err.Error(), Err: true}
		renderJSON(w, response)
		return
	}

	// Check the password
	if err := server.checkPassword(userRegistrationForm.Password); err != nil {
		response := Response{Data: err.Error(), Err: true}
		renderJSON(w, response)
		return
	}

	id, err := server.store.RegisterUser(userRegistrationForm)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	response := Response{
		Data: id,
		Err:  false,
	}
	renderJSON(w, response)
}

func (server *Server) checkEmailAddressAvailable(emailAddr string) error {
	for _, user := range server.store.Users {
		if user.EmailAddr == emailAddr {
			if user.Active {
				return errors.New("errEmailAddressAlreadyActive")
			} else {
				return errors.New("errEmailAddressNotConfirmed")
			}
		}
	}
	return nil
}

func (server *Server) checkUsernameAvailable(username string) error {
	for _, user := range server.store.Users {
		if user.Username == username {
			return errors.New("errUsernameExists")
		}
	}
	return nil
}

func (server *Server) checkPassword(password string) error {
	log.Printf("Password is: %s\n", password)
	return nil
}

func (server *Server) updateStore() error {
	if err := server.store.GetAllUsers(); err != nil {
		return err
	}
	return nil
}

func (server *Server) GetUsers(w http.ResponseWriter, req *http.Request) {
	if err := server.updateStore(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	renderJSON(w, server.store.Users)
}

type UserLoginForm struct {
	UsernameOrEmailAddr string `json:"usernameOrEmailAddr"`
	Password            string `json:"password"`
}

func (server *Server) LoginUserHandler(w http.ResponseWriter, req *http.Request) {
	// Decode the JSON body
	if err := parseBodyJSON(w, req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()

	// Decode to the User Login Form to a go struct
	var userLoginForm UserLoginForm
	if err := dec.Decode(&userLoginForm); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update the store
	if err := server.updateStore(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Find user id if the email address or username exist
	userId, err := server.findUserIdByUsernameOrEmailAddress(userLoginForm.UsernameOrEmailAddr)
	if err != nil {
		response := Response{Data: err.Error(), Err: true}
		renderJSON(w, response)
		return
	}

	// Check the password
	if err = server.store.CheckPassword(userId, userLoginForm.Password); err != nil {
		response := Response{Data: err.Error(), Err: true}
		renderJSON(w, response)
		return
	}

	// Set the authentication cookie
	session, _ := server.cookieStore.Get(req, "authentication")
	session.Values["userId"] = userId
	session.Values["authenticated"] = true
	session.Save(req, w)

	// Send the response
	response := Response{Data: "logInSuccessfulCookieSet", Err: false}
	renderJSON(w, response)
}

func (server *Server) findUserIdByUsernameOrEmailAddress(usernameOrEmailAddr string) (string, error) {
	for key, user := range server.store.Users {
		if user.Username == usernameOrEmailAddr || user.EmailAddr == usernameOrEmailAddr {
			if user.Active == false {
				return "", errors.New("errEmailAddressNotConfirmed")
			}
			return key, nil
		}
	}
	return "", errors.New("errUsernameOrEmailAddressNotFound")
}

func (server *Server) LogoutUserHandler(w http.ResponseWriter, req *http.Request) {
	session, _ := server.cookieStore.Get(req, "auth-session")
	session.Options.MaxAge = -1
	session.Save(req, w)
	response := Response{Data: "logOutSuccessful", Err: false}
	renderJSON(w, response)
}

type UserVerifyForm struct {
	EmailAddr string `json:"emailAddr"`
	Code      string `json:"code"`
}

func (server *Server) VerifyUserHandler(w http.ResponseWriter, req *http.Request) {
	// Decode the JSON body
	if err := parseBodyJSON(w, req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	dec := json.NewDecoder(req.Body)
	dec.DisallowUnknownFields()

	// Decode to the User Registration Form to a go struct
	var userVerifyForm UserVerifyForm
	if err := dec.Decode(&userVerifyForm); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update the store
	if err := server.updateStore(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Find user id if the email address or username exist
	userId, err := server.isEmailAddressUnverified(userVerifyForm.EmailAddr)
	if err != nil {
		response := Response{Data: err.Error(), Err: true}
		renderJSON(w, response)
		return
	}

	// Verify the user
	if err = server.store.VerifyUser(userId, userVerifyForm.Code); err != nil {
		response := Response{Data: err.Error(), Err: true}
		renderJSON(w, response)
		return
	}

	// Send the response
	response := Response{Data: "verificationSuccessful", Err: false}
	renderJSON(w, response)
}

func (server *Server) isEmailAddressUnverified(emailAddr string) (string, error) {
	for key, user := range server.store.Users {
		if user.EmailAddr == emailAddr {
			if user.Active == true {
				return "", errors.New("errEmailAddressAlreadyActive")
			}
			return key, nil
		}
	}
	return "", errors.New("errEmailAddressNotFound")
}

func (server *Server) SecretHandler(w http.ResponseWriter, req *http.Request) {
	server.store.GetAllUsers()
	session, err := server.cookieStore.Get(req, "auth-session")
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
