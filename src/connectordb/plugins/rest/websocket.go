package rest

import (
	"connectordb/streamdb/operator"
	"io"
	"net/http"
	"time"

	"github.com/apcera/nats"
	"github.com/gorilla/websocket"

	log "github.com/Sirupsen/logrus"
)

const (
	//The max size of a websocket message
	messageSizeLimit = 1 * Mb

	//The time allowed to write a message
	writeWait = 2 * time.Second

	//Ping pong stuff - making sure that the connection still exists
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10

	//The number of messages to buffer
	messageBuffer = 3

	webSocketClosed         = "EXIT"
	webSocketClosedNonClean = "@EXIT"
)

//The websocket upgrader
var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		// Allow from all origins
		CheckOrigin: func(r *http.Request) bool { return true },
	}
)

//WebsocketConnection is the general connection with a websocket that is run.
//Loosely based on github.com/gorilla/websocket/blob/master/examples/chat/conn.go
//No need for mutex because only reader reads and implements commands
type WebsocketConnection struct {
	ws *websocket.Conn

	subscriptions map[string]*nats.Subscription

	c chan operator.Message

	logger *log.Entry //logrus uses a mutex internally
	o      operator.Operator
}

//NewWebsocketConnection creates a new websocket connection based on the operators and stuff
func NewWebsocketConnection(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (*WebsocketConnection, error) {
	logger = logger.WithField("op", "ws")

	ws, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		logger.Errorln(err)
		return nil, err
	}

	ws.SetReadLimit(messageSizeLimit)

	return &WebsocketConnection{ws, make(map[string]*nats.Subscription), make(chan operator.Message, messageBuffer), logger, o}, nil
}

func (c *WebsocketConnection) write(obj interface{}) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteJSON(obj)
}

//Close the websocket connection
func (c *WebsocketConnection) Close() {
	c.UnsubscribeAll()
	close(c.c)
	c.ws.Close()
	c.logger.WithField("cmd", "close").Debugln()
}

//Insert a datapoint using the websocket
func (c *WebsocketConnection) Insert(ws *websocketCommand) {
	logger := c.logger.WithFields(log.Fields{"cmd": "insert", "arg": ws.Arg})
	logger.Debugln("Inserting", len(ws.D), "dp")
	err := c.o.InsertStream(ws.Arg, ws.D)
	if err != nil {
		//TODO: Notify user of insert failure
		logger.Warn(err.Error())
	}
}

//Subscribe to the given data stream
func (c *WebsocketConnection) Subscribe(s string) {
	logger := c.logger.WithFields(log.Fields{"cmd": "subscribe", "arg": s})
	if _, ok := c.subscriptions[s]; !ok {
		subs, err := c.o.Subscribe(s, c.c)
		if err != nil {
			logger.Warningln(err)
		} else {
			logger.Debugln()
			c.subscriptions[s] = subs
		}
	} else {
		logger.Warningln("Already subscribed")
	}
}

//Unsubscribe from the given data stream
func (c *WebsocketConnection) Unsubscribe(s string) {
	logger := c.logger.WithFields(log.Fields{"cmd": "unsubscribe", "arg": s})
	if val, ok := c.subscriptions[s]; ok {
		logger.Debugln()
		val.Unsubscribe()
		delete(c.subscriptions, s)
	} else {
		logger.Warningln("subscription DNE")
	}
}

//UnsubscribeAll from all streams of data
func (c *WebsocketConnection) UnsubscribeAll() {
	c.logger.WithField("cmd", "unsubscribeALL").Debugln()
	for _, val := range c.subscriptions {
		val.Unsubscribe()
	}
	c.subscriptions = make(map[string]*nats.Subscription)
}

//A command is a cmd and the arg operation
type websocketCommand struct {
	Cmd string
	Arg string
	D   []operator.Datapoint //If the command is "insert", it needs an additional datapoint
}

//RunReader runs the reading routine. It also maps the commands to actual subscriptions
func (c *WebsocketConnection) RunReader(readmessenger chan string) {

	//Set up the heartbeat reader(makes sure that sockets are alive)
	c.ws.SetReadDeadline(time.Now().Add(pongWait))
	c.ws.SetPongHandler(func(string) error {
		//c.logger.WithField("cmd", "PingPong").Debugln()
		c.ws.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	var cmd websocketCommand
	for {
		err := c.ws.ReadJSON(&cmd)
		if err != nil {
			if err == io.EOF {
				readmessenger <- webSocketClosed
				return //On EOF, do nothing - it is just a close
			}
			c.logger.Warningln(err)
			break
		}
		switch cmd.Cmd {
		default:
			c.logger.Warningln("Command not recognized:", cmd.Cmd)
			//Do nothing - the command is not recognized
		case "insert":
			c.Insert(&cmd)
		case "subscribe":
			c.Subscribe(cmd.Arg)
		case "unsubscribe":
			c.Unsubscribe(cmd.Arg)
		case "unsubscribe_all":
			c.UnsubscribeAll()
		}
	}
	//Since the reader is exiting, notify the writer to send close message
	readmessenger <- webSocketClosedNonClean
}

//RunWriter writes the subscription data as well as the heartbeat pings.
func (c *WebsocketConnection) RunWriter(readmessenger chan string, exitchan chan bool) {
	ticker := time.NewTicker(pingPeriod)
	defer ticker.Stop()
loop:
	for {
		select {
		case dp, ok := <-c.c:
			if !ok {
				c.ws.SetWriteDeadline(time.Now().Add(writeWait))
				c.ws.WriteMessage(websocket.CloseMessage, []byte{})
				break loop
			}
			c.logger.WithFields(log.Fields{"cmd": "MSG", "arg": dp.Stream}).Debugln()
			if err := c.write(dp); err != nil {
				break loop
			}
		case <-ticker.C:
			//c.logger.WithField("cmd", "PING").Debugln()
			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.ws.WriteMessage(websocket.PingMessage, []byte{}); err != nil {
				break loop
			}
		case msg := <-readmessenger:
			if msg == webSocketClosed {
				break loop
			} else if msg == webSocketClosedNonClean {
				c.ws.SetWriteDeadline(time.Now().Add(writeWait))
				c.ws.WriteMessage(websocket.CloseMessage, []byte{})
				break loop
			}
			c.ws.SetWriteDeadline(time.Now().Add(writeWait))
			c.ws.WriteMessage(websocket.TextMessage, []byte(msg))
		}
	}
	exitchan <- true
}

//Run the websocket operations
func (c *WebsocketConnection) Run() error {
	c.logger.Debugln("Running websocket...")

	//The reader can communicate with the writer through the channel
	msgchn := make(chan string, 1)
	exitchan := make(chan bool, 1)
	go c.RunWriter(msgchn, exitchan)
	c.RunReader(msgchn)
	//Wait for writer to exit, or for the exit timeout to happen
	go func() {
		time.Sleep(writeWait)
		exitchan <- false
	}()

	if !<-exitchan {
		c.logger.Error("writer exit timeout")
	}
	return nil
}

//RunWebsocket runs the websocket handler
func RunWebsocket(o operator.Operator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) error {
	conn, err := NewWebsocketConnection(o, writer, request, logger)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return err
	}
	defer conn.Close()

	return conn.Run()
}
