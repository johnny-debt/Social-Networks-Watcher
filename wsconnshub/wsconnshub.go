package wsconnshub

import (
	"github.com/gorilla/websocket"
	"log"
	"fmt"
)

/* 
 * Provides functions to work with WebSocket connectionsMap hub
 * Hub allows:
 * - Register / unregister connection
 * - Get list of all currently existing connections
 * - Subscribe / unsubscribe connection to the Topic
 * - Get list of connection that are subscribed to the certain Topic
 */

type Connection *websocket.Conn
type Topic string
type connectionsMap map[Connection]bool


type WsConnectionsHub struct {
	connectionsMap   connectionsMap
	subscriptionsMap map[Topic]connectionsMap
}

func NewWsConnectionsHub() WsConnectionsHub {
	hub := WsConnectionsHub{}
	hub.connectionsMap = make(connectionsMap)
	hub.subscriptionsMap = make(map[Topic]connectionsMap)
	return hub
}

func (hub WsConnectionsHub) GetAllConnections () []Connection {
	return connectionsMapToSlice(hub.connectionsMap)
}

func (hub WsConnectionsHub) AddConnection (connection Connection) {
	hub.connectionsMap[connection] = true
}

func (hub WsConnectionsHub) RemoveConnection (connection Connection)  {
	delete(hub.connectionsMap, connection)
	// @TODO Unsubscribe connection from
}

func (hub WsConnectionsHub) Subscribe (connection Connection, topic Topic) {
	if _, connExists := hub.connectionsMap[connection]; !connExists {
		log.Fatal("Can not subscribe unknown connection to the Topic")
		return
	}
	if _, topicExists := hub.subscriptionsMap[topic]; !topicExists {
		hub.subscriptionsMap[topic] = make(connectionsMap)
	}
	hub.subscriptionsMap[topic][connection] = true
}

func (hub WsConnectionsHub) Unsubscribe (connection Connection, topic Topic) {
	if _, topicExists := hub.subscriptionsMap[topic]; !topicExists {
		log.Fatal(fmt.Sprintf("Can not unsubscribe from unknown Topic [%s]", topic))
	}
	delete(hub.subscriptionsMap[topic], connection)
}

func (hub WsConnectionsHub) GetSubscribedConnections (topic Topic) []Connection {
	connectionsMap, topicExists := hub.subscriptionsMap[topic]
	if !topicExists {
		log.Fatal(fmt.Sprintf("Can not get subscribers for every Topic from unknown Topic [%s]", topic))
	}
	return connectionsMapToSlice(connectionsMap)
}

func connectionsMapToSlice (connectionsMap connectionsMap) []Connection {
	connections := make([]Connection, 0, len(connectionsMap))
	for k := range connectionsMap {
		connections = append(connections, k)
	}
	return connections
}