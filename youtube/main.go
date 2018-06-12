package main

import (
	//"./cinterfaces"
	//"./crss"
	//"./cstorage"
	"fmt"
	"log"
	//"github.com/mmcdole/gofeed"
	//"gopkg.in/mgo.v2"
	//"reflect"
	"github.com/streadway/amqp"
	"encoding/json"
)

type FeedSubscription struct {
	Parser      string	`json:"parser"`
	RequestType	string	`json:"parsertype"`
	Request     string	`json:"request"`
	LastDelta   int		`json:"lastdelta"`
	LastUpdate  int		`json:"lastupdate"`
}



type Test struct{
	Field string `json:"field"`
	Another int `json:"another"`
}
func failOnError(err error, msg string) {
  if err != nil {
    log.Fatalf("%s: %s", msg, err)
    panic(fmt.Sprintf("%s: %s", msg, err))
  }
}

var test struct{
	Field string `json:"field"`
	Another int `json:"another"`
}

func main() {
	conn, err := amqp.Dial("amqp://watchforyou:watchforyou@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
			"rpc_queue", // name
			false,       // durable
			false,       // delete when usused
			false,       // exclusive
			false,       // no-wait
			nil,         // arguments
	)
	failOnError(err, "Failed to declare a queue")

	err = ch.Qos(
			1,     // prefetch count
			0,     // prefetch size
			false, // global
	)
	failOnError(err, "Failed to set QoS")

	msgs, err := ch.Consume(
			q.Name, // queue
			"",     // consumer
			false,  // auto-ack
			false,  // exclusive
			false,  // no-local
			false,  // no-wait
			nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
			for d := range msgs {
					var n Test
					err1 := json.Unmarshal(d.Body, &n)
					failOnError(err1, "Failed to convert body to integer")

					log.Printf(" [.] (%d)", n)
					pub, err2 := json.Marshal(n)
					failOnError(err2, "Failed to jsonify")
					err = ch.Publish(
							"",        // exchange
							d.ReplyTo, // routing key
							false,     // mandatory
							false,     // immediate
							amqp.Publishing{
									ContentType:   "text/plain",
									CorrelationId: d.CorrelationId,
									Body:          pub,
							})
					failOnError(err, "Failed to publish a message")

					d.Ack(false)
			}
	}()

	log.Printf(" [*] Awaiting RPC requests")
	<-forever
}

	
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
	