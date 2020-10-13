package main

import (
	"context"
	"container/list"
	"google.golang.org/grpc"
	"log"
	"../proto"
	"os"
	"bufio"
	"fmt"
	"io"
	"encoding/csv"
	"strconv"
)

func main(){
	conn, err := grpc.Dial("localhost:4040", grpc.WithInsecure())
	if err != nil{
		log.Fatalf("failed to connect: %s", err)
	}

	client := proto.NewAddServiceClient(conn)

	/*
	message := proto.Request{
		A:2,
		B:4,
	}
	*/

	
	fmt.Println("press p for pymes, r for retail")

	reader := bufio.NewReader(os.Stdin)
	char, _, err := reader.ReadRune()
	var tipo string

	//lista de códigos de seguimiento
	listaSeguimiento := list.New()

	if err != nil {
	fmt.Println(err)
	}
	
	switch char {
	case 'p':
	fmt.Println("p Key Pressed")
	csvfile, err := os.Open("pymes.csv")
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}
	r := csv.NewReader(csvfile)
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
		}else{
			tipo="normal"
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
		if err != nil{
			log.Fatalf("error when calling Order: %s", err)
		}
		listaSeguimiento.PushBack(response.Seguimiento)
	}
	
	break

	case 'r':
	fmt.Println("r Key Pressed")
	csvfile, err := os.Open("retail.csv")
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}
	r := csv.NewReader(csvfile)
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

		response, err := client.Order(context.Background(), &request)
		if err != nil{
			log.Fatalf("error when calling Order: %s", err)
		}
		listaSeguimiento.PushBack(response.Seguimiento)
	}
	break
	}
	
	

}

func pedirEstado(idSeguimiento string){
		
}