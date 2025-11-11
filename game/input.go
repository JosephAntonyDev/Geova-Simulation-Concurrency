package game

import (
	"geova-simulation/simulation"
	"geova-simulation/state"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func (g *Game) handleInput() {
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}

	x, y := ebiten.CursorPosition()
	clickPoint := image.Pt(x, y)

	g.isBotonPressed = g.BotonRect.Bounds().Canon().Overlaps(
		image.Rectangle{Min: clickPoint, Max: clickPoint.Add(image.Pt(1, 1))},
	) && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)

	if ebiten.IsKeyPressed(ebiten.KeyLeft) && g.State.CurrentTilt > -15.0 {
		g.State.CurrentTilt -= 0.5
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) && g.State.CurrentTilt < 15.0 {
		g.State.CurrentTilt += 0.5
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && !g.State.SimulacionIniciada {
		if g.BotonRect.Bounds().Canon().Overlaps(
			image.Rectangle{Min: clickPoint, Max: clickPoint.Add(image.Pt(1, 1))},
		) {
			g.startSimulation()
		}
	}
}

func (g *Game) startSimulation() {
	g.State.Mutex.Lock()

	g.State.Packets = make(map[string]*state.PacketState)
	g.State.DisplayDistancia = 0
	g.State.DisplayNitidez = 0
	g.State.DisplayRoll = 0
	g.State.SimulacionIniciada = true
	g.State.PythonAPITimer = 0
	g.State.RabbitMQTimer = 0
	g.State.WebsocketAPITimer = 0

	tilt := g.State.CurrentTilt
	g.State.Mutex.Unlock()

	go simulation.SendPOSTRequest(
		"http://localhost:8000/tfluna/sensor",
		simulation.GenerateRandomTFLunaData(),
		"tfluna", g.State, 180.0, color.RGBA{R: 255, G: 50, B: 50, A: 255},
	)
	go simulation.SendPOSTRequest(
		"http://localhost:8000/mpu/sensor",
		simulation.GenerateRandomMPUData(tilt),
		"mpu", g.State, 200.0, color.RGBA{R: 50, G: 150, B: 255, A: 255},
	)
	go simulation.SendPOSTRequest(
		"http://localhost:8000/imx477/sensor",
		simulation.GenerateRandomIMXData(),
		"imx", g.State, 220.0, color.RGBA{R: 50, G: 255, B: 50, A: 255},
	)
}
