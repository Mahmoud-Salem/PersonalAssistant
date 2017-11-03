
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"github.com/gorilla/mux"
	"os"
)

func routes(w http.ResponseWriter, req *http.Request) {
	auth := req.Header.Get("Authorization")
	if auth == "" || req.Header["Authorization"] == nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(`{ "message" : "You have to include your id ."}`)
		return
	}
	if req.Body == nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(`{ "message" : "Ask me Something ."}`)
		return
	}

	req.ParseForm()
	if req.Form.Get("message") == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(`{ "message" : "The message can't be empty."}`)
		return
	}
	body := req.Form.Get("message")

	if strings.Contains(body, "calendar") {
		HandleCalendar(w, req, body)
	} else if strings.Contains(body, "memo") {
		HandleMemo(w, req, body)
	} else if strings.Contains(body, "key") || strings.Contains(body, "keys") {
		HandleKey(w, req, body)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(`{ "message" : "You have to specify what service you want calendar , keys or memos ."}`)
		return
	}

}

func Welcome(w http.ResponseWriter, req *http.Request) {

	w.WriteHeader(200)
	json.NewEncoder(w).Encode(`{ "message" : "Welcome to our personal assistant to register fill in your email , name and password in /register request"}`)

}
func main() {

	//router setup
	router := mux.NewRouter()

	// welcoming 
	router.HandleFunc("/welcome", Welcome).Methods("GET")
	
	// Register and login
	router.HandleFunc("/register", Register).Methods("POST")
	router.HandleFunc("/login", Login).Methods("POST")

	// Services
	router.HandleFunc("/chat", routes).Methods("POST")
	port := os.Getenv("PORT") 
	if port == "" {
		port = "8000"
	}
	log.Fatal(http.ListenAndServe(":"+port, router))

}
