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
    username, _, _ := reader.ReadLine()

    fmt.Print("Enter admin email: ")
    email, _, _ := reader.ReadLine()

    fmt.Print("Enter admin password: ")
    pass, _, _ := reader.ReadLine()

    userid, err := users.CreateUser(string(username), string(email), string(pass))

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
