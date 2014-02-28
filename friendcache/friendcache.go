package friendcache

import (
	. "github.com/GamingRobot/steamgo/internal"
	. "github.com/GamingRobot/steamgo/steamid"
	"sync"
)

// Friend and group lists are implemented as doubly-linked lists for thread-safety.
// They can be iterated over like so:
// 	for friend := client.Social.Friends.First(); friend != nil; friend = friend.Next() {
// 		log.Println(friend.SteamId())
// 	}

// A thread-safe friend list which contains references to its predecessor and successor.
// It is mutable and will be changed by Social.
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

func (f *FriendsList) First() *lockingFriend {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.first
}

func (f *FriendsList) Last() *lockingFriend {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.last
}

func (f *FriendsList) ById(id SteamId) *lockingFriend {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.byId[id]
}

func (f *FriendsList) Count() int {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	return len(f.byId)
}

// A thread-safe friend in a friend list which contains references to its predecessor and successor.
// It is mutable and will be changed by Social.
type Friend struct {
	SteamId           SteamId
	Name              string
	Relationship      EFriendRelationship
	PersonaStateFlags EPersonaStateFlag
	GameAppId         uint64
}

type lockingFriend struct {
	mutex *sync.RWMutex

	prev *lockingFriend
	next *lockingFriend

	friend *Friend
}

func (f *lockingFriend) Next() *lockingFriend {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.next
}

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

// A thread-safe group list which contains references to its predecessor and successor.
// It is mutable and will be changed by Social.
type GroupsList struct {
	mutex sync.RWMutex

	first *lockingGroup
	last  *lockingGroup

	byId map[SteamId]*lockingGroup // fast lookup by ID
}

func NewGroupsList() *GroupsList {
	return &GroupsList{byId: make(map[SteamId]*lockingGroup)}
}

func (list *GroupsList) Add(group *Group) {
	lockgroup := &lockingGroup{group: group}
	list.mutex.Lock()
	defer list.mutex.Unlock()

	lockgroup.mutex = &list.mutex

	list.byId[group.SteamId] = lockgroup

	if list.first == nil {
		list.first = lockgroup
		list.last = lockgroup
	} else {
		lockgroup.prev = list.last
		list.last.next = lockgroup
		list.last = lockgroup
	}
}

func (list *GroupsList) Remove(id SteamId) {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	lockgroup := list.byId[id]
	if lockgroup == nil {
		return
	}

	if list.first == lockgroup {
		list.first = nil
	} else {
		lockgroup.prev.next = lockgroup.next
	}

	if list.last == lockgroup {
		list.last = nil
	} else {
		lockgroup.next.prev = lockgroup.prev
	}
}

// Returns the first group in the group list or nil if the list is empty.
func (list *GroupsList) First() *lockingGroup {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	return list.first
}

// Returns the last group in the group list or nil if the list is empty.
func (list *GroupsList) Last() *lockingGroup {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	return list.last
}

// Returns the group by a SteamId or nil if there is no such group.
func (list *GroupsList) ById(id SteamId) *lockingGroup {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	return list.byId[id]
}

func (list *GroupsList) Count() int {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	return len(list.byId)
}

// A thread-safe group in a group list which contains references to its predecessor and successor.
type Group struct {
	SteamId      SteamId
	Name         string
	Relationship EClanRelationship
}

type lockingGroup struct {
	mutex *sync.RWMutex

	prev *lockingGroup
	next *lockingGroup

	group *Group
}

// Returns the previous element in the group list or nil if this group is the first in the list.
func (g *lockingGroup) Prev() *lockingGroup {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.prev
}

// Returns the next element in the group list or nil if this group is the last in the list.
func (g *lockingGroup) Next() *lockingGroup {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.next
}

func (g *lockingGroup) SteamId() SteamId {
	// the steam id of a group never changes, so we don't need to lock here
	return g.group.SteamId
}

func (g *lockingGroup) Name() string {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.group.Name
}

func (g *lockingGroup) Relationship() EClanRelationship {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.group.Relationship
}
