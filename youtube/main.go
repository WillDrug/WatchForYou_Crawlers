package main
import (
	"fmt"
	"log"
	"encoding/json"
	"github.com/spf13/viper"
	"github.com/streadway/amqp"
	"./crss"
)
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
type RSSGetter interface {
	GetRSSLinkByName(query string) (string, error)
	GetRSSLinkByLink(url string) (string, error)
	GetRSSLinkUniversal(req string) (string, error)
	CheckRSSLink(url string) (bool, error)
}
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
	Entries []crss.Entry	
}
// #utility again
func onAtomicError(err error, message string, d amqp.Delivery) bool {
	if err != nil {
		go ReturnAnswer(RPCResponse{false, message, make([]crss.Entry, 0)}, d)
		return true
	}
	return false
}

// #variables
var conn *amqp.Connection
var ch 	 *amqp.Channel
var q	 amqp.Queue
var msgs <-chan amqp.Delivery


// #init
func init() {
	var err error
	// initializing #config TODO: move to separate package for all crawlers
	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/host/WatchForYou/WatchForYou_Crawlers/youtube/")
	err = viper.ReadInConfig() // Find and read the config file
	onError(err, "Failed to initialize config", true)
	
	log.Printf("Initializing %s crawler", viper.GetString("modulename"))
	// #default -ing viper
	viper.SetDefault("rmqconn", "amqp://watchforyou:watchforyou@localhost:5672/")
	viper.SetDefault("rmqqname", "crawlpc")
	
	// initializing #rabbitmq RPC interfaces TODO: move to separate package with deferred connection closing
	log.Printf("Dialing %s", viper.GetString("rmqconn"))
	conn, err = amqp.Dial(viper.GetString("rmqconn"))
	onError(err, "Failed to connect to RabbitMQ", true)
	
	log.Printf("Creating a channel")
	ch, err = conn.Channel()
	onError(err, "Failed to open a channel", true)
	
	// #declare ing queue. Consumer always declares!
	log.Printf("Declaring %s queue for RPC requests", viper.GetString("rmqqname"))
	q, err = ch.QueueDeclare(
			viper.GetString("rmqqname"), // name TODO: fix those configs, all of them, this is bollocks
			true,       // durable
			true,        // delete when usused
			false,       // exclusive
			false,       // no-wait
			nil,         // arguments
	)
	onError(err, "Failed to declare a queue", true)
	
	log.Printf("Setting Qos to 1 prefetch full") // TODO: configurable?
	err = ch.Qos(
			1,     // prefetch count
			0,     // prefetch size
			false, // global
	)
	onError(err, "Failed to set QoS", false)
	
	log.Printf("Starting a consumer channel")
	msgs, err = ch.Consume(
			q.Name, // queue
			"",     // consumer
			false,  // auto-ack
			false,  // exclusive
			false,  // no-local
			false,  // no-wait
			nil,    // args
	)
	onError(err, "Failed to register a consumer", true)
		
}
// #main
func main() {
	// Be tidy
	defer conn.Close()
	defer ch.Close()
	// Be weird
	forever := make(chan bool)
	// GO!
	go parseRPCRequest()
	log.Printf(" [*] Awaiting RPC requests")
	// AND DON'T STOP EVER!
	<-forever
}
// #return
func ReturnAnswer(res RPCResponse, d amqp.Delivery) bool {
	pub, err := json.Marshal(res)
	if onError(err, "Failed to jsonify", false) {
		return false
	} // TODO : you can double down on error logs and exception handling	
	// Publish response
	err = ch.Publish(
			"",        // exchange
			d.ReplyTo, // routing key // is handled by sender.
			false,     // mandatory
			false,     // immediate
			amqp.Publishing{
					ContentType:   "application/json",
					CorrelationId: d.CorrelationId, // CorrelationId is handled by sender.
					Body:          pub,
			})
	if !onError(err, "Failed to publish a message", false) {
		d.Ack(false)
	}
	return true
}

// #parsers
func parseRPCRequest() {
	// Create some space for the request
	// WARNING: If errors occur - move this inside the loop.
	var req RPCRequest
	var err error
	var res RPCResponse
	var getter RSSGetter
	getter = crss.YRSSGetter{}
	
	// loop prefetched in channel
	// inside there is GO lexem just if prefetch size changes
	for d := range msgs {
		// Parse request into struct
		log.Printf("Got me a message")
		err = json.Unmarshal(d.Body, &req)
		onAtomicError(err, "Failed to convert body to *RPCRequest", d)
		log.Printf("Got me a %s request type %s", req.Request, req.RequestType) // TODO: MUST HAVE log-levels here
		
		// Get data.
		// basing on request type, but actually can just run this:
		// getter.GetRSSLinkUniversal()
		if req.Parser != "youtube" {
			res = RPCResponse{false, "Wrong parser for message", make([]crss.Entry, 0)}
		} else {
			var entries []crss.Entry
			if req.RequestType == "query" { // TODO : change to config
				url, err := getter.GetRSSLinkUniversal(req.Request)
				if onAtomicError(err, "Failed to fetch RSS feed URL", d) {
					continue
				}
				entries, err = crss.GetFeedUpdatesWithPOSIX(url, req.LastUpdate)
				if onAtomicError(err, "Failed to fetch RSS feed from URL", d) {
					continue
				}
				res = RPCResponse{true, "", make([]crss.Entry, 0)}
			} else if req.RequestType == "URL" {
				entries, err = crss.GetFeedUpdatesWithPOSIX(req.Request, req.LastUpdate)
				if onAtomicError(err, "Failed to fetch RSS feed from URL", d) {
					continue
				}
				res = RPCResponse{true, "", make([]crss.Entry, 0)}
			} else {
				entries = make([]crss.Entry, 0)
				res = RPCResponse{false, "Uncrecognized request type", make([]crss.Entry, 0)}
			}
			res.Entries = entries
		}
		go ReturnAnswer(res, d)
	}
}

