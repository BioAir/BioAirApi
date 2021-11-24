package main

import (
	"bufio"
	"encoding/csv"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
)

var localhostReg string
var localhostNot string
var localhostHp string
var remotehost string
var bitacoraAddr []string
var bitacoraAddr2 []string
var lista = [732]Aire{}

func readCSVFromUrl(url string) ([][]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	reader := csv.NewReader(resp.Body)
	reader.Comma = ','
	reader.LazyQuotes = true
	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return data, nil

}

type Aire struct {
	Ruido       float64 `json:"ruido"`
	UV          float64 `json:"uv"`
	Humedad     float64 `json:"humedad"`
	Presion     float64 `json:"presion"`
	Temperatura float64 `json:"temperatura"`
}

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

// imprime
func aire(rango1 int, rango2 int) {

	url := "https://raw.githubusercontent.com/gonzalouria/monitoreo/main/Monitoreo_octubre.csv"
	data, err := readCSVFromUrl(url)
	if err != nil {
		panic(err)
	}

	for idx, row := range data {

		var auxaire Aire

		if idx == 0 {

			continue
		}
		auxaire.Ruido, _ = strconv.ParseFloat(row[0], 32)
		auxaire.UV, _ = strconv.ParseFloat(row[1], 32)
		auxaire.Humedad, _ = strconv.ParseFloat(row[2], 32)
		auxaire.Presion, _ = strconv.ParseFloat(row[3], 32)
		auxaire.Temperatura, _ = strconv.ParseFloat(row[4], 32)

		lista[idx] = auxaire

		// skip header

	}
	var sum float64 = 0
	var cont float64 = 0
	for i := rango1; i < rango2; i++ {

		fmt.Println(lista[i])
		cont++
		sum = sum + lista[i].Temperatura

	}
	fmt.Println(sum/cont, "temperatura final")

}

func manejadorHP(con net.Conn) {
	defer con.Close()
	bufferIn := bufio.NewReader(con)
	strNum, _ := bufferIn.ReadString('\n')
	strNum = strings.TrimSpace(strNum)
	num, _ := strconv.Atoi(strNum)
	//lógica

	fmt.Println("Número recibido = ", num)

	rango1 := num / 2
	rango2 := num
	aire(rango1, rango2)

	if num == 0 {
		fmt.Println("Proceso finalizado!!!")
	} else {
		enviarProximo(rango1)
	}

}

func enviarProximo(num int) {
	indice := rand.Intn(len(bitacoraAddr2))
	fmt.Printf("Enviando %d hacia %s", num, bitacoraAddr2[indice])

	con, _ := net.Dial("tcp", bitacoraAddr2[indice])
	defer con.Close()
	aire(1, num)
	fmt.Fprintln(con, num-1)

}

// algoritmo
const problabilidadFallo = 0.00000000001

var errorU = errors.New("error detectado")

type Class string

type Clasificador struct {
	Classes         []Class
	aprendido       int
	visto           int32
	datas           map[Class]*classData
	tfIdf           bool
	DidConvertTfIdf bool
}

type serializableClasificador struct {
	Classes         []Class
	aprendido       int
	visto           int
	Datas           map[Class]*classData
	TfIdf           bool
	DidConvertTfIdf bool
}

type classData struct {
	Freqs   map[string]float64
	FreqTfs map[string][]float64
	Total   int
}

func nuevaDataClass() *classData {
	return &classData{
		Freqs:   make(map[string]float64),
		FreqTfs: make(map[string][]float64),
	}
}

func (d *classData) problPalabra(palabra string) float64 {
	valor, ok := d.Freqs[palabra]
	if !ok {
		return problabilidadFallo
	}
	return float64(valor) / float64(d.Total)
}

func (d *classData) problPalabras(palabras []string) (probl float64) {
	probl = 1
	for _, palabra := range palabras {
		probl *= d.problPalabra(palabra)
	}
	return
}

func nuevoClasificadorTfIdf(classes ...Class) (c *Clasificador) {
	n := len(classes)

	if n < 2 {
		panic("al menos dos clases pls")
	}

	check := make(map[Class]bool, n)
	for _, class := range classes {
		check[class] = true
	}
	if len(check) != n {
		panic("la clase debe ser única")
	}

	c = &Clasificador{
		Classes: classes,
		datas:   make(map[Class]*classData, n),
		tfIdf:   true,
	}
	for _, class := range classes {
		c.datas[class] = nuevaDataClass()
	}
	return
}

func nuevoClasificador(classes ...Class) (c *Clasificador) {
	n := len(classes)

	if n < 2 {
		panic("al menos dos clases")
	}

	check := make(map[Class]bool, n)
	for _, class := range classes {
		check[class] = true
	}
	if len(check) != n {
		panic("clase debe ser única")
	}

	c = &Clasificador{
		Classes:         classes,
		datas:           make(map[Class]*classData, n),
		tfIdf:           false,
		DidConvertTfIdf: false,
	}
	for _, class := range classes {
		c.datas[class] = nuevaDataClass()
	}
	return
}

func nuevoClasificadorArchivo(name string) (c *Clasificador, err error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return nuevoClasificadorLector(file)
}

func nuevoClasificadorLector(r io.Reader) (c *Clasificador, err error) {
	//dec := gob.DecodificadorNuevo(r)
	w := new(serializableClasificador)
	//err = dec.Decode(w)

	return &Clasificador{w.Classes, w.aprendido, int32(w.visto), w.Datas, w.TfIdf, w.DidConvertTfIdf}, err
}

func (c *Clasificador) obtenerPrioridades() (Prioridades []float64) {
	n := len(c.Classes)
	Prioridades = make([]float64, n, n)
	sum := 0
	for index, class := range c.Classes {
		total := c.datas[class].Total
		Prioridades[index] = float64(total)
		sum += total
	}
	if sum != 0 {
		for i := 0; i < n; i++ {
			Prioridades[i] /= float64(sum)
		}
	}
	return
}

//func (c *Clasificador) aprendido() int {
//return c.aprendido
///}

//func (c *Clasificador) visto() int {
//return int(atomic.LoadInt32(&c.visto))
//}

func (c *Clasificador) IsTfIdf() bool {
	return c.tfIdf
}

func (c *Clasificador) ContadorPalabra() (result []int) {
	result = make([]int, len(c.Classes))
	for inx, class := range c.Classes {
		data := c.datas[class]
		result[inx] = data.Total
	}
	return
}

func (c *Clasificador) Observe(palabra string, count int, which Class) {
	data := c.datas[which]
	data.Freqs[palabra] += float64(count)
	data.Total += count
}

func (c *Clasificador) Aprender(document []string, which Class) {

	if c.tfIdf {
		if c.DidConvertTfIdf {
			panic("no se puede connvertir más de una vez")
		}

		docTf := make(map[string]float64)
		for _, palabra := range document {
			docTf[palabra]++
		}

		docLen := float64(len(document))

		for wIndex, wCount := range docTf {
			docTf[wIndex] = wCount / docLen

			c.datas[which].FreqTfs[wIndex] = append(c.datas[which].FreqTfs[wIndex], docTf[wIndex])
		}

	}

	data := c.datas[which]
	for _, palabra := range document {
		data.Freqs[palabra]++
		data.Total++
	}
	c.aprendido++
}

func (c *Clasificador) ConvertTermsFreqToTfIdf() {

	if c.DidConvertTfIdf {
		panic("no se puede connvertir más de una vez.")
	}

	for className := range c.datas {

		for wIndex := range c.datas[className].FreqTfs {
			tfIdfAdder := float64(0)

			for tfSampleIndex := range c.datas[className].FreqTfs[wIndex] {

				tf := c.datas[className].FreqTfs[wIndex][tfSampleIndex]
				c.datas[className].FreqTfs[wIndex][tfSampleIndex] = math.Log1p(tf) * math.Log1p(float64(c.aprendido)/float64(c.datas[className].Total))
				tfIdfAdder += c.datas[className].FreqTfs[wIndex][tfSampleIndex]
			}

			c.datas[className].Freqs[wIndex] = tfIdfAdder
		}

	}

	c.DidConvertTfIdf = true

}

func (c *Clasificador) LogScores(document []string) (scores []float64, inx int, strict bool) {
	if c.tfIdf && !c.DidConvertTfIdf {
		panic("Using a TF-IDF Clasificador. Please call ConvertTermsFreqToTfIdf before calling LogScores.")
	}

	n := len(c.Classes)
	scores = make([]float64, n, n)
	Prioridades := c.obtenerPrioridades()

	for index, class := range c.Classes {
		data := c.datas[class]

		score := math.Log(Prioridades[index])
		for _, palabra := range document {
			score += math.Log(data.problPalabra(palabra))
		}
		scores[index] = score
	}
	inx, strict = Max(scores)
	atomic.AddInt32(&c.visto, 1)
	return scores, inx, strict
}

func (c *Clasificador) ProbPuntaje(doc []string) (scores []float64, inx int, strict bool) {
	if c.tfIdf && !c.DidConvertTfIdf {
		panic("convertir antes de llamar a la probabliidad")
	}
	n := len(c.Classes)
	scores = make([]float64, n, n)
	Prioridades := c.obtenerPrioridades()
	sum := float64(0)

	for index, class := range c.Classes {
		data := c.datas[class]

		score := Prioridades[index]
		for _, palabra := range doc {
			score *= data.problPalabra(palabra)
		}
		scores[index] = score
		sum += score
	}
	for i := 0; i < n; i++ {
		scores[i] /= sum
	}
	inx, strict = Max(scores)
	atomic.AddInt32(&c.visto, 1)
	return scores, inx, strict
}
func (c *Clasificador) probPuntajeSafe(doc []string) (scores []float64, inx int, strict bool, err error) {
	if c.tfIdf && !c.DidConvertTfIdf {
		panic("convierta antews de usar safe.")
	}

	n := len(c.Classes)
	scores = make([]float64, n, n)
	logScores := make([]float64, n, n)
	Prioridades := c.obtenerPrioridades()
	sum := float64(0)

	for index, class := range c.Classes {
		data := c.datas[class]

		score := Prioridades[index]
		logScore := math.Log(Prioridades[index])
		for _, palabra := range doc {
			p := data.problPalabra(palabra)
			score *= p
			logScore += math.Log(p)
		}
		scores[index] = score
		logScores[index] = logScore
		sum += score
	}
	for i := 0; i < n; i++ {
		scores[i] /= sum
	}
	inx, strict = Max(scores)
	logInx, logStrict := Max(logScores)

	if inx != logInx || strict != logStrict {
		err = errorU
	}
	atomic.AddInt32(&c.visto, 1)
	return scores, inx, strict, err
}

func (c *Clasificador) frecuenciaPalabra(palabras []string) (freqMatrix [][]float64) {
	n, l := len(c.Classes), len(palabras)
	freqMatrix = make([][]float64, n)
	for i := range freqMatrix {
		arr := make([]float64, l)
		data := c.datas[c.Classes[i]]
		for j := range arr {
			arr[j] = data.problPalabra(palabras[j])
		}
		freqMatrix[i] = arr
	}
	return
}

func (c *Clasificador) palabrasPorClase(class Class) (freqMap map[string]float64) {
	freqMap = make(map[string]float64)
	for palabra, cnt := range c.datas[class].Freqs {
		freqMap[palabra] = float64(cnt) / float64(c.datas[class].Total)
	}

	return freqMap
}

func (c *Clasificador) escribirArchivo(name string) (err error) {
	file, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	return c.EscribirA(file)
}

func (c *Clasificador) escribirClasesArchivo(rootPath string) (err error) {
	for name := range c.datas {
		c.WriteClassToFile(name, rootPath)
	}
	return
}

func (c *Clasificador) WriteClassToFile(name Class, rootPath string) (err error) {
	data := c.datas[name]
	fileName := filepath.Join(rootPath, string(name))
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer file.Close()

	enc := gob.NewEncoder(file)
	err = enc.Encode(data)
	return
}

func (c *Clasificador) EscribirA(w io.Writer) (err error) {
	enc := gob.NewEncoder(w)
	err = enc.Encode(&serializableClasificador{c.Classes, c.aprendido, int(c.visto), c.datas, c.tfIdf, c.DidConvertTfIdf})

	return
}

func (c *Clasificador) escribirClaseDeArchivo(class Class, location string) (err error) {
	fileName := filepath.Join(location, string(class))
	file, err := os.Open(fileName)

	if err != nil {
		return err
	}
	defer file.Close()

	//dec := gob.DecodificadorNuevo(file)
	w := new(classData)
	//err = dec.Decode(w)

	c.aprendido++
	c.datas[class] = w
	return
}

func Max(scores []float64) (inx int, strict bool) {
	inx = 0
	strict = true
	for i := 1; i < len(scores); i++ {
		if scores[inx] < scores[i] {
			inx = i
			strict = true
		} else if scores[inx] == scores[i] {
			strict = false
		}
	}
	return
}

//
