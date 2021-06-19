package services

import (
	kmeansalgoritmo "API/kmeans/kmeansAlgoritmo"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Variables struct {
	Edad            int `json:"edad"`
	Nivel_educativo int `json:"nivel_educativo"`
	Ocupacion       int `json:"ocupacion"`
	Entidad         int `json:"entidad"`
	Tamanio_entidad int `json:"tamanio_entidad"`
}
type Predicted struct {
	Cluster string `json:"Cluster"`
}

func HomeRoute(writer http.ResponseWriter, req *http.Request) {
	fmt.Fprint(writer, "bienvenidos al himalaya")
}

func PostInputs(writer http.ResponseWriter, req *http.Request) {
	var inputs Variables
	body, err := ioutil.ReadAll(req.Body)

	if err != nil {
		fmt.Fprint(writer, "Variables incorrectas")
	}

	json.Unmarshal(body, &inputs)
	fmt.Println(inputs)

	kmeansobject := entrenarAlgoritmo()
	predicted_cluster := predecir(inputs, kmeansobject)

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(predicted_cluster)

}

func entrenarAlgoritmo() *kmeansalgoritmo.KmeansObject {
	data, _ := kmeansalgoritmo.ImportDataFile("https://raw.githubusercontent.com/GeraldineNisbeth/KMeansParallel/main/dataset.csv", 2, 6)
	fmt.Println(len(data))
	kmeansobject := kmeansalgoritmo.Kmeans(300, 4)

	col_clusters, err := kmeansobject.Entrenamiento(data)

	if err == nil {
		fmt.Println(col_clusters)
		fmt.Println(kmeansobject.TamanioClusters())
	}
	return kmeansobject
}

func predecir(inputs Variables, kmeansobject *kmeansalgoritmo.KmeansObject) Predicted {
	array := []float64{float64(inputs.Edad), float64(inputs.Nivel_educativo), float64(inputs.Ocupacion), float64(inputs.Entidad), float64(inputs.Tamanio_entidad)}
	var predicted_cluster Predicted
	predicted_cluster.Cluster = kmeansobject.Predecir(array)
	fmt.Println(predicted_cluster)
	return predicted_cluster
}
