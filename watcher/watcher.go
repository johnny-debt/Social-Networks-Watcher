package watcher

import (
	"fmt"
	"time"
)

type WatchedObject interface {
	Identifier() string
	items() []interface{}
}

type WatchingResultsReceiver interface {
	receive(interface{}, WatchedObject)
}

// Each watched object has a map with Watched Object Identifier as a key and Subscribed users number as a value.
type WatchedObjectsList struct {
	identifiers map[string]int
	watchers map[string]chan bool
	receiver WatchingResultsReceiver
}

func NewWatchedObjectsList (receiver WatchingResultsReceiver) WatchedObjectsList {
	list := WatchedObjectsList{}
	list.identifiers = map[string]int {}
	list.watchers = map[string]chan bool {}
	list.receiver = receiver
	return list
}

func watcher(stop chan bool, object WatchedObject, receiver WatchingResultsReceiver) {
	for {
		select {
		default:
			fmt.Printf("Watcher of [%v] has new iteration\n", object.Identifier())
			items := object.items()
			for _, item := range items {
				receiver.receive(item, object)
			}
		case <-stop:
			fmt.Printf("Signal stop received for watcher of [%v]\n", object.Identifier())
			return
		}
		time.Sleep(time.Millisecond * 500)
	}
}

func (list *WatchedObjectsList) Watch(object WatchedObject) {
	fmt.Printf("Start watching %s\n", object.Identifier())
	list.identifiers[object.Identifier()]++
	list.runWatcher(object)
}


func (list *WatchedObjectsList) Unwatch(object WatchedObject) {
	fmt.Printf("Stop watching %s\n", object.Identifier())
	channel := list.watchers[object.Identifier()]
	close(channel)
}

func (list *WatchedObjectsList) runWatcher (object WatchedObject) {
	fmt.Printf("Run watcher for %s\n", object.Identifier())
	channel := make(chan bool)
	list.watchers[object.Identifier()] = channel
	go watcher(channel, object, list.receiver)
}