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
func HandleMemo(w http.ResponseWriter, req *http.Request, body string) {

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
		newBody = newBody + tokens[i] + "."
		}
	body = newBody
	//route to a handler based on the request
	if strings.Contains(body, "make") {
		MakeMemoHandler(w, req, body, auth)
	} else if strings.Contains(body, "edit") {
		EditMemoHandler(w, req, body, auth)
	} else if strings.Contains(body, "delete") {
		DeleteMemoHandler(w, req, body, auth)
	} else if strings.Contains(body, "showAll") {
		ShowAllMemosHandler(w, req, body, auth)
	} else if strings.Contains(body, "show") {
		ShowMemoHandler(w, req, body, auth)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Not a valid instruction for Memo operations {make,edit,delete,showAll}"}
		json.NewEncoder(w).Encode(e)
		return
	}

}

// a function that validates the make memo request and extracts the information from the request
func MakeMemoHandler(w http.ResponseWriter, req *http.Request, body string, auth string) {
	tokens := strings.Split(body, ".")
	if len(tokens) != 4 {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format"}
		json.NewEncoder(w).Encode(e)
		return
	}
	if strings.Contains(tokens[2], "name:") && strings.Contains(tokens[3], "content:") == false {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format the format should be in the form of memo.make.name:name of your memo .content:content of your memo"}
		json.NewEncoder(w).Encode(e)
		return
	}
	if len(tokens[3]) < 5 || len(tokens[4]) < 8 {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format"}
		json.NewEncoder(w).Encode(e)
		return
	}
	if tokens[0] != strings.ToLower("memo") || tokens[1] != strings.ToLower("make") || (tokens[2])[0:5] != strings.ToLower("name:") || (tokens[3])[0:8] != strings.ToLower("content:") {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format the format should be in the form of memo.make.name:name of your memo .content:content of your memo"}
		json.NewEncoder(w).Encode(e)
		return
	}
	if strings.Replace(strings.Split(tokens[2], ":")[1], " ", "", -1) == "" ||
		strings.Replace(strings.Split(tokens[3], ":")[1], " ", "", -1) == "" {

		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Neither the name nor the content can be spaces or empty string"}
		json.NewEncoder(w).Encode(e)
		return
	}
	name := strings.Split(tokens[2], ":")[1]
	content := strings.Split(tokens[3], ":")[1]
	MakeMemo(w, req, name, content, auth)

}

// adds a new memo to the list of memos of the user who made the request
func MakeMemo(w http.ResponseWriter, req *http.Request, name string, content string, auth string) {

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

	newMemo := Memo{
		Id:      ai.Next("memos"),
		Name:    name,
		Content: content,
	}
	// inserting the Memo
	colQuerier := bson.M{"unique": auth}
	change := bson.M{"$set": bson.M{"memos": append(foundUser.Memos, newMemo)}}
	err2 := users.Update(colQuerier, change)
	if err2 != nil {
		w.WriteHeader(http.StatusInternalServerError)
		e := map[string]string{"message": "Can't Update Memos"}
		json.NewEncoder(w).Encode(e)
		return
	}

	w.WriteHeader(http.StatusOK)
	e := map[string]string{"message": "Memo Added Successfully"}
	json.NewEncoder(w).Encode(e)
	return

}

///////////

func EditMemoHandler(w http.ResponseWriter, req *http.Request, body string, auth string) {
	tokens := strings.Split(body, ".")
	if len(tokens) != 5 {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format"}
		json.NewEncoder(w).Encode(e)
		return
	}
	if len(tokens[2]) < 3 || len(tokens[3]) < 5 || len(tokens[4]) < 8 {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format"}
		json.NewEncoder(w).Encode(e)
		return
	}
	if strings.Contains(tokens[2], "id:") && strings.Contains(tokens[3], "name:") && strings.Contains(tokens[4], "content:") == false {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format the format should be in the form of memo.edit.id:id of your memo .name:new name .content:new content"}
		json.NewEncoder(w).Encode(e)
		return
	}

	if tokens[0] != strings.ToLower("memo") || tokens[1] != strings.ToLower("edit") || (tokens[2])[0:3] != strings.ToLower("id:") || (tokens[3])[0:5] != strings.ToLower("name:") || (tokens[4])[0:8] != strings.ToLower("content:") {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format the format should be in the form of memo.edit.id:id of your memo .name:new name .content:new content"}
		json.NewEncoder(w).Encode(e)
		return
	}
	if strings.Replace(strings.Split(tokens[2], ":")[1], " ", "", -1) == "" ||
		strings.Replace(strings.Split(tokens[3], ":")[1], " ", "", -1) == "" ||
		strings.Replace(strings.Split(tokens[4], ":")[1], " ", "", -1) == "" {

		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Neither the id, the name nor the content can be spaces or empty string"}
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
	content := strings.Split(tokens[4], ":")[1]
	EditMemo(w, req, id, name, content, auth)
}

func EditMemo(w http.ResponseWriter, req *http.Request, id int, name string, content string, auth string) {

	// validity check using Authorization key
	// database configuration
	session, err := mgo.Dial("mongodb://mahmoud.salem:123a456@ds145223.mlab.com:45223/personalassistant")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
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

	currentMemos := foundUser.Memos
	inputId := uint64(id)
	found := 0
	for i := 0; i < len(currentMemos); i++ {
		if currentMemos[i].Id == inputId {
			currentMemos[i].Content = content
			currentMemos[i].Name = name
			found++
			break
		}

	}
	if found == 0 {
		w.WriteHeader(http.StatusForbidden)
		e := map[string]string{"message": "You don't have a Memo with this ID"}
		json.NewEncoder(w).Encode(e)
		return
	}
	colQuerier := bson.M{"unique": auth}
	change := bson.M{"$set": bson.M{"memos": currentMemos}}
	err2 := users.Update(colQuerier, change)
	if err2 != nil {
		w.WriteHeader(http.StatusInternalServerError)
		e := map[string]string{"message": "Can't Update Memos"}
		json.NewEncoder(w).Encode(e)
		return
	}
	w.WriteHeader(http.StatusOK)
	e := map[string]string{"message": "Memo Updated Successfully"}
	json.NewEncoder(w).Encode(e)
	return

}

////////
func DeleteMemoHandler(w http.ResponseWriter, req *http.Request, body string, auth string) {
	tokens := strings.Split(body, ".")
	if len(tokens) != 3 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Invalid Format")
		return
	}
	if strings.Contains(tokens[2], "id:") == false {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format the format should be in the form of memo.delete.id:id of your memo"}
		json.NewEncoder(w).Encode(e)
		return
	}
	if len(tokens[2]) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format"}
		json.NewEncoder(w).Encode(e)
		return
	}

	if tokens[0] != strings.ToLower("memo") || tokens[1] != strings.ToLower("delete") || (tokens[2])[0:3] != "id:" {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format the format should be in the form of memo.delete.id:id of your memo"}
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
	DeleteMemo(w, req, id, auth)
}

func DeleteMemo(w http.ResponseWriter, req *http.Request, id int, auth string) {

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
	currentMemos := foundUser.Memos
	inputId := uint64(id)
	found := 0
	for i := 0; i < len(currentMemos); i++ {
		if currentMemos[i].Id == inputId {
			currentMemos = append(currentMemos[:i], currentMemos[i+1:]...)
			found++
			break
		}

	}
	if found == 0 {
		w.WriteHeader(http.StatusForbidden)
		e := map[string]string{"message": "You don't have a Memo with this ID"}
		json.NewEncoder(w).Encode(e)
		return
	}
	colQuerier := bson.M{"unique": auth}
	change := bson.M{"$set": bson.M{"memos": currentMemos}}
	err2 := users.Update(colQuerier, change)
	if err2 != nil {
		w.WriteHeader(http.StatusInternalServerError)
		e := map[string]string{"message": "Can't Delete Memo due to a Database Error"}
		json.NewEncoder(w).Encode(e)
		return
	}
	w.WriteHeader(http.StatusOK)
	e := map[string]string{"message": "Memo Deleted Successfully"}
	json.NewEncoder(w).Encode(e)
	return

}

func ShowAllMemosHandler(w http.ResponseWriter, req *http.Request, body string, auth string) {
	tokens := strings.Split(body, ".")
	if len(tokens) != 2 {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format"}
		json.NewEncoder(w).Encode(e)
		return
	}

	if tokens[0] != "memo" || tokens[1] != "showAll" {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format the format should be in the form of memo.showAll"}
		json.NewEncoder(w).Encode(e)
		return
	}
	showAllMemos(w, req, auth)
}

func showAllMemos(w http.ResponseWriter, req *http.Request, auth string) {
	session, err := mgo.Dial("mongodb://mahmoud.salem:123a456@ds145223.mlab.com:45223/personalassistant")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
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
	currentMemos := foundUser.Memos
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(currentMemos)
	return
}

func ShowMemoHandler(w http.ResponseWriter, req *http.Request, body string, auth string) {
	tokens := strings.Split(body, ".")
	if len(tokens) != 3 {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format"}
		json.NewEncoder(w).Encode(e)
		return
	}
	if strings.Contains(tokens[2], "id:") == false {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format the format should be in the form of memo.show.id:id of your memo"}
		json.NewEncoder(w).Encode(e)
		return
	}
	if len(tokens[2]) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format"}
		json.NewEncoder(w).Encode(e)
		return
	}
	if tokens[0] != "memo" || tokens[1] != "show" || (tokens[2])[0:3] != "id:" {
		w.WriteHeader(http.StatusBadRequest)
		e := map[string]string{"message": "Invalid Format the format should be in the form of memo.show.id:id of your memo"}
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
	ShowMemo(w, req, id, auth)
}

func ShowMemo(w http.ResponseWriter, req *http.Request, id int, auth string) {
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

	currentMemos := foundUser.Memos
	inputId := uint64(id)
	found := 0
	for i := 0; i < len(currentMemos); i++ {
		if currentMemos[i].Id == inputId {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(currentMemos[i])
			return
		}

	}
	if found == 0 {
		w.WriteHeader(http.StatusForbidden)
		e := map[string]string{"message": "You don't have a Memo with this ID"}
		json.NewEncoder(w).Encode(e)
		return
	}
}