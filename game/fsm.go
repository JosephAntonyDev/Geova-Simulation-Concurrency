package game

import (
	"geova-simulation/simulation"
	"geova-simulation/state"
	"math"
)

func (g *Game) updatePacketFSM() {
	g.State.Mutex.Lock()
	defer g.State.Mutex.Unlock()

	if g.State.PythonAPITimer > 0 {
		g.State.PythonAPITimer--
	}
	if g.State.RabbitMQTimer > 0 {
		g.State.RabbitMQTimer--
	}
	if g.State.WebsocketAPITimer > 0 {
		g.State.WebsocketAPITimer--
	}

	allDone := true

	for _, packet := range g.State.Packets {
		if packet.Status == state.Error || packet.Status == state.Done {
			continue
		}

		allDone = false

		dx := packet.TargetX - packet.X
		dy := packet.TargetY - packet.Y
		distance := math.Sqrt(dx*dx + dy*dy)

		if distance > packetSpeed {
			packet.X += (dx / distance) * packetSpeed
			packet.Y += (dy / distance) * packetSpeed
		} else {
			packet.X = packet.TargetX
			packet.Y = packet.TargetY
			g.handlePacketArrival(packet)
		}
	}

	if allDone && len(g.State.Packets) > 0 {
		g.State.SimulacionIniciada = false
	}
}

func (g *Game) handlePacketArrival(packet *state.PacketState) {
	switch packet.Status {
	case state.SendingToAPI:

	case state.ArrivedAtAPI:
		g.State.PythonAPITimer = processingDelay
		packet.ProcessingTimer = processingDelay
		packet.Status = state.ProcessingAtAPI

	case state.ProcessingAtAPI:
		if packet.ProcessingTimer > 0 {
			packet.ProcessingTimer--
		} else {
			packet.Status = state.SendingToRabbit
			packet.TargetX = iconRabbitX
			packet.TargetY = iconRabbitY
		}

	case state.SendingToRabbit:
		if packet.X == packet.TargetX && packet.Y == packet.TargetY {
			g.State.RabbitMQTimer = processingDelay
			packet.ProcessingTimer = processingDelay
			packet.Status = state.ProcessingAtRabbit
		}

	case state.ProcessingAtRabbit:
		if packet.ProcessingTimer > 0 {
			packet.ProcessingTimer--
		} else {
			packet.Status = state.SendingToWebsocket
			packet.TargetX = iconWebsocketX
			packet.TargetY = iconWebsocketY
		}

	case state.SendingToWebsocket:
		if packet.X == packet.TargetX && packet.Y == packet.TargetY {
			g.State.WebsocketAPITimer = processingDelay
			packet.ProcessingTimer = processingDelay
			packet.Status = state.ProcessingAtWebsocket
		}

	case state.ProcessingAtWebsocket:
		if packet.ProcessingTimer > 0 {
			packet.ProcessingTimer--
		} else {
			packet.Status = state.SendingToFrontend
			packet.TargetX = monitorX
			packet.TargetY = monitorY
		}

	case state.SendingToFrontend:
		if packet.X == packet.TargetX && packet.Y == packet.TargetY {
			packet.Status = state.Done
			packet.Active = false
			g.updateDashboard(packet)
		}
	}
}

func (g *Game) updateDashboard(packet *state.PacketState) {
	switch data := packet.Payload.(type) {
	case simulation.TFLunaData:
		g.State.DisplayDistancia = data.DistanciaM
	case simulation.MPUData:
		g.State.DisplayRoll = data.Roll
	case simulation.IMXData:
		g.State.DisplayNitidez = data.Nitidez
	}
}
