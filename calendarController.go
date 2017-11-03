package main 

import(
"net/http"
"encoding/json"
"gopkg.in/mgo.v2"
"gopkg.in/mgo.v2/bson"
"google.golang.org/api/calendar/v3"
"time"
"strings"

)

func HandleCalendar(w http.ResponseWriter, req *http.Request ,body string) {




    if strings.Contains(body, "add") {
        AddEvent(w, req ,body)
    } else if strings.Contains(body, "delete") {
        DeleteEvent(w,req,body)
    } else if strings.Contains(body, "show") {
        ShowCalendar(w,req,body)
    } else if strings.Contains(body, "modify"){
        ModifyEvent(w,req,body)
    } else {
        w.WriteHeader(http.StatusBadRequest)
        e := map[string]string{"message":"You have to specify what you want to do with the calendar whether it is add event , delete event , modify event or show all ."}		
        json.NewEncoder(w).Encode(e)
        return;
    }
}

func AddEvent(w http.ResponseWriter, req *http.Request ,  body string){
    // Check the existance of the needed attributes 
    auth := req.Header.Get("Authorization")
    session, err := mgo.Dial("mongodb://mahmoud.salem:123a456@ds145223.mlab.com:45223/personalassistant")   
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
            panic(err)
    }
    defer session.Close()
    session.SetMode(mgo.Monotonic, true)

    // Login 
    users := session.DB("test").C("users")
    foundUser := User{}
    err = users.Find(bson.M{"unique": string(auth)}).One(&foundUser)
    if(err != nil){
        w.WriteHeader(http.StatusUnauthorized)
        e := map[string]string{"message":"No Such ID."}		
        json.NewEncoder(w).Encode(e)
        return
      }

    email := foundUser.Email
    cal := foundUser.CalendarId
    srv ,err3 := Calendar();
        
	if err3 != nil {	
        w.WriteHeader(http.StatusInternalServerError)
        e := map[string]string{"message":"Server Can't use google calendar."}		
        json.NewEncoder(w).Encode(e)
		return	
        }

    if cal == "" {
        call, err := srv.Calendars.Insert(&calendar.Calendar{
            Summary:     email,
            TimeZone:    "Africa/Cairo",
        }).Do()
        if err != nil {
                w.WriteHeader(http.StatusInternalServerError)
                e := map[string]string{"message":"Server problem ."}		
                json.NewEncoder(w).Encode(e)
                return	
        }
        cal = call.Id
        colQuerier := bson.M{"email": email}
        change := bson.M{"$set": bson.M{"calendarid": cal}}
        
        err2:= users.Update(colQuerier, change)
        if err2 != nil {	
        w.WriteHeader(http.StatusInternalServerError)
        e := map[string]string{"message":"Server problem ."}		
        json.NewEncoder(w).Encode(e)
        return	
        }
    }

    // Creating new Event 

    tokens := strings.Split(body, ".")
    
    starttime := "" 
    endtime := "" 
    name := "" 
    description := "" 
    for i:= 0 ; i<len(tokens) ; i++ {
        req := strings.Split(tokens[i], "/")

        if strings.TrimSpace(req[0]) == "start time" {
            starttime = strings.TrimSpace(req[1])
        }
        if strings.TrimSpace(req[0]) == "end time" {
            endtime = strings.TrimSpace(req[1])
        }
        if strings.TrimSpace(req[0]) == "name" {
            name = strings.TrimSpace(req[1])
        }
        if strings.TrimSpace(req[0]) == "description" {
            description = strings.TrimSpace(req[1])
        }
    }
    t := time.Now().Format(time.RFC3339)
    if !(starttime != "" && endtime != "" && name != "" && description != "") {
        w.WriteHeader(http.StatusBadRequest)
        str := "Your request for adding event should be as follows : add calendar event . start time /"+t+" . end time /"+t+". name / anything . description / anything."
        e := map[string]string{"message":str}		
        json.NewEncoder(w).Encode(e)
        return	
    }
    
	event:= &calendar.Event{
			Summary: name,
			Description: description,
			Start: &calendar.EventDateTime{
			  DateTime: starttime,
			  TimeZone: "Africa/Cairo",
			},
			End: &calendar.EventDateTime{
				DateTime: endtime,
				TimeZone: "Africa/Cairo",
			  },
		}

              calendarId := cal
             // fmt.Println(cal)
		  event, err = srv.Events.Insert(calendarId, event).Do()
		  
		  if err != nil {
            w.WriteHeader(http.StatusBadRequest)
            str := "Your request for adding event should be as follows : add calendar event . start time /"+t+" . end time /"+t+". name / anything . description / anything."
            e := map[string]string{"message":str}		
            json.NewEncoder(w).Encode(e)
            return	
		  }
w.WriteHeader(http.StatusOK)
sstr := "Event Added Successfully you can use the event id to modify or delete event , event id :"+event.Id+"."
e := map[string]string{"message":sstr}		
json.NewEncoder(w).Encode(e)
return

}
////////////////////////////////////////////////////////////////////////////////


func ShowCalendar(w http.ResponseWriter, req *http.Request , body string){
    auth := req.Header.Get("Authorization")
    session, err := mgo.Dial("mongodb://mahmoud.salem:123a456@ds145223.mlab.com:45223/personalassistant")   
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
            panic(err)
    }
    defer session.Close()
    session.SetMode(mgo.Monotonic, true)

    // Login 
    users := session.DB("test").C("users")
    foundUser := User{}
    err = users.Find(bson.M{"unique": string(auth)}).One(&foundUser)
    if(err != nil){
        w.WriteHeader(http.StatusUnauthorized)
        e := map[string]string{"message":"No Such ID."}		
        json.NewEncoder(w).Encode(e)
        return
      }

    //email := foundUser.Email
    //cal := foundUser.CalendarId
    srv ,err3 := Calendar();
        
	if err3 != nil {	
        w.WriteHeader(http.StatusInternalServerError)
        e := map[string]string{"message":"Server Can't use google calendar."}		
        json.NewEncoder(w).Encode(e)
		return	
        }


		  calendarId := foundUser.CalendarId
          if calendarId == "" {	
            w.WriteHeader(http.StatusOK)
            e := map[string]string{"message":"No Upcoming events."}		
            json.NewEncoder(w).Encode(e)
            return	
            }
          t := time.Now().Format(time.RFC3339)
          
          events, err := srv.Events.List(calendarId).ShowDeleted(false).SingleEvents(true).TimeMin(t).OrderBy("startTime").Do()
          if err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            e := map[string]string{"message":"Can't Show Events."}		
            json.NewEncoder(w).Encode(e)
			return	
          }
        ev := ""
        if len(events.Items) > 0 {
          ev = ""
          for _, i := range events.Items{
            var when string
            var when2 string
            if i.Start.DateTime != "" {
              when = i.Start.DateTime
            } else {
              when = i.Start.Date
            }
            if i.End.DateTime != "" {
                when2 = i.End.DateTime
              } else {
                when2 = i.End.Date
              }
            ev +="{name:"+i.Summary+", starttime:"+when+", endtime:"+when2+",description:"+i.Description+",Id:"+i.Id+"},"
            }
          
        } else {
            w.WriteHeader(http.StatusOK)
            e := map[string]string{"message":"No Upcoming events."}		
            json.NewEncoder(w).Encode(e)
            return

        }
        if(ev == ""){
            w.WriteHeader(http.StatusOK)
            e := map[string]string{"message":"No Upcoming events."}		
            json.NewEncoder(w).Encode(e)
            return 
        }
        ev = ev[0:len(ev)-1]
w.WriteHeader(http.StatusOK)
e := map[string]string{"message":ev}		
json.NewEncoder(w).Encode(e)
return

}



///////////////////////////////////////////////////////////////////

func DeleteEvent(w http.ResponseWriter, req *http.Request , body string){
    auth := req.Header.Get("Authorization")
    session, err := mgo.Dial("mongodb://mahmoud.salem:123a456@ds145223.mlab.com:45223/personalassistant")   
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
            panic(err)
    }
    defer session.Close()
    session.SetMode(mgo.Monotonic, true)

    // Login 
    users := session.DB("test").C("users")
    foundUser := User{}
    err = users.Find(bson.M{"unique": string(auth)}).One(&foundUser)
    if(err != nil){
        w.WriteHeader(http.StatusUnauthorized)
        e := map[string]string{"message":"No Such ID."}		
        json.NewEncoder(w).Encode(e)
        return
      }

    //email := foundUser.Email
    cal := foundUser.CalendarId



    if cal == "" {
        w.WriteHeader(422)
        e := map[string]string{"message":"No events with this id ."}		
        json.NewEncoder(w).Encode(e)
        return
    }

	srv ,err3 := Calendar()
	if err3 != nil {	
		w.WriteHeader(http.StatusInternalServerError)
        e := map[string]string{"message":"Server Can't use google calendar."}		
        json.NewEncoder(w).Encode(e)
		return	
        }

        tokens := strings.Split(body, ".")
        id := ""
    for  i:=0 ; i<len(tokens) ; i++ {
            req := strings.Split(tokens[i], ":")
            if strings.TrimSpace(req[0]) == "event id" {
                id = strings.TrimSpace(req[1])
            }
        }
        if id == "" {
            w.WriteHeader(http.StatusBadRequest)
            e := map[string]string{"message":"Your request for adding event should be as follows : delete calendar event . event id :id."}		
            json.NewEncoder(w).Encode(e)
            return	
        }
         errr:=srv.Events.Delete(cal,id).Do()
          if errr != nil {
            w.WriteHeader(422)
            e := map[string]string{"message":"No events with this id ."}		
            json.NewEncoder(w).Encode(e)
            return	
          }
          w.WriteHeader(http.StatusOK)
          e := map[string]string{"message":"event deleted successfully."}		
          json.NewEncoder(w).Encode(e)
          return

}

///////////////////////////////////////////////////

func ModifyEvent(w http.ResponseWriter, req *http.Request,body string){
    auth := req.Header.Get("Authorization")
    session, err := mgo.Dial("mongodb://mahmoud.salem:123a456@ds145223.mlab.com:45223/personalassistant")   
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
            panic(err)
    }
    defer session.Close()
    session.SetMode(mgo.Monotonic, true)

    // Login 
    users := session.DB("test").C("users")
    foundUser := User{}
    err = users.Find(bson.M{"unique": string(auth)}).One(&foundUser)
    if(err != nil){
        w.WriteHeader(http.StatusUnauthorized)
        e := map[string]string{"message":"No Such ID."}		
        json.NewEncoder(w).Encode(e)
        return
      }

    //email := foundUser.Email
    cal := foundUser.CalendarId
    srv ,err3 := Calendar();
        
	if err3 != nil {	
        w.WriteHeader(http.StatusInternalServerError)
        e := map[string]string{"message":"Server Can't use google calendar."}		
        json.NewEncoder(w).Encode(e)
		return	
        }


    if foundUser.CalendarId == "" {
        w.WriteHeader(422)
        e := map[string]string{"message":"No events with this id ."}		
        json.NewEncoder(w).Encode(e)
        return
    }

    starttime := ""
    endtime := ""
    name := ""
    description := ""
    eventId := ""

    tokens := strings.Split(body, ".")
    
        for i:=0 ; i<len(tokens) ; i++ {
            req := strings.Split(tokens[i], "/")
            if strings.TrimSpace(req[0]) == "start time" {
                starttime = strings.TrimSpace(req[1])
            }
            if strings.TrimSpace(req[0]) == "end time" {
                endtime = strings.TrimSpace(req[1])
            }
            if strings.TrimSpace(req[0]) == "name" {
                name = strings.TrimSpace(req[1])
            }
            if strings.TrimSpace(req[0]) == "description" {
                description = strings.TrimSpace(req[1])
            }
            if strings.TrimSpace(req[0]) == "event id" {
                eventId = strings.TrimSpace(req[1])
            }
        }
        t := time.Now().Format(time.RFC3339)
        if eventId == "" {
            w.WriteHeader(http.StatusBadRequest)
            str := "Your request for modifying event should be as follows : modify calendar event .(Obligatory) event id / id . (Optional) start time / "+t+" . (Optional) end time / "+t+". (Optional) name / anything . (Optional) description / anything."
            e := map[string]string{"message":str}		
            json.NewEncoder(w).Encode(e)
            return	
        }


        event,err5 := srv.Events.Get(cal, eventId).Do()
        if err5 != nil {
			w.WriteHeader(422)
            e := map[string]string{"message":"No events with this id ."}		
            json.NewEncoder(w).Encode(e)
			return	
		  }
        if endtime != "" {
                event.End.DateTime = endtime
        }
        if starttime != "" {
                event.Start.DateTime = starttime

        }
        if description != "" {
           // description = event.Description
            event.Description = description
        }
        if name != "" {
           // summary =event.Summary 
           event.Summary = name
        }
        
        _,errr:=srv.Events.Update(cal,eventId,event).Do()
        if errr != nil {
          w.WriteHeader(http.StatusBadRequest)
          str := "Your request for modifying event should be as follows : modify calendar event .(Obligatory) event id / id . (Optional) start time / "+t+" . (Optional) end time / "+t+". (Optional) name / anything . (Optional) description / anything."
          e := map[string]string{"message":str}		
          json.NewEncoder(w).Encode(e)
        return	
        }
	
w.WriteHeader(http.StatusOK)
e := map[string]string{"message":"Event Modified Successfully ."}		
json.NewEncoder(w).Encode(e)
return

}
