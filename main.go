package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
)

type Data struct {
	Hora        string  `json:"hora"`
	Distancia   float64 `json:"distancia"`
	Temperatura float64 `json:"temperatura"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Cambiar la extensión del archivo a .csv
const csvPath = `C:\Users\Jose Ramon\Desktop\Escuela\SEMESTRE7\InteligenciaArtificial\InteligenciaAritificalGrafica\bin\Debug\datos.csv`

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error al establecer WebSocket:", err)
		http.Error(w, "No se pudo establecer conexión WebSocket", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	for {
		var data Data
		err := conn.ReadJSON(&data)
		if err != nil {
			log.Println("Error al leer datos de WebSocket:", err)
			break
		}

		// Si no se proporciona hora, se asigna la hora actual
		if data.Hora == "" {
			data.Hora = time.Now().Format("15:04:05")
		}

		// Guardar los datos en la base de datos y en el CSV
		if err := saveToDatabase(data); err != nil {
			log.Println("Error al guardar en la base de datos:", err)
			continue
		}
		if err := saveToCSV(data); err != nil {
			log.Println("Error al guardar en CSV:", err)
			continue
		}

		// Responder al cliente WebSocket
		if err := conn.WriteMessage(websocket.TextMessage, []byte("Datos recibidos correctamente")); err != nil {
			log.Println("Error al enviar respuesta WebSocket:", err)
			break
		}
	}
}

func saveToDatabase(data Data) error {
	dsn := "root:123456@tcp(127.0.0.1:3306)/integrador7"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO mediciones (hora, distancia, temperatura) VALUES (?, ?, ?)",
		data.Hora, data.Distancia, data.Temperatura)
	return err
}

func calculateProbability(hora string, distancia float64) float64 {
	decimalHour, err := horaADecimal(hora)
	if err != nil {
		log.Println("Error al convertir hora a decimal:", err)
		return 0.0
	}

	var baseProbability float64

	switch {
	case decimalHour >= 8 && decimalHour <= 10:
		baseProbability = 0.8
	case decimalHour >= 16 && decimalHour <= 18:
		baseProbability = 0.7
	case decimalHour >= 12 && decimalHour <= 14:
		baseProbability = 0.6
	default:
		baseProbability = 0.3
	}

	distanceAdjustment := 1.0 - (distancia / 10.0)
	if distanceAdjustment < 0 {
		distanceAdjustment = 0
	}

	probability := baseProbability * distanceAdjustment
	if probability > 1.0 {
		probability = 1.0
	}

	return probability
}

func horaADecimal(hora string) (float64, error) {
	t, err := time.Parse("15:04:05", hora)
	if err != nil {
		return 0.0, err
	}

	// Convertir la hora, minuto y segundo a un valor decimal
	h := float64(t.Hour())
	m := float64(t.Minute()) / 60.0
	s := float64(t.Second()) / 3600.0

	return h + m + s, nil
}

func saveToCSV(data Data) error {
	// Asegurarse de que el directorio existe
	dir := filepath.Dir(csvPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creando directorio: %v", err)
	}

	// Verificar si el archivo existe
	var file *os.File
	if _, err := os.Stat(csvPath); os.IsNotExist(err) {
		// Crear el archivo si no existe
		file, err = os.Create(csvPath)
		if err != nil {
			return fmt.Errorf("error creando archivo CSV: %v", err)
		}
	} else {
		// Abrir el archivo en modo append si ya existe
		file, err = os.OpenFile(csvPath, os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return fmt.Errorf("error abriendo archivo CSV: %v", err)
		}
	}
	defer file.Close()

	// Crear el writer CSV
	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Convertir la hora a formato decimal
	decimalHora, err := horaADecimal(data.Hora)
	if err != nil {
		return fmt.Errorf("error convirtiendo hora a decimal: %v", err)
	}

	// Calcular la probabilidad
	probability := calculateProbability(data.Hora, data.Distancia)

	// Preparar el registro CSV con la hora convertida a decimal
	record := []string{
		strconv.FormatFloat(decimalHora, 'f', 6, 64), // Escribir la hora en formato decimal
		strconv.FormatFloat(data.Distancia, 'f', 2, 64),
		strconv.FormatFloat(probability, 'f', 2, 64),
	}

	// Escribir el registro
	if err := writer.Write(record); err != nil {
		return fmt.Errorf("error escribiendo en CSV: %v", err)
	}

	return nil
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	fmt.Println("Servidor WebSocket escuchando en el puerto 8080...")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}
