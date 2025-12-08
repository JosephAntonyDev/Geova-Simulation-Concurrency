package game

import (
	"fmt"
	"geova-simulation/state"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

func (g *Game) Draw(screen *ebiten.Image) {
	g.drawBackground(screen)
	g.drawTripode(screen)
	g.drawTiltMeter(screen)
	g.drawIcons(screen)
	g.drawPackets(screen)
	g.drawButton(screen)
	g.drawDashboard(screen)
	ebitenutil.DebugPrintAt(screen, "Controles:  Flechas <- -> para inclinar ANTES de crear  |  Click en CREAR  |  F11 pantalla completa", 10, 10)
}

func (g *Game) drawBackground(screen *ebiten.Image) {
	if g.Assets.Background != nil {
		op := &ebiten.DrawImageOptions{}
		screenW, screenH := screen.Bounds().Dx(), screen.Bounds().Dy()
		bgW, bgH := g.Assets.Background.Bounds().Dx(), g.Assets.Background.Bounds().Dy()

		scaleX := float64(screenW) / float64(bgW)
		scaleY := float64(screenH) / float64(bgH)

		op.GeoM.Scale(scaleX, scaleY)
		screen.DrawImage(g.Assets.Background, op)
	} else {
		screen.Fill(color.RGBA{R: 0x1a, G: 0x1a, B: 0x1a, A: 255})
	}
}

func (g *Game) drawTripode(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(tripodeX, tripodeY)

	frameIndex := g.getTripodeFrame(g.State.CurrentTilt)
	sx := frameIndex * tripodeFrameWidth
	rect := image.Rect(sx, 0, sx+tripodeFrameWidth, tripodeFrameHeight)

	screen.DrawImage(g.Assets.UITiltMeter.SubImage(rect).(*ebiten.Image), op)
}

func (g *Game) getTripodeFrame(tilt float64) int {
	switch {
	case tilt <= -12.5:
		return 0
	case tilt <= -7.5:
		return 1
	case tilt <= -2.5:
		return 2
	case tilt < 2.5:
		return 3
	case tilt < 7.5:
		return 4
	case tilt < 12.5:
		return 5
	default:
		return 6
	}
}

func (g *Game) drawTiltMeter(screen *ebiten.Image) {
	ebitenutil.DebugPrintAt(screen,
		fmt.Sprintf("Inclinación Actual: %.1f°", g.State.CurrentTilt),
		int(tiltMeterX), int(tiltMeterY))

	meterX := int(tiltMeterX) + 200
	meterY := int(tiltMeterY)

	for i := -15; i <= 15; i++ {
		x := meterX + i*3
		ebitenutil.DebugPrintAt(screen, "|", x, meterY)
	}

	markerX := meterX + int(g.State.CurrentTilt*3)
	ebitenutil.DebugPrintAt(screen, "▼", markerX-2, meterY-15)
}

func (g *Game) drawButton(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(float64(g.BotonRect.Min.X), float64(g.BotonRect.Min.Y))

	if g.State.SimulacionIniciada {
		if g.isBotonPressed {
			screen.DrawImage(g.Assets.ButtonStopDown, op)
		} else {
			screen.DrawImage(g.Assets.ButtonStopUp, op)
		}
	} else {
		if g.isBotonPressed {
			screen.DrawImage(g.Assets.ButtonStartDown, op)
		} else {
			screen.DrawImage(g.Assets.ButtonStartUp, op)
		}
	}
}

func (g *Game) drawMonitor(screen *ebiten.Image) {
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(monitorX, monitorY)

	if g.State.SimulacionIniciada {
		frameIndex := (g.animIconCounter / monitorAnimSpeed) % monitorFrameCount
		sx := frameIndex * monitorFrameWidth
		rect := image.Rect(sx, 0, sx+monitorFrameWidth, monitorFrameHeight)
		screen.DrawImage(g.Assets.MonitorAnim.SubImage(rect).(*ebiten.Image), op)
	} else {
		screen.DrawImage(g.Assets.IconMonitor, op)
	}
}

func (g *Game) drawIcons(screen *ebiten.Image) {
	g.drawIcon(screen, g.Assets.IconPythonIdle, g.Assets.IconPythonActiveAnim,
		g.State.PythonAPITimer, iconPythonX, iconPythonY)
	g.drawIcon(screen, g.Assets.IconRabbitIdle, g.Assets.IconRabbitActiveAnim,
		g.State.RabbitMQTimer, iconRabbitX, iconRabbitY)
	g.drawIcon(screen, g.Assets.IconWebsocketIdle, g.Assets.IconWebsocketActiveAnim,
		g.State.WebsocketAPITimer, iconWebsocketX, iconWebsocketY)

	g.drawMonitor(screen)
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

		c := packet.Color.(color.RGBA)
		op.ColorScale.SetR(float32(c.R) / 255)
		op.ColorScale.SetG(float32(c.G) / 255)
		op.ColorScale.SetB(float32(c.B) / 255)

		screen.DrawImage(packetFrame, op)

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

		if packet.Status == state.Error {
			ebitenutil.DebugPrintAt(screen, "✗ ERROR", int(packet.X)-10, int(packet.Y)+25)
		}
	}
}

func (g *Game) drawDashboard(screen *ebiten.Image) {
	y := int(dashboardY)

	ebitenutil.DebugPrintAt(screen, "--- Dashboard de Resultados ---", int(dashboardX), y)
	y += 20

	distText := fmt.Sprintf("  Distancia (TFLuna): %.2f m", g.State.DisplayDistancia)
	if g.State.DisplayDistancia == 0 {
		distText = "  Distancia (TFLuna): --"
	}
	ebitenutil.DebugPrintAt(screen, distText, int(dashboardX), y)
	y += 25

	nitText := "  Nitidez (IMX477):"
	if g.State.DisplayNitidez == 0 {
		nitText = "  Nitidez (IMX477): --"
	}
	ebitenutil.DebugPrintAt(screen, nitText, int(dashboardX), y)

	if g.State.DisplayNitidez > 0 {
		opBarBG := &ebiten.DrawImageOptions{}
		opBarBG.GeoM.Translate(dashboardX+180, float64(y))
		screen.DrawImage(g.Assets.UIProgressBG, opBarBG)

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

		ebitenutil.DebugPrintAt(screen,
			fmt.Sprintf("%.2f", g.State.DisplayNitidez),
			int(dashboardX)+330, y)
	}

	y += 25

	rollText := fmt.Sprintf("  Inclinacion Roll (MPU): %.1f°", g.State.DisplayRoll)
	if g.State.DisplayRoll == 0 {
		rollText = "  Inclinacion Roll (MPU): --"
	}
	ebitenutil.DebugPrintAt(screen, rollText, int(dashboardX), y)

	y += 30

	if g.State.SimulacionIniciada {
		ebitenutil.DebugPrintAt(screen, ">> Procesando solicitudes...", int(dashboardX), y)
	} else {
		ebitenutil.DebugPrintAt(screen, ">> Listo para nueva simulacion", int(dashboardX), y)
	}
}
