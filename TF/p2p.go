package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
)

var localhostRegistro string
var localhostNotificar string
var localhostHP string
var remotehost string
var bitacoraAddr []string
var bitacoraAddr2 []string

const (
	registrarPort = 9100
	notificarPort = 9101
)

func main() {
	//saber su dirección del nodo
	bufferIn := bufio.NewReader(os.Stdin)
	fmt.Print("Ingrese el puerto de registro:")
	port, _ := bufferIn.ReadString('\n')
	port = strings.TrimSpace(port)
	localhostRegistro = fmt.Sprintf("localhost:%s", port) //reemplazar ip de cada nodo

	fmt.Print("Ingrese el puerto de notificación:")
	port, _ = bufferIn.ReadString('\n')
	port = strings.TrimSpace(port)
	localhostNotificar = fmt.Sprintf("localhost:%s", port)

	fmt.Print("Ingrese el puerto de proceso hot potato:")
	port, _ = bufferIn.ReadString('\n')
	port = strings.TrimSpace(port)
	localhostHP = fmt.Sprintf("localhost:%s", port)

	//configurar roll del server concurrente
	go RegistrarServer() //servicio de escucha para nuevas solicitudes

	//solicitud del nodo para unirse a la red

	fmt.Print("Ingrese el puerto del nodo:")
	puerto, _ := bufferIn.ReadString('\n')
	puerto = strings.TrimSpace(puerto)
	remotehost = fmt.Sprintf("localhost:%s", puerto)

	//consulta si es el primer nodo de la red
	if puerto != "" {
		registrarSolicitud(remotehost)
	}

	notificarServer()   //escuchando peticiones
	notificarServerHP() //escuchando las notificacionens HP que llegan

}

func RegistrarServer() {
	//cuál va a ser el puerto de escucha
	//localhost = fmt.Sprintf("localhost:%d", registrarPort)
	ln, _ := net.Listen("tcp", localhostRegistro)

	defer ln.Close()

	for {
		con, _ := ln.Accept()
		go manejadorRegistro(con)
	}

}
func manejadorRegistro(con net.Conn) {
	defer con.Close()
	//leer
	bufferIn := bufio.NewReader(con)
	ip, _ := bufferIn.ReadString('\n')
	ip = strings.TrimSpace(ip)

	ip2, _ := bufferIn.ReadString('\n')
	ip2 = strings.TrimSpace(ip2) //localhost puerto

	//responder al solicitante con la bitácora que tiene este nodo
	bytes, _ := json.Marshal(append(bitacoraAddr, localhostNotificar))
	fmt.Fprintln(con, string(bytes)) //envía la bitácora

	bytes, _ = json.Marshal(append(bitacoraAddr2, localhostHP)) //para servicio
	fmt.Fprintln(con, string(bytes))                            //envía la bitácora

	//comunicar a todos los nodos la llegada de uno nuevo
	comunicarTodos(ip)
	comunicarTodosHP(ip2)

	//actualizar la bitácora con el nuevo ip
	bitacoraAddr = append(bitacoraAddr, ip)
	bitacoraAddr2 = append(bitacoraAddr2, ip2)

	fmt.Println(bitacoraAddr)
	fmt.Println(bitacoraAddr2)

}
func comunicarTodos(ip string) {
	//recorrer toda la bitácora para comunicar
	for _, addr := range bitacoraAddr {
		notificar(addr, ip)
	}

}
func comunicarTodosHP(ip string) {
	for _, addr := range bitacoraAddr2 {
		notificar(addr, ip)
	}
}

func notificar(addr string, ip string) {
	con, _ := net.Dial("tcp", addr)

	defer con.Close()
	fmt.Fprintln(con, ip)

}

func registrarSolicitud(remotehost string) {

	con, _ := net.Dial("tcp", remotehost)
	defer con.Close()
	fmt.Fprintln(con, localhostNotificar) //enviamos el puerto de notificación
	fmt.Fprintln(con, localhostHP)

	//recuperar lo que responde el server
	bufferIn := bufio.NewReader(con)
	bitacoraServer, _ := bufferIn.ReadString('\n')

	var bitacoraTemp []string
	json.Unmarshal([]byte(bitacoraServer), &bitacoraTemp)

	bitacoraAddr = bitacoraTemp //agregamos al final de la bitácora su dirección

	bitacoraServer, _ = bufferIn.ReadString('\n')

	var bitacoraTemp2 []string
	json.Unmarshal([]byte(bitacoraServer), &bitacoraTemp2)

	bitacoraAddr2 = bitacoraTemp2

	fmt.Println(bitacoraAddr)
	fmt.Println(bitacoraAddr2)

}
func notificarServer() {
	ln, _ := net.Listen("tcp", localhostNotificar)
	defer ln.Close()

	for {
		con, _ := ln.Accept()
		go manejarRecibeNotificar(con)
	}

}
func manejarRecibeNotificar(con net.Conn) {
	defer con.Close()
	bufferIn := bufio.NewReader(con)
	ip, _ := bufferIn.ReadString('\n')
	ip = strings.TrimSpace(ip)
	bitacoraAddr = append(bitacoraAddr, ip)

	fmt.Println(bitacoraAddr)
}

func notificarServerHP() {
	ln, _ := net.Listen("tcp", localhostHP)
	defer ln.Close()

	for {
		con, _ := ln.Accept()
		go manejarRecibeNotificarHP(con)
	}

}
func manejarRecibeNotificarHP(con net.Conn) {
	defer con.Close()
	bufferIn := bufio.NewReader(con)
	ip, _ := bufferIn.ReadString('\n')
	ip = strings.TrimSpace(ip)
	bitacoraAddr2 = append(bitacoraAddr2, ip)

	fmt.Println(bitacoraAddr)
}

func servicioHP() {
	ln, _ := net.Listen("tcp", localhostHP)
	defer ln.Close()
	for {
		con, _ := ln.Accept()
		go manejadorHP(con)
	}
}

func manejadorHP(con net.Conn) {
	defer con.Close()
	bufferIn := bufio.NewReader(con)
	strnum, _ := bufferIn.ReadString('\n')
	strnum = strings.TrimSpace(strnum)
	num, _ := strconv.Atoi(strnum)

	fmt.Println("Número recibido:", num)

	if num == 0 {
		fmt.Println("Proceso finalizado!:")
	} else {
		enviarProximo(num)
	}

}
func enviarProximo(num int) {
	indice := rand.Intn(len(bitacoraAddr2))
	fmt.Printf("Enviando %d hacia %s", num, bitacoraAddr2[indice])
	con, _ := net.Dial("tcp", bitacoraAddr2[indice])
	defer con.Close()
	fmt.Fprintln(con, num-1)
}
