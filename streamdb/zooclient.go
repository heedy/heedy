package main

import (
    "fmt"
    "time"
    "os"
    "strconv"
    "strings"
    "errors"
    "github.com/samuel/go-zookeeper/zk"
)

type ConnectorClient struct {
    zoo *zk.Conn
}

func (cc *ConnectorClient) Close() {
    cc.zoo.Close()
    fmt.Printf("Done")
}

/*
The only needed argument is really the name - the other 2 can be "" and 0, in which case it
uses the computer's predefined hostname.
If the service does not bind to a port, port should be set to 0
*/
func (cc *ConnectorClient) RegisterMe(name string,hostname string,port int) (err error) {
    acl := zk.WorldACL(zk.PermAll)  //This defines the zookeeper node permissions.
    basepath := "/connector/"+name  //The base path that needs to exist - it defines the program "type"
    if (hostname=="") {             //Make sure that hostname exists - use the default hostname if not set
        hostname,err = os.Hostname()
        if (err!=nil) {
            return err
        }
    }
    nodename := "/"+hostname        //The nodename is of format hostname:port:pid, but if no port used, hostname:pid.
    if (port > 0) {
        nodename = nodename +":"+strconv.Itoa(port)
    }
    nodename = nodename +":"+ strconv.Itoa(os.Getpid())

    exists,_,err := cc.zoo.Exists(basepath) //Make sure that the basepath exists - and create it if it doesn't
    if (err != nil) {
        return err
    }
    if (exists == false) {
        cc.zoo.Create(basepath,[]byte(""),int32(0),acl)
    }

    //Finally, create the ephemeral node associated with this specific client.
    _,err = cc.zoo.Create(basepath+nodename,[]byte(""),int32(zk.FlagEphemeral),acl)

    return err
}

/*
Given name of a server, returns the address of the first server which has that name.
*/
func (cc *ConnectorClient) ServerAddress(name string) (hostname string,err error) {
    children,_,err := cc.zoo.Children("/connector/"+name)   //Get the list of children of given name node
    if (err != nil) {       //Make sure the name node exists
        return "",err
    }
    if (len(children)<1) {  //Make sure that there is at least one server connected
        return "",errors.New("No server found")
    }
    //For now, the implementation only cares about one server: the first one
    i:=strings.LastIndex(children[0],":")   //We want to strip the pid from the server address
    if (i<0) {
        return "",errors.New("Malformed server data")
    }
    return children[0][0:i],nil
}

/*
Given address of zookeeper server, connects and maintains a connection with the zookeeper server.
*/
func Connect(loc string) (cc *ConnectorClient,err error) {
    conn, _, err := zk.Connect([]string{loc}, time.Second)
    if err != nil {
        return nil,err
    }
    cc = &ConnectorClient{conn}
    return cc,nil
}

func main() {
    cc,err := Connect("localhost:1337")
    defer cc.Close()

    err = cc.RegisterMe("gotest","localhost",0)

    host,err := cc.ServerAddress("mongodb")
    fmt.Printf("Server address: %s %s\n",host,err)


    if (err != nil) {
        fmt.Printf("Failed")
    } else {
        fmt.Printf("cool")
    }
}
