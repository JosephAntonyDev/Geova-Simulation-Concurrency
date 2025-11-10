package game

import (
	"fmt"
	"geova-simulation/assets"
	"geova-simulation/simulation"
	"geova-simulation/state"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// --- Constantes de Posición CORREGIDAS ---
const (
	tripodeX = 80.0
	tripodeY = 200.0

	// Iconos alineados horizontalmente
	iconPythonX    = 250.0
	iconPythonY    = 200.0
	iconRabbitX    = 400.0
	iconRabbitY    = 200.0
	iconWebsocketX = 550.0
	iconWebsocketY = 200.0

	monitorX = 620.0
	monitorY = 180.0

	tiltMeterX = 100.0
	tiltMeterY = 50.0

	dashboardX = 50.0
	dashboardY = 450.0

	packetSpeed = 3.0 // Velocidad aumentada para mejor fluidez

	// Delay en frames para que el icono "procese" antes de enviar
	processingDelay = 30 // 0.5 segundos a 60 FPS
)

type Game struct {
	Assets *assets.Assets
	State  *state.VisualState

	BotonRect      image.Rectangle
	isBotonPressed bool

	animPacketCounter int
	animIconCounter   int
}

func NewGame(assets *assets.Assets, state *state.VisualState, btnRect image.Rectangle) *Game {
	return &Game{
		Assets:    assets,
		State:     state,
		BotonRect: btnRect,
	}
}

func (g *Game) Update() error {
	g.animPacketCounter = (g.animPacketCounter + 1) % 360
	g.animIconCounter = (g.animIconCounter + 1) % 360

	g.handleInput()
	g.updatePacketFSM()

	return nil
}

func (g *Game) handleInput() {
	// Pantalla completa con F11
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}

	x, y := ebiten.CursorPosition()
	clickPoint := image.Pt(x, y)

	g.isBotonPressed = g.BotonRect.Bounds().Canon().Overlaps(
		image.Rectangle{Min: clickPoint, Max: clickPoint.Add(image.Pt(1, 1))},
	) && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)

	// Control de inclinación con teclado
	if ebiten.IsKeyPressed(ebiten.KeyLeft) && g.State.CurrentTilt > -15.0 {
		g.State.CurrentTilt -= 0.5
	}
	if ebiten.IsKeyPressed(ebiten.KeyRight) && g.State.CurrentTilt < 15.0 {
		g.State.CurrentTilt += 0.5
	}

	// Iniciar simulación con click en botón
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

	// Resetear todo el estado
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

	// Lanzar las 3 goroutines con colores distintivos
	go simulation.SendPOSTRequest(
		"http://localhost:8000/tfluna/sensor",
		simulation.GenerateRandomTFLunaData(),
		"tfluna", g.State, 180.0, color.RGBA{R: 255, G: 50, B: 50, A: 255}, // Rojo
	)
	go simulation.SendPOSTRequest(
		"http://localhost:8000/mpu/sensor",
		simulation.GenerateRandomMPUData(tilt),
		"mpu", g.State, 200.0, color.RGBA{R: 50, G: 150, B: 255, A: 255}, // Azul
	)
	go simulation.SendPOSTRequest(
		"http://localhost:8000/imx477/sensor",
		simulation.GenerateRandomIMXData(),
		"imx", g.State, 220.0, color.RGBA{R: 50, G: 255, B: 50, A: 255}, // Verde
	)
}

func (g *Game) updatePacketFSM() {
	g.State.Mutex.Lock()
	defer g.State.Mutex.Unlock()

	// Decrementar timers de animación
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
		// Ignorar paquetes que ya terminaron o fallaron
		if packet.Status == state.Error || packet.Status == state.Done {
			continue
		}

		allDone = false

		// Mover el paquete hacia su objetivo
		dx := packet.TargetX - packet.X
		dy := packet.TargetY - packet.Y
		distance := math.Sqrt(dx*dx + dy*dy)

		if distance > packetSpeed {
			// Mover hacia el objetivo
			packet.X += (dx / distance) * packetSpeed
			packet.Y += (dy / distance) * packetSpeed
		} else {
			// Llegó al objetivo
			packet.X = packet.TargetX
			packet.Y = packet.TargetY

			// Transición de estado según la FSM
			g.handlePacketArrival(packet)
		}
	}

	// Si todos los paquetes terminaron, reactivar el botón
	if allDone && len(g.State.Packets) > 0 {
		g.State.SimulacionIniciada = false
	}
}

func (g *Game) handlePacketArrival(packet *state.PacketState) {
	switch packet.Status {
	case state.SendingToAPI:
		// No hacer nada, esperar que el worker cambie el estado a ArrivedAtAPI

	case state.ArrivedAtAPI:
		// Activar animación del icono Python
		g.State.PythonAPITimer = processingDelay
		packet.ProcessingTimer = processingDelay
		packet.Status = state.ProcessingAtAPI

	case state.ProcessingAtAPI:
		// Esperar el timer de procesamiento
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

			// Actualizar el dashboard con los datos
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

// ============ DRAW METHODS ============

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{R: 0x1a, G: 0x1a, B: 0x1a, A: 255})

	// Dibujar elementos en orden de profundidad
	g.drawTripode(screen)
	g.drawTiltMeter(screen)
	g.drawIcons(screen)
	g.drawPackets(screen)
	g.drawButton(screen)
	g.drawDashboard(screen)

	// Instrucciones mejoradas
	ebitenutil.DebugPrintAt(screen, "Controles:  Flechas <- -> para inclinar ANTES de crear  |  Click en CREAR  |  F11 pantalla completa", 10, 10)
}

func (g *Game) drawTripode(screen *ebiten.Image) {
	opTripode := &ebiten.DrawImageOptions{}
	opTripode.GeoM.Translate(tripodeX, tripodeY)

	// El sprite geova_tilt_anim.png tiene 7 frames horizontales (128x128 cada uno)
	// Total: 896x128 (7 frames de 128x128)
	frameWidth := 128
	frameHeight := 128
	frameIndex := 3 // Por defecto, nivelado (centro - frame 3 de 7)

	// Seleccionar frame según inclinación (7 frames: 0 a 6)
	// Frame 0: Máxima inclinación izquierda (-15°)
	// Frame 1: Inclinación izquierda media-alta (-10°)
	// Frame 2: Inclinación izquierda media-baja (-5°)
	// Frame 3: Nivelado (0°)
	// Frame 4: Inclinación derecha media-baja (+5°)
	// Frame 5: Inclinación derecha media-alta (+10°)
	// Frame 6: Máxima inclinación derecha (+15°)
	tilt := g.State.CurrentTilt
	if tilt <= -12.5 {
		frameIndex = 0
	} else if tilt <= -7.5 {
		frameIndex = 1
	} else if tilt <= -2.5 {
		frameIndex = 2
	} else if tilt < 2.5 {
		frameIndex = 3
	} else if tilt < 7.5 {
		frameIndex = 4
	} else if tilt < 12.5 {
		frameIndex = 5
	} else {
		frameIndex = 6
	}

	// Calcular región del sprite (frames horizontales)
	sx := frameIndex * frameWidth
	rect := image.Rect(sx, 0, sx+frameWidth, frameHeight)

	// Dibujar el frame correcto de geova_tilt_anim.png
	screen.DrawImage(g.Assets.UITiltMeter.SubImage(rect).(*ebiten.Image), opTripode)
}

func (g *Game) drawTiltMeter(screen *ebiten.Image) {
	// Mostrar inclinación actual en tiempo real
	ebitenutil.DebugPrintAt(screen,
		fmt.Sprintf("Inclinación Actual: %.1f°", g.State.CurrentTilt),
		int(tiltMeterX), int(tiltMeterY))

	// Indicador visual de barra
	meterX := int(tiltMeterX) + 200
	meterY := int(tiltMeterY)

	// Dibujar línea base
	for i := -15; i <= 15; i++ {
		x := meterX + i*3
		ebitenutil.DebugPrintAt(screen, "|", x, meterY)
	}

	// Dibujar marcador de posición actual
	markerX := meterX + int(g.State.CurrentTilt*3)
	ebitenutil.DebugPrintAt(screen, "▼", markerX-2, meterY-15)
}

func (g *Game) drawButton(screen *ebiten.Image) {
	opBoton := &ebiten.DrawImageOptions{}
	opBoton.GeoM.Translate(float64(g.BotonRect.Min.X), float64(g.BotonRect.Min.Y))

	if g.State.SimulacionIniciada {
		// Mostrar botón deshabilitado con opacidad reducida
		opBoton.ColorScale.Scale(0.5, 0.5, 0.5, 1.0) // RGB atenuado, Alpha normal
		screen.DrawImage(g.Assets.ButtonCreateUp, opBoton)
	} else if g.isBotonPressed {
		screen.DrawImage(g.Assets.ButtonCreateDown, opBoton)
	} else {
		screen.DrawImage(g.Assets.ButtonCreateUp, opBoton)
	}
}

func (g *Game) drawIcons(screen *ebiten.Image) {
	g.drawIcon(screen, g.Assets.IconPythonIdle, g.Assets.IconPythonActiveAnim,
		g.State.PythonAPITimer, iconPythonX, iconPythonY)
	g.drawIcon(screen, g.Assets.IconRabbitIdle, g.Assets.IconRabbitActiveAnim,
		g.State.RabbitMQTimer, iconRabbitX, iconRabbitY)
	g.drawIcon(screen, g.Assets.IconWebsocketIdle, g.Assets.IconWebsocketActiveAnim,
		g.State.WebsocketAPITimer, iconWebsocketX, iconWebsocketY)

	opMonitor := &ebiten.DrawImageOptions{}
	opMonitor.GeoM.Translate(monitorX, monitorY)
	screen.DrawImage(g.Assets.IconMonitor, opMonitor)
}

func (g *Game) drawIcon(screen *ebiten.Image, idle *ebiten.Image, anim *ebiten.Image,
	timer int, x, y float64) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y)

	if timer > 0 {
		frameWidth := 64
		frameCount := 6
		frameIndex := (g.animIconCounter / 6) % frameCount
		sx := frameIndex * frameWidth
		rect := image.Rect(sx, 0, sx+frameWidth, 64)
		screen.DrawImage(anim.SubImage(rect).(*ebiten.Image), op)
	} else {
		screen.DrawImage(idle, op)
	}
}

func (g *Game) drawPackets(screen *ebiten.Image) {
	g.State.Mutex.Lock()
	defer g.State.Mutex.Unlock()

	frameWidth := 32
	frameCount := 6
	frameIndex := (g.animPacketCounter / 6) % frameCount
	sx := frameIndex * frameWidth
	rect := image.Rect(sx, 0, sx+frameWidth, 32)
	packetFrame := g.Assets.DataPacketAnim.SubImage(rect).(*ebiten.Image)

	for _, packet := range g.State.Packets {
		if !packet.Active {
			continue
		}

		op := &ebiten.DrawImageOptions{}
		op.GeoM.Translate(packet.X, packet.Y)

		// Aplicar color
		c := packet.Color.(color.RGBA)
		op.ColorScale.SetR(float32(c.R) / 255)
		op.ColorScale.SetG(float32(c.G) / 255)
		op.ColorScale.SetB(float32(c.B) / 255)

		screen.DrawImage(packetFrame, op)

		// Etiqueta identificadora encima del paquete
		labelX := int(packet.X) - 15
		labelY := int(packet.Y) - 10

		var label string
		switch packet.ID {
		case "tfluna":
			label = "TFL"
		case "mpu":
			label = "MPU"
		case "imx":
			label = "IMX"
		}

		ebitenutil.DebugPrintAt(screen, label, labelX, labelY)

		// Mostrar X si hay error
		if packet.Status == state.Error {
			ebitenutil.DebugPrintAt(screen, "✗ ERROR", int(packet.X)-10, int(packet.Y)+25)
		}
	}
}

func (g *Game) drawDashboard(screen *ebiten.Image) {
	y := int(dashboardY)

	ebitenutil.DebugPrintAt(screen, "--- Dashboard de Resultados ---", int(dashboardX), y)
	y += 20

	// Distancia (TF-Luna) - Rojo
	distText := fmt.Sprintf("  Distancia (TFLuna): %.2f m", g.State.DisplayDistancia)
	if g.State.DisplayDistancia == 0 {
		distText = "  Distancia (TFLuna): --"
	}
	ebitenutil.DebugPrintAt(screen, distText, int(dashboardX), y)
	y += 25

	// Nitidez (IMX477) - Verde con barra
	nitText := "  Nitidez (IMX477):"
	if g.State.DisplayNitidez == 0 {
		nitText = "  Nitidez (IMX477): --"
	}
	ebitenutil.DebugPrintAt(screen, nitText, int(dashboardX), y)

	if g.State.DisplayNitidez > 0 {
		opBarBG := &ebiten.DrawImageOptions{}
		opBarBG.GeoM.Translate(dashboardX+180, float64(y))
		screen.DrawImage(g.Assets.UIProgressBG, opBarBG)

		// Normalizar nitidez de 4.0-6.0 a 0.0-1.0
		normalizedNitidez := (g.State.DisplayNitidez - 4.0) / 2.0
		if normalizedNitidez < 0 {
			normalizedNitidez = 0
		}
		if normalizedNitidez > 1 {
			normalizedNitidez = 1
		}

		opBarFill := &ebiten.DrawImageOptions{}
		opBarFill.GeoM.Scale(normalizedNitidez, 1.0)
		opBarFill.GeoM.Translate(dashboardX+180, float64(y))
		screen.DrawImage(g.Assets.UIProgressFill, opBarFill)

		// Mostrar valor numérico
		ebitenutil.DebugPrintAt(screen,
			fmt.Sprintf("%.2f", g.State.DisplayNitidez),
			int(dashboardX)+330, y)
	}

	y += 25

	// Inclinación (MPU) - Azul
	rollText := fmt.Sprintf("  Inclinacion Roll (MPU): %.1f°", g.State.DisplayRoll)
	if g.State.DisplayRoll == 0 {
		rollText = "  Inclinacion Roll (MPU): --"
	}
	ebitenutil.DebugPrintAt(screen, rollText, int(dashboardX), y)

	y += 30

	// Estado de la simulación
	if g.State.SimulacionIniciada {
		ebitenutil.DebugPrintAt(screen, ">> Procesando solicitudes...", int(dashboardX), y)
	} else {
		ebitenutil.DebugPrintAt(screen, ">> Listo para nueva simulacion", int(dashboardX), y)
	}
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 900, 650
}
