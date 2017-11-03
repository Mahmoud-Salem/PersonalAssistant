package main

import (
	"encoding/json"
	"net/http"
	"strconv"

	"strings"

	"github.com/night-codes/mgo-ai"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// the general function that routes to the another function based on the request and Authorize the request
func HandleKey(w http.ResponseWriter, req *http.Request, body string) {

	// validate the request header
	if req.Header.Get("Authorization") == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Please Provide The Authorization Key")
		return
	}
	// validate the database connection
	auth := req.Header.Get("Authorization")
	session, err := mgo.Dial("localhost")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	// validat the id
	users := session.DB("test").C("users")
	foundUser := User{}
	err = users.Find(bson.M{"unique": string(auth)}).One(&foundUser)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(`{ "message" :"No Such an Authorization ID."}`)
		return
	}
	// validate the request body
	if body == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Please Provide a Body for your Request")
		return
	}
	//route to a handler based on the request
	if strings.Contains(body, "make") {
		MakeKeyHandler(w, req, body)
	} else if strings.Contains(body, "edit") {
		EditKeyHandler(w, req, body)
	} else if strings.Contains(body, "delete") {
		DeleteKeyHandler(w, req, body)
	} else if strings.Contains(body, "showAll") {
		ShowAllKeysHandler(w, req, body)
	} else if strings.Contains(body, "show") {
		ShowKeyHandler(w, req, body)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Not a valid instruction for Key operations {make,edit,delete,show,showAll}")
		return
	}

}

// a function that validates the make key request and extracts the information from the request
func MakeKeyHandler(w http.ResponseWriter, req *http.Request, body string) {
	tokens := strings.Split(body, ".")
	if len(tokens) != 4 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Invalid Format")
		return
	}
	if len(tokens[2]) < 5 || len(tokens[3]) < 6 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Invalid Format")
		return
	}
	if strings.Contains(tokens[2], "name:") && strings.Contains(tokens[3], "value:") == false {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Invalid Format the format should be in the form of key.make.name:name of your key .value:value of your key")
		return
	}
	if tokens[0] != "key" || tokens[1] != "make" || (tokens[2])[0:5] != "name:" || (tokens[3])[0:6] != "value:" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Invalid Format the format should be in the form of key.make.name:name of your key .value:value of your key")
		return
	}
	if strings.Replace(strings.Split(tokens[2], ":")[1], " ", "", -1) == "" ||
		strings.Replace(strings.Split(tokens[3], ":")[1], " ", "", -1) == "" {

		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Neither the name nor the value can be spaces or empty string")
		return
	}
	name := strings.Split(tokens[2], ":")[1]
	key := strings.Split(tokens[3], ":")[1]
	MakeKey(w, req, name, key)

}

// adds a new key to the list of keyss of the user who made the request
func MakeKey(w http.ResponseWriter, req *http.Request, name string, value string) {

	// database configuration
	session, err := mgo.Dial("localhost")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	ai.Connect(session.DB("test").C("counters"))
	users := session.DB("test").C("users")
	foundUser := User{}
	err = users.Find(bson.M{"unique": req.Header.Get("Authorization")}).One(&foundUser)

	newKey := Key{
		Id:    ai.Next("keys"),
		Name:  name,
		Value: value,
	}
	// inserting the Key
	colQuerier := bson.M{"unique": req.Header.Get("Authorization")}
	change := bson.M{"$set": bson.M{"keys": append(foundUser.Keys, newKey)}}
	err2 := users.Update(colQuerier, change)
	if err2 != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Can't Update Keys")
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("Key Added Successfully")
	return

}

///////////

func EditKeyHandler(w http.ResponseWriter, req *http.Request, body string) {
	tokens := strings.Split(body, ".")
	if len(tokens) != 5 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Invalid Format")
		return

	}
	if len(tokens[2]) < 3 || len(tokens[4]) < 6 || len(tokens[3]) < 5 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Invalid Format")
		return
	}
	if strings.Contains(tokens[2], "id:") && strings.Contains(tokens[3], "name:") && strings.Contains(tokens[4], "value::") == false {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Invalid Format the format should be in the form of => key.edit.id:id of your key .name:new name .value:new value")
		return
	}

	if tokens[0] != "key" || tokens[1] != "edit" || (tokens[2])[0:3] != strings.ToLower("id:") || (tokens[3])[0:5] != strings.ToLower("name:") || (tokens[4])[0:6] != strings.ToLower("value:") {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Invalid Format the format should be in the form of => key.edit.id:id of your key .name:new name .value:new value")
		return
	}
	if strings.Replace(strings.Split(tokens[2], ":")[1], " ", "", -1) == "" ||
		strings.Replace(strings.Split(tokens[3], ":")[1], " ", "", -1) == "" ||
		strings.Replace(strings.Split(tokens[4], ":")[1], " ", "", -1) == "" {

		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Neither the id, the name nor the value can be spaces or empty string")
		return
	}
	id, errInt := strconv.Atoi(strings.Replace(strings.Split(tokens[2], ":")[1], " ", "", -1))
	if errInt != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Id must be a number")
		return
	}
	name := strings.Split(tokens[3], ":")[1]
	value := strings.Split(tokens[4], ":")[1]
	EditKey(w, req, id, name, value)
}

func EditKey(w http.ResponseWriter, req *http.Request, id int, name string, value string) {

	// validity check using Authorization key
	// database configuration
	session, err := mgo.Dial("localhost")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	// Authentication
	users := session.DB("test").C("users")
	foundUser := User{}
	err = users.Find(bson.M{"unique": req.Header.Get("Authorization")}).One(&foundUser)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode("Invalid Authorization ID")
		return
	}

	currentKeys := foundUser.Keys
	inputId := uint64(id)

	found := 0
	for i := 0; i < len(currentKeys); i++ {
		if currentKeys[i].Id == inputId {
			currentKeys[i].Value = value
			currentKeys[i].Name = name
			found++
			break
		}

	}
	if found == 0 {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode("You don't have a Key with this ID")
		return
	}
	colQuerier := bson.M{"unique": req.Header.Get("Authorization")}
	change := bson.M{"$set": bson.M{"keys": currentKeys}}
	err2 := users.Update(colQuerier, change)
	if err2 != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Can't Update Keys")
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("Key Updated Successfully")
	return

}

////////
func DeleteKeyHandler(w http.ResponseWriter, req *http.Request, body string) {
	tokens := strings.Split(body, ".")
	if len(tokens) != 3 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Invalid Format")
		return
	}
	if strings.Contains(tokens[2], "id:") == false {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Invalid Format the format should be in the form of key.delete.id:id of your key")
		return
	}
	if len(tokens[2]) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Invalid Format")
		return
	}
	if tokens[0] != "key" || tokens[1] != "delete" || (tokens[2])[0:3] != "id:" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Invalid Format the format should be in the form of key.delete.id:id of your key")
		return
	}
	if strings.Replace(strings.Split(tokens[2], ":")[1], " ", "", -1) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("The id  can't be spaces or empty string")
		return
	}
	id, errInt := strconv.Atoi(strings.Replace(strings.Split(tokens[2], ":")[1], " ", "", -1))
	if errInt != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Id must be a number")
		return
	}
	DeleteKey(w, req, id)
}

func DeleteKey(w http.ResponseWriter, req *http.Request, id int) {

	// validity check using Authentication key
	// database configuration
	session, err := mgo.Dial("localhost")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	// Login
	users := session.DB("test").C("users")
	foundUser := User{}
	err = users.Find(bson.M{"unique": req.Header.Get("Authorization")}).One(&foundUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Error in the Database Conncetion")
		return
	}
	currentKeys := foundUser.Keys
	inputId := uint64(id)
	found := 0
	for i := 0; i < len(currentKeys); i++ {
		if currentKeys[i].Id == inputId {
			currentKeys = append(currentKeys[:i], currentKeys[i+1:]...)
			found++
			break
		}

	}
	if found == 0 {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode("You don't have a Key with this ID")
		return
	}
	colQuerier := bson.M{"unique": req.Header.Get("Authorization")}
	change := bson.M{"$set": bson.M{"keys": currentKeys}}
	err2 := users.Update(colQuerier, change)
	if err2 != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Can't Delete Key due to a Database Error")
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("Key Deleted Successfully")
	return

}

func ShowAllKeysHandler(w http.ResponseWriter, req *http.Request, body string) {
	tokens := strings.Split(body, ".")
	if len(tokens) != 2 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Invalid Format")
		return
	}

	if tokens[0] != "key" || tokens[1] != "showAll" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Invalid Format the format should be in the form of key.showAll")
		return
	}
	showAllKeys(w, req)
}

func showAllKeys(w http.ResponseWriter, req *http.Request) {
	session, err := mgo.Dial("localhost")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	// Authentication
	users := session.DB("test").C("users")
	foundUser := User{}
	err = users.Find(bson.M{"unique": req.Header.Get("Authorization")}).One(&foundUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Error in the Database Conncetion")
		return
	}
	currentKeys := foundUser.Keys
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(currentKeys)
	return
}

////
func ShowKeyHandler(w http.ResponseWriter, req *http.Request, body string) {
	tokens := strings.Split(body, ".")
	if len(tokens) != 3 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Invalid Format")
		return
	}
	if strings.Contains(tokens[2], "id:") == false {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Invalid Format the format should be in the form of key.show.id:id of your key")
		return
	}
	if len(tokens[2]) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Invalid Format")
		return
	}
	if tokens[0] != "key" || tokens[1] != "show" || (tokens[2])[0:3] != "id:" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Invalid Format the format should be in the form of key.show.id:id of your key")
		return
	}
	if strings.Replace(strings.Split(tokens[2], ":")[1], " ", "", -1) == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("The id  can't be spaces or empty string")
		return
	}
	id, errInt := strconv.Atoi(strings.Replace(strings.Split(tokens[2], ":")[1], " ", "", -1))
	if errInt != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Id must be a number")
		return
	}
	ShowKey(w, req, id)
}

func ShowKey(w http.ResponseWriter, req *http.Request, id int) {
	// validity check using Authentication key
	// database configuration
	session, err := mgo.Dial("localhost")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	// Authorization
	users := session.DB("test").C("users")
	foundUser := User{}
	err = users.Find(bson.M{"unique": req.Header.Get("Authorization")}).One(&foundUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode("Error in the Database Conncetion")
		return
	}
	currentKeys := foundUser.Keys
	inputId := uint64(id)
	found := 0
	for i := 0; i < len(currentKeys); i++ {
		if currentKeys[i].Id == inputId {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(currentKeys[i])
			return
		}

	}
	if found == 0 {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode("You don't have a Key with this ID")
		return
	}
}
