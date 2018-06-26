
	// How to getRSS feed example
	// ------------------------------------
	//var reader interfaces.RSSGetter
	//reader = rss.YRSSGetter{}
	//rss_url, err := reader.GetRSSLinkByName("Tom Scott")
	//if err != nil {
	//	log.Fatal(err)
	//	return
	//}
	//fmt.Println(rss_url)
	
// send
package main

import (
        "fmt"
        "log"
		"encoding/json"
		       "math/rand"
        "github.com/streadway/amqp"
		"./cinterfaces"
)


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

func failOnError(err error, msg string) {
        if err != nil {
                log.Fatalf("%s: %s", msg, err)
                panic(fmt.Sprintf("%s: %s", msg, err))
        }
}
func randomString(l int) string {
        bytes := make([]byte, l)
        for i := 0; i < l; i++ {
                bytes[i] = byte(randInt(65, 90))
        }
        return string(bytes)
}
func randInt(min int, max int) int {
        return min + rand.Intn(max-min)
}

func youtubeRPC(n RPCRequest) (res RPCResponse, err error) {
        conn, err := amqp.Dial("amqp://watchforyou:watchforyou@localhost:5672/")
        failOnError(err, "Failed to connect to RabbitMQ")
        defer conn.Close()

        ch, err := conn.Channel()
        failOnError(err, "Failed to open a channel")
        defer ch.Close()

        q, err := ch.QueueDeclare(
                "callback_q_lol",    // name
                false, // durable
                false, // delete when usused
                true,  // exclusive
                false, // noWait
                nil,   // arguments
        )
        failOnError(err, "Failed to declare a queue")

        msgs, err := ch.Consume(
                q.Name, // queue
                "",     // consumer
                false,   // auto-ack
                true,  // exclusive
                false,  // no-local
                false,  // no-wait
                nil,    // args
        )
        failOnError(err, "Failed to register a consumer")

        corrId := randomString(32)
		pub, _ := json.Marshal(n)
        err = ch.Publish(
                "",          // exchange
                "youtube_rpc", // routing key
                false,       // mandatory
                false,       // immediate
                amqp.Publishing{
                        ContentType:   "application/json",
                        CorrelationId: corrId,
                        ReplyTo:       q.Name,
                        Body:          pub,
                })
        failOnError(err, "Failed to publish a message")
        for d := range msgs {
                if corrId == d.CorrelationId {
						err = json.Unmarshal(d.Body, &res)
                        failOnError(err, "Failed to convert body")
                        break
                }
        }
        return res, nil
}

func main() {
	res, err := youtubeRPC(RPCRequest{"youtube","query","Some Channel",0,0})
	failOnError(err, "Failed to handle RPC request")
	log.Printf(" [.] Got %v", res)
}

