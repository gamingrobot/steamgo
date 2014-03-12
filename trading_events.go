package steamgo

import (
	. "github.com/gamingrobot/steamgo/internal"
	. "github.com/gamingrobot/steamgo/steamid"
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
