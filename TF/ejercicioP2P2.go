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

var localhostReg string
var localhostNot string
var localhostHp string
var remotehost string
var bitacoraAddr []string
var bitacoraAddr2 []string

func main() {
	//saber su dirección del nodo
	bufferIn := bufio.NewReader(os.Stdin)
	fmt.Print("Ingrese el puerto de registro: ")
	port, _ := bufferIn.ReadString('\n')
	port = strings.TrimSpace(port)
	localhostReg = fmt.Sprintf("localhost:%s", port) //reemplazar por la ip de cada nodo

	fmt.Print("Ingrese el puerto de notificacion: ")
	port, _ = bufferIn.ReadString('\n')
	port = strings.TrimSpace(port)
	localhostNot = fmt.Sprintf("localhost:%s", port) //reemplazar por la ip de cada nodo

	fmt.Print("Ingrese el puerto de proceso HP: ")
	port, _ = bufferIn.ReadString('\n')
	port = strings.TrimSpace(port)
	localhostHp = fmt.Sprintf("localhost:%s", port) //reemplazar por la ip de cada nodo

	//configurar rol de server concurrente
	go registrarServer() //servicio de escucha para nuevas solicitudes
	//lógica solicitud del nodo para unirse a la red
	go servicioHP()

	fmt.Print("Ingrese puerto del nodo a solicitar registro: ")
	puerto, _ := bufferIn.ReadString('\n')
	puerto = strings.TrimSpace(puerto)
	remotehost = fmt.Sprintf("localhost:%s", puerto)
	//consulta si es el primer nodo de la red
	if puerto != "" {
		registrarSolicitud(remotehost) //envio de solicitudes
	}
	recibeNotificarServer() //escuchando las notificaciones que llegan
	//recibeNotificarServerHP() //escuchando las notificaciones HP que llegan
}

func registrarServer() {
	//cual va a ser el puerto de escucha
	//localhost = fmt.Sprintf("localhost:%d", registrarPort)
	ln, _ := net.Listen("tcp", localhostReg)
	defer ln.Close()

	for {
		con, _ := ln.Accept()
		go manejadorRegistro(con) //concurrente
	}

}

func manejadorRegistro(con net.Conn) {
	defer con.Close()
	//leer
	bufferIn := bufio.NewReader(con)
	ip, _ := bufferIn.ReadString('\n')
	ip = strings.TrimSpace(ip) ///localhost:puerto
	//responder al solicitante con la bitácora que tiene este nodo

	bytes, _ := json.Marshal(append(bitacoraAddr, localhostNot))
	fmt.Fprintln(con, string(bytes)) //envia la bitácora

	ip2, _ := bufferIn.ReadString('\n')
	ip2 = strings.TrimSpace(ip2) ///localhost:puerto

	fmt.Println("IP2:", ip2)

	bytes, _ = json.Marshal(append(bitacoraAddr2, localhostHp)) //para servicio HP
	fmt.Fprintln(con, string(bytes))                            //envia la bitácora
	////////////////////
	//comunicar al todos los nodos la llegada de uno nuevo
	comunicarTodos(ip, ip2)

	//actualizar la bitácora con el nuevo ip
	bitacoraAddr = append(bitacoraAddr, ip)
	bitacoraAddr2 = append(bitacoraAddr2, ip2)

	fmt.Println(bitacoraAddr)
	fmt.Println(bitacoraAddr2)
}

func comunicarTodos(ip, ip2 string) {
	//recorrer toda la bitácora para comunicar
	for _, addr := range bitacoraAddr {
		notificar(addr, ip, ip2)
	}
}

func notificar(addr, ip, ip2 string) {
	con, _ := net.Dial("tcp", addr)
	defer con.Close()
	fmt.Fprintln(con, ip)
	fmt.Fprintln(con, ip2)
}

func registrarSolicitud(remotehost string) {
	con, _ := net.Dial("tcp", remotehost)
	defer con.Close()
	fmt.Fprintln(con, localhostNot) //enviamos el puerto de notificacion

	//recuperar lo que responde el server
	bufferIn := bufio.NewReader(con)
	bitacoraServer, _ := bufferIn.ReadString('\n')

	var bitacoraTemp []string
	json.Unmarshal([]byte(bitacoraServer), &bitacoraTemp)

	//bitacoraAddr = append(bitacoraTemp, localhostNot) //agregamos al final de la bitácora su direccion
	bitacoraAddr = bitacoraTemp
	/////////////////

	fmt.Fprintln(con, localhostHp) //enviamos el puerto de notificacion

	//recuperar lo que responde el server
	bitacoraServer, _ = bufferIn.ReadString('\n')

	var bitacoraTemp2 []string
	json.Unmarshal([]byte(bitacoraServer), &bitacoraTemp2)

	bitacoraAddr2 = bitacoraTemp2

	fmt.Println(bitacoraAddr)
	fmt.Println(bitacoraAddr2)

}

func recibeNotificarServer() {
	ln, _ := net.Listen("tcp", localhostNot)
	defer ln.Close()
	for {
		con, _ := ln.Accept()
		go manejadorRecibeNotificar(con)
	}
}

func manejadorRecibeNotificar(con net.Conn) {
	defer con.Close()
	bufferIn := bufio.NewReader(con)
	ip, _ := bufferIn.ReadString('\n')
	ip = strings.TrimSpace(ip)
	bitacoraAddr = append(bitacoraAddr, ip)

	ip2, _ := bufferIn.ReadString('\n')
	ip2 = strings.TrimSpace(ip2)
	bitacoraAddr2 = append(bitacoraAddr2, ip2)

	fmt.Println(bitacoraAddr)
	fmt.Println(bitacoraAddr2)
}

////////////////////////////

func servicioHP() {
	ln, _ := net.Listen("tcp", localhostHp)
	defer ln.Close()
	for {
		con, _ := ln.Accept()
		go manejadorHP(con)
	}
}

func manejadorHP(con net.Conn) {
	defer con.Close()
	bufferIn := bufio.NewReader(con)
	strNum, _ := bufferIn.ReadString('\n')
	strNum = strings.TrimSpace(strNum)
	num, _ := strconv.Atoi(strNum)
	//lógica
	fmt.Println("Número recibido = ", num)

	if num == 0 {
		fmt.Println("Proceso finalizado!!!")
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
