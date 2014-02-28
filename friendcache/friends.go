package friendcache

import (
	. "github.com/GamingRobot/steamgo/internal"
	. "github.com/GamingRobot/steamgo/steamid"
	"sync"
)

// Friend list is implemented as doubly-linked lists for thread-safety.
// They can be iterated over like so:
// 	for friend := client.Social.Friends.First(); friend != nil; friend = friend.Next() {
// 		log.Println(friend.SteamId())
// 	}

// A thread-safe friend list which contains references to its predecessor and successor.
type FriendsList struct {
	mutex sync.RWMutex

	first *lockingFriend
	last  *lockingFriend

	byId map[SteamId]*lockingFriend // fast lookup by ID
}

func NewFriendsList() *FriendsList {
	return &FriendsList{byId: make(map[SteamId]*lockingFriend)}
}

func (list *FriendsList) Add(friend *Friend) {
	lockfriend := &lockingFriend{friend: friend}
	list.mutex.Lock()
	defer list.mutex.Unlock()

	lockfriend.mutex = &list.mutex

	list.byId[friend.SteamId] = lockfriend

	if list.first == nil {
		list.first = lockfriend
		list.last = lockfriend
	} else {
		lockfriend.prev = list.last
		list.last.next = lockfriend
		list.last = lockfriend
	}
}

func (list *FriendsList) Remove(id SteamId) {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	lockfriend := list.byId[id]
	if lockfriend == nil {
		return
	}

	if list.first == lockfriend {
		list.first = nil
	} else {
		lockfriend.prev.next = lockfriend.next
	}

	if list.last == lockfriend {
		list.last = nil
	} else {
		lockfriend.next.prev = lockfriend.prev
	}
}

// Returns the first friend in the friend list or nil if the list is empty.
func (f *FriendsList) First() *lockingFriend {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.first
}

// Returns the last friend in the friend list or nil if the list is empty.
func (f *FriendsList) Last() *lockingFriend {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.last
}

// Returns the friend by a SteamId or nil if there is no such friend.
func (f *FriendsList) ById(id SteamId) *lockingFriend {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.byId[id]
}

// Returns the number of friends
func (f *FriendsList) Count() int {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return len(f.byId)
}

// A friend
type Friend struct {
	SteamId           SteamId
	Name              string
	Relationship      EFriendRelationship
	PersonaStateFlags EPersonaStateFlag
	GameAppId         uint64
}

// A internal type for keeping things threadsafe
type lockingFriend struct {
	mutex *sync.RWMutex

	prev *lockingFriend
	next *lockingFriend

	friend *Friend
}

// Returns the previous element in the friend list or nil if this friend is the first in the list.
func (f *lockingFriend) Next() *lockingFriend {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.next
}

// Returns the next element in the friend list or nil if this friend is the last in the list.
func (f *lockingFriend) Prev() *lockingFriend {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.prev
}

func (f *lockingFriend) SteamId() SteamId {
	// the steam id of a friend never changes, so we don't need to lock here
	return f.friend.SteamId
}

func (f *lockingFriend) Name() string {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.friend.Name
}

func (f *lockingFriend) Relationship() EFriendRelationship {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.friend.Relationship
}

func (f *lockingFriend) PersonaStateFlags() EPersonaStateFlag {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.friend.PersonaStateFlags
}

func (f *lockingFriend) GameAppId() uint64 {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.friend.GameAppId
}
