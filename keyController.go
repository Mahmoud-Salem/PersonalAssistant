package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"fmt"
	"github.com/night-codes/mgo-ai"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

// the general function that routes to the another function based on the request and Authorize the request
func HandleKey(w http.ResponseWriter, req *http.Request, body string) {

	var tokens = strings.Split(body, ".")
	var lastToken = (tokens[len(tokens)-1])
	if len(strings.Split(lastToken, ":")) < 2 {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Please Provide The Authorization Key as loggedin_id"}
		json.NewEncoder(w).Encode(e)
		return
	}
	if strings.Split(lastToken, ":")[0] != "loggedin_id" || strings.Split(lastToken, ":")[1] == "" {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Please Provide The Authorization Key as loggedin_id"}
		json.NewEncoder(w).Encode(e)
		return
	}
	var idToken = strings.Split(lastToken, ":")[1]

	// validate the request header
	if idToken == "" {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Please Provide The Authorization Key"}
		json.NewEncoder(w).Encode(e)
		return
	}
	// validate the database connection
	auth := idToken
	session, err := mgo.Dial("mongodb://mahmoud.salem:123a456@ds145223.mlab.com:45223/personalassistant")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		e := map[string]string{"message": "Internal Error"}
		json.NewEncoder(w).Encode(e)
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	// validat the id
	users := session.DB("personalassistant").C("users")
	foundUser := User{}
	err = users.Find(bson.M{"unique": string(auth)}).One(&foundUser)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		e := map[string]string{"message": "No Such an Authorization ID."}
		json.NewEncoder(w).Encode(e)
		return
	}
	var newBody = ""
	for i := 0; i < len(tokens)-1; i++ {
		if i == len(tokens)-2 {
			newBody = newBody + tokens[i]
		} else {
			newBody = newBody + tokens[i] + "."
		}
		}

	body = newBody
	fmt.Println(body)
	//route to a handler based on the request
	if strings.Contains(body, "make") {
		MakeKeyHandler(w, req, body, auth)
	} else if strings.Contains(body, "edit") {
		EditKeyHandler(w, req, body, auth)
	} else if strings.Contains(body, "delete") {
		DeleteKeyHandler(w, req, body, auth)
	} else if strings.Contains(body, "showAll") {
		ShowAllKeysHandler(w, req, body, auth)
	} else if strings.Contains(body, "show") {
		ShowKeyHandler(w, req, body, auth)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Not a valid instruction for Key operations {make,edit,delete,show,showAll}"}
		json.NewEncoder(w).Encode(e)
		return
	}

}

// a function that validates the make key request and extracts the information from the request
func MakeKeyHandler(w http.ResponseWriter, req *http.Request, body string, auth string) {
	tokens := strings.Split(body, ".")
	if len(tokens) != 4 {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format"}
		json.NewEncoder(w).Encode(e)
		return
	}
	if len(tokens[2]) < 5 || len(tokens[3]) < 6 {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format"}
		json.NewEncoder(w).Encode(e)
		return
	}
	if strings.Contains(tokens[2], "name:") && strings.Contains(tokens[3], "value:") == false {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format the format should be in the form of key.make.name:name of your key .value:value of your key"}
		json.NewEncoder(w).Encode(e)
		return
	}
	if tokens[0] != "key" || tokens[1] != "make" || (tokens[2])[0:5] != "name:" || (tokens[3])[0:6] != "value:" {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format the format should be in the form of key.make.name:name of your key .value:value of your key"}
		json.NewEncoder(w).Encode(e)
		return
	}
	if strings.Replace(strings.Split(tokens[2], ":")[1], " ", "", -1) == "" ||
		strings.Replace(strings.Split(tokens[3], ":")[1], " ", "", -1) == "" {

		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Neither the name nor the value can be spaces or empty string"}
		json.NewEncoder(w).Encode(e)
		return
	}
	name := strings.Split(tokens[2], ":")[1]
	key := strings.Split(tokens[3], ":")[1]
	MakeKey(w, req, name, key, auth)

}

// adds a new key to the list of keyss of the user who made the request
func MakeKey(w http.ResponseWriter, req *http.Request, name string, value string, auth string) {

	// database configuration
	session, err := mgo.Dial("mongodb://mahmoud.salem:123a456@ds145223.mlab.com:45223/personalassistant")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		e := map[string]string{"message": "Internal Error"}
		json.NewEncoder(w).Encode(e)
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)
	ai.Connect(session.DB("personalassistant").C("counters"))
	users := session.DB("personalassistant").C("users")
	foundUser := User{}
	err = users.Find(bson.M{"unique": auth}).One(&foundUser)

	newKey := Key{
		Id:    ai.Next("keys"),
		Name:  name,
		Value: value,
	}
	// inserting the Key
	colQuerier := bson.M{"unique": auth}
	change := bson.M{"$set": bson.M{"keys": append(foundUser.Keys, newKey)}}
	err2 := users.Update(colQuerier, change)
	if err2 != nil {
		w.WriteHeader(http.StatusInternalServerError)
		e := map[string]string{"message": "Can't Update Keys"}
		json.NewEncoder(w).Encode(e)
		return
	}

	w.WriteHeader(http.StatusOK)
	e := map[string]string{"message": "Key Added Successfully"}
	json.NewEncoder(w).Encode(e)
	return

}

///////////

func EditKeyHandler(w http.ResponseWriter, req *http.Request, body string, auth string) {
	tokens := strings.Split(body, ".")
	if len(tokens) != 5 {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format"}
		json.NewEncoder(w).Encode(e)
		return

	}
	if len(tokens[2]) < 3 || len(tokens[4]) < 6 || len(tokens[3]) < 5 {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format"}
		json.NewEncoder(w).Encode(e)
		return
	}
	if strings.Contains(tokens[2], "id:") && strings.Contains(tokens[3], "name:") && strings.Contains(tokens[4], "value::") == false {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format the format should be in the form of => key.edit.id:id of your key .name:new name .value:new value"}
		json.NewEncoder(w).Encode(e)
		return
	}

	if tokens[0] != "key" || tokens[1] != "edit" || (tokens[2])[0:3] != strings.ToLower("id:") || (tokens[3])[0:5] != strings.ToLower("name:") || (tokens[4])[0:6] != strings.ToLower("value:") {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format the format should be in the form of => key.edit.id:id of your key .name:new name .value:new value"}
		json.NewEncoder(w).Encode(e)
		return
	}
	if strings.Replace(strings.Split(tokens[2], ":")[1], " ", "", -1) == "" ||
		strings.Replace(strings.Split(tokens[3], ":")[1], " ", "", -1) == "" ||
		strings.Replace(strings.Split(tokens[4], ":")[1], " ", "", -1) == "" {

		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Neither the id, the name nor the value can be spaces or empty string"}
		json.NewEncoder(w).Encode(e)
		return
	}
	id, errInt := strconv.Atoi(strings.Replace(strings.Split(tokens[2], ":")[1], " ", "", -1))
	if errInt != nil {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Id must be a number"}
		json.NewEncoder(w).Encode(e)
		return
	}
	name := strings.Split(tokens[3], ":")[1]
	value := strings.Split(tokens[4], ":")[1]
	EditKey(w, req, id, name, value, auth)
}

func EditKey(w http.ResponseWriter, req *http.Request, id int, name string, value string, auth string) {

	// validity check using Authorization key
	// database configuration
	session, err := mgo.Dial("mongodb://mahmoud.salem:123a456@ds145223.mlab.com:45223/personalassistant")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		e := map[string]string{"message": "Invalid Format"}
		json.NewEncoder(w).Encode(e)
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	// Authentication
	users := session.DB("personalassistant").C("users")
	foundUser := User{}
	err = users.Find(bson.M{"unique": auth}).One(&foundUser)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		e := map[string]string{"message": "Invalid Authorization ID"}
		json.NewEncoder(w).Encode(e)
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
		e := map[string]string{"message": "You don't have a Key with this ID"}
		json.NewEncoder(w).Encode(e)
		return
	}
	colQuerier := bson.M{"unique": auth}
	change := bson.M{"$set": bson.M{"keys": currentKeys}}
	err2 := users.Update(colQuerier, change)
	if err2 != nil {
		w.WriteHeader(http.StatusInternalServerError)
		e := map[string]string{"message": "Can't Update Keys"}
		json.NewEncoder(w).Encode(e)
		return
	}
	w.WriteHeader(http.StatusOK)
	e := map[string]string{"message": "Key Updated Successfully"}
	json.NewEncoder(w).Encode(e)
	return

}

////////
func DeleteKeyHandler(w http.ResponseWriter, req *http.Request, body string, auth string) {
	tokens := strings.Split(body, ".")
	if len(tokens) != 3 {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format"}
		json.NewEncoder(w).Encode(e)
		return
	}
	if strings.Contains(tokens[2], "id:") == false {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format the format should be in the form of key.delete.id:id of your key"}
		json.NewEncoder(w).Encode(e)
		return
	}
	if len(tokens[2]) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format"}
		json.NewEncoder(w).Encode(e)
		return
	}
	if tokens[0] != "key" || tokens[1] != "delete" || (tokens[2])[0:3] != "id:" {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format the format should be in the form of key.delete.id:id of your key"}
		json.NewEncoder(w).Encode(e)
		return
	}
	if strings.Replace(strings.Split(tokens[2], ":")[1], " ", "", -1) == "" {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "The id  can't be spaces or empty string"}
		json.NewEncoder(w).Encode(e)
		return
	}
	id, errInt := strconv.Atoi(strings.Replace(strings.Split(tokens[2], ":")[1], " ", "", -1))
	if errInt != nil {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Id must be a number"}
		json.NewEncoder(w).Encode(e)
		return
	}
	DeleteKey(w, req, id, auth)
}

func DeleteKey(w http.ResponseWriter, req *http.Request, id int, auth string) {

	// validity check using Authentication key
	// database configuration
	session, err := mgo.Dial("mongodb://mahmoud.salem:123a456@ds145223.mlab.com:45223/personalassistant")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	// Login
	users := session.DB("personalassistant").C("users")
	foundUser := User{}
	err = users.Find(bson.M{"unique": auth}).One(&foundUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		e := map[string]string{"message": "Error in the Database Conncetion"}
		json.NewEncoder(w).Encode(e)
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
		e := map[string]string{"message": "You don't have a Key with this ID"}
		json.NewEncoder(w).Encode(e)
		return
	}
	colQuerier := bson.M{"unique": auth}
	change := bson.M{"$set": bson.M{"keys": currentKeys}}
	err2 := users.Update(colQuerier, change)
	if err2 != nil {
		w.WriteHeader(http.StatusInternalServerError)
		e := map[string]string{"message": "Can't Delete Key due to a Database Error"}
		json.NewEncoder(w).Encode(e)
		return
	}
	w.WriteHeader(http.StatusOK)
	e := map[string]string{"message": "Key Deleted Successfully"}
	json.NewEncoder(w).Encode(e)
	return

}

func ShowAllKeysHandler(w http.ResponseWriter, req *http.Request, body string, auth string) {
	tokens := strings.Split(body, ".")
	if len(tokens) != 2 {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format"}
		json.NewEncoder(w).Encode(e)
		return
	}

	if tokens[0] != "key" || tokens[1] != "showAll" {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format the format should be in the form of key.showAll"}
		json.NewEncoder(w).Encode(e)
		return
	}
	showAllKeys(w, req, auth)
}

func showAllKeys(w http.ResponseWriter, req *http.Request, auth string) {
	session, err := mgo.Dial("mongodb://mahmoud.salem:123a456@ds145223.mlab.com:45223/personalassistant")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		e := map[string]string{"message": "Internal Error"}
		json.NewEncoder(w).Encode(e)
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	// Authentication
	users := session.DB("personalassistant").C("users")
	foundUser := User{}
	err = users.Find(bson.M{"unique": auth}).One(&foundUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		e := map[string]string{"message": "Error in the Database Conncetion"}
		json.NewEncoder(w).Encode(e)
		return
	}

	keys := foundUser.Keys
	
	w.WriteHeader(http.StatusOK)
	result := ""
	for i := 0; i < len(keys); i++ {
		result = result + strconv.Itoa(int(keys[i].Id)) + "; " + keys[i].Name + "; " + keys[i].Value + "    "
	}
	e := map[string]string{"message": result}
	json.NewEncoder(w).Encode(e)
	return
}

////
func ShowKeyHandler(w http.ResponseWriter, req *http.Request, body string, auth string) {
	tokens := strings.Split(body, ".")
	if len(tokens) != 3 {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format"}
		json.NewEncoder(w).Encode(e)
		return
	}
	if strings.Contains(tokens[2], "id:") == false {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format the format should be in the form of key.show.id:id of your key"}
		json.NewEncoder(w).Encode(e)
		return
	}
	if len(tokens[2]) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format"}
		json.NewEncoder(w).Encode(e)
		return
	}
	if tokens[0] != "key" || tokens[1] != "show" || (tokens[2])[0:3] != "id:" {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format the format should be in the form of key.show.id:id of your key"}
		json.NewEncoder(w).Encode(e)
		return
	}
	if strings.Replace(strings.Split(tokens[2], ":")[1], " ", "", -1) == "" {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "The id  can't be spaces or empty string"}
		json.NewEncoder(w).Encode(e)
		return
	}
	id, errInt := strconv.Atoi(strings.Replace(strings.Split(tokens[2], ":")[1], " ", "", -1))
	if errInt != nil {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Id must be a number"}
		json.NewEncoder(w).Encode(e)
		return
	}
	ShowKey(w, req, id, auth)
}

func ShowKey(w http.ResponseWriter, req *http.Request, id int, auth string) {
	// validity check using Authentication key
	// database configuration
	session, err := mgo.Dial("mongodb://mahmoud.salem:123a456@ds145223.mlab.com:45223/personalassistant")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		e := map[string]string{"message": "Internal Error"}
		json.NewEncoder(w).Encode(e)
		panic(err)
	}
	defer session.Close()
	session.SetMode(mgo.Monotonic, true)

	// Authorization
	users := session.DB("personalassistant").C("users")
	foundUser := User{}
	err = users.Find(bson.M{"unique": auth}).One(&foundUser)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		e := map[string]string{"message": "Error in the Database Conncetion"}
		json.NewEncoder(w).Encode(e)
		return
	}
	currentKeys := foundUser.Keys
	inputId := uint64(id)
	found := 0
	for i := 0; i < len(currentKeys); i++ {
		if currentKeys[i].Id == inputId {

			w.WriteHeader(http.StatusOK)
			result := strconv.Itoa(int(currentKeys[i].Id)) + "; " + currentKeys[i].Name + "; " + currentKeys[i].Value + "    "
			e := map[string]string{"message": result}
			json.NewEncoder(w).Encode(e)

			return
		}

	}
	if found == 0 {
		w.WriteHeader(http.StatusForbidden)
		e := map[string]string{"message": "You don't have a Key with this ID"}
		json.NewEncoder(w).Encode(e)
		return
	}
}