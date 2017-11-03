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
    e := map[string]string{"message":"provide your name , email and password in the following format to register . name : mahmoud . email : ms@gmail.com . password :123456 . "}		
    json.NewEncoder(w).Encode(e)
    return
}
    req.ParseForm()
    if req.Form.Get("message") == "" {
		w.WriteHeader(http.StatusBadRequest)
        e := map[string]string{"message":"provide your name , email and password in the following format to register . name : mahmoud . email : ms@gmail.com . password :123456 . "}		
        json.NewEncoder(w).Encode(e)
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
        e := map[string]string{"message":"provide your name , email and password in the following format to register . name : mahmoud . email : ms@gmail.com . password :123456 . "}		
        json.NewEncoder(w).Encode(e)
		return
    }
            // database configuration 
            session, err := mgo.Dial("mongodb://mahmoud.salem:123a456@ds145223.mlab.com:45223/personalassistant")   
            ai.Connect(session.DB("test").C("counters"))
            if err != nil {
                w.WriteHeader(http.StatusInternalServerError)
                e := map[string]string{"message":"Internal Error "}		
                json.NewEncoder(w).Encode(e)
                    panic(err)
            }
            defer session.Close()
            session.SetMode(mgo.Monotonic, true)
                        // Registeration
            users := session.DB("personalassistant").C("users")
            currentUsers := []User{User{}}
        err = users.Find(bson.M{"email": email}).All(&currentUsers)
        if(err != nil){
          w.WriteHeader(http.StatusInternalServerError)  
          e := map[string]string{"message":"Error with the Database Connection"}		
          json.NewEncoder(w).Encode(e)
          return
        }
        if(len(currentUsers)>0){
        w.WriteHeader(http.StatusForbidden)
        e := map[string]string{"message":"This Email Already Exists"}		
        json.NewEncoder(w).Encode(e)
          return
        }
        // encrypt the coming pasword 
        hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
        if(err != nil){
            w.WriteHeader(http.StatusInternalServerError)
            e := map[string]string{"message":"Error with the Encryption Tool"}		
            json.NewEncoder(w).Encode(e)
           return
        }
        hashedemail, err := bcrypt.GenerateFromPassword([]byte(email), bcrypt.DefaultCost)
        if(err != nil){
            w.WriteHeader(http.StatusInternalServerError)
            e := map[string]string{"message":"Error with the Encryption Tool"}		
            json.NewEncoder(w).Encode(e)
           return
        }
        error := users.Insert(&User{string(hashedemail),name,email,string(hashedPassword),
            "",nil,nil})
        if error != nil {
            w.WriteHeader(http.StatusForbidden)
                e := map[string]string{"message":"Internal Error"}		
                json.NewEncoder(w).Encode(e)
                log.Fatal(error)
                return
        }else{
            w.WriteHeader(http.StatusOK)
            e := map[string]string{"message":"User Registered Successfully" , "uuid": string(hashedemail)}		
            json.NewEncoder(w).Encode(e)
        }

}
