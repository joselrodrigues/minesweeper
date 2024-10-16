package main

import (
	"log"

	_ "image/png"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

type Game struct {
	sprites []*Sprite
}

type Sprite struct {
	image *ebiten.Image
	x, y  float64
}

func (g *Game) Update() error {
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	return 320, 240
}

func main() {
	ebiten.SetWindowSize(640, 480)
	ebiten.SetWindowTitle("Hello, World!")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	img, _, err := ebitenutil.NewImageFromFile("assets/sprite/board.png")
	if err != nil {
		log.Fatal(err)
	}

	if err := ebiten.RunGame(&Game{}); err != nil {
		log.Fatal(err)
	}
}
