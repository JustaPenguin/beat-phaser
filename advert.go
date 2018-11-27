package main

import (
	"fmt"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/font/basicfont"
)

type advert struct {
	pos pixel.Vec
	atlas *text.Atlas

	i float64
	maxWidth int

	text string
	shownText string
}

func (a *advert) init() {
	a.atlas = text.NewAtlas(
		basicfont.Face7x13,
		text.ASCII,
	)

	a.text = "Welcome to Beat Phaser. Try not to die, won't you?     "
}

func (a *advert) update(dt float64) {
	i := int(a.i)

	if i+a.maxWidth < len(a.text) {
		a.shownText = a.text[i:i+a.maxWidth]

		a.i += dt * 10
	} else if i < len(a.text) {
		a.shownText = a.text[i:]

		a.i += dt * 10
	} else {
		a.i = 0
	}

	j := 0

	for len(a.shownText) < a.maxWidth {
		a.shownText += string(a.text[j])
		j++
	}
}

func (a *advert) draw(t pixel.Target) {
	tx := text.New(a.pos, a.atlas)
	tx.Color = pixel.RGB(235.0/255.0,35.0/255.0,208.0/255.0)

	_, err := fmt.Fprintf(tx, a.shownText)

	if err != nil {
		panic(err)
	}

	tx.Draw(t, pixel.IM)
}