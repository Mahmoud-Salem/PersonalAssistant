
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"github.com/gorilla/mux"
	"os"
	cors "github.com/heppu/simple-cors"
	"crypto/md5"
	"strconv"
	"time"
	"encoding/hex"
	"fmt"
	
	
)
type (

	// JSON Holds a JSON object
	JSON map[string]interface{}

)


func routes(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")	


	if req.Body == nil {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message":"Ask me Something ."}		
		json.NewEncoder(w).Encode(e)
		return
	}

	data := JSON{}
	if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
		http.Error(w, fmt.Sprintf("Couldn't decode JSON: %v.", err), http.StatusBadRequest)
		return
	}
	defer req.Body.Close()


	_, messageFound := data["message"]
	if !messageFound {
		http.Error(w, "Missing message key in body.", http.StatusBadRequest)
		return
	}
	body := data["message"].(string)



	if strings.Contains(body, "login") {
		Login(w, req, body)
	} else if strings.Contains(body, "register") {
		Register(w, req, body)
	} else {
	
	auth := req.Header.Get("Authorization")
	if auth == "" || req.Header["Authorization"] == nil {
		w.WriteHeader(http.StatusUnauthorized)
		e := map[string]string{"message":"You have to include your id ."}		
		json.NewEncoder(w).Encode(e)
		return
	}

	if strings.Contains(body, "calendar") {
		HandleCalendar(w, req, body)
	} else if strings.Contains(body, "memo") {
		HandleMemo(w, req, body)
	} else if strings.Contains(body, "key") || strings.Contains(body, "keys") {
		HandleKey(w, req, body)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message":" if you are logged in . You have to specify what service you want calendar , keys or memos . else register or login"}		
		json.NewEncoder(w).Encode(e)
		return
	}
}
}

func handle(w http.ResponseWriter, r *http.Request) {
	body :=
		"<!DOCTYPE html><html><head><title>Chatbot</title></head><body><pre style=\"font-family: monospace;\">\n" +
			"Available Routes:\n\n" +
			"  GET  /welcome -> Welcome\n" +
			"  POST /chat    -> routes\n" +
			"  GET  /        -> handle        (current)\n" +
			"</pre></body></html>"
	w.Header().Add("Content-Type", "text/html")
	fmt.Fprintln(w, body)
	
}

func Welcome(w http.ResponseWriter, req *http.Request) {
		// Generate a UUID.
	hasher := md5.New()
	hasher.Write([]byte(strconv.FormatInt(time.Now().Unix(), 10)))
	uuid := hex.EncodeToString(hasher.Sum(nil))
	w.Header().Set("Content-Type", "application/json")
	
	w.WriteHeader(200)
	d := map[string]string{"message":"Welcome to our personal assistant to register fill in your email , name and password in /register request" , "uuid" : uuid}
	json.NewEncoder(w).Encode(d)
}
func main() {

	//router setup2
	router := mux.NewRouter()

	// welcoming 
	router.HandleFunc("/welcome", Welcome).Methods("GET")
	router.HandleFunc("/", handle).Methods("GET")
	// Register and login
//	router.HandleFunc("/register", Register).Methods("POST")
//	router.HandleFunc("/login", Login).Methods("POST")

	// Services
	router.HandleFunc("/chat", routes).Methods("POST")
	port := os.Getenv("PORT") 
	if port == "" {
		port = "8000"
	}

	log.Printf("Listening on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, cors.CORS(router)))

}
