package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	//"https://github.com/BioAir/BioAirApi/blob/main/Monitoreo_octubre.csv"
)

type Clima struct {
	ID          int     `json:"id,omitempty"`
	Ruido       float64 `json:Ruido(dB)`
	UV          float64 `json:UV`
	Humedad     float64 `json:Humedad(%)`
	Presion     float64 `json:Presion(hPa)`
	Temperatura float64 `json:Temperatura(C)`
}

var Climas []Clima
var nextID = 1

func GetClimas() []Clima {
	return Climas
}

func AddClima(Clima Clima) int {
	Clima.ID = nextID
	nextID++
	Climas = append(Climas, Clima)
	return Clima.ID

}
func HandleRequest(w http.ResponseWriter, r *http.Request) {
	log.Println("Request received:", r.Method)

	switch r.Method {
	case http.MethodGet:
		list(w, r)
		break
	case http.MethodPost:
		add(w, r)
		break
	default:
		w.WriteHeader(405)
		w.Write([]byte("Method not allowed"))
		break
	}
}
func list(w http.ResponseWriter, r *http.Request) {
	Climas := GetClimas()
	json, _ := json.Marshal(Climas)

	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(200)
	w.Write(json)

	log.Println("Response returned:", 200)

}

func add(w http.ResponseWriter, r *http.Request) {
	payload, _ := ioutil.ReadAll(r.Body)

	var Clima Clima
	err := json.Unmarshal(payload, &Clima)

	if err != nil || Clima.Ruido == 0 || Clima.UV == 0 {
		w.WriteHeader(400)
		w.Write([]byte("Bad Request"))
		return
	}

	Clima.ID = AddClima(Clima)

	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(201)

	json, _ := json.Marshal(Clima)
	w.Write(json)
	log.Println("Response returned:", 201)

}
func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
func readCSVFromUrl(url string) ([][]string, error) {
	resp, err := http.Get(url)
	check(err)
	defer resp.Body.Close()
	r := csv.NewReader(resp.Body)
	r.Comma = ','
	if _, err := r.Read(); err != nil {
		return [][]string{}, err
	}
	data, err := r.ReadAll()
	if err != nil {
		return [][]string{}, err
	}
	return data, nil
}
func convertThread(tuples [][]string, ch chan cl.Recurso) {
	var latitud, longitud *float64 = new(float64), new(float64)
	for _, value := range tuples {
		codigo, err := strconv.Atoi(strings.Replace(strings.TrimSpace(value[3]), ",", "", -1))
		check(err)
		if value[9] != "" {
			aux, err := strconv.ParseFloat(strings.TrimSpace(value[9]), 64)
			check(err)
			latitud = &aux
		} else {
			latitud = nil
		}
		if value[10] != "" {
			aux, err := strconv.ParseFloat(strings.TrimSpace(value[10]), 64)
			check(err)
			longitud = &aux
		} else {
			longitud = nil
		}
		recursosingle := cl.Recurso{
			REGIÓN:             value[0],
			PROVINCIA:          value[1],
			DISTRITO:           value[2],
			Codigo_del_Recurso: codigo,
			Nombre_del_Recurso: value[4],
			CATEGORIA:          value[5],
			Tipo_de_Categoria:  value[6],
			Sub_tipo_Categoria: value[7],
			URL:                value[8],
			LATITUD:            latitud,
			LONGITUD:           longitud,
		}
		//put to channel
		ch <- recursosingle
		//recursos = append(recursos, recursosingle)
	}
	close(ch)
}
func Min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func LoadRecursos() []cl.Recurso {
	data, err := readCSVFromUrl("https://raw.githubusercontent.com/Alextron0102/TA2-Go-API/main/files/Inventario_recursos_turisticos.csv")
	check(err)
	channels := make([]chan cl.Recurso, cl.NUM_CPU+1)
	limit := len(data) / cl.NUM_CPU
	fmt.Print("lineas en total: ")
	fmt.Println(len(data))
	iteratoraux := 0
	for i := 0; i < len(data); i += limit {
		chunk := data[i:Min(i+limit, len(data))]
		channels[iteratoraux] = make(chan cl.Recurso)
		go convertThread(chunk, channels[iteratoraux])
		iteratoraux++
	}
	var recursos []cl.Recurso
	for _, channel := range channels {
		for recurso := range channel {
			//PrintRecurso(recurso)
			recursos = append(recursos, recurso)
		}
	}
	return recursos
}

func main() {
	AddClima(Clima{
		Ruido:       100.00,
		UV:          100.00,
		Humedad:     100.00,
		Presion:     100.00,
		Temperatura: 100.00,
	})

	AddClima(Clima{
		Ruido:       200.00,
		UV:          200.00,
		Humedad:     200.00,
		Presion:     200.00,
		Temperatura: 200.00,
	})

	http.HandleFunc("/", HandleRequest)
	err := http.ListenAndServe(":9000", nil)

	if err != nil {
		log.Println(err)
		return
	}

}
