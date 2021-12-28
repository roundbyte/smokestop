package main

import (
	"log"
	"github.com/roundbyte/smokestopper/smokerstore"
	"github.com/gorilla/mux"
	"encoding/json"
	"net/http"
	"mime"
)

type smokerServer struct {
	store *smokerstore.SmokerStore
}

func NewSmokerServer() *smokerServer {
	store := smokerstore.New()
	return &smokerServer{store: store}
}

type Smoker struct {
	EmailAddress string `json:"email_addr"`
}

type Response struct {
	Key string `json:"key"`
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

func (ss *smokerServer) addSmokerHandler(w http.ResponseWriter, req *http.Request) {
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
	var smoker Smoker
	if err := dec.Decode(&smoker); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ss.store.AddSmoker(smoker.EmailAddress)
	renderJSON(w, Response{Key: smoker.EmailAddress})
}

func (ss *smokerServer) getAllSmokersHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling getAllSmokers at %s\n", req.URL.Path)

	ss.store.GetSmokers()
	renderJSON(w, ss.store.Smokers)
}

func (ss *smokerServer) deleteSmokerHandler(w http.ResponseWriter, req *http.Request) {
	log.Printf("handling deleteSmoker at %s\n", req.URL.Path)

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
	var smoker Smoker
	if err := dec.Decode(&smoker); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ss.store.DeleteSmoker(smoker.EmailAddress)
	renderJSON(w, Response{Key: smoker.EmailAddress})
}

func main() {
	router := mux.NewRouter()
	//router.StrictSlash(true)
	server := NewSmokerServer()

	router.HandleFunc("/api/smoker/", server.addSmokerHandler).Methods("POST")
	router.HandleFunc("/api/smoker/", server.getAllSmokersHandler).Methods("GET")
	router.HandleFunc("/api/smoker/", server.deleteSmokerHandler).Methods("DELETE")
	log.Fatal(http.ListenAndServe(":3000", router))
}
