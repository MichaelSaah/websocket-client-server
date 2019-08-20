// heavily based on https://github.com/gorilla/websocket/tree/master/examples/echo

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

// specify a custom host:port with the addr flag if you don't like the default
var addr = flag.String("addr", "localhost:8888", "http service address")

var upgrader = websocket.Upgrader{} // use default options

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
		log.Printf("recv: %s", message)
	}
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/", listen)

	// run the http server in a goroutine
	// this allows the program to be the client _and_ the server
	go http.ListenAndServe(*addr, nil)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	for i := 0; i < 10; i++ {
		err = c.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("test message %d", i)))
		if err != nil {
			log.Println("write:", err)
			return
		}

		time.Sleep(1 * time.Second)
	}
}
