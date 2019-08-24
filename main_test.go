// heavily based on https://github.com/gorilla/websocket/tree/master/examples/echo

package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

var addr = "localhost:8888"

var upgrader = websocket.Upgrader{} // use default options

var responseChan = make(chan []byte)

// handle and upgrade incoming websocket requests
func listen(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}

		responseChan <- message
	}
}

// dummy route to establish that server is ready
func ready(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "")
}

func startServer(host string) {
	http.HandleFunc("/", listen)
	http.HandleFunc("/ready", ready)

	// run the http server in a goroutine
	// this allows the program to be the client _and_ the server
	go http.ListenAndServe(addr, nil)

	// server can take a little time to be ready to serve
	// block in this function until we get a 200 back
	for {
		resp, err := http.Get(strings.Join([]string{"http://", host, "/ready"}, ""))
		if err != nil {
			continue
		}

		if resp.StatusCode == 200 {
			// ready to roll
			break
		}
	}
}

func TestCase(t *testing.T) {
	startServer(addr)

	u := url.URL{Scheme: "ws", Host: addr, Path: "/"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}

	for i := 1; i <= 10; i++ {
		message := []byte(fmt.Sprintf("test message %d/10", i))
		err = c.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Println("write:", err)
			return
		}

		// server will pass the messages it receive back to the test case via the responseChan
		received_message := <-responseChan

		// we got the message! assert that it was what we expected
		log.Printf("server received: \"%s\"", received_message)
		if bytes.Compare(message, received_message) != 0 {
			t.Error(fmt.Sprintf("we sent %s, but the server got %s", message, received_message))
		}

		time.Sleep(500 * time.Millisecond)
	}
}
