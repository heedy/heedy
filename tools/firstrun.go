package main

import(
    "streamdb/users"
    "flag"
    "bufio"
    "fmt"
    "os"
)




func main() {
    flag.Parse()
    
    
    users, err := users.NewSqliteUserDatabase("production.sqlite")

    if err != nil {
        panic("Cannot open user database")
    }
    
    reader := bufio.NewReader(os.Stdin)
    
    fmt.Print("Enter admin username: ")
    username, _ := reader.ReadString('\n')
    
    fmt.Print("Enter admin email: ")
    email, _ := reader.ReadString('\n')
    
    fmt.Print("Enter admin password: ")
    pass, _ := reader.ReadString('\n')
    
    userid, err := users.CreateUser(username, email, pass)
    
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    fmt.Printf("New user created with id: %v", userid)
    
    fmt.Println("Trying to grant admin rights...")
    
    usr, err := users.ReadUserById(userid)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    
    usr.Admin = true
    
    err = users.UpdateUser(usr)
    
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
}
