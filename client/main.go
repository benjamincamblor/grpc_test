package main

import (
	"context"
	"container/list"
	"google.golang.org/grpc"
	"log"
	"../proto"
	"time"
	"os"
	"bufio"
	"fmt"
	"io"
	"encoding/csv"
	"strconv"
	"sync"
)

//lista de códigos de seguimiento
var listaSeguimiento = list.New()

var mutexSeguimiento = &sync.Mutex{}

func main(){
	
	var segundos int
	fmt.Println("Intervalo en segundos entre órdenes del cliente")
	_, err := fmt.Scanf("%d", &segundos)
	if err != nil{
		fmt.Println(err)
	}
	
	conn, err := grpc.Dial("10.6.40.247:50051", grpc.WithInsecure())
	if err != nil{
		log.Fatalf("failed to connect: %s", err)
	}

	var client = proto.NewAddServiceClient(conn)

	
	/*
	message := proto.Request{
		A:2,
		B:4,
	}
	*/

	
	fmt.Println("P para pymes, R para retail")

	reader := bufio.NewReader(os.Stdin)
	char, _, err := reader.ReadRune()
	var tipo string


	if err != nil {
		fmt.Println(err)
	}
	
	switch char {
	case 'p':
	fmt.Println("p Key Pressed")
	csvfile, err := os.Open("../pymes.csv")
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}
	r := csv.NewReader(csvfile)

	go func(){
		time.Sleep(1 * 30 * time.Second)
		for {
			mutexSeguimiento.Lock()
			for e := listaSeguimiento.Front(); e != nil; e = e.Next(){
				message:= proto.EstadoRequest{
					Seguimiento: e.Value.(string),
				}
				response, err:=client.RequestEstado(context.Background(), &message)
				if err != nil{
					log.Fatalf("error when calling Order: %s", err)
				}
				fmt.Println("Order número de seguimiento "+e.Value.(string)+", estado:"+response.Seguimiento)
				mutexSeguimiento.Unlock()
				time.Sleep(time.Duration(segundos)*60*time.Second)
				mutexSeguimiento.Lock()
			}
			mutexSeguimiento.Unlock()
		}
	}()
	
	_, _ = r.Read() //Lee la primera linea del archivo csv, de otra manera el cliente enviara una orden basura

	for{	
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		if  record[5]=="1"{
			tipo="prioritario"
			fmt.Printf("Enviado prioritario\n")
		}else{
			tipo="normal"
			fmt.Printf("Enviado normal\n")
		}
		valor, err := strconv.ParseInt(record[2], 10, 64)
		if err != nil {
			fmt.Printf("Error converting valor")
		}

		request := proto.ClientRequest{
			Tipo:tipo,
			NombreProducto: record[1],
			Valor: valor,
			Origen: record[3],
			Destino: record[4],
		}

		response, err := client.Order(context.Background(), &request)
		fmt.Println("ID:",response.Seguimiento)	
		if err != nil{
			log.Fatalf("error when calling Order: %s", err)
		}

		mutexSeguimiento.Lock()
		listaSeguimiento.PushBack(response.Seguimiento)
		mutexSeguimiento.Unlock()
		time.Sleep(time.Duration(segundos)*time.Second)
	}
	
	break

	case 'r':
	fmt.Println("r Key Pressed")
	csvfile, err := os.Open("../retail.csv")
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}
	
	r := csv.NewReader(csvfile)
	
	_,_ = r.Read()
	for{
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		valor, err := strconv.ParseInt(record[2], 10, 64)
		if err != nil {
			fmt.Printf("Error converting valor")
		}

		request := proto.ClientRequest{
			Tipo:"retail",
			NombreProducto: record[1],
			Valor: valor,
			Origen: record[3],
			Destino: record[4],
		}

		_, err = client.Order(context.Background(), &request)
		if err != nil{
			log.Fatalf("error when calling Order: %s", err)
		}
		fmt.Printf("Enviado: Retail\n")
		time.Sleep(time.Duration(segundos)*time.Second)
	}
	break
	}
	
	

}
/*
func pedirEstado(){
	time.Sleep(1 * 60 * time.Second)
	for {
		for e := listaSeguimiento.Front(); e != nil; e = e.Next(){
			message:= proto.EstadoRequest{
				Seguimiento: e.Value.(string),
			}
			response, err:=client.RequestEstado(context.Background(), &message)
			if err != nil{
				log.Fatalf("error when calling Order: %s", err)
			}
		}
	}
}
*/
