package main

import (
	"container/list"
	"context"
	"log"
	"net"
	"time"
	"fmt"
	"sync"
	"../proto"
	guuid "github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"github.com/golang/protobuf/ptypes"
	"github.com/streadway/amqp"
	"encoding/json"
)

type server struct{}

type registroOrden struct {
	timestamp      	time.Time
	//id            guuid.UUID
	id		string
	tipo          	string
	nombreProducto 	string
	valor          	int64
	origen         	string
	destino        	string
	seguimiento	string
	estado string
	//seguimiento   guuid.UUID
}

type paquete struct {
	//id           guuid.UUID
	id	     string
	tipo         string
	valor        int64
	origen       string
	destino      string
	camion	     int64
	intentos     int64
	fechaentrega time.Time
}

type messageFinanzas struct{
	Id           string	
	Intentos     int64	
	Estado 	     string	
	Valor        int64	
	Tipo         string	
}

//mutex
var mutexColas = &sync.Mutex{}
var mutexRegistro = &sync.Mutex{}
var mutexCamion = &sync.Mutex{}

//gestion de sync
var cond_colas = sync.NewCond(mutexColas)
var cond_camion = sync.NewCond(mutexCamion) 

//crear las colas
//var retail = list.New()
var retail[]paquete
var prioritario[]paquete
//var normal = list.New()
var normal[]paquete
//lista de ordenes
var registroOrdenes = list.New()

//lista registros finanzas
var registroFinanzas = list.New()

//mutex
var mutexFinanzas = &sync.Mutex{}
var Cond *sync.Cond=sync.NewCond(mutexFinanzas)

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func escuchar(llave_colas *sync.Cond, llave_camion *sync.Cond, cola_paquetes *[]paquete, clase_cola string){
	conn, err:= grpc.Dial("10.6.40.249:50052",grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %s", err)
	}
	cliente := proto.NewServicioCamionClient(conn)
	//var paquete proto.Paquete
	for{	
		llave_colas.L.Lock()
		for len(*cola_paquetes) == 0{//cola agotada
			fmt.Printf("Esperando paquetes cola %s\n", clase_cola)
			llave_colas.Wait()
		}
		fmt.Printf("Voy: %s\n", clase_cola)
		var paquete proto.Paquete
		var revisando int64
		ultimo := (*cola_paquetes)[len(*cola_paquetes) - 1]
		var revisando_id string = ultimo.id 
		fecha, _ := ptypes.TimestampProto(ultimo.fechaentrega)
		if clase_cola == "normal"{ //***************************************************************************************************************************
			for len((prioritario)) > 0 {//la cola normal debe dar preferencia a la cola prioritaria
				fmt.Printf("Normal packets waiting until priorized queue is empty.\n")
				llave_colas.Wait() //espera que se agote la cola de paquetes prioritarios
			}
			normal = normal[:len(normal)-1]//Se elimina el paquete al final de la fila.
			llave_colas.L.Unlock()

			request := proto.Disponibilidad{ //consulta si el camion encargado esta disponible
				Camion : 2,
			} 
			revisando = 2
			llave_camion.L.Lock()
			response, err := cliente.Consultar(context.Background(), &request)
			if err != nil{
				log.Fatalf("Error al invocar el metodo: %s", err)
			}
			fmt.Printf("Consultado camion\n")
			flag := response.GetRespuesta()
			for flag == false{ //camion no disponible
				llave_camion.Wait() // espera que regrese algun camion
				for len(prioritario) > 0 { // luego de despertar verifica que no hayan llegado nuevos paquete a la cola de prioridad
					fmt.Printf("Esperando al camion 2\n")
					llave_camion.Wait()	
				}
 
				response, err := cliente.Consultar(context.Background(), &request) //pregunta otra vez
				if err != nil{
					log.Fatalf("Error al invocar el metodo: %s", err)
				}
				flag = response.GetRespuesta()
				
			}
			
			paquete = proto.Paquete{
				Tipo: ultimo.tipo,
				Id: ultimo.id,
				Valor: ultimo.valor,
				Origen: ultimo.origen,
				Destino: ultimo.destino,
				Camion: 2,
				Intentos: ultimo.intentos,
				Fechaentrega: fecha,
			}

			//_, err = cliente.Despachar(context.Background(), &paquete)
			//if err != nil{
			//	log.Fatalf("Error al invocar el metodo: %s", err)
			//}

		}else{//paquetes en cola prioritario o retail *************************************************************************************************************
			if clase_cola == "prioritario"{
				//fmt.Printf("Paquetes en cola prioritaria: %v\n", len(prioritario))
				prioritario = prioritario[:len(prioritario) - 1]
				llave_colas.L.Unlock()
				if len(retail) > 0 {
					fmt.Printf("Can't use trucks 0 and 1 until all retail packets are sent.\n")
					request := proto.Disponibilidad{//pregunta por el camion numero 3
						Camion : 2,
					}
					llave_camion.L.Lock()
					response, err := cliente.Consultar(context.Background(), &request)
					if err != nil{
						log.Fatalf("Error al invocar el metodo Consultar: %s", err)
					}
					flag := response.GetRespuesta()
					for flag == false {
						llave_camion.Wait() //Espera el regreso de algun camion
						response, err := cliente.Consultar(context.Background(), &request)
						if err != nil {
							log.Fatalf("Error al invocar el metodo Consultar: %s", err)
						}
						flag = response.GetRespuesta()
					}
					paquete = proto.Paquete{
						Tipo: ultimo.tipo,
						Id: ultimo.id,
						Valor: ultimo.valor,
						Origen: ultimo.origen,
						Destino: ultimo.destino,
						Camion: 2,
						Intentos: ultimo.intentos,
						Fechaentrega: fecha,
					}
					revisando = 2

				}else{ //we can ask for trucks 0, 1 and 2 
					var camion int64 = 99
					llave_camion.L.Lock()
					for i := 0; i < 3; i++{
						var pos int64 = int64(i) 
						request := proto.Disponibilidad{
							Camion: pos,
						}
						response, err := cliente.Consultar(context.Background(), &request)
						if err != nil{
							log.Fatalf("Error al invocar el metodo Consultar: %s", err)
						}
						flag := response.GetRespuesta()
						if flag == true {
							camion = pos
							break
						}
						if i == 2{ //llegar a este punto implica que los tres camiones estan en uso
							i = 0 //reinicia el contador
							llave_camion.Wait() //espera que retorne algun camion
						}
						
						
					}
					paquete = proto.Paquete{
						Tipo: ultimo.tipo,
						Id: ultimo.id,
						Valor: ultimo.valor,
						Origen: ultimo.origen,
						Destino: ultimo.destino,
						Camion: camion,
						Intentos: ultimo.intentos,
						Fechaentrega: fecha,
					}
					revisando = camion

				} 
				//_, err = cliente.Despachar(context.Background(), &paquete)
				
				//if err != nil{
				//	log.Fatalf("Error al invocar el metodo Despachar: %s", err)
				//}

				//fmt.Printf("Paquete enviado: %s\n", clase_cola)

			}else{ //Paquete de clase retail ***********************************************************************
				//ultimo := retail[len(retail)]
				retail = retail[:len(retail) - 1]
				var camion int64 = 99
				llave_colas.L.Unlock()
 				llave_camion.L.Lock()
				for i := 0 ; i < 2 ; i++{
					var pos int64 = int64(i)
					request := proto.Disponibilidad{
						Camion: pos,
					} 
					response, err := cliente.Consultar(context.Background(), &request)
					if err != nil{
						log.Fatalf("Error al invocar la funcion Consultar: %s", err)
					}
					flag := response.GetRespuesta()
					if flag == true{
						camion = pos
						break
					}
					if i == 1{
						i = 0
						llave_camion.Wait()
					}
				}
				
				paquete = proto.Paquete{
					Tipo: ultimo.tipo,
					Id: ultimo.id,
					Valor: ultimo.valor,
					Origen: ultimo.origen,
					Destino: ultimo.destino,
					Camion: camion,
					Intentos: ultimo.intentos,
					Fechaentrega: fecha,
				}
				revisando = camion
				//_, err = cliente.Despachar(context.Background(), &paquete)
				//if err != nil{
				//	log.Fatalf("Error al invocar el metodo Despachar: %s", err)
				//}
				//fmt.Printf("Paquete enviado: %s\n", clase_cola)


			}
			
		}
		_, err = cliente.Despachar(context.Background(), &paquete)
		if err != nil{
			log.Fatalf("Error al invocar el metodo Despachar: %s", err)
		}
		fmt.Printf("Paquete",clase_cola," id",revisando_id ,"enviado al camion", revisando)
		mutexRegistro.Lock()
		for e := registroOrdenes.Front(); e != nil; e = e.Next(){
			if e.Value.(registroOrden).id==paquete.Id{
				temp:=e.Value.(registroOrden)
				registroOrdenes.Remove(e)
				temp.estado="en camino"
				registroOrdenes.PushBack(temp)
				break
			}
	
		}
		mutexRegistro.Unlock()
		llave_camion.L.Unlock()
		llave_camion.Broadcast()
		llave_colas.Broadcast()
		
	}
}





func main() {
	//hebras permanentes que revisan el estado de las colas
	go escuchar(cond_colas, cond_camion, &normal, "normal")
	go escuchar(cond_colas, cond_camion, &prioritario, "prioritario")
	go escuchar(cond_colas, cond_camion, &retail, "retail")

	//hebra que reporta a finanzas
	go connectToFinances()

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen on port 50051: %v", err)
	}

	srv := grpc.NewServer()
	proto.RegisterAddServiceServer(srv, &server{})
	reflection.Register(srv)

	if e := srv.Serve(listener); e != nil {
		log.Fatalf("failed to Serve on port 4040: %v", e)
	}

}

func (s *server) Add(ctx context.Context, request *proto.Request) (*proto.Response, error) {
	a, b := request.GetA(), request.GetB()
	result := a + b
	time.Sleep(1 * 60 * time.Second)
	return &proto.Response{Result: result}, nil
}

func (s *server) Multiply(ctx context.Context, request *proto.Request) (*proto.Response, error) {
	a, b := request.GetA(), request.GetB()
	result := a * b

	return &proto.Response{Result: result}, nil
}

func (s *server) Order(ctx context.Context, request *proto.ClientRequest) (*proto.ResponseToClient, error) {
	id := guuid.New().String()
	seguimiento := guuid.New().String()
	orden := registroOrden{
		timestamp:      time.Now(),
		id:             id,
		tipo:           request.GetTipo(),
		nombreProducto: request.GetNombreProducto(),
		valor:          request.GetValor(),
		origen:         request.GetOrigen(),
		destino:        request.GetDestino(),
		seguimiento:    seguimiento,
		estado:			"en bodega",
	}
	mutexRegistro.Lock()
	registroOrdenes.PushBack(orden)
	mutexRegistro.Unlock()
	paquete :=  paquete{
		id:       id,
		tipo:     request.GetTipo(),
		valor:    request.GetValor(),
		origen:   request.GetOrigen(),
		destino:  request.GetDestino(),
		intentos: 0,
		fechaentrega: time.Time{},//los paquetes se inicializan con un valor por defecto
	}
	switch request.GetTipo() {
	case "retail":
		mutexColas.Lock()
		//retail.PushBack(paquete)
		retail = append(retail,paquete)
		mutexColas.Unlock()
		break
	case "prioritario":
		mutexColas.Lock()
		//prioritario.PushBack(paquete)
		prioritario = append(prioritario,paquete)
		//fmt.Printf("Nuevo paquete: %v\n", len(prioritario))
		mutexColas.Unlock()
		break
	case "normal":
		mutexColas.Lock()
		//normal.PushBack(paquete)
		normal = append(normal,paquete)
		mutexColas.Unlock()
		break
	}
	cond_colas.Broadcast()
	//cond_camion.Broadcast()
	return &proto.ResponseToClient{Seguimiento: seguimiento}, nil
}

func (s *server) RequestEstado(ctx context.Context, request *proto.EstadoRequest) (*proto.ResponseToClient, error){
	mutexRegistro.Lock()
	codigo, err := guuid.Parse(request.GetSeguimiento())
	if err != nil{
		log.Fatalf("failed to parse uuid")
	}
	for e := registroOrdenes.Front(); e != nil; e = e.Next(){
		if e.Value.(registroOrden).seguimiento==codigo.String(){
			mutexRegistro.Unlock()
			return &proto.ResponseToClient{Seguimiento: e.Value.(registroOrden).estado}, nil// retorna tipo por mientras, ya que el registro no lleva el estado de la orden segun el pdf wtf
		}

	}
	return &proto.ResponseToClient{Seguimiento: "NOT FOUND"}, nil
}

func (s *server) ReportarDespacho(ctx context.Context, request *proto.Reporte) (*proto.Response, error){
	//reportar a finanzas
	var estado string
	if request.GetEntregado(){
		estado= "recibido"
	}else{
		estado= "no recibido"
	}
	mutexRegistro.Lock()
		for e := registroOrdenes.Front(); e != nil; e = e.Next(){
			if e.Value.(registroOrden).id==request.GetId(){
				temp:=e.Value.(registroOrden)
				registroOrdenes.Remove(e)
				temp.estado=estado
				registroOrdenes.PushBack(temp)
				break
			}
	
		}
	mutexRegistro.Unlock()
	message := toFinance(request.GetId(),request.GetIntentos(),estado,request.GetValor(),request.GetTipo())
	mutexFinanzas.Lock()
	registroFinanzas.PushBack(message)
	mutexFinanzas.Unlock()
	Cond.Signal()
	cond_camion.Broadcast()
	return &proto.Response{Result: 1}, nil	
}

func toFinance(id string, intentos int64, estado string, valor int64, tipo string)([]byte){
	//fmt.Println("Parametros********** Intentos:",intentos,"Valor:",valor,"Estado",estado)
	m:=messageFinanzas{Id: id,
			Intentos: intentos,
			Estado: estado, 
			Valor: valor, 
			Tipo: tipo,}
	//fmt.Println("Id:",id,"Intentos:",intentos,"Estado:",estado)
	message, err := json.Marshal(m)
	failOnError(err, "Failed to encode a message")
	
	return message
}

func connectToFinances() {
	conn, err := amqp.Dial("amqp://admin:admin@10.6.40.249:5672/")
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
				//ContentType: "text/plain",
				Body:        message,
			})
		
		failOnError(err, "Failed to publish a message")
		mutexFinanzas.Unlock()
	}

		
}
