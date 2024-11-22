package main

import (
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	url := "ws://localhost:8080/ws"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("Error al conectar:", err)
	}
	defer conn.Close()

	// Ejemplo de datos de prueba
	data := map[string]interface{}{
		"temperatura": 32.5,
		"distancia":   0.5,
	}
	for {
		err := conn.WriteJSON(data)
		if err != nil {
			log.Println("Error al enviar datos:", err)
			return
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error al recibir respuesta:", err)
			return
		}
		log.Printf("Respuesta recibida: %s\n", message)

		time.Sleep(5 * time.Second) // Espera antes de enviar el siguiente mensaje
	}
}
