package steam

import (
	. "github.com/GamingRobot/steamgo/internal"
)

type LoggedOnEvent struct{}

type LoginKeyEvent struct {
	UniqueId uint32
	LoginKey string
}

type LoggedOffEvent struct {
	Result EResult
}

type MachineAuthUpdateEvent struct {
	Hash []byte
}

type AccountInfoEvent struct {
	PersonaName          string
	Country              string
	PasswordSalt         []byte
	PasswordSHADisgest   []byte
	CountAuthedComputers int32
	LockedWithIpt        bool
	AccountFlags         EAccountFlags
	FacebookId           uint64
	FacebookName         string
}
