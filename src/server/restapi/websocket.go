/**
Copyright (c) 2015 The ConnectorDB Contributors (see AUTHORS)
Licensed under the MIT license.
**/
package restapi

import (
	"connectordb/authoperator"
	"connectordb/datastream"
	"connectordb/messenger"
	"connectordb/query"
	"errors"
	"io"
	"net/http"
	"server/restapi/restcore"
	"server/webcore"
	"sync"
	"sync/atomic"
	"time"

	"github.com/connectordb/pipescript"
	"github.com/gorilla/websocket"
	"github.com/nats-io/nats"

	log "github.com/Sirupsen/logrus"
)

const (
	//The max size of a websocket message
	messageSizeLimit = 1 * restcore.Mb

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

	//websocketWaitGroup is the WaitGroup of websockets that are currently open
	websocketWaitGroup = sync.WaitGroup{}
)

type Subscription struct {
	sync.Mutex //The transform mutex

	nats *nats.Subscription //The nats subscription

	transform map[string]*pipescript.Script //the transforms associated with the subscription - this allows us to run transforms on the data!
}

func NewSubscription(subs *nats.Subscription) *Subscription {
	return &Subscription{
		nats:      subs,
		transform: make(map[string]*pipescript.Script),
	}
}

//Close shuts down the subscription
func (s *Subscription) Close() {
	s.Lock()
	defer s.Unlock()
	s.nats.Unsubscribe()
}

//Size is the number of subscriptions to the stream (using different transforms)
func (s *Subscription) Size() int {
	s.Lock()
	defer s.Unlock()
	return len(s.transform)
}

//Add a transform subscription to the string
func (s *Subscription) AddTransform(transform string) (err error) {
	if _, ok := s.transform[transform]; ok {
		return errors.New("Subscription to the transform already exists")
	}

	//First, attempt to generate the transform
	var t *pipescript.Script
	if transform != "" {
		t, err = pipescript.Parse(transform)
		if err != nil {
			return err
		}
	}

	s.Lock()
	s.transform[transform] = t
	s.Unlock()

	return nil
}

//RemTransform deletes a transform from the subscriptions
func (s *Subscription) RemTransform(transform string) (err error) {
	s.Lock()
	delete(s.transform, transform)
	s.Unlock()
	return nil
}

//WebsocketConnection is the general connection with a websocket that is run.
//Loosely based on github.com/gorilla/websocket/blob/master/examples/chat/conn.go
//No need for mutex because only reader reads and implements commands
type WebsocketConnection struct {
	ws *websocket.Conn

	subscriptions map[string]*Subscription

	c chan messenger.Message

	logger *log.Entry //logrus uses a mutex internally
	o      *authoperator.AuthOperator
}

//NewWebsocketConnection creates a new websocket connection based on the operators and stuff
func NewWebsocketConnection(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (*WebsocketConnection, error) {

	ws, err := upgrader.Upgrade(writer, request, nil)
	if err != nil {
		logger.Errorln(err)
		return nil, err
	}

	ws.SetReadLimit(messageSizeLimit)

	return &WebsocketConnection{ws, make(map[string]*Subscription), make(chan messenger.Message, messageBuffer), logger, o}, nil
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

//CheckSubscriptions checks the current subscriptions, making sure that there are no issues (such as lost permissions or deleted streams)
//which trigger unsubscribes
func (c *WebsocketConnection) CheckSubscriptions() error {
	for stream, val := range c.subscriptions {
		if _, err := c.o.ReadStream(stream); err != nil {
			c.logger.Warnf("Invalidated: %s", stream)
			val.Close()
			delete(c.subscriptions, stream)
		}
	}
	return nil
}

//Insert a datapoint using the websocket
func (c *WebsocketConnection) Insert(ws *websocketCommand) {
	logger := c.logger.WithFields(log.Fields{"cmd": "insert", "arg": ws.Arg})
	logger.Debugln("-> insert ", len(ws.D), "dp")
	err := c.o.InsertStream(ws.Arg, ws.D, true)
	if err != nil {
		//TODO: Notify user of insert failure
		logger.Warn(err.Error())
	} else {
		atomic.AddUint32(&webcore.StatsInserts, uint32(len(ws.D)))
	}
}

//Subscribe to the given data stream
func (c *WebsocketConnection) Subscribe(s, transform string) {
	logger := c.logger.WithFields(log.Fields{"cmd": "subscribe", "arg": s})

	//Next check if nats is subscribed
	if _, ok := c.subscriptions[s]; !ok {
		subs, err := c.o.Subscribe(s, c.c)
		if err != nil {
			logger.Warningln(err)
		} else {

			logger.Debugln("Initializing subscription")
			c.subscriptions[s] = NewSubscription(subs)
		}
	}

	err := c.subscriptions[s].AddTransform(transform)
	if err != nil {
		logger.Warningln(err)
	}
}

//Unsubscribe from the given data stream
func (c *WebsocketConnection) Unsubscribe(s, transform string) {
	logger := c.logger.WithFields(log.Fields{"cmd": "unsubscribe", "arg": s})
	if val, ok := c.subscriptions[s]; ok {
		val.RemTransform(transform)
		if val.Size() == 0 {
			logger.Debugln("stop subscription")
			val.Close()
			delete(c.subscriptions, s)
		} else {
			logger.Debugln()
		}

	} else {
		logger.Warningln("subscription DNE")
	}
}

//UnsubscribeAll from all streams of data
func (c *WebsocketConnection) UnsubscribeAll() {
	for key, val := range c.subscriptions {
		c.logger.Debugf("Unsubscribe: %s", key)
		val.Close()
	}
	c.subscriptions = make(map[string]*Subscription)
}

//A command is a cmd and the arg operation
type websocketCommand struct {
	Cmd       string
	Arg       string
	Transform string //Allows subscribing with a transform

	D []datastream.Datapoint //If the command is "insert", it needs an additional datapoint
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
			c.Subscribe(cmd.Arg, cmd.Transform)
		case "unsubscribe":
			c.Unsubscribe(cmd.Arg, cmd.Transform)
		case "unsubscribe_all":
			c.UnsubscribeAll()
		}
	}
	//Since the reader is exiting, notify the writer to send close message
	readmessenger <- webSocketClosedNonClean
}

func (c *WebsocketConnection) updateDeadline(messageCode int, message string) error {
	c.ws.SetWriteDeadline(time.Now().Add(writeWait))
	return c.ws.WriteMessage(messageCode, []byte(message))
}

// returns false if we failed, we should quit
func (c *WebsocketConnection) processDatapoint(datapoint messenger.Message) error {

	logger := c.logger.WithFields(log.Fields{"stream": datapoint.Stream})

	//Now loop through all transforms for the datapoint array
	subs, ok := c.subscriptions[datapoint.Stream]
	if !ok {
		return c.write(datapoint)
	}

	subs.Lock()
	defer subs.Unlock()

	for transform, tf := range subs.transform {
		if transform == "" {
			logger.Debugln("<- send")
			if err := c.write(datapoint); err != nil {
				return err
			}
			continue
		}

		datapointArray, err := query.TransformArray(tf, &datapoint.Data)
		logger.Debugf("<- send %s", transform)
		if err != nil || datapointArray.Length() <= 0 {
			continue
		}

		message := messenger.Message{
			datapoint.Stream,
			transform,
			*datapointArray,
		}

		if err := c.write(message); err != nil {
			return err
		}
	}

	return c.write(datapoint)
}

//RunWriter writes the subscription data as well as the heartbeat pings.
func (c *WebsocketConnection) RunWriter(readmessenger chan string, exitchan chan bool) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		exitchan <- true

	}()

	for {
		select {
		case datapoint, ok := <-c.c:
			if !ok {
				c.updateDeadline(websocket.CloseMessage, "")
				return
			}

			if c.processDatapoint(datapoint) != nil {
				return
			}

		case <-ticker.C:
			//This is the ping timer - ping messages are sent here
			if c.updateDeadline(websocket.PingMessage, "") != nil {
				return
			}

			//Now, let's make sure that the active subscriptions are still valid
			c.CheckSubscriptions()

		case msg := <-readmessenger:
			switch msg {
			case webSocketClosed:
				return
			case webSocketClosedNonClean:
				c.updateDeadline(websocket.CloseMessage, "")
				return
			default:
				c.updateDeadline(websocket.TextMessage, msg)
			}

		case <-webcore.ShutdownChannel:
			webcore.ShutdownChannel <- true
			c.updateDeadline(websocket.CloseMessage, "")
			return
		}
	}
}

//Run the websocket operations
func (c *WebsocketConnection) Run() error {
	c.logger.Debugln("Running websocket...")
	websocketWaitGroup.Add(1)

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
	websocketWaitGroup.Done()
	return nil
}

//RunWebsocket runs the websocket handler
func RunWebsocket(o *authoperator.AuthOperator, writer http.ResponseWriter, request *http.Request, logger *log.Entry) (int, string) {
	conn, err := NewWebsocketConnection(o, writer, request, logger)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return 3, err.Error()
	}
	defer conn.Close()
	err = conn.Run()
	if err != nil {
		return 2, err.Error()
	}
	return 0, "Websocket closed"
}
