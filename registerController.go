package main 

import(
    "encoding/json"
    "net/http"
    "github.com/night-codes/mgo-ai"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "log"
    "golang.org/x/crypto/bcrypt"
    "strings"    
)

 

func Register (w http.ResponseWriter , req *http.Request){
   
   // Validity checking    
   if req.Body == nil {
    w.WriteHeader(http.StatusBadRequest)
    json.NewEncoder(w).Encode(`{ "message" : "provide your name , email and password in the following format to register . name : mahmoud . email : ms@gmail.com . password :123456 . "}`)
    return
}
    req.ParseForm()
    if req.Form.Get("message") == "" {
		w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(`{ "message" : "provide your name , email and password in the following format to register . name : mahmoud . email : ms@gmail.com . password :123456 . "}`)
		return
    }
    body := req.Form.Get("message")

    tokens := strings.Split(body, ".")
    
    name := "" 
    email := "" 
    password := ""     
    for i:= 0 ; i<len(tokens) ; i++ {
        req := strings.Split(tokens[i], ":")

        if strings.TrimSpace(req[0]) == "name" {
            name = strings.TrimSpace(req[1])
        }
        if strings.TrimSpace(req[0]) == "email" {
            email = strings.TrimSpace(req[1])
        }
        if strings.TrimSpace(req[0]) == "password" {
            password = strings.TrimSpace(req[1])
        }
    }
    if !(name != "" && email != "" && password != "") {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(`{ "message" : "provide your name , email and password in the following format to register . name : mahmoud . email : ms@gmail.com . password :123456 . "}`)
		return
    }
            // database configuration 
session, err := mgo.Dial("localhost")   
        ai.Connect(session.DB("test").C("counters"))
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
                panic(err)
        }
        defer session.Close()
        session.SetMode(mgo.Monotonic, true)
                        // Registeration
users := session.DB("test").C("users")
        currentUsers := []User{User{}}
        err = users.Find(bson.M{"email": email}).All(&currentUsers)
        if(err != nil){
          w.WriteHeader(http.StatusInternalServerError)  
          json.NewEncoder(w).Encode("Error with the Database Connection")
          return
        }
        if(len(currentUsers)>0){
        w.WriteHeader(http.StatusForbidden)
          json.NewEncoder(w).Encode("This Email Already Exists")
          return
        }
        // encrypt the coming pasword 
        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
        if(err != nil){
            w.WriteHeader(http.StatusInternalServerError)
           json.NewEncoder(w).Encode("Error with the Encryption Tool")
           return
        }
        hashedemail, err := bcrypt.GenerateFromPassword([]byte(email), bcrypt.DefaultCost)
        if(err != nil){
            w.WriteHeader(http.StatusInternalServerError)
           json.NewEncoder(w).Encode("Error with the Encryption Tool")
           return
        }
        error := users.Insert(&User{string(hashedemail),name,email,string(hashedPassword),
            "",nil,nil})
        if error != nil {
            w.WriteHeader(http.StatusForbidden)
                log.Fatal(error)
        }else{
            w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode("User Registered Successfully , your id is : "+ string(hashedemail))

        }

}
