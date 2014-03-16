package steamgo

import (
	. "github.com/gamingrobot/steamgo/internal"
	. "github.com/gamingrobot/steamgo/steamid"
)

type TradeProposedEvent struct {
	RequestId TradeRequestId
	Other     SteamId `json:",string"`
	OtherName string
}

type TradeResultEvent struct {
	RequestId TradeRequestId
	Response  EEconTradeResponse
	Other     SteamId `json:",string"`
}

type TradeSessionStartEvent struct {
	Other SteamId `json:",string"`
}
