package main


type User struct {
        Unique string
        Name string
        Email string 
        Password string
        CalendarId string  
        Memos  []Memo   
        Keys   []Key   
        
}


type Memo struct{
        Id   uint64
        Name string
        Content string
}

type Key struct{
        Id   uint64
        Name string
        Value string 
}
