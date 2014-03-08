package steam

import (
	. "github.com/GamingRobot/steamgo/internal"
	. "github.com/GamingRobot/steamgo/steamid"
)

type TradeProposedEvent struct {
	RequestId TradeRequestId
	Other     SteamId
	OtherName string
}

type TradeResultEvent struct {
	RequestId TradeRequestId
	Response  EEconTradeResponse
	Other     SteamId
}

type TradeSessionStartEvent struct {
	Other SteamId
}
