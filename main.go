package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
	"github.com/xuri/excelize/v2"
)

type Data struct {
	Temperatura float64 `json:"temperatura"`
	Distancia   float64 `json:"distancia"`
}

// Configuración de WebSocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Acepta conexiones desde cualquier origen (ajustar para producción)
	},
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Actualizar conexión HTTP a WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Error al establecer WebSocket:", err)
		http.Error(w, "No se pudo establecer conexión WebSocket", http.StatusBadRequest)
		return
	}
	defer conn.Close()

	for {
		var data Data
		// Lee el mensaje de WebSocket
		err := conn.ReadJSON(&data)
		if err != nil {
			log.Println("Error al leer datos de WebSocket:", err)
			break
		}

		// Guarda los datos en la base de datos y en el archivo Excel
		if err := saveToDatabase(data); err != nil {
			log.Println("Error al guardar en la base de datos:", err)
			continue
		}
		if err := saveToExcel(data); err != nil {
			log.Println("Error al guardar en Excel:", err)
			continue
		}

		// Responde con un mensaje de éxito
		if err := conn.WriteMessage(websocket.TextMessage, []byte("Datos recibidos correctamente")); err != nil {
			log.Println("Error al enviar respuesta WebSocket:", err)
			break
		}
	}
}

func saveToDatabase(data Data) error {
	dsn := "root:itsoeh23@tcp(127.0.0.1:3306)/integrador7"
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec("INSERT INTO mediciones (temperatura, distancia) VALUES (?, ?)", data.Temperatura, data.Distancia)
	return err
}

func saveToExcel(data Data) error {
	file, err := excelize.OpenFile("datos.xlsx")
	if err != nil {
		file = excelize.NewFile()
		file.SetCellValue("Sheet1", "A1", "Temperatura")
		file.SetCellValue("Sheet1", "B1", "Distancia")
	}

	row := 2
	for {
		cell, _ := file.GetCellValue("Sheet1", fmt.Sprintf("A%d", row))
		if cell == "" {
			break
		}
		row++
	}

	file.SetCellValue("Sheet1", fmt.Sprintf("A%d", row), data.Temperatura)
	file.SetCellValue("Sheet1", fmt.Sprintf("B%d", row), data.Distancia)
	return file.SaveAs("datos.xlsx")
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)
	fmt.Println("Servidor WebSocket escuchando en el puerto 8080...")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}
