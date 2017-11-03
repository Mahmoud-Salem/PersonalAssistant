package main

import(
    "encoding/json"
    "net/http"
    "gopkg.in/mgo.v2"
    "gopkg.in/mgo.v2/bson"
    "golang.org/x/crypto/bcrypt"
    "strings"
)

func Login (w http.ResponseWriter , req *http.Request){

   // Validity checking    
   if req.Body == nil {
    w.WriteHeader(http.StatusBadRequest)
    json.NewEncoder(w).Encode(`{ "message" : "provide your email and password in the following format to login . email : ms@gmail.com . password :123456 . "}`)
    return
}
    req.ParseForm()
    if req.Form.Get("message") == "" {
		w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(`{ "message" : "provide your email and password in the following format to login . email : ms@gmail.com . password :123456 . "}`)
		return
    }
    body := req.Form.Get("message")

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
        json.NewEncoder(w).Encode(`{ "message" : "provide your email and password in the following format to login . email : ms@gmail.com . password :123456 . "}`)
		return
    }
// database configuration 
    session, err := mgo.Dial("mongodb://<mahmoud.salem>:<Mesqueunclub12>@ds145223.mlab.com:45223/personalassistant")   
        if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
                panic(err)
        }
        defer session.Close()
        session.SetMode(mgo.Monotonic, true)

        // Login 
        users := session.DB("test").C("users")
        foundUser := User{}
        err = users.Find(bson.M{"email": email}).One(&foundUser)
        if(err != nil){
          w.WriteHeader(http.StatusUnauthorized)
          json.NewEncoder(w).Encode("No Such Email")
          return
        }
     	error := bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(password))
    if(error != nil){
      w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode("Wrong Password")
        return
    }else{
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode("Logged-in Succesfully , here is your id. Use it to perform any action " +foundUser.Unique)
        return
    }

 
}
