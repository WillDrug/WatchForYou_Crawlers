package crabbit

import (
	"github.com/streadway/amqp"
	"log"
	"fmt"
//	"../cconfig" -- not needed to be a proper interface
//	"../cinterfaces"
)
//type Message amqp.Delivery



type RabbitConnector struct {
	conn *amqp.Connection
	ch 	 *amqp.Channel
	q	 amqp.Queue
	Msgs <-chan amqp.Delivery
}

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



func (rc *RabbitConnector) Connect(cstr string) bool {
	var err error
	// initializing #rabbitmq RPC interfaces TODO: move to separate package with deferred connection closing
	log.Printf("Dialing %s", cstr)
	rc.conn, err = amqp.Dial(cstr)
	onError(err, "Failed to connect to RabbitMQ", true)
	
	log.Printf("Creating a channel")
	rc.ch, err = rc.conn.Channel()
	onError(err, "Failed to open a channel", true)
	return true
}

func (rc *RabbitConnector) Disconnect() bool {
	rc.conn.Close()
	rc.ch.Close()
	return true
}

func (rc *RabbitConnector) DeclareAndConsume(qname string) bool {
	var err error
	
	// #declare ing queue. Consumer always declares!
	log.Printf("Declaring %s queue for RPC requests", qname)
	rc.q, err = rc.ch.QueueDeclare(
			qname, // name TODO: fix those configs, all of them, this is bollocks
			true,       // durable
			true,        // delete when usused
			false,       // exclusive
			false,       // no-wait
			nil,         // arguments
	)
	onError(err, "Failed to declare a queue", true)
	
	log.Printf("Setting Qos to 1 prefetch full") // TODO: configurable?
	err = rc.ch.Qos(
			1,     // prefetch count
			0,     // prefetch size
			false, // global
	)
	onError(err, "Failed to set QoS", false)
	
	log.Printf("Starting a consumer channel")
	rc.Msgs, err = rc.ch.Consume(
			rc.q.Name, // queue
			"",     // consumer
			false,  // auto-ack
			false,  // exclusive
			false,  // no-local
			false,  // no-wait
			nil,    // args
	)
	
	
	
	onError(err, "Failed to register a consumer", true)
	return true
}

func (rc *RabbitConnector) Reply(pub []byte, d amqp.Delivery) error {
	return rc.ch.Publish(
			"",        // exchange
			d.ReplyTo, // routing key // is handled by sender.
			false,     // mandatory
			false,     // immediate
			amqp.Publishing{
					ContentType:   "application/json",
					CorrelationId: d.CorrelationId, // CorrelationId is handled by sender.
					Body:          pub,
			})
}