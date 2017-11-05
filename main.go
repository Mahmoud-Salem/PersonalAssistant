
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"github.com/gorilla/mux"
	"os"
	cors "github.com/heppu/simple-cors"
	
)

func routes(w http.ResponseWriter, req *http.Request) {
	auth := req.Header.Get("Authorization")
	if auth == "" || req.Header["Authorization"] == nil {
		w.WriteHeader(http.StatusUnauthorized)
		e := map[string]string{"message":"You have to include your id ."}		
		json.NewEncoder(w).Encode(e)
		return
	}
	if req.Body == nil {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message":"Ask me Something ."}		
		json.NewEncoder(w).Encode(e)
		return
	}

	req.ParseForm()
	if req.Form.Get("message") == "" {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message":"The message can't be empty."}		
		json.NewEncoder(w).Encode(e)
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
		e := map[string]string{"message":"You have to specify what service you want calendar , keys or memos ."}		
		json.NewEncoder(w).Encode(e)
		return
	}

}

func Welcome(w http.ResponseWriter, req *http.Request) {

	w.WriteHeader(200)
//	json.NewEncoder(w).Encode(`{ "message" : "Welcome to our personal assistant to register fill in your email , name and password in /register request"}`)

//json.NewEncoder(w).Encode({"message":"Welcome to our personal assistant to register fill in your email , name and password in /register request"}) 
d := map[string]string{"message":"Welcome to our personal assistant to register fill in your email , name and password in /register request"}
json.NewEncoder(w).Encode(d)
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

	log.Printf("Listening on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, cors.CORS(router)))

}
