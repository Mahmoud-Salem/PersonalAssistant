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
    "crypto/md5"
	"strconv"
	"time"
	"encoding/hex"
)

 

func Register (w http.ResponseWriter , req *http.Request , body string){
   


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
        e := map[string]string{"message":"provide your name , email and password in the following format to register . register request . name : mahmoud . email : ms@gmail.com . password :123456 . "}		
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

        hasher := md5.New()
        hasher.Write([]byte(strconv.FormatInt(time.Now().Unix(), 10)))
        uuid := hex.EncodeToString(hasher.Sum(nil))
        error := users.Insert(&User{uuid,name,email,string(hashedPassword),
            "",nil,nil})
        if error != nil {
            w.WriteHeader(http.StatusForbidden)
                e := map[string]string{"message":"Internal Error"}		
                json.NewEncoder(w).Encode(e)
                log.Fatal(error)
                return
        }else{
            w.WriteHeader(http.StatusOK)
            e := map[string]string{"message":"User Registered Successfully , your id to perform actions : "+uuid }		
            json.NewEncoder(w).Encode(e)
        }

}
