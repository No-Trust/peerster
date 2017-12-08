// Structure for storing (rumor) messages

package main

import (
	"errors"
	"github.com/No-Trust/peerster/common"
	"sync"
)

// Messages, a set of messages
type Messages struct {
	M     map[string]map[uint32]RumorMessage
	mutex *sync.Mutex
}

// Retrieve a message with an origin and ID.
func (messages *Messages) Get(peerID string, messageID uint32) (*RumorMessage, bool) {
	// return the message from peerID of ID messageID, or nil if not present
	messages.mutex.Lock()
	if m, present := messages.M[peerID]; present {
		if val, present := m[messageID]; present {
			messages.mutex.Unlock()
			return &val, true
		}
	}
	messages.mutex.Unlock()
	return nil, false
}

// Check if the Messages contains a certain message.
func (messages *Messages) Contains(rumor *RumorMessage) bool {
	// Check if rumor is already stored in messages
	messages.mutex.Lock()
	if _, ok := messages.M[rumor.Origin]; !ok {
		// messages does not contain any messages from rumor.Origin
		messages.mutex.Unlock()
		return false
	} else if _, ok := messages.M[rumor.Origin][rumor.ID]; !ok {
		// messages does not contain this specific message from rumor.Origin
		messages.mutex.Unlock()
		return false
	} else {
		messages.mutex.Unlock()
		return true
	}
}

// Add a message to the set Messages.
func (messages *Messages) Add(rumor *RumorMessage) {
	// add a rumor
	messages.mutex.Lock()
	if _, ok := messages.M[rumor.Origin]; !ok {
		// initializing messages.M[rumor.Origin]
		messages.M[rumor.Origin] = make(map[uint32]RumorMessage)
	} else if _, ok := messages.M[rumor.Origin][rumor.ID]; ok {
		// a rumorMessage is already stored with this id
		common.CheckRead(errors.New("Trying to add already existing message to received messages"))
	}
	messages.M[rumor.Origin][rumor.ID] = *rumor
	messages.mutex.Unlock()
}
