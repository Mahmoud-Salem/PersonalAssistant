package main

import(
    "encoding/json"
    "net/http"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "golang.org/x/crypto/bcrypt"
    "strings"
)

func Login (w http.ResponseWriter , req *http.Request , body string){


    tokens := strings.Split(body, ".")
    
    email := "" 
    password := ""     
    for i:= 0 ; i<len(tokens) ; i++ {
        req := strings.Split(tokens[i], ":")

        if strings.TrimSpace(req[0]) == "email" {
            email = strings.TrimSpace(req[1])
        }
        if strings.TrimSpace(req[0]) == "password" {
            password = strings.TrimSpace(req[1])
        }
    }
    if !(email != "" && password != "") {
        w.WriteHeader(http.StatusBadRequest)
        e := map[string]string{"message":"provide your email and password in the following format to login . Login request . email : ms@gmail.com . password :123456 . "}		
        json.NewEncoder(w).Encode(e)
		return
    }
// database configuration 
    session, err := mgo.Dial("mongodb://mahmoud.salem:123a456@ds145223.mlab.com:45223/personalassistant")   
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            e := map[string]string{"message":"Internal Error "}		
            json.NewEncoder(w).Encode(e)
                panic(err)
        }
        defer session.Close()
        session.SetMode(mgo.Monotonic, true)

        // Login 
        users := session.DB("personalassistant").C("users")
        foundUser := User{}
        err = users.Find(bson.M{"email": email}).One(&foundUser)
        if(err != nil){
          w.WriteHeader(http.StatusUnauthorized)
          e := map[string]string{"message":"No Such Email "}		
          json.NewEncoder(w).Encode(e)
          return
        }
     	error := bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(password))
    if(error != nil){
      w.WriteHeader(http.StatusUnauthorized)
        e := map[string]string{"message":"Wrong Password "}		
        json.NewEncoder(w).Encode(e)
        return
    }else{
        w.WriteHeader(http.StatusOK)
        e := map[string]string{"message":"Logged-in Succesfully your id to perform actions : "+foundUser.Unique ,"uuid":foundUser.Unique}		
        json.NewEncoder(w).Encode(e)
        return
    }

 
}
