package friendcache

import (
	. "github.com/GamingRobot/steamgo/internal"
	. "github.com/GamingRobot/steamgo/steamid"
	"sync"
)

// Group list is implemented as doubly-linked lists for thread-safety.
// They can be iterated over like so:
// 	for group := client.Social.Groups.First(); group != nil; group = group.Next() {
// 		log.Println(group.SteamId())
// 	}

// A thread-safe group list which contains references to its predecessor and successor.
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

// Returns the number of groups
func (list *GroupsList) Count() int {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	return len(list.byId)
}

// A group
type Group struct {
	SteamId      SteamId
	Name         string
	Relationship EClanRelationship
}

// A internal type for keeping things threadsafe
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
