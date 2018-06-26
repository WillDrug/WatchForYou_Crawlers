package main
import (
	"fmt"
	"log"
	"encoding/json"
//	"github.com/streadway/amqp"
	"./cinterfaces"
	"./crss"
	"./cconfig"
	"./crabbit"
)

// TODO: Connector interface
// #helpers
func onError(err error, message string, sever bool) bool {
	if err != nil {
		if sever {
			log.Panic(fmt.Sprintf("Paniced while > %d < with > %d <", message, err))
			return true
		} else {
			log.Panic(fmt.Sprintf("Fatal error while > %d < with > %d <", message, err))
			return true
		}
	}
	return false
}

// #types


type RPCRequest struct {
	Parser      string	`json:"parser"`
	RequestType	string	`json:"parsertype"`
	Request     string	`json:"request"`
	LastDelta   int		`json:"lastdelta"`
	LastUpdate  int64	`json:"lastupdate"`
}

type RPCResponse struct {
	Success bool `json:"updated"`
	Message string `json:"message"`
	Entries []cinterfaces.Entry	
}

// #variables
var cfg cinterfaces.ConfigLoader
var cn cinterfaces.Connection //crabbit.RabbitConnector
var prs cinterfaces.RSSGetter

// #init
func init() {
	// init cfg
	cfg = cconfig.ConfigLoader{}
	log.Printf("Initialized %s crawler", cfg.GetString("modulename"))
}

// #main
func main() {
	// initialize connection
	cn = &crabbit.RabbitConnector{}
	cn.Connect(cfg.GetString("rmqconn"))
	cn.Consume(cfg.GetString("rmqqname"))
	// Be tidy
	defer cn.Disconnect()

	// Be weird
	forever := make(chan bool)
	
	// GO!
	go parseRPCRequest()
	log.Printf(" [*] Awaiting RPC requests")
	// AND DON'T STOP EVER!
	<-forever
}

func ReturnAnswer(res RPCResponse, d cinterfaces.Message) bool {
	pub, err := json.Marshal(res)
	if onError(err, "Failed to jsonify", false) {
		return false
	} // TODO : you can double down on error logs and exception handling	
	// Publish response
	err = cn.ReplyToMessage(pub, d)
	onError(err, "Failed to publish a message", false)
	return true
}

func onAtomicError(err error, message string, d cinterfaces.Message) bool {
	if err != nil {
		go ReturnAnswer(RPCResponse{false, message, make([]cinterfaces.Entry, 0)}, d)
		return true
	}
	return false
}

// #parsers
func parseRPCRequest() {

	// Create some space for the request
	// WARNING: If errors occur - move this inside the loop.
	var req RPCRequest
	var err error
	var res RPCResponse
	prs = crss.YRSSGetter{}
	inbound := cn.InboundMessages(1)
	// loop prefetched in channel
	// inside there is GO lexem just if prefetch size changes
	for d := range inbound {
		// Parse request into struct
		log.Printf("Got me a message")
		err = json.Unmarshal(d.Content, &req)
		onAtomicError(err, "Failed to convert body to *RPCRequest", d)
		log.Printf("Got me a %s request type %s", req.Request, req.RequestType) // TODO: MUST HAVE log-levels here
		
		// Get data.
		// basing on request type, but actually can just run this:
		// prs.GetRSSLinkUniversal()
		
		if req.Parser != "youtube" {
			res = RPCResponse{false, "Wrong parser for message", make([]cinterfaces.Entry, 0)}
		} else {
			var entries []cinterfaces.Entry
			if req.RequestType == "query" { // TODO : change to config
				url, err := prs.GetRSSLinkUniversal(req.Request)
				if onAtomicError(err, "Failed to fetch RSS feed URL", d) {
					continue
				}
				entries, err = prs.GetFeedUpdatesWithPOSIX(url, req.LastUpdate)
				if onAtomicError(err, "Failed to fetch RSS feed from URL", d) {
					continue
				}
				res = RPCResponse{true, "", make([]cinterfaces.Entry, 0)}
			} else if req.RequestType == "URL" {
				entries, err = prs.GetFeedUpdatesWithPOSIX(req.Request, req.LastUpdate)
				if onAtomicError(err, "Failed to fetch RSS feed from URL", d) {
					continue
				}
				res = RPCResponse{true, "", make([]cinterfaces.Entry, 0)}
			} else {
				entries = make([]cinterfaces.Entry, 0)
				res = RPCResponse{false, "Uncrecognized request type", make([]cinterfaces.Entry, 0)}
			}
			res.Entries = entries
		}
		go ReturnAnswer(res, d)
	}
}

