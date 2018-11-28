package main

import (
	"fmt"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/font/basicfont"
	"math/rand"
	"time"
)

const advertMessageSpacing = "     "
const textSpeed = 5

var advertMessages = []string{
	"Welcome to Beat Phaser. Try not to die, won't you?",
	"This is another hilarious message",
	"Another hilarious message that is much longer than the previous hilarious message hahaha",
}

type advert struct {
	pos   pixel.Vec
	atlas *text.Atlas

	i        float64
	maxWidth int

	text      string
	shownText string

	tick <-chan time.Time
}

func (a *advert) init() {
	a.atlas = text.NewAtlas(
		basicfont.Face7x13,
		text.ASCII,
	)

	a.text = advertMessages[0] + advertMessageSpacing
	a.tick = time.Tick(time.Second * 20)
}

func (a *advert) pickRandomMessage() {
	r := rand.Intn(len(advertMessages))

	a.text = advertMessages[r] + advertMessageSpacing
	a.shownText = ""
	a.i = 0
}

func (a *advert) update(dt float64) {
	i := int(a.i)

	if i == len(a.text)-1 {
		select {
		case <-a.tick:
			a.pickRandomMessage()
			return
		default:

		}
	}

	if i+a.maxWidth < len(a.text) {
		a.shownText = a.text[i : i+a.maxWidth]

		a.i += dt * textSpeed
	} else if i < len(a.text) {
		a.shownText = a.text[i:]

		a.i += dt * textSpeed
	} else {
		a.i = 0
	}

	j := 0

	for len(a.shownText) < a.maxWidth && j < len(a.text) {
		a.shownText += string(a.text[j])
		j++
	}
}

func (a *advert) draw(t pixel.Target) {
	tx := text.New(a.pos, a.atlas)
	tx.Color = pixel.RGB(235.0/255.0, 35.0/255.0, 208.0/255.0)

	_, err := fmt.Fprintf(tx, a.shownText)

	if err != nil {
		panic(err)
	}

	tx.Draw(t, pixel.IM)
}
