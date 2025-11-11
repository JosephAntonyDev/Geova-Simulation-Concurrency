package game

import (
	"geova-simulation/assets"
	"geova-simulation/state"
	"image"
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

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 900, 650
}
