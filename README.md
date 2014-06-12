Steam for Go
=======

This library allows you to interact with Steam as if it was an actual Steam client.
It's a port of [SteamKit2](https://github.com/SteamRE/SteamKit) to Go.

## Installation

    go get github.com/gamingrobot/steamgo

## Usage

You can view the documentation with the `godoc` tool or
[online on godoc.org](http://godoc.org/github.com/gamingrobot/steamgo).

## License

Steam for Go is licensed under the New BSD License. More information can be found in LICENSE

## Example

```go
package main

import (
	"github.com/gamingrobot/steamgo"
	"github.com/gamingrobot/steamgo/internal"
	"io/ioutil"
	"log"
)

func main() {
	myLoginInfo := steamgo.LogOnDetails{}
	myLoginInfo.Username = "UserName"
	myLoginInfo.Password = "PassWord"

	client := steamgo.NewClient()
	client.Connect()
	for event := range client.Events() {
		switch e := event.(type) {
		case steamgo.ConnectedEvent:
			client.Auth.LogOn(myLoginInfo)
		case steamgo.MachineAuthUpdateEvent:
			ioutil.WriteFile("sentry", e.Hash, 0666)
		case steamgo.LoggedOnEvent:
			client.Social.SetPersonaState(internal.EPersonaState_Online)
		case steamgo.FatalError:
			client.Connect() // please do some real error handling here
			log.Print(e)
		case error:
			log.Print(e)
		}
	}
}

```

##Zephyr
Web based steam client that uses steamgo  
https://github.com/gamingrobot/zephyr
