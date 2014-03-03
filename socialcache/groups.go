package socialcache

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

// Returns a new groups list
func NewGroupsList() *GroupsList {
	return &GroupsList{byId: make(map[SteamId]*lockingGroup)}
}

func (list *GroupsList) Add(group *Group) {
	group.ChatMembers = &chatMemberList{mutex: &list.mutex, byId: make(map[SteamId]*lockingChatMember)}
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
	id = id.ChatToClan()
	return list.byId[id]
}

// Returns the number of groups
func (list *GroupsList) Count() int {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	return len(list.byId)
}

//used to fix the clan SteamId to a chat SteamId
func fixChatId(id SteamId) SteamId {
	if id.GetAccountType() == int32(EAccountType_Clan) {
		id = id.SetAccountInstance(uint32(Clan))
		id = id.SetAccountType(EAccountType_Chat)
	}
	return id
}

// A group
type Group struct {
	SteamId      SteamId
	Name         string
	Relationship EClanRelationship
	ChatMembers  *chatMemberList
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

func (g *lockingGroup) ChatMembers() *chatMemberList {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.group.ChatMembers
}

// A thread-safe Chat Member list which contains references to its predecessor and successor.
type chatMemberList struct {
	mutex *sync.RWMutex

	first *lockingChatMember
	last  *lockingChatMember

	byId map[SteamId]*lockingChatMember // fast lookup by ID
}

func (list *chatMemberList) Add(member *ChatMember) {
	lockmember := &lockingChatMember{member: member}
	list.mutex.Lock()
	defer list.mutex.Unlock()

	lockmember.mutex = list.mutex

	list.byId[member.SteamId] = lockmember

	if list.first == nil {
		list.first = lockmember
		list.last = lockmember
	} else {
		lockmember.prev = list.last
		list.last.next = lockmember
		list.last = lockmember
	}
}

func (list *chatMemberList) Remove(id SteamId) {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	lockmember := list.byId[id]
	if lockmember == nil {
		return
	}

	if list.first == lockmember {
		list.first = nil
	} else {
		lockmember.prev.next = lockmember.next
	}

	if list.last == lockmember {
		list.last = nil
	} else {
		lockmember.next.prev = lockmember.prev
	}
}

// Returns the first chat member in the chat member list or nil if the list is empty.
func (list *chatMemberList) First() *lockingChatMember {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	return list.first
}

// Returns the last chat member in the chat member list or nil if the list is empty.
func (list *chatMemberList) Last() *lockingChatMember {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	return list.last
}

// Returns the chat member by a SteamId or nil if there is no such chat member.
func (list *chatMemberList) ById(id SteamId) *lockingChatMember {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	return list.byId[id]
}

// Returns the number of groups
func (list *chatMemberList) Count() int {
	list.mutex.Lock()
	defer list.mutex.Unlock()
	return len(list.byId)
}

// A Chat Member
type ChatMember struct {
	SteamId     SteamId
	Permissions EChatPermission
	Rank        EClanRank
}

// A internal type for keeping things threadsafe
type lockingChatMember struct {
	mutex *sync.RWMutex

	prev *lockingChatMember
	next *lockingChatMember

	member *ChatMember
}

// Returns the previous element in the chat member list or nil if this chat member is the first in the list.
func (c *lockingChatMember) Prev() *lockingChatMember {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.prev
}

// Returns the next element in the chat member list or nil if this chat member is the last in the list.
func (c *lockingChatMember) Next() *lockingChatMember {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.next
}

func (c *lockingChatMember) SteamId() SteamId {
	// the steam id of a chat member never changes, so we don't need to lock here
	return c.member.SteamId
}

func (c *lockingChatMember) Permissions() EChatPermission {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.member.Permissions
}

func (c *lockingChatMember) Rank() EClanRank {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.member.Rank
}
