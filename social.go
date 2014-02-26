package steam

import (
	"bytes"
	"code.google.com/p/goprotobuf/proto"
	"encoding/binary"
	. "github.com/GamingRobot/steamgo/internal"
	. "github.com/GamingRobot/steamgo/steamid"
	"sync"
)

// Provides access to social aspects of Steam.
//
// Friend and group lists are implemented as doubly-linked lists for thread-safety.
// They can be iterated over like so:
// 	for friend := client.Social.Friends.First(); friend != nil; friend = friend.Next() {
// 		log.Println(friend.SteamId())
// 	}
type Social struct {
	mutex sync.RWMutex

	personaState EPersonaState

	Friends *FriendsList
	Groups  *GroupsList

	client *Client
}

func newSocial(client *Client) *Social {
	return &Social{
		Friends: &FriendsList{byId: make(map[SteamId]*Friend)},
		Groups:  &GroupsList{byId: make(map[SteamId]*Group)},
		client:  client,
	}
}

func (s *Social) SetPersonaState(state EPersonaState) {
	s.client.Write(NewClientMsgProtobuf(EMsg_ClientChangeStatus, &CMsgClientChangeStatus{
		PersonaState: proto.Uint32(uint32(state)),
	}))
}

func (s *Social) SendMessage(to SteamId, entryType EChatEntryType, message string) {
	if to.GetAccountType() == int32(EAccountType_Individual) || to.GetAccountType() == int32(EAccountType_ConsoleUser) {
		s.SendChatMessage(to, entryType, message)
	} else if to.GetAccountType() == int32(EAccountType_Clan) || to.GetAccountType() == int32(EAccountType_Chat) {
		s.SendChatRoomMessage(to, entryType, message)
	}
}

// Sends a chat message to the given friend.
func (s *Social) SendChatMessage(to SteamId, entryType EChatEntryType, message string) {
	s.client.Write(NewClientMsgProtobuf(EMsg_ClientFriendMsg, &CMsgClientFriendMsg{
		Steamid:       proto.Uint64(to.ToUint64()),
		ChatEntryType: proto.Int32(int32(entryType)),
		Message:       []byte(message),
	}))
}

// Adds a friend to your friends list or accepts a friend. You'll receive a FriendStateEvent
// for every new/changed friend.
func (s *Social) AddFriend(id SteamId) {
	s.client.Write(NewClientMsgProtobuf(EMsg_ClientAddFriend, &CMsgClientAddFriend{
		SteamidToAdd: proto.Uint64(uint64(id)),
	}))
}

func (s *Social) RemoveFriend(id SteamId) {
	s.client.Write(NewClientMsgProtobuf(EMsg_ClientRemoveFriend, &CMsgClientRemoveFriend{
		Friendid: proto.Uint64(uint64(id)),
	}))
}

// Attempts to join a chat room.
func (s *Social) JoinChat(id SteamId) {
	chatId := SteamId(id)
	if chatId.GetAccountType() == int32(EAccountType_Clan) {
		chatId = chatId.SetAccountInstance(uint32(Clan))
		chatId = chatId.SetAccountType(EAccountType_Chat)
	}
	s.client.Write(NewClientMsg(&MsgClientJoinChat{SteamIdChat: chatId}, make([]byte, 0)))
}

// Attempts to leave a chat room.
func (s *Social) LeaveChat(id SteamId) {
	chatId := SteamId(id)
	if chatId.GetAccountType() == int32(EAccountType_Clan) {
		chatId = chatId.SetAccountInstance(uint32(Clan))
		chatId = chatId.SetAccountType(EAccountType_Chat)
	}
	payload := new(bytes.Buffer)
	binary.Write(payload, binary.LittleEndian, s.client.SteamId().ToUint64())       // ChatterActedOn
	binary.Write(payload, binary.LittleEndian, uint32(EChatMemberStateChange_Left)) // StateChange
	binary.Write(payload, binary.LittleEndian, s.client.SteamId().ToUint64())       // ChatterActedBy
	s.client.Write(NewClientMsg(&MsgClientChatMemberInfo{
		SteamIdChat: chatId,
		Type:        EChatInfoType_StateChange,
	}, payload.Bytes()))
}

// Sends a chat message to a chat room
func (s *Social) SendChatRoomMessage(to SteamId, entryType EChatEntryType, message string) {
	chatId := SteamId(to)
	if chatId.GetAccountType() == int32(EAccountType_Clan) {
		chatId = chatId.SetAccountInstance(uint32(Clan))
		chatId = chatId.SetAccountType(EAccountType_Chat)
	}
	s.client.Write(NewClientMsg(&MsgClientChatMsg{
		ChatMsgType:     entryType,
		SteamIdChatRoom: chatId,
		SteamIdChatter:  s.client.SteamId(),
	}, []byte(message)))
}

func (s *Social) HandlePacket(packet *PacketMsg) {
	switch packet.EMsg {
	case EMsg_ClientPersonaState:
		s.handlePersonaState(packet)
	case EMsg_ClientClanState:
		s.handleClanState(packet)
	case EMsg_ClientFriendsList:
		s.handleFriendsList(packet)
	case EMsg_ClientFriendMsgIncoming:
		s.handleFriendMsg(packet)
	case EMsg_ClientAccountInfo:
		s.handleAccountInfo(packet)
	case EMsg_ClientAddFriendResponse:
		s.handleFriendResponse(packet)
	case EMsg_ClientChatEnter:
		s.handleChatEnter(packet)
	case EMsg_ClientChatMsg:
		s.handleChatMsg(packet)
	case EMsg_ClientChatMemberInfo:
		s.handleChatMemberInfo(packet)
	case EMsg_ClientChatActionResult:
		s.handleChatActionResult(packet)
	case EMsg_ClientChatInvite:
		s.handleChatInvite(packet)
	case EMsg_ClientSetIgnoreFriendResponse:
		s.handleIgnoreFriendResponse(packet)
	case EMsg_ClientFriendProfileInfoResponse:
		s.handleProfileInfoResponse(packet)
	}
}

//TODO: handleAccountInfo
func (s *Social) handleAccountInfo(packet *PacketMsg) {
	body := new(CMsgClientAccountInfo)
	packet.ReadProtoMsg(body)
	//fmt.Printf("%+v\n", body)
}

type FriendListEvent struct{}

type FriendStateEvent struct {
	SteamId      SteamId
	Relationship EFriendRelationship
}

func (f *FriendStateEvent) IsFriend() bool {
	return f.Relationship == EFriendRelationship_Friend
}

type GroupStateEvent struct {
	SteamId      SteamId
	Relationship EClanRelationship
}

func (f *FriendStateEvent) IsMember() bool {
	return f.Relationship == EClanRelationship_Member
}

func (s *Social) handleFriendsList(packet *PacketMsg) {
	list := new(CMsgClientFriendsList)
	packet.ReadProtoMsg(list)

	for _, friend := range list.GetFriends() {
		steamId := SteamId(friend.GetUlfriendid())
		isClan := steamId.GetAccountType() == int32(EAccountType_Clan)

		if isClan {
			rel := EClanRelationship(friend.GetEfriendrelationship())
			if rel == EClanRelationship_None {
				s.Groups.remove(steamId)
			} else {
				s.Groups.add(&Group{
					steamId:      steamId,
					relationship: rel,
				})
			}
			if list.GetBincremental() {
				s.client.Emit(&GroupStateEvent{steamId, rel})
			}
		} else {
			rel := EFriendRelationship(friend.GetEfriendrelationship())
			if rel == EFriendRelationship_None {
				s.Friends.remove(steamId)
			} else {
				s.Friends.add(&Friend{
					steamId:      steamId,
					relationship: rel,
				})
			}
			if list.GetBincremental() {
				s.client.Emit(&FriendStateEvent{steamId, rel})
			}
		}
	}

	if !list.GetBincremental() {
		s.client.Emit(new(FriendListEvent))
	}
}

//TODO: handlePersonaState
func (s *Social) handlePersonaState(packet *PacketMsg) {
	body := new(CMsgClientPersonaState)
	packet.ReadProtoMsg(body)
	//fmt.Printf("%+v\n", body)
}

//TODO: handleClanState
func (s *Social) handleClanState(packet *PacketMsg) {
	body := new(CMsgClientClanState)
	packet.ReadProtoMsg(body)
	//fmt.Printf("%+v\n", body)
}

//TODO: handleFriendResponse
func (s *Social) handleFriendResponse(packet *PacketMsg) {
	body := new(CMsgClientAddFriendResponse)
	packet.ReadProtoMsg(body)
	//fmt.Printf("%+v\n", body)
}

type ChatMsgEvent struct {
	Chatroom SteamId // not set for friend messages
	Sender   SteamId
	Message  string
	Type     EChatEntryType
}

// Whether the type is ChatMsg
func (c *ChatMsgEvent) IsMessage() bool {
	return c.Type == EChatEntryType_ChatMsg
}

func (s *Social) handleFriendMsg(packet *PacketMsg) {
	body := new(CMsgClientFriendMsgIncoming)
	packet.ReadProtoMsg(body)
	message := string(bytes.Split(body.GetMessage(), []byte{0x0})[0])

	s.client.Emit(&ChatMsgEvent{
		Sender:  SteamId(body.GetSteamidFrom()),
		Message: message,
		Type:    EChatEntryType(body.GetChatEntryType()),
	})
}

func (s *Social) handleChatMsg(packet *PacketMsg) {
	body := new(MsgClientChatMsg)
	payload := packet.ReadClientMsg(body).Payload
	message := string(bytes.Split(payload, []byte{0x0})[0])
	s.client.Emit(&ChatMsgEvent{
		Chatroom: SteamId(body.SteamIdChatRoom),
		Sender:   SteamId(body.SteamIdChatter),
		Message:  message,
		Type:     EChatEntryType(body.ChatMsgType),
	})
}

type ChatEnterEvent struct{} //TODO: Make a useful event

func (s *Social) handleChatEnter(packet *PacketMsg) {
	body := new(MsgClientChatEnter)
	packet.ReadMsg(body)
	s.client.Emit(&ChatEnterEvent{})
}

//TODO: handleChatMemberInfo
func (s *Social) handleChatMemberInfo(packet *PacketMsg) {
	body := new(MsgClientChatMemberInfo)
	packet.ReadClientMsg(body)
	//payload := packet.ReadClientMsg(body).Payload //commented out for now
	//fmt.Printf("%+v %v\n", body, payload)
}

//TODO: handleChatActionResult
func (s *Social) handleChatActionResult(packet *PacketMsg) {
	body := new(MsgClientChatActionResult)
	packet.ReadClientMsg(body)
	//fmt.Printf("%+v\n", body)
}

//TODO: handleChatInvite
func (s *Social) handleChatInvite(packet *PacketMsg) {
	body := new(CMsgClientChatInvite)
	packet.ReadProtoMsg(body)
	//fmt.Printf("%+v\n", body)
}

//TODO: handleIgnoreFriendResponse
func (s *Social) handleIgnoreFriendResponse(packet *PacketMsg) {
	body := new(MsgClientSetIgnoreFriendResponse)
	packet.ReadMsg(body)
	//fmt.Printf("%+v\n", body)
}

//TODO: handleProfileInfoResponse
func (s *Social) handleProfileInfoResponse(packet *PacketMsg) {
	body := new(CMsgClientFriendProfileInfoResponse)
	packet.ReadProtoMsg(body)
	//fmt.Printf("%+v\n", body)
}

// A thread-safe friend list which contains references to its predecessor and successor.
// It is mutable and will be changed by Social.
type FriendsList struct {
	mutex sync.RWMutex

	first *Friend
	last  *Friend

	byId map[SteamId]*Friend // fast lookup by ID
}

func (list *FriendsList) add(friend *Friend) {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	friend.mutex = &list.mutex

	list.byId[friend.steamId] = friend

	if list.first == nil {
		list.first = friend
		list.last = friend
	} else {
		friend.prev = list.last
		list.last.next = friend
		list.last = friend
	}
}

func (list *FriendsList) remove(id SteamId) {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	friend := list.byId[id]
	if friend == nil {
		return
	}

	if list.first == friend {
		list.first = nil
	} else {
		friend.prev.next = friend.next
	}

	if list.last == friend {
		list.last = nil
	} else {
		friend.next.prev = friend.prev
	}
}

func (f *FriendsList) First() *Friend {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.first
}

func (f *FriendsList) Last() *Friend {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.last
}

func (f *FriendsList) ById(id SteamId) *Friend {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.byId[id]
}

// A thread-safe friend in a friend list which contains references to its predecessor and successor.
// It is mutable and will be changed by Social.
type Friend struct {
	mutex *sync.RWMutex

	prev *Friend
	next *Friend

	steamId           SteamId
	name              string
	relationship      EFriendRelationship
	personaStateFlags EPersonaStateFlag

	gameAppId uint64
}

func (f *Friend) Next() *Friend {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.next
}

func (f *Friend) Prev() *Friend {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.prev
}

func (f *Friend) SteamId() SteamId {
	// the steam id of a friend never changes, so we don't need to lock here
	return f.steamId
}

func (f *Friend) Name() string {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.name
}

func (f *Friend) Relationship() EFriendRelationship {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.relationship
}

func (f *Friend) PersonaStateFlags() EPersonaStateFlag {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.personaStateFlags
}

func (f *Friend) GameAppId() uint64 {
	f.mutex.RLock()
	defer f.mutex.RUnlock()
	return f.gameAppId
}

// A thread-safe group list which contains references to its predecessor and successor.
// It is mutable and will be changed by Social.
type GroupsList struct {
	mutex sync.RWMutex

	first *Group
	last  *Group

	byId map[SteamId]*Group // fast lookup by ID
}

func (list *GroupsList) add(group *Group) {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	list.byId[group.steamId] = group

	if list.first == nil {
		list.first = group
		list.last = group
	} else {
		group.prev = list.last
		list.last.next = group
		list.last = group
	}
}

func (list *GroupsList) remove(id SteamId) {
	list.mutex.Lock()
	defer list.mutex.Unlock()

	group := list.byId[id]
	if group == nil {
		return
	}

	if list.first == group {
		list.first = nil
	} else {
		group.prev.next = group.next
	}

	if list.last == group {
		list.last = nil
	} else {
		group.next.prev = group.prev
	}
}

// Returns the first group in the group list or nil if the list is empty.
func (list *GroupsList) First() *Group {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	return list.first
}

// Returns the last group in the group list or nil if the list is empty.
func (list *GroupsList) Last() *Group {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	return list.last
}

// Returns the group by a SteamId or nil if there is no such group.
func (list *GroupsList) ById(id SteamId) *Group {
	list.mutex.RLock()
	defer list.mutex.RUnlock()
	return list.byId[id]
}

// A thread-safe group in a group list which contains references to its predecessor and successor.
// It is mutable and will be changed by Social.
type Group struct {
	mutex sync.RWMutex

	prev *Group
	next *Group

	steamId      SteamId
	name         string
	relationship EClanRelationship
}

// Returns the previous element in the group list or nil if this group is the first in the list.
func (g *Group) Prev() *Group {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.prev
}

// Returns the next element in the group list or nil if this group is the last in the list.
func (g *Group) Next() *Group {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.next
}

func (g *Group) SteamId() SteamId {
	// the steam id of a group never changes, so we don't need to lock here
	return g.steamId
}

func (g *Group) Name() string {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.name
}

func (g *Group) Relationship() EClanRelationship {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.relationship
}
