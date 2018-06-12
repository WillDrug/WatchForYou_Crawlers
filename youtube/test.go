package main

import (
        "fmt"
        "log"
        "os"
        "strconv"
        "strings"
		"encoding/json"
		       "math/rand"
        "github.com/streadway/amqp"
)

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

func fibonacciRPC(n Test) (res Test, err error) {
        conn, err := amqp.Dial("amqp://watchforyou:watchforyou@localhost:5672/")
        failOnError(err, "Failed to connect to RabbitMQ")
        defer conn.Close()

        ch, err := conn.Channel()
        failOnError(err, "Failed to open a channel")
        defer ch.Close()

        q, err := ch.QueueDeclare(
                "",    // name
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
                true,   // auto-ack
                false,  // exclusive
                false,  // no-local
                false,  // no-wait
                nil,    // args
        )
        failOnError(err, "Failed to register a consumer")

        corrId := randomString(32)
		pub, _ := json.Marshal(n)
        err = ch.Publish(
                "",          // exchange
                "rpc_queue", // routing key
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
                        failOnError(err, "Failed to convert body to integer")
                        break
                }
        }

        return
}

func main() {
        res, err := fibonacciRPC(Test{"Strings are good", 5})
        failOnError(err, "Failed to handle RPC request")

        log.Printf(" [.] Got %d", res)
}

func bodyFrom(args []string) int {
        var s string
        if (len(args) < 2) || os.Args[1] == "" {
                s = "30"
        } else {
                s = strings.Join(args[1:], " ")
        }
        n, err := strconv.Atoi(s)
        failOnError(err, "Failed to convert arg to integer")
        return n
}