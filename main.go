package main

import (
	"plugin"
	"log"
	"fmt"
	"./winterfaces"
	"./wconfig"
	"./wconnector"
)


// variables
var cfg winterfaces.ConfigLoader
var conn winterfaces.Connection
var prs winterfaces.Parser

// init
func init() {
	// init config and set defaults
	cfg = &wconfig.ConfigLoader{}
	cfg.SetDefaults(map[string]interface{}{
		"connstring": "amqp://watchforyou:watchforyou@localhost:5672/",
		"connpoint": "youtube",
	})
	// init connector
	conn = &wconnector.Connector{}

	// init parser
	pname := cfg.GetString("connpoint")
	mod := fmt.Sprintf("./wparsers/%v/%v.so", pname, pname)
	plug, err := plugin.Open(mod)
	if err != nil {
		log.Fatal(err)
		panic("Parser not found")
	}
	prsSym, err := plug.Lookup("Parser")
	if err != nil {
		log.Fatal(err)
		panic("Selected plugin does not implement Parser interface")
	}
	prs = prsSym.(winterfaces.Parser)
	log.Printf("Loaded as %v parser\n", pname)
}

func main() {
	var err error
	// Connect
	err = conn.Connect(cfg.GetString("connstring"))
	if err != nil {
		panic(err)
	}
	// Be tidy
	defer conn.Disconnect()
	// Declare and start populating a channel
	var msgs *chan winterfaces.RPCRequest
	msgs, err = conn.Consume(cfg.GetString("connpoint"))
	log.Printf("Got channel %#v", msgs)
	if err != nil {
		panic(err)
	}

	go func() {
		var entries []winterfaces.Entry
		var err error
		log.Printf("Main started listening")
		for msg := range *msgs { // expecting winterfaces.RPCRequest
			log.Printf("Main got a message")
			entries, err = prs.CheckUpdates(msg.Request, msg.RequestType, msg.LastUpdate) // using int64 version under hood ([]winterfaces.Entry, error)
			if err != nil {
				log.Printf("Failed to asnwer to %#v", msg)
				msg.ReplyChannel <- winterfaces.RPCResponse{false, err.Error(), make([]winterfaces.Entry, 0)}
			} else {
				// Reply
				msg.ReplyChannel <- winterfaces.RPCResponse{true, "", entries}
			}
		}
	}()		

	forever := make(chan bool)
	<-forever
}

/*type RPCRequest struct {
	Parser      string	`json:"parser"`
	RequestType	string	`json:"parsertype"`
	Request     string	`json:"request"`
	LastDelta   int		`json:"lastdelta"`
	LastUpdate  int64	`json:"lastupdate"`
	ReplyChannel chan RPCResponse
}*/