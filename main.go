package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	_ "github.com/go-sql-driver/mysql"
)

type Data struct {
	Temperatura float64 `json:"temperatura"`
	Distancia   float64 `json:"distancia"`
}

var (
	csvPath         = `C:\Users\Intel\Desktop\SEMESTRE7\Inteligencia Artificial\U1\InteligenciaArtificalGrafica\InteligenciaArtificalGrafica\bin\Debug\datos.csv`
	activationCount = make(map[int]int)
)

// Manejador de la ruta de recepción de datos
func handleDataReceive(w http.ResponseWriter, r *http.Request) {
	// Verificar que sea una solicitud POST
	if r.Method != http.MethodPost {
		http.Error(w, "Método no permitido", http.StatusMethodNotAllowed)
		return
	}

	// Decodificar los datos JSON
	var data Data
	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil {
		http.Error(w, "Error al decodificar datos", http.StatusBadRequest)
		log.Println("Error al decodificar datos:", err)
		return
	}

	// Guardar los datos en la base de datos
	insertedID, err := saveToDatabase(data)
	if err != nil {
		http.Error(w, "Error al guardar en la base de datos", http.StatusInternalServerError)
		log.Println("Error al guardar en la base de datos:", err)
		return
	}

	// Obtener la hora asociada a la medición
	hora, err := getHoraFromDatabase(insertedID)
	if err != nil {
		http.Error(w, "Error al obtener la hora", http.StatusInternalServerError)
		log.Println("Error al obtener la hora:", err)
		return
	}

	// Actualizar contador de activaciones por hora
	horaInt := extractHour(hora)
	activationCount[horaInt]++

	// Guardar las activaciones en el archivo CSV
	if err := saveActivationsToCSV(); err != nil {
		http.Error(w, "Error al guardar activaciones", http.StatusInternalServerError)
		log.Println("Error al guardar activaciones:", err)
		return
	}

	// Responder al cliente
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Datos recibidos y activaciones actualizadas",
	})
}

// Guardar datos en la base de datos
func saveToDatabase(data Data) (int64, error) {
	dsn := "root:itsoeh23@tcp(0.0.0.0:3306)/integrador7"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return 0, err
	}
	defer db.Close()

	result, err := db.Exec("INSERT INTO mediciones (hora, distancia, temperatura) VALUES (NOW(), ?, ?)",
		data.Distancia, data.Temperatura)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

// Obtener la hora de la base de datos
func getHoraFromDatabase(id int64) (string, error) {
	dsn := "root:itsoeh23@tcp(0.0.0.0:3306)/integrador7"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return "", err
	}
	defer db.Close()

	var hora string
	err = db.QueryRow("SELECT hora FROM mediciones WHERE id = ?", id).Scan(&hora)
	if err != nil {
		return "", err
	}

	return hora, nil
}

// Extraer la hora en formato entero (1-24)
func extractHour(hora string) int {
	components := strings.Split(hora, ":")
	hour, _ := strconv.Atoi(components[0]) // Tomar solo la parte de la hora
	return hour
}

// Guardar activaciones por hora en un archivo CSV
func saveActivationsToCSV() error {
	// Asegurarse de que el directorio existe
	dir := filepath.Dir(csvPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creando directorio: %v", err)
	}

	// Abre el archivo CSV en modo de apéndice (agregar nuevas líneas)
	file, err := os.OpenFile(csvPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("error abriendo archivo CSV: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Si el archivo está vacío, escribir el encabezado
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("error obteniendo información del archivo: %v", err)
	}
	if fileInfo.Size() == 0 {
		if err := writer.Write([]string{"Hora", "Activaciones", "Probabilidad (%)"}); err != nil {
			return fmt.Errorf("error escribiendo encabezado CSV: %v", err)
		}
	}

	// Escribir datos de activaciones
	for hora, activaciones := range activationCount {
		probability := float64(activaciones) // Ajuste de probabilidad si es necesario
		record := []string{
			strconv.Itoa(hora),
			strconv.Itoa(activaciones),
			fmt.Sprintf("%.2f", probability),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("error escribiendo fila CSV: %v", err)
		}
	}

	return nil
}

func main() {
	// Configurar el manejador para la ruta de recepción de datos
	http.HandleFunc("/datos", handleDataReceive)

	// Configurar un manejador CORS básico
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Permitir solicitudes desde cualquier origen
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Manejar solicitudes OPTIONS para CORS
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
	})

	fmt.Println("Servidor escuchando en 0.0.0.0:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
