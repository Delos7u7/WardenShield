package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	// Cambiar la URL para conectar con la IP del servidor
	url := "ws://192.168.137.40:8080/ws"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("Error al conectar:", err)
	}
	defer conn.Close()

	// Generador de números aleatorios
	rand.Seed(time.Now().UnixNano())

	// Bucle para enviar datos al servidor cada 5 segundos
	for {
		// Generar temperatura aleatoria entre 9.0 y 13.0 grados
		temperatura := 9.0 + rand.Float64()*4.0

		// Generar distancia aleatoria entre 48.0 y 50.0 cm
		distancia := 48.0 + rand.Float64()*2.0

		// Generar hora actual
		hora := time.Now().Format("15:04:05")

		// Crear los datos a enviar
		data := map[string]interface{}{
			"hora":        hora,
			"temperatura": temperatura,
			"distancia":   distancia,
		}

		// Enviar los datos al servidor
		err := conn.WriteJSON(data)
		if err != nil {
			log.Println("Error al enviar datos:", err)
			return
		}

		fmt.Println(data)

		// Esperar la respuesta del servidor
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error al recibir respuesta:", err)
			return
		}
		log.Printf("Respuesta recibida: %s\n", message)

		// Espera antes de enviar el siguiente mensaje
		time.Sleep(5 * time.Second)
	}
}
