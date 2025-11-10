package simulation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"geova-simulation/state"
	"image/color"
	"math/rand"
	"net/http"
	"time"
)

// --- Generadores de Datos Aleatorios ---

func GenerateRandomIMXData() IMXData {
	return IMXData{
		IDProject:      1,
		Resolution:     "640x480",
		Luminosidad:    5.0 + rand.Float64()*10.0,
		Nitidez:        4.0 + rand.Float64()*2.0,
		LaserDetectado: rand.Intn(2) == 1,
		CalidadFrame:   20.0,
		Confiabilidad:  0.8 + rand.Float64()*0.2,
		Event:          true,
		Timestamp:      time.Now().Format("2006-01-02 15:04:05"),
	}
}

func GenerateRandomMPUData(tilt float64) MPUData {
	return MPUData{
		IDProject: 1,
		Ax:        0.1 + rand.Float64()*0.1,
		Ay:        -0.05 + rand.Float64()*0.1,
		Az:        9.8 + rand.Float64()*0.1,
		Gx:        0.01 + rand.Float64()*0.02,
		Gy:        0.02 + rand.Float64()*0.02,
		Gz:        0.03 + rand.Float64()*0.02,
		Roll:      tilt, // Usa la inclinación actual del trípode
		Pitch:     0.5 + rand.Float64()*1.0,
		Apertura:  tilt * 1.5,
		Event:     true,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}
}

func GenerateRandomTFLunaData() TFLunaData {
	distCm := 150 + rand.Intn(150) // 150-300 cm
	return TFLunaData{
		IDProject:   1,
		DistanciaCm: distCm,
		DistanciaM:  float64(distCm) / 100.0,
		FuerzaSenal: 5000 + rand.Intn(1000),
		Temperatura: 50.0 + rand.Float64()*5.0,
		Event:       true,
		Timestamp:   time.Now().Format("2006-01-02 15:04:05"),
	}
}

// --- Worker Concurrente (Mejorado) ---

func SendPOSTRequest(url string, payload interface{}, packetID string, 
	visState *state.VisualState, startY float64, c color.Color) {
	
	// 1. Crear el paquete visual en posición inicial (trípode)
	visState.Mutex.Lock()
	packet := &state.PacketState{
		ID:       packetID,
		Active:   true,
		X:        80.0,  // Posición del trípode
		Y:        startY,
		TargetX:  250.0, // Python API
		TargetY:  200.0,
		Color:    c,
		Status:   state.SendingToAPI,
		Payload:  payload,
		ProcessingTimer: 0,
	}
	visState.Packets[packetID] = packet
	visState.Mutex.Unlock()

	// 2. Preparar JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("[%s] Error al serializar JSON: %v\n", packetID, err)
		visState.Mutex.Lock()
		visState.Packets[packetID].Status = state.Error
		visState.Mutex.Unlock()
		return
	}

	// 3. Simular latencia de red (500-1000ms)
	time.Sleep(time.Duration(500+rand.Intn(500)) * time.Millisecond)

	// 4. Hacer la petición HTTP POST
	fmt.Printf("[%s] Enviando POST a %s\n", packetID, url)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))

	// 5. Actualizar el estado según el resultado
	visState.Mutex.Lock()
	defer visState.Mutex.Unlock()

	if err != nil {
		fmt.Printf("[%s] Error en HTTP: %v\n", packetID, err)
		visState.Packets[packetID].Status = state.Error
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		fmt.Printf("[%s] Error HTTP %d\n", packetID, resp.StatusCode)
		visState.Packets[packetID].Status = state.Error
		return
	}

	// ¡Éxito! Cambiar a "Arrived"
	fmt.Printf("[%s] ✓ Petición exitosa (HTTP %d)\n", packetID, resp.StatusCode)
	visState.Packets[packetID].Status = state.ArrivedAtAPI
}