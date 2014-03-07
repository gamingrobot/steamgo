package steam

import (
	"bytes"
	"code.google.com/p/goprotobuf/proto"
	"encoding/binary"
	. "github.com/GamingRobot/steamgo/internal"
	"github.com/GamingRobot/steamgo/socialcache"
	. "github.com/GamingRobot/steamgo/steamid"
	"io"
	"sync"
)

// Provides access to social aspects of Steam.
type Social struct {
	mutex sync.RWMutex

	name         string
	avatarHash   []byte
	personaState EPersonaState

	Friends *socialcache.FriendsList
	Groups  *socialcache.GroupsList

	client *Client
}

func newSocial(client *Client) *Social {
	return &Social{
		Friends: socialcache.NewFriendsList(),
		Groups:  socialcache.NewGroupsList(),
		client:  client,
	}
}

// Gets the local user's persona name
func (s *Social) GetPersonaName() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.name
}

// Sets the local user's persona name and broadcasts it over the network
func (s *Social) SetPersonaName(name string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.name = name
	s.client.Write(NewClientMsgProtobuf(EMsg_ClientChangeStatus, &CMsgClientChangeStatus{
		PersonaState: proto.Uint32(uint32(s.personaState)),
		PlayerName:   proto.String(name),
	}))
}

// Gets the local user's persona state
func (s *Social) GetPersonaState() EPersonaState {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.personaState
}

// Sets the local user's persona state and broadcasts it over the network
func (s *Social) SetPersonaState(state EPersonaState) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.personaState = state
	s.client.Write(NewClientMsgProtobuf(EMsg_ClientChangeStatus, &CMsgClientChangeStatus{
		PersonaState: proto.Uint32(uint32(state)),
	}))
}

// Sends a chat message to ether a room or friend
func (s *Social) SendMessage(to SteamId, entryType EChatEntryType, message string) {
	if to.GetAccountType() == int32(EAccountType_Individual) || to.GetAccountType() == int32(EAccountType_ConsoleUser) {
		s.SendChatMessage(to, entryType, message)
	} else if to.GetAccountType() == int32(EAccountType_Clan) || to.GetAccountType() == int32(EAccountType_Chat) {
		s.SendChatRoomMessage(to, entryType, message)
	}
}

// Sends a chat message to a friend
func (s *Social) SendChatMessage(to SteamId, entryType EChatEntryType, message string) {
	s.client.Write(NewClientMsgProtobuf(EMsg_ClientFriendMsg, &CMsgClientFriendMsg{
		Steamid:       proto.Uint64(to.ToUint64()),
		ChatEntryType: proto.Int32(int32(entryType)),
		Message:       []byte(message),
	}))
}

// Adds a friend to your friends list or accepts a friend. You'll receive a FriendStateEvent
// for every new/changed friend
func (s *Social) AddFriend(id SteamId) {
	s.client.Write(NewClientMsgProtobuf(EMsg_ClientAddFriend, &CMsgClientAddFriend{
		SteamidToAdd: proto.Uint64(uint64(id)),
	}))
}

// Removes a friend from your friends list
func (s *Social) RemoveFriend(id SteamId) {
	s.client.Write(NewClientMsgProtobuf(EMsg_ClientRemoveFriend, &CMsgClientRemoveFriend{
		Friendid: proto.Uint64(uint64(id)),
	}))
}

// Ignores or unignores a friend on Steam
func (s *Social) IgnoreFriend(id SteamId, setIgnore bool) {
	ignore := uint8(1) //True
	if !setIgnore {
		ignore = uint8(0) //False
	}
	s.client.Write(NewClientMsg(&MsgClientSetIgnoreFriend{
		MySteamId:     s.client.SteamId(),
		SteamIdFriend: id,
		Ignore:        ignore,
	}, make([]byte, 0)))
}

// Attempts to join a chat room
func (s *Social) JoinChat(id SteamId) {
	chatId := id.ClanToChat()
	s.client.Write(NewClientMsg(&MsgClientJoinChat{
		SteamIdChat: chatId,
	}, make([]byte, 0)))
}

// Attempts to leave a chat room
func (s *Social) LeaveChat(id SteamId) {
	chatId := id.ClanToChat()
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
func (s *Social) SendChatRoomMessage(room SteamId, entryType EChatEntryType, message string) {
	chatId := room.ClanToChat()
	s.client.Write(NewClientMsg(&MsgClientChatMsg{
		ChatMsgType:     entryType,
		SteamIdChatRoom: chatId,
		SteamIdChatter:  s.client.SteamId(),
	}, []byte(message)))
}

// Kicks the specified chat member from the given chat room
func (s *Social) KickChatMember(room SteamId, user SteamId) {
	chatId := room.ClanToChat()
	s.client.Write(NewClientMsg(&MsgClientChatAction{
		SteamIdChat:        chatId,
		SteamIdUserToActOn: user,
		ChatAction:         EChatAction_Kick,
	}, make([]byte, 0)))
}

// Bans the specified chat member from the given chat room
func (s *Social) BanChatMember(room SteamId, user SteamId) {
	chatId := room.ClanToChat()
	s.client.Write(NewClientMsg(&MsgClientChatAction{
		SteamIdChat:        chatId,
		SteamIdUserToActOn: user,
		ChatAction:         EChatAction_Ban,
	}, make([]byte, 0)))
}

// Unbans the specified chat member from the given chat room
func (s *Social) UnbanChatMember(room SteamId, user SteamId) {
	chatId := room.ClanToChat()
	s.client.Write(NewClientMsg(&MsgClientChatAction{
		SteamIdChat:        chatId,
		SteamIdUserToActOn: user,
		ChatAction:         EChatAction_UnBan,
	}, make([]byte, 0)))
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

func (s *Social) handleAccountInfo(packet *PacketMsg) {
	body := new(CMsgClientAccountInfo)
	packet.ReadProtoMsg(body)
	s.mutex.Lock()
	defer s.mutex.Unlock()
	//just set the name Auth handles the callback
	s.name = body.GetPersonaName()
}

type FriendsListEvent struct{}

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

func (g *GroupStateEvent) IsMember() bool {
	return g.Relationship == EClanRelationship_Member
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
				s.Groups.Remove(steamId)
			} else {
				s.Groups.Add(socialcache.Group{
					SteamId:      steamId,
					Relationship: rel,
				})
			}
			if list.GetBincremental() {
				s.client.Emit(&GroupStateEvent{steamId, rel})
			}
		} else {
			rel := EFriendRelationship(friend.GetEfriendrelationship())
			if rel == EFriendRelationship_None {
				s.Friends.Remove(steamId)
			} else {
				s.Friends.Add(socialcache.Friend{
					SteamId:      steamId,
					Relationship: rel,
				})
			}
			if list.GetBincremental() {
				s.client.Emit(&FriendStateEvent{steamId, rel})
			}
		}
	}

	if !list.GetBincremental() {
		s.client.Emit(&FriendsListEvent{})
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

type ChatEnterEvent struct {
	SteamIdChat   SteamId
	SteamIdFriend SteamId
	ChatRoomType  EChatRoomType
	SteamIdOwner  SteamId
	SteamIdClan   SteamId
	ChatFlags     byte
	EnterResponse EChatRoomEnterResponse
	Name          string
}

func (s *Social) handleChatEnter(packet *PacketMsg) {
	body := new(MsgClientChatEnter)
	payload := packet.ReadClientMsg(body).Payload
	reader := bytes.NewBuffer(payload)
	count, _ := ReadInt32(reader)
	name, _ := ReadString(reader)
	ReadByte(reader) //0
	chatId := SteamId(body.SteamIdChat)
	for i := 0; i < int(count); i++ {
		id, permissions, rank := readChatMember(reader)
		ReadBytes(reader, 6) //No idea what this is
		s.Groups.AddChatMember(chatId, socialcache.ChatMember{
			SteamId:     SteamId(id),
			Permissions: permissions,
			Rank:        rank,
		})
	}
	s.client.Emit(&ChatEnterEvent{
		SteamIdChat:   SteamId(body.SteamIdChat),
		SteamIdFriend: SteamId(body.SteamIdFriend),
		ChatRoomType:  EChatRoomType(body.ChatRoomType),
		SteamIdOwner:  SteamId(body.SteamIdOwner),
		SteamIdClan:   SteamId(body.SteamIdClan),
		ChatFlags:     byte(body.ChatFlags),
		EnterResponse: EChatRoomEnterResponse(body.EnterResponse),
		Name:          name,
	})
}

type StateChangeDetails struct {
	ChatterActedOn SteamId
	StateChange    EChatMemberStateChange
	ChatterActedBy SteamId
}

type ChatMemberInfoEvent struct {
	SteamIdChat     SteamId
	Type            EChatInfoType
	StateChangeInfo StateChangeDetails
}

func (s *Social) handleChatMemberInfo(packet *PacketMsg) {
	body := new(MsgClientChatMemberInfo)
	payload := packet.ReadClientMsg(body).Payload
	reader := bytes.NewBuffer(payload)
	chatId := SteamId(body.SteamIdChat)
	if body.Type == EChatInfoType_StateChange {
		actedOn, _ := ReadUint64(reader)
		state, _ := ReadInt32(reader)
		actedBy, _ := ReadUint64(reader)
		ReadByte(reader) //0
		stateChange := EChatMemberStateChange(state)
		if stateChange == EChatMemberStateChange_Entered {
			_, permissions, rank := readChatMember(reader)
			s.Groups.AddChatMember(chatId, socialcache.ChatMember{
				SteamId:     SteamId(actedOn),
				Permissions: permissions,
				Rank:        rank,
			})
		} else if stateChange == EChatMemberStateChange_Banned || stateChange == EChatMemberStateChange_Kicked ||
			stateChange == EChatMemberStateChange_Disconnected || stateChange == EChatMemberStateChange_Left {
			s.Groups.RemoveChatMember(chatId, SteamId(actedOn))
		}
		stateInfo := StateChangeDetails{
			ChatterActedOn: SteamId(actedOn),
			StateChange:    EChatMemberStateChange(stateChange),
			ChatterActedBy: SteamId(actedBy),
		}
		s.client.Emit(&ChatMemberInfoEvent{
			SteamIdChat:     SteamId(body.SteamIdChat),
			Type:            EChatInfoType(body.Type),
			StateChangeInfo: stateInfo,
		})
	}
}

func readChatMember(r io.Reader) (SteamId, EChatPermission, EClanRank) {
	ReadString(r) // MessageObject
	ReadByte(r)   // 7
	ReadString(r) //steamid
	id, _ := ReadUint64(r)
	ReadByte(r)   // 2
	ReadString(r) //Permissions
	permissions, _ := ReadInt32(r)
	ReadByte(r)   // 2
	ReadString(r) //Details
	rank, _ := ReadInt32(r)
	if rank == 4 { //Fix rank to match EClanRank
		rank = EClanRank_Member
	} else if rank == 8 {
		rank = EClanRank_Moderator
	}
	return SteamId(id), EChatPermission(permissions), EClanRank(rank)
}

type ChatActionResultEvent struct {
	SteamIdChat        SteamId
	SteamIdUserActedOn SteamId
	ChatAction         EChatAction
	ActionResult       EChatActionResult
}

func (s *Social) handleChatActionResult(packet *PacketMsg) {
	body := new(MsgClientChatActionResult)
	packet.ReadClientMsg(body)
	s.client.Emit(&ChatActionResultEvent{
		SteamIdChat:        SteamId(body.SteamIdChat),
		SteamIdUserActedOn: SteamId(body.SteamIdUserActedOn),
		ChatAction:         EChatAction(body.ChatAction),
		ActionResult:       EChatActionResult(body.ActionResult),
	})
}

type ChatInviteEvent struct {
	SteamIdInvited    SteamId
	SteamIdChat       SteamId
	SteamIdPatron     SteamId
	ChatRoomType      EChatRoomType
	SteamIdFriendChat SteamId
	ChatName          string
	GameId            uint64
}

func (s *Social) handleChatInvite(packet *PacketMsg) {
	body := new(CMsgClientChatInvite)
	packet.ReadProtoMsg(body)
	s.client.Emit(&ChatInviteEvent{
		SteamIdInvited:    SteamId(body.GetSteamIdInvited()),
		SteamIdChat:       SteamId(body.GetSteamIdChat()),
		SteamIdPatron:     SteamId(body.GetSteamIdPatron()),
		ChatRoomType:      EChatRoomType(body.GetChatroomType()),
		SteamIdFriendChat: SteamId(body.GetSteamIdFriendChat()),
		ChatName:          body.GetChatName(),
		GameId:            body.GetGameId(),
	})
}

//TODO: handleIgnoreFriendResponse
func (s *Social) handleIgnoreFriendResponse(packet *PacketMsg) {
	body := new(MsgClientSetIgnoreFriendResponse)
	packet.ReadClientMsg(body)
	//fmt.Printf("%+v\n", body)
}

//TODO: handleProfileInfoResponse
func (s *Social) handleProfileInfoResponse(packet *PacketMsg) {
	body := new(CMsgClientFriendProfileInfoResponse)
	packet.ReadProtoMsg(body)
	//fmt.Printf("%+v\n", body)
}
