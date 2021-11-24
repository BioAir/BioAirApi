package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
)

//funcion cors
func enableCors(res http.ResponseWriter) {
	(res).Header().Set("Access-Control-Allow-Origin", "*")
	
}

//estructura

var data []Aire

type Aire struct {
	Ruido       float64 `json:"ruido"`
	Uv          float64 `json:"uv"`
	Humedad     float64 `json:"humedad"`
	Presion     float64 `json:"presion"`
	Temperatura float64 `json:"temperatura"`
}

//read
func readCSVFromUrl(url string) ([][]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	reader := csv.NewReader(resp.Body)
	reader.Comma = ','
	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return data, nil
}
func loadDataAire() {

	url := "https://raw.githubusercontent.com/gonzalouria/monitoreo/main/Monitoreo_octubre.csv"
	datos, err := readCSVFromUrl(url)
	if err != nil {
		panic(err)
	}

	var aire Aire

	for idx, rec := range datos {

		if idx == 0 {
			continue
		}
		aire.Ruido, _ = strconv.ParseFloat(rec[0], 32)
		aire.Uv, _ = strconv.ParseFloat(rec[1], 32)
		aire.Humedad, _ = strconv.ParseFloat(rec[2], 32)
		aire.Presion, _ = strconv.ParseFloat(rec[3], 32)
		aire.Temperatura, _ = strconv.ParseFloat(rec[4], 32)

		data = append(data, aire)
	}
	log.Println(data)
}

func getDataAire(res http.ResponseWriter, req *http.Request) {
	
	enableCors(res)
	res.Header().Set("Content-Type", "application/json; charset=utf-8")
	jsonBytes, _ := json.MarshalIndent(data, "", " ")
	io.WriteString(res, string(jsonBytes))
}

//read

func handleContextos() {
	route := http.NewServeMux()
	//endpoints del API
	route.HandleFunc("/home", mostrarHome)
	route.HandleFunc("/calcular_temperatura", calcular_temp)
	route.HandleFunc("/aire_data", getDataAire)

	log.Fatal(http.ListenAndServe(":9000", route))
}

func calcular_temp(resp http.ResponseWriter, req *http.Request) {

	enviar_cluster(332)

	//tipo de contenido
	resp.Header().Set("Content-Type", "text/html")
	io.WriteString(resp, `
		<html>
		<head></head>
		<body><h4>Calcular promedio de Temperatura</h4></body>
		</html>
	`)

}

func mostrarHome(resp http.ResponseWriter, req *http.Request) {
	log.Println("Ingresando al método mostrarHome")

	//tipo de contenido
	resp.Header().Set("Content-Type", "text/html")
	io.WriteString(resp, `
		<html>
		<head></head>
		<body><h2>Mi API de Aire</h2></body>
		</html>
	`)

}

func enviar_cluster(num int) {

	//bufferIn := bufio.NewReader(os.Stdin)
	//fmt.Print("Ingrese el puerto remoto: ")
	//puerto, _ := bufferIn.ReadString('\n')
	//puerto = strings.TrimSpace(puerto)
	remotehost := fmt.Sprintf("localhost:%s", "9003")
	con, _ := net.Dial("tcp", remotehost)
	defer con.Close()

	fmt.Fprintln(con, num)
	//fmt.Fprintln(con, 35)

}

func main() {

	loadDataAire()
	handleContextos()
}
