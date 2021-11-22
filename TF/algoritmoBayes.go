package bayesian

import (
	"encoding/gob"
	"errors"
	"io"
	"math"
	"os"
	"path/filepath"
	"sync/atomic"
)

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
	dec := gob.DecodificadorNuevo(r)
	w := new(serializableClasificador)
	err = dec.Decode(w)

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

func (c *Clasificador) aprendido() int {
	return c.aprendido
}

func (c *Clasificador) visto() int {
	return int(atomic.LoadInt32(&c.visto))
}

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

	dec := gob.DecodificadorNuevo(file)
	w := new(classData)
	err = dec.Decode(w)

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
