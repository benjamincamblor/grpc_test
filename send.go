package main

import (
	//"strconv"
  "log"
  "container/list"
  "github.com/streadway/amqp"
  "sync"
  "encoding/json"
)

type messageFinanzas struct{
	id           string
	intentos     int
	estado 		 string
	valor        int64
	tipo         string
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

//lista registros finanzas
var registroFinanzas = list.New()

//mutex
var mutexFinanzas = &sync.Mutex{}
var Cond *sync.Cond=sync.NewCond(mutexFinanzas)

func main() {
	conn, err := amqp.Dial("amqp://admin:admin@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	failOnError(err, "Failed to declare a queue")

	for {
		mutexFinanzas.Lock()
		if registroFinanzas.Len()==0 {
			Cond.Wait()
		}
		message:=registroFinanzas.Remove(registroFinanzas.Front()).([]byte)
	
		err = ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        message,
			})
		
		failOnError(err, "Failed to publish a message")
		mutexFinanzas.Unlock()
	}

		
}

func toFinance(id string, intentos int, estado string, valor int64, tipo string)([]byte){
	m:=messageFinanzas{id,intentos,estado, valor, tipo}
	message, err := json.Marshal(m)
	failOnError(err, "Failed to encode a message")

	return message
}