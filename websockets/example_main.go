package main

import (
	"fmt"
	"time"

	"github.com/fym201/bigo"
	"github.com/fym201/bigo/websockets"
	"github.com/gorilla/websocket"
)

type Message struct {
	A string
}

func main() {
	m := bigo.Classic()

	m.Get("/sockets", websockets.JSON(Message{}), func(receiver <-chan *Message, sender chan<- *Message, done <-chan bool, disconnect chan<- int, errorChannel <-chan error) {
		ticker := time.After(30 * time.Minute)
		for {
			select {
			case msg := <-receiver:
				// here we simply echo the received message to the sender for demonstration purposes
				// In your app, collect the senders of different clients and do something useful with them
				fmt.Println(msg)
				sender <- msg
			case <-ticker:
				// This will close the connection after 30 minutes no matter what
				// To demonstrate use of the disconnect channel
				// You can use close codes according to RFC 6455
				disconnect <- websocket.CloseNormalClosure
			case <-done:
				// the client disconnected, so you should return / break if the done channel gets sent a message
				return
			case err := <-errorChannel:
				fmt.Sprintln(err)
				// Uh oh, we received an error. This will happen before a close if the client did not disconnect regularly.
				// Maybe useful if you want to store statistics
			}
		}
	})

	m.Run()
}
