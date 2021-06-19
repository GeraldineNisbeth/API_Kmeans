package kmeansalgoritmo

import (
	"encoding/csv"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"gonum.org/v1/gonum/floats"
)

type KmeansObject struct {
	numIteraciones, numClusters int
	distancia                   DistanceFunc

	temp1, temp2, contador, umbral int          //variables para manejar la convergencia de la clusterización
	semaforo                       sync.RWMutex // semáforo que nos ayudará con la sincronización
	a, b                           []int        // a contiene el cluster correspondiente de cada dato y b contiene el # de datos de cada cluster
	centroides, centroidesTemp     [][]float64  //arrays que contienen los centroides de cada cluster
	data                           [][]float64
}

var (
	DistanciaEuclidiana = func(puntoA, puntoB []float64) float64 {
		var sum, temp float64

		for i, _ := range puntoA {
			temp = puntoA[i] - puntoB[i]
			sum += math.Pow(temp, 2)
		}
		return math.Sqrt(sum)
	}
)

type DistanceFunc func([]float64, []float64) float64

func Kmeans(numIteraciones, numClusters int) *KmeansObject {

	return &KmeansObject{
		numIteraciones: numIteraciones,
		numClusters:    numClusters,
		distancia:      DistanciaEuclidiana,
	}
}

func (kobj *KmeansObject) Entrenamiento(data [][]float64) ([]int, error) {

	kobj.semaforo.Lock()

	kobj.data = data
	kobj.a = make([]int, len(data))
	kobj.b = make([]int, kobj.numClusters)
	kobj.umbral = 2
	kobj.temp1 = 0
	kobj.temp2 = 0

	kobj.cargarCentroides()

	for i := 0; i < kobj.numIteraciones && kobj.contador != kobj.umbral; i++ {
		kobj.entrenar()

		if kobj.temp1 == kobj.temp2 { // convergió
			kobj.contador++
		}
		kobj.temp2 = kobj.temp1
	}

	kobj.centroidesTemp = nil

	kobj.semaforo.Unlock()
	return kobj.a, nil
}

func (kobj *KmeansObject) cargarCentroides() {
	kobj.centroides = make([][]float64, kobj.numClusters)     //array que contiene los centroides de cada dato
	kobj.centroidesTemp = make([][]float64, kobj.numClusters) //array que contiene los centroides de cada dato

	rand.Seed(time.Now().UTC().Unix())

	var (
		k          int
		s, t, l, f float64
		dataTemp   []float64 = make([]float64, len(kobj.data))
	)

	kobj.centroides[0] = kobj.data[rand.Intn(len(kobj.data)-1)] //inicializamos el centroide con un dato aleatorio

	for i := 1; i < kobj.numClusters; i++ {
		s = 0
		t = 0
		for j := 0; j < len(kobj.data); j++ {
			l = kobj.distancia(kobj.centroides[0], kobj.data[j]) //calculamos la distancia entre el centroide y los datos
			for h := 1; h < i; h++ {
				f = kobj.distancia(kobj.centroides[h], kobj.data[j]) //calculamos la distancia con los otros centroides
				if f < l {
					l = f //distancia minima
				}
			}
			dataTemp[j] = math.Pow(l, 2)
			s += dataTemp[j]
		}
		t = rand.Float64() * s
		k = 0
		for s = dataTemp[0]; s < t; s += dataTemp[k] {
			k++
		}
		kobj.centroides[i] = kobj.data[k]
	}

	for i := 0; i < kobj.numClusters; i++ {
		kobj.centroidesTemp[i] = make([]float64, len(kobj.centroides[0]))
	}

}

func (kobj *KmeansObject) entrenar() {

	var (
		a, b, c    int = len(kobj.centroides[0]), 0, 0
		dmin, dist float64
	)

	for i := 0; i < kobj.numClusters; i++ {
		kobj.b[i] = 0
	}
	for i := 0; i < len(kobj.data); i++ {
		dmin = kobj.distancia(kobj.data[i], kobj.centroides[0])
		c = 0

		for j := 1; j < kobj.numClusters; j++ {

			dist = kobj.distancia(kobj.data[i], kobj.centroides[j])

			if dist < dmin {
				dmin = dist
				c = j
			}
		}

		b = c + 1
		if kobj.a[i] != b {
			kobj.temp1++
		}
		kobj.a[i] = b
		kobj.b[c]++
		floats.Add(kobj.centroidesTemp[c], kobj.data[i])
	}
	for i := 0; i < kobj.numClusters; i++ {
		floats.Scale(1/float64(kobj.b[i]), kobj.centroidesTemp[i]) // normalizar el valor de los centroides
		for j := 0; j < a; j++ {
			kobj.centroides[i][j] = kobj.centroidesTemp[i][j] //liberamos centroidesTemp
			kobj.centroidesTemp[i][j] = 0
		}
	}
}

func (kobj *KmeansObject) TamanioClusters() []int {
	kobj.semaforo.RLock()
	defer kobj.semaforo.RUnlock()
	return kobj.b
}

func (kobj *KmeansObject) Predecir(input []float64) string {

	var (
		cluster   int = 1
		dist      float64
		centroide = kobj.distancia(input, kobj.centroides[0])
	)

	for i := 1; i < kobj.numClusters; i++ {
		dist = kobj.distancia(input, kobj.centroides[i])

		if dist < centroide {
			centroide = dist
			cluster = i + 1
		}
	}

	map_sueldos := make(map[int]string)
	map_sueldos[1] = "menor a 950"
	map_sueldos[2] = "950 a 1500"
	map_sueldos[3] = "1500 a 2500"
	map_sueldos[4] = "2500 a 3500"

	return map_sueldos[cluster]
}

func ImportDataFile(url string, inicio, fin int) ([][]float64, error) {
	var (
		data = make([][]float64, 0)
		s    = fin - inicio + 1 //rango de columnas
		goal []float64
	)

	respuesta, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer respuesta.Body.Close()

	reader := csv.NewReader(respuesta.Body)
	dataFilas, err := reader.ReadAll()

	for j, fila := range dataFilas {
		if j == 0 { //eliminar cabecera
			continue
		}

		if err != nil {
			return [][]float64{}, err
		}

		goal = make([]float64, 0, s)

		for i := inicio; i <= fin; i++ {
			f, err := strconv.ParseFloat(fila[i], 64)
			if err == nil {
				goal = append(goal, f)
			} else {
				continue
			}
		}

		data = append(data, goal)
	}
	return data, nil
}

/*
func main() {

	data, _ := ImportDataFile("https://raw.githubusercontent.com/GeraldineNisbeth/KMeansParallel/main/dataset.csv", 2, 7)

	fmt.Println(len(data))

	KmeansObject := Kmeans(300, 3)
	fmt.Println(KmeansObject.numClusters)

	col_clusters, err := KmeansObject.Entrenamiento(data)

	if err == nil {
		fmt.Println(col_clusters)
		fmt.Println(KmeansObject.TamanioClusters())
	}

	input := []float64{23.0, 8.0, 3.0, 4.0, 2.0, 1500.0}
	predicted_cluster := KmeansObject.Predecir(input)
	fmt.Println(predicted_cluster)
}*/
