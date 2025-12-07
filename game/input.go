package game

import (
	"fmt"
	"geova-simulation/simulation"
	"geova-simulation/state"
	"image"
	"image/color"
	"time"

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

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if g.BotonRect.Bounds().Canon().Overlaps(
			image.Rectangle{Min: clickPoint, Max: clickPoint.Add(image.Pt(1, 1))},
		) {
			g.toggleSimulation()
		}
	}
}

func (g *Game) toggleSimulation() {
	g.State.Mutex.Lock()
	if g.State.SimulacionIniciada {
		if g.State.StopChan != nil {
			close(g.State.StopChan)
			g.State.StopChan = nil
		}
		g.State.SimulacionIniciada = false
		g.State.Mutex.Unlock()
		fmt.Println("[SIMULACIÓN] Detenida")
		return
	}

	g.State.Packets = make(map[string]*state.PacketState)
	g.State.DisplayDistancia = 0
	g.State.DisplayNitidez = 0
	g.State.DisplayRoll = 0
	g.State.SimulacionIniciada = true
	g.State.PacketID = 0
	g.State.StopChan = make(chan struct{})
	stopChan := g.State.StopChan
	g.State.Mutex.Unlock()

	fmt.Println("[SIMULACIÓN] Iniciada - Click de nuevo para detener")

	go g.runContinuousSimulation(stopChan)
}

func (g *Game) runContinuousSimulation(stopChan chan struct{}) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	g.sendBatchRequests()

	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			g.sendBatchRequests()
		}
	}
}

func (g *Game) sendBatchRequests() {
	g.State.Mutex.Lock()
	tilt := g.State.CurrentTilt
	g.State.PacketID++
	id := g.State.PacketID
	g.State.Mutex.Unlock()

	go simulation.SendPOSTRequest(
		"http://localhost:8000/tfluna/sensor",
		simulation.GenerateRandomTFLunaData(),
		fmt.Sprintf("tfluna_%d", id), g.State, 180.0, color.RGBA{R: 255, G: 50, B: 50, A: 255},
	)
	go simulation.SendPOSTRequest(
		"http://localhost:8000/mpu/sensor",
		simulation.GenerateRandomMPUData(tilt),
		fmt.Sprintf("mpu_%d", id), g.State, 200.0, color.RGBA{R: 50, G: 150, B: 255, A: 255},
	)
	go simulation.SendPOSTRequest(
		"http://localhost:8000/imx477/sensor",
		simulation.GenerateRandomIMXData(),
		fmt.Sprintf("imx_%d", id), g.State, 220.0, color.RGBA{R: 50, G: 255, B: 50, A: 255},
	)
}
