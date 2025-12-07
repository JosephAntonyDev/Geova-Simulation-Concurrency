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
	ProcessingAtAPI
	SendingToRabbit
	ProcessingAtRabbit
	SendingToWebsocket
	ProcessingAtWebsocket
	SendingToFrontend
	Done
	Error
)

type PacketState struct {
	ID               string
	Active           bool
	X, Y             float64
	TargetX, TargetY float64
	Color            color.Color
	Status           PacketStatus
	Payload          interface{}
	ProcessingTimer  int
}

type VisualState struct {
	Mutex   sync.Mutex
	Packets map[string]*PacketState

	PythonAPITimer    int
	RabbitMQTimer     int
	WebsocketAPITimer int

	DisplayDistancia   float64
	DisplayRoll        float64
	DisplayNitidez     float64
	CurrentTilt        float64
	SimulacionIniciada bool

	StopChan chan struct{}
	PacketID int
}
