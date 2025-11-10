package state

import (
	"image/color"
	"sync"
)

type PacketStatus int

const (
	Idle PacketStatus = iota
	SendingToAPI
	ArrivedAtAPI
	ProcessingAtAPI       // NUEVO: Estado intermedio para procesar
	SendingToRabbit
	ProcessingAtRabbit    // NUEVO
	SendingToWebsocket
	ProcessingAtWebsocket // NUEVO
	SendingToFrontend
	Done
	Error
)

type PacketState struct {
	ID        string
	Active    bool
	X, Y      float64
	TargetX, TargetY float64
	Color     color.Color
	Status    PacketStatus
	Payload   interface{}
	ProcessingTimer int // NUEVO: Para contar frames de espera
}

type VisualState struct {
	Mutex   sync.Mutex
	Packets map[string]*PacketState

	// Timers para animaciones de iconos
	PythonAPITimer    int
	RabbitMQTimer     int
	WebsocketAPITimer int

	// Datos para el dashboard
	DisplayDistancia float64
	DisplayRoll      float64
	DisplayNitidez   float64
	CurrentTilt      float64
	SimulacionIniciada bool
}