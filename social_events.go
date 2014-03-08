package steam

import (
	. "github.com/GamingRobot/steamgo/internal"
	. "github.com/GamingRobot/steamgo/steamid"
)

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

// Fired when someone changing their friend details
type PersonaStateEvent struct {
	StatusFlags            EClientPersonaStateFlag
	FriendId               SteamId
	State                  EPersonaState
	StateFlags             EPersonaStateFlag
	GameAppId              uint32
	GameId                 uint64
	GameName               string
	GameServerIp           uint32
	GameServerPort         uint32
	QueryPort              uint32
	SourceSteamId          SteamId
	GameDataBlob           []byte
	Name                   string
	AvatarHash             []byte
	LastLogOff             uint32
	LastLogOn              uint32
	ClanRank               uint32
	ClanTag                string
	OnlineSessionInstances uint32
	PublishedSessionId     uint32
	PersonaSetByUser       bool
	FacebookName           string
	FacebookId             uint64
}

// Fired when a clan's state has been changed
type ClanStateEvent struct {
	ClandId             SteamId
	StateFlags          EClientPersonaStateFlag
	AccountFlags        EAccountFlags
	ClanName            string
	AvatarHash          []byte
	MemberTotalCount    uint32
	MemberOnlineCount   uint32
	MemberChattingCount uint32
	MemberInGameCount   uint32
	Events              []ClanEventDetails
	Announcements       []ClanEventDetails
}

type ClanEventDetails struct {
	Id         uint64
	EventTime  uint32
	Headline   string
	GameId     uint64
	JustPosted bool
}

// Fired in response to adding a friend to your friends list
type FriendAddedEvent struct {
	Result      EResult
	SteamId     SteamId
	PersonaName string
}

// Fired when the client receives a message from either a friend or a chat room
type ChatMsgEvent struct {
	ChatRoomId SteamId // not set for friend messages
	ChatterId  SteamId
	Message    string
	EntryType  EChatEntryType
}

// Whether the type is ChatMsg
func (c *ChatMsgEvent) IsMessage() bool {
	return c.EntryType == EChatEntryType_ChatMsg
}

// Fired in response to joining a chat
type ChatEnterEvent struct {
	ChatRoomId    SteamId
	FriendId      SteamId
	ChatRoomType  EChatRoomType
	OwnerId       SteamId
	ClanId        SteamId
	ChatFlags     byte
	EnterResponse EChatRoomEnterResponse
	Name          string
}

// Fired in response to a chat member's info being received
type ChatMemberInfoEvent struct {
	ChatRoomId      SteamId
	Type            EChatInfoType
	StateChangeInfo StateChangeDetails
}

type StateChangeDetails struct {
	ChatterActedOn SteamId
	StateChange    EChatMemberStateChange
	ChatterActedBy SteamId
}

// Fired when a chat action has completed
type ChatActionResultEvent struct {
	ChatRoomId SteamId
	ChatterId  SteamId
	Action     EChatAction
	Result     EChatActionResult
}

// Fired when a chat invite is received
type ChatInviteEvent struct {
	InvitedId    SteamId
	ChatRoomId   SteamId
	PatronId     SteamId
	ChatRoomType EChatRoomType
	FriendChatId SteamId
	ChatRoomName string
	GameId       uint64
}

// Fired in response to ignoring a friend
type IgnoreFriendEvent struct {
	Result EResult
}

// Fired in response to requesting profile info for a user
type ProfileInfoEvent struct {
	Result      EResult
	SteamId     SteamId
	TimeCreated uint32
	RealName    string
	CityName    string
	StateName   string
	CountryName string
	Headline    string
	Summary     string
}
