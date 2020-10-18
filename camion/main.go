package main

import(
	"container/list"
	"context"
	"log"
	"net"
	"time"
	"fmt"
	"sync"
	"../proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"math/rand"
	"github.com/golang/protobuf/ptypes"

)
	//"os"
	//"bufio"	
	//guuid "github.com/google/uuid"


type server struct{}

type paquete struct{	
	//id		guuid.UUID
	id		string
	tipo		string
	valor		int64
	origen		string
	destino		string
	camion		int64
	intentos	int64
	fechaentrega	time.Time
}

type reporte struct{
	id		string
	tipo		string
	valor		int64
	intentos	int64
	entregado	bool
	fechaentrega	time.Time			
}
var registroDespachos = list.New()

//Flags que permiten gestionar la asignacion de camiones. Cada uno representa la disponibilidad de un camion
var flag = [3]bool{true,true,true}

//Contador de paquetes para cada camion
var contadores = [3]int{0,0,0}

var camion_cero [2]paquete
var camion_uno [2]paquete
var camion_dos [2]paquete

var registro_cero []paquete
var registro_uno []paquete
var registro_dos []paquete

//Contador que asigna secuencialmente id a cada paquete nuevo
var contador_id_paquete int = 1

//Sincronizadores para manejar el tiempo de espera de los camiones
var mutexCero = &sync.Mutex{}
var mutexUno = &sync.Mutex{}
var mutexDos = &sync.Mutex{}

var cond_cero = sync.NewCond(mutexCero)
var cond_uno = sync.NewCond(mutexUno)
var cond_dos = sync.NewCond(mutexDos)

//Valores ingresados por el operador para tiempos de espera
var tiempo_espera int
var tiempo_entrega int



func gestionenvios(registro_camion *[]paquete, paquetes_camion *[2]paquete, camion int, llave_camion *sync.Cond, flag_camion *[3]bool){
	conn, err := grpc.Dial ("10.6.40.247:50051", grpc.WithInsecure())
	if err != nil{
		log.Fatalf("Failed to connect: %s", err)
	}
	client := proto.NewAddServiceClient(conn)	
	

	for{
		var reportes []proto.Reporte
		//fmt.Println("Numero reportes: ", len(reportes))
		llave_camion.L.Lock()
		for contadores[camion] == 0 {
			fmt.Println("Im chilling now: ", camion)
			llave_camion.Wait() //la sincronizacion permite al camion comenzar a operar al momento apropiado
		}
	
		for i := 0; i < tiempo_espera && (flag[camion]==true); i++{
			fmt.Println("Camion ", camion," Esperando: ", i+1)
			llave_camion.L.Unlock()
			time.Sleep(1*time.Second)
			llave_camion.L.Lock()
		}
		llave_camion.L.Unlock()
		fmt.Println("Camion", camion, "on the move")
		var paquete_elegido int
		if (*paquetes_camion)[0].valor > (*paquetes_camion)[1].valor {
			paquete_elegido = 0	
		}else{
			paquete_elegido = 1
		}
		
		//Una vez cargados los paquetes al camion, se inicia el despacho
		for {
			fmt.Println("Paquete",paquete_elegido,"elegido para el camion",camion)
			if(*paquetes_camion)[paquete_elegido].id == ""{//el paquete actual ya fue entregado
				paquete_elegido = (paquete_elegido+1)%2
				if (*paquetes_camion)[paquete_elegido].id == ""{//el segundo paquete tambien lo fue
					fmt.Println("Camion",camion,"regresando a la central")
					break
				}
			}
			tipo_paquete := (*paquetes_camion)[paquete_elegido].tipo
			intentos_paquete := (*paquetes_camion)[paquete_elegido].intentos
			id_paquete := (*paquetes_camion)[paquete_elegido].id
			if ((tipo_paquete != "retail") && (intentos_paquete) < 2) || ((tipo_paquete == "retail") && (intentos_paquete < 3) ){
				//simular envio
				fmt.Println("Camion", camion , " entregando paquete ",id_paquete)
				var exito float64 = 0.8
				//(*paquetes_camion)[paquete_elegido].intentos++
				simulado := rand.Float64()
				//fmt.Println("Simulado: ",simulado)
				(*paquetes_camion)[paquete_elegido].intentos++
				time.Sleep(time.Duration(tiempo_entrega)*time.Second)
				if(simulado <= exito){
					//fmt.Printf("%f is greater than %f", 0.8, simulado)
					fecha , _ := ptypes.TimestampProto(time.Now())
					reporte := proto.Reporte{
						Id:		(*paquetes_camion)[paquete_elegido].id,
						Tipo:		(*paquetes_camion)[paquete_elegido].tipo,
						Valor:		(*paquetes_camion)[paquete_elegido].valor,
						Entregado: 	true,
						FechaEntrega:	fecha,
					}
					reportes = append(reportes,reporte)
					fmt.Println("Camion", camion ,"entrego", id_paquete)
					//agregar paquete al registro historico
					(*paquetes_camion)[paquete_elegido].id = "" //flag que remueve el paquete del camion
				}else{
					//fmt.Println("%f is lower than %f", simulado, 0.8)
					fmt.Println("Camion", camion , "fallo", intentos_paquete+1, "veces entregar el paquete", id_paquete)
					//(*paquetes_camion)[paquete_elegido].intentos++
					//paquete_elegido = (paquete_elegido+1)%2
				}			
			}else{//el paquete alcanzo su maximo de intentos.
				//agregar al registro historico
				fmt.Println("Camion ", camion," fallo", intentos_paquete," veces con la entrega del paquete", id_paquete,". Paquete no entregado")
				fecha,_ := ptypes.TimestampProto(time.Now())
				reporte := proto.Reporte{
					Id: 		(*paquetes_camion)[paquete_elegido].id,
					Tipo:		(*paquetes_camion)[paquete_elegido].tipo,
					Valor:		(*paquetes_camion)[paquete_elegido].valor,
					Intentos:	(*paquetes_camion)[paquete_elegido].intentos,
					Entregado:	false,
					FechaEntrega:	fecha,
					
				}
				reportes = append(reportes,reporte)
				fmt.Println("Paquete no recibido")
				(*paquetes_camion)[paquete_elegido].id = ""
			}
			paquete_elegido = (paquete_elegido+1)%2
	
		}
		contadores[camion] = 0
		flag[camion] = true
		for i:=0 ; i < len(reportes) ; i++{ //Se envian los reportes elaborados
			_, err = client.ReportarDespacho(context.Background(), &reportes[i])
			
		}
		
	}
}

func main(){
	fmt.Println("Ingrese tiempo de espera camion en segundos")
	_, err := fmt.Scanf("%d", &tiempo_espera)

	fmt.Println("Ingrese tiempo de entrega de pedidos en segundos")
	_, err = fmt.Scanf("%d", &tiempo_entrega)
	
	go gestionenvios(&registro_cero, &camion_cero, 0, cond_cero, &flag)
	go gestionenvios(&registro_uno, &camion_uno, 1, cond_uno, &flag)
	go gestionenvios(&registro_dos, &camion_dos, 2, cond_dos, &flag)

	listener, err := net.Listen("tcp",":50052")
	if err != nil{
		log.Fatalf("Failed to listen on port 50052: %v\n", err)
	}

	srvr := grpc.NewServer()
	proto.RegisterServicioCamionServer(srvr, &server{})
	reflection.Register(srvr)

	if e:= srvr.Serve(listener); e != nil{
		log.Fatalf("Failed to Serve on port 50051: %v\n", e)
	}

	//reader := bufio.NewReader(os.Stdin)
 
}

func (s *server) Despachar(ctx context.Context, request *proto.Paquete) (*proto.RespuestaPedido, error){
	camion := request.GetCamion()
	fecha, _ := ptypes.Timestamp(request.GetFechaentrega())
	paquete := paquete{
		tipo: 		request.GetTipo(),
		id: 		request.GetId(),
		valor: 		request.GetValor(),
		origen: 	request.GetOrigen(),
		destino: 	request.GetDestino(),
		camion: 	camion,
		intentos: 	request.GetIntentos(),
		fechaentrega: 	fecha,
	}
	//fmt.Println("Camion paquete: ", paquete.camion)
	cond_cero.L.Lock()
	if camion == 0 && flag[0] == true{
		if contadores[0] == 1{
			//fmt.Println("Enviando!")
			camion_cero[1] = paquete
			flag[0] = false
			contadores[0] = 0
			cond_cero.Signal()
			cond_cero.L.Unlock()
		}else{
			contadores[0]++
			camion_cero[0] = paquete
			cond_cero.Signal()
			cond_cero.L.Unlock()	
		}	
	}else{
		cond_cero.L.Unlock()
		cond_uno.L.Lock()
		if camion == 1 && flag[1] == true{
			if contadores[1] == 1{
				camion_uno[1] = paquete
				flag[1] = false
				contadores[1] = 0
				cond_uno.Signal()
				cond_uno.L.Unlock()
			}else{
				contadores[1]++
				camion_uno[0] = paquete
				cond_uno.Signal()
				cond_uno.L.Unlock()
			}
		}else{//camion == 2
			cond_uno.L.Unlock()
			cond_dos.L.Lock()
			if camion == 2 && flag[2] == true{
				if contadores[2] == 1{
					camion_dos[1] = paquete
					flag[2] = false
					contadores[2] = 0
					cond_dos.Signal()
					cond_dos.L.Unlock()
				}else{
					contadores[2]++
					camion_dos[0] = paquete
					cond_dos.Signal()
					cond_dos.L.Unlock()
				}
			
			}

		}
		
	}
	//fmt.Println("We got that sucka")
	return &proto.RespuestaPedido{Esperando: []bool{true}}, nil
}

func (s *server) Consultar(ctx context.Context, request *proto.Disponibilidad) (*proto.RespuestaDisponibilidad, error){
	camion := request.GetCamion()
	respuesta := flag[camion]
	return &proto.RespuestaDisponibilidad{Respuesta: respuesta}, nil
}