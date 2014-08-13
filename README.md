Steam for Go
=======

Do not use this library it is in the process of getting merged into the original repo
https://github.com/Philipp15b/go-steam


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
