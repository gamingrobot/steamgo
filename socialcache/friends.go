package socialcache

import (
	. "github.com/GamingRobot/steamgo/internal"
	. "github.com/GamingRobot/steamgo/steamid"
	"sync"
)

// Friend list is implemented as doubly-linked lists for thread-safety.
// They can be iterated over like so:
// 	for id, friend := range client.Social.Friends.GetCopy() {
// 		log.Println(id, friend.Name)
// 	}

// A thread-safe friend list
type FriendsList struct {
	mutex sync.RWMutex
	byId  map[SteamId]*Friend
}

// Returns a new friends list
func NewFriendsList() *FriendsList {
	return &FriendsList{byId: make(map[SteamId]*Friend)}
}

// Adds a friend to the friend list
func (list *FriendsList) Add(friend *Friend) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	list.byId[friend.SteamId] = friend
}

// Removes a friend from the friend list
func (list *FriendsList) Remove(id SteamId) {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	delete(list.byId, id)
}

// Returns a copy of the friends map
func (list *FriendsList) GetCopy() map[SteamId]Friend {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	flist := make(map[SteamId]Friend)
	for key, friend := range list.byId {
		flist[key] = *friend
	}
	return flist
}

// Returns a copy of the friend of a given SteamId
func (list *FriendsList) ById(id SteamId) Friend {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	return *list.byId[id]
}

// Returns the number of friends
func (list *FriendsList) Count() int {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	return len(list.byId)
}

// A Friend
type Friend struct {
	SteamId           SteamId
	Name              string
	Relationship      EFriendRelationship
	PersonaStateFlags EPersonaStateFlag
	GameAppId         uint64
}
