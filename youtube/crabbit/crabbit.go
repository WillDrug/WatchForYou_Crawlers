package crabbit

import (
	"github.com/streadway/amqp"
	"log"
//	"fmt"
//	"../cconfig" // not needed to be a proper interface
	"../cinterfaces"
)


type RabbitConnector struct {
	conn *amqp.Connection
	ch 	 *amqp.Channel
	q	 amqp.Queue
	iMsgs <-chan amqp.Delivery
	Msgs chan cinterfaces.Message
}

func (rc *RabbitConnector) Connect(cstr string) error {
	var err error
	
	log.Printf("Dialing %s", cstr)
	rc.conn, err = amqp.Dial(cstr)
	if err != nil { return err }
	
	log.Printf("Creating a channel")
	rc.ch, err = rc.conn.Channel()
	return err
}

func (rc *RabbitConnector) Disconnect() {
	rc.conn.Close()
	rc.ch.Close()
}

func (rc *RabbitConnector) Consume(qname string) error {
	//rc := &rrc
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
	if err != nil { return err }
	
	log.Printf("Setting Qos to 1 prefetch full") // TODO: configurable?
	err = rc.ch.Qos(
			1,     // prefetch count
			0,     // prefetch size
			false, // global
	)
	if err != nil { return err }
	
	log.Printf("Starting a consumer channel")
	rc.iMsgs, err = rc.ch.Consume(
			rc.q.Name, // queue
			"",     // consumer
			false,  // auto-ack
			false,  // exclusive
			false,  // no-local
			false,  // no-wait
			nil,    // args
	)
	
	return err
}

// testing! TODO: review
var replier map[string]amqp.Delivery

func (rc *RabbitConnector) InboundMessages(workers int) chan cinterfaces.Message {
	replier = make(map[string]amqp.Delivery)
	if workers != 1 {
		panic("Don't know how to handle more than 1 thread yet")
	}
	go func() {
		for d := range rc.iMsgs {
			replier[d.CorrelationId] = d
			rc.Msgs<-cinterfaces.Message{d.Body, d.CorrelationId}
		}
	}()
	return rc.Msgs
}
	

func (rc *RabbitConnector) ReplyToMessage(pub []byte, msg cinterfaces.Message) error {
	original := replier[msg.Context]
	err := rc.ch.Publish(
			"",        // exchange
			original.ReplyTo, // routing key // is handled by sender.
			false,     // mandatory
			false,     // immediate
			amqp.Publishing{
					ContentType:   "application/json",
					CorrelationId: msg.Context, // CorrelationId is handled by sender.
					Body:          pub,
			})
	if err != nil {
		original.Ack(false)
	} else {
		original.Ack(true)
	}	
	delete(replier, msg.Context)
	return err
}