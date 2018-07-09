package main

import (
	"log"
	"net/http"
	"github.com/gorilla/websocket"
	"encoding/json"
	"fmt"
	"github.com/johnny-debt/instascrap"
	"github.com/johnny-debt/social-networks-watcher/watcher"
	"time"
	"github.com/johnny-debt/social-networks-watcher/wsconnshub"
	"sort"
)

var wsConnectionsHub = wsconnshub.NewWsConnectionsHub()
var receiver = hashtagWatchingResultsReceiver{}
var list = watcher.NewWatchedObjectsList(receiver)
var maxIds = make(map[string]string)


func processCommand(conn *websocket.Conn) bool {
	// Read raw bytes from the connection
	messageType, payload, err := conn.ReadMessage()
	if err != nil {
		log.Printf("Payload read error: %v\n", err)
		return false
	}
	log.Printf("Payload read [%d]: %s\n", messageType, payload)
	// Parse raw bytes to the internal command struct
	var command ClientCommand
	err = json.Unmarshal(payload, &command);
	if err != nil {
		log.Printf("Payload parsing error: %v\n", err)
		return true
	}
	log.Printf("Command parsed: %v\n", command)
	if command.Command == "watch" {
		hashtag := watchedHashtag{slug: command.Hashtag}
		wsConnectionsHub.Subscribe(conn, wsconnshub.Topic(hashtag.Identifier()))
		list.Watch(hashtag)
	}
	return true
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
	// Configure WebSocket route
	http.HandleFunc("/ws", handleConnections)
	// Start the server on localhost port 8000 and log any errors
	log.Println("http server started on :8000")
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Upgrade initial GET request to a WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	// Make sure we close the connection when the function returns
	defer wsConnectionsHub.RemoveConnection(conn)

	// Register our new client
	wsConnectionsHub.AddConnection(conn)

	for {
		if !processCommand(conn) {
			break
		}
	}
}

type watchedHashtag struct {
	slug string
	maxId string
}

func (hashtag watchedHashtag) Identifier () string {
	return hashtag.slug
}

func (hashtag watchedHashtag) Items() []interface{} {
	medias, _:= instascrap.GetHashtagMedia(hashtag.slug)
	// Sort medias be ascending ID
	sort.SliceStable(medias, func(i, j int) bool {
		return medias[i].ID < medias[j].ID
	})

	var items []interface{}
	for _, v := range medias {
		maxId, exists := maxIds[hashtag.slug]
		if !exists {
			fmt.Printf("Hashtag maxId is empty string!\n")
		}
		if !exists || v.ID > maxId {
			maxIds[hashtag.slug] = v.ID
			items = append(items, v)
		} else {
			fmt.Printf("Item is skipped [%s]\n", v.ID)
		}
	}
	return items
}

func (hashtag watchedHashtag) GetInterval () time.Duration {
	return time.Second * 2
}

type hashtagWatchingResultsReceiver struct {

}

func (receiver hashtagWatchingResultsReceiver) Receive (item interface{}, object watcher.WatchedObject) {
	switch item.(type) {
	case instascrap.Media:
		fmt.Printf("Media #%s received for source %s\n", item.(instascrap.Media).ID, object.Identifier())
		// Send this media to all subscribers
		subscribers := wsConnectionsHub.GetSubscribedConnections(wsconnshub.Topic(object.Identifier()))
		for _, subscriber := range subscribers {
			conn := (*websocket.Conn)(subscriber)
			conn.WriteJSON(item)
		}
	default:
		fmt.Printf("Unknown object received (%T)\n", item)
	}
}