package main

import (
	"log"
	"net/http"
	"github.com/gorilla/websocket"
	"encoding/json"
	"fmt"
	"github.com/johnny-debt/instascrap"
	"github.com/johnny-debt/social-networks-watcher/watcher"
)

var clients = make(map[*websocket.Conn]bool) // connected clients
var subscriptions = make(map[string]map[*websocket.Conn]bool)
var broadcast = make(chan Message) // broadcast channel
var receiver = hashtagWatchingResultsReceiver{}
var list = watcher.NewWatchedObjectsList(receiver)

func processCommand(conn *websocket.Conn) {
	// Read raw bytes from the connection
	messageType, payload, err := conn.ReadMessage()
	if err != nil {
		log.Printf("Payload read error: %v\n", err)
		return
	}
	log.Printf("Payload read [%d]: %s\n", messageType, payload)
	// Parse raw bytes to the internal command struct
	var command ClientCommand
	err = json.Unmarshal(payload, &command);
	if err != nil {
		log.Printf("Payload parsing error: %v\n", err)
		return
	}
	log.Printf("Command parsed: %v\n", command)
	hashtag := watchedHashtag{slug: command.Hashtag}
	list.Watch(hashtag)
}

// Configure the upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Message interface {}

// Define our message object
type ClientCommand struct {
	Command string `json:"command"` // "watch" or "unwatch"
	Hashtag string `json:"hashtag"`
}

func main() {
	// Configure websocket route
	http.HandleFunc("/ws", handleConnections)

	// Start listening for incoming chat messages
	go handleMessages()

	// Start the server on localhost port 8000 and log any errors
	log.Println("http server started on :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a WebSocket
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer ws.Close()

	// Register our new client
	clients[ws] = true

	for {
		// Read in a new message as JSON and map it to a Message object
		//err := ws.ReadJSON(&msg)
		//if err != nil {
		//	log.Printf("error: %v (ws.ReadJSON)", err)
		//	delete(clients, ws)
		//	break
		//}
		// Send the newly received message to the broadcast channel
		processCommand(ws)
	}
}

func handleMessages() {
	for {
		// Grab the next message from the broadcast channel
		msg := <-broadcast
		// Send it out to every client that is currently connected
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("error: %v (client.WriteJSON)", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}

type watchedHashtag struct {
	slug string
}

func (hashtag watchedHashtag) Identifier () string {
	return hashtag.slug
}

func (hashtag watchedHashtag) items() []interface{} {
	medias, _:= instascrap.GetHashtagMedia("beer")
	items := make([]interface{}, len(medias))
	for i, v := range medias {
		items[i] = v
	}
	return items
}

type hashtagWatchingResultsReceiver struct {

}

func (receiver hashtagWatchingResultsReceiver) receive (item interface{}, object watcher.WatchedObject) {
	switch item.(type) {
	case instascrap.Media:
		fmt.Printf("Media #%s received for source %s\n", item.(instascrap.Media).ID, object.Identifier())
	default:
		fmt.Printf("Unknown object received (%T)\n", item)
	}
}