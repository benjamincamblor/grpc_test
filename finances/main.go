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
	"sync"
)

type messageFinanzas struct{
	Id           string  	
	Intentos     int64	
	Estado 		 string
	Valor        int64
	Tipo         string
}

var end bool 
var candado = &sync.Mutex{}

var cond_final = sync.NewCond(candado)
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func main() {
	end = false
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

	go func(){
		fmt.Println("press t to end execution")

		reader := bufio.NewReader(os.Stdin)
		char, _, err := reader.ReadRune()
		if err != nil {
			fmt.Println(err)
			}
		if char=='t'{
			end=true
			cond_final.Signal()
		}
	}()

//	forever := make(chan bool)
	go func() {
	for end == false{

		for d := range msgs {
			sum=0
			err:=json.Unmarshal(d.Body,&m)
			fmt.Println("Tipo:",m.Tipo," Id:",m.Id,"Intentos:",m.Intentos,"Estado:",m.Estado,"Valor:",m.Valor)
			failOnError(err, "Failed to decode json")
			switch m.Tipo{
			case"retail":
				sum+=float64(m.Valor-(m.Intentos-1)*10)
				ingresos+=float64(m.Valor)
				gastos+=-float64((m.Intentos-1)*10)
				break
			case "prioritario":
				if m.Estado=="no recibido"{
					sum+=float64(m.Valor)*0.3-float64((m.Intentos-1)*10)
					ingresos+=float64(m.Valor)*0.3
					gastos+=-float64((m.Intentos-1)*10)
				}else{
					sum+=float64(m.Valor)*1.3-float64((m.Intentos-1)*10)
					ingresos+=float64(m.Valor)*1.3
					gastos+=-float64((m.Intentos-1)*10)
				}
				break
			case "normal":
				if m.Estado=="no recibido"{
					sum+=-float64((m.Intentos-1)*10)
					gastos+=-float64((m.Intentos-1)*10)
				}else{
					sum+=float64(m.Valor)-float64((m.Intentos-1)*10)
					ingresos+=float64(m.Valor)
					gastos+=-float64((m.Intentos-1)*10)
				}
				break
			}
			cuenta+=sum
			err = writer.Write([]string{m.Id+";"+m.Estado+","+strconv.FormatInt(m.Intentos,10)+","+fmt.Sprintf("%f", sum)})
        	failOnError(err,"Cannot write to file")
		}
	
	}
	}()


	cond_final.L.Lock()
	for end == false{
		cond_final.Wait()
	}



	
	//fmt.Println("Variable fea:",end)
	//log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	
	defer file.Close()
	defer writer.Flush()
	log.Printf("Gastos: %f", gastos)
	log.Printf("Ingresos: %f", ingresos)
	log.Printf("Balance final: %f", cuenta)
	
}
