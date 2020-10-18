package main

import (
	"bufio"
	"log"
	"github.com/streadway/amqp"
	"encoding/json"
	"encoding/csv"
	"os"
	"strconv"
	"fmt"
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

func main() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
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

	file, err := os.Create("result.csv")
    failOnError(err,"Cannot create file")
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()


	var cuenta float64 = 0
	var gastos float64 = 0
	var ingresos float64 = 0
	var m messageFinanzas
	var sum float64
	var end bool = false

	go func(){
		fmt.Println("press t to end execution")

		reader := bufio.NewReader(os.Stdin)
		char, _, err := reader.ReadRune()
		if err != nil {
			fmt.Println(err)
			}
		if char=='t'{
			end=true
		}
	}()

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			sum=0
			err:=json.Unmarshal(d.Body,&m)
			failOnError(err, "Failed to decode json")
			switch m.tipo{
			case"retail":
				sum+=float64(m.valor-(int64(m.intentos)-1)*10)
				ingresos+=float64(m.valor)
				gastos+=-float64((int64(m.intentos)-1)*10)
				break
			case "prioritario":
				if m.estado=="no recibido"{
					sum+=float64(m.valor)*0.3-float64((int64(m.intentos)-1)*10)
					ingresos+=float64(m.valor)*0.3
					gastos+=-float64((int64(m.intentos)-1)*10)
				}else{
					sum+=float64(m.valor)*1.3-float64((int64(m.intentos)-1)*10)
					ingresos+=float64(m.valor)*1.3
					gastos+=-float64((int64(m.intentos)-1)*10)
				}
				break
			case "normal":
				if m.estado=="no recibido"{
					sum+=-float64((int64(m.intentos)-1)*10)
					gastos+=-float64((int64(m.intentos)-1)*10)
				}else{
					sum+=float64(m.valor)-float64((int64(m.intentos)-1)*10)
					ingresos+=float64(m.valor)
					gastos+=-float64((int64(m.intentos)-1)*10)
				}
				break
			}
			cuenta+=sum
			err = writer.Write([]string{m.estado+","+strconv.Itoa(m.intentos)+","+fmt.Sprintf("%f", sum)})
        	failOnError(err,"Cannot write to file")
		}
	}()

	//log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	if !end{
		<-forever
	}else{
		log.Printf("Gastos: %f", gastos)
		log.Printf("Ingresos: %f", ingresos)
		log.Printf("Balance final: %f", cuenta)
	}
	
}