package wconnector


import (
	"../winterfaces"
	"github.com/streadway/amqp"
	"log"
	"encoding/json"
)


type Connector struct {
	conn  *amqp.Connection
	ch 	 *amqp.Channel
	q	 amqp.Queue
}


func (cn *Connector) Connect(cstr string) error {
	var err error
	log.Printf("Dialing %s", cstr)
	cn.conn, err = amqp.Dial(cstr)
	if err != nil { return err }
	
	log.Printf("Creating a channel")
	cn.ch, err = cn.conn.Channel()
	
	return err
}

func (cn *Connector) Disconnect() {
	cn.conn.Close()
	cn.ch.Close()
}

func (cn *Connector) Consume(qname string) (*chan winterfaces.RPCRequest, error) {
	var err error
	
	// #declare ing queue. Consumer always declares!
	log.Printf("Declaring %s queue for RPC requests", qname)
	cn.q, err = cn.ch.QueueDeclare(
			qname, // name TODO: fix those configs, all of them, this is bollocks
			true,       // durable
			true,        // delete when usused
			false,       // exclusive
			false,       // no-wait
			nil,         // arguments
	)
	if err != nil { 
		return nil, err 
	}
	
	log.Printf("Setting Qos to 1 prefetch full") // TODO: configurable?
	err = cn.ch.Qos(
			1,     // prefetch count
			0,     // prefetch size
			false, // global
	)
	if err != nil { return nil, err }
	
	log.Printf("Starting a consumer channel")
	//var msgs chan amqp.Delivery
	msgs, err := cn.ch.Consume(
			cn.q.Name, // queue
			"",     // consumer
			false,  // auto-ack
			false,  // exclusive
			false,  // no-local
			false,  // no-wait
			nil,    // args
	)
	var retc chan winterfaces.RPCRequest

	retc = make(chan winterfaces.RPCRequest)

	var retv winterfaces.RPCRequest
	go func(msgs <-chan amqp.Delivery, retc chan winterfaces.RPCRequest) {
		log.Printf("Connector started listening")
		for d := range msgs {
			log.Printf("Connector got a message %#v", d.CorrelationId)
			err = json.Unmarshal(d.Body, &retv)
			retv.ReplyChannel = make(chan winterfaces.RPCResponse)
			if err != nil {
				log.Printf("Failed to marsahll %#v", d)
				go cn.ReplyToMessage(retv.ReplyChannel, d)
				retv.ReplyChannel <- winterfaces.RPCResponse{false, "Failed to parse request", make([]winterfaces.Entry, 0)}

			} else {
				retc <- retv
				go cn.ReplyToMessage(retv.ReplyChannel, d)
			}
		}
	}(msgs, retc)
	return &retc, err
}

func (cn *Connector) ReplyToMessage(respc chan winterfaces.RPCResponse, context amqp.Delivery) {
	log.Printf("Listening for a reply to %#v", context.CorrelationId)
	var err error
	resp := <- respc
	pub, err := json.Marshal(resp)
	log.Printf("Publishing reply to %#v", context.CorrelationId)
	if err == nil {
		err = cn.ch.Publish(
				"",        // exchange
				context.ReplyTo, // routing key // is handled by sender.
				false,     // mandatory
				false,     // immediate
				amqp.Publishing{
						ContentType:   "application/json",
						CorrelationId: context.CorrelationId, // CorrelationId is handled by sender.
						Body:          pub,
				})
	}
	if err != nil {
		context.Ack(false)
	} else {
		context.Ack(true)
	}	
	// TODO: Context passing to alert main code?
}
