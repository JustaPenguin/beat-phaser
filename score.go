package main

import (
	"fmt"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
	"time"
)

type score struct {
	multiplier int
	increment  int
	pos        pixel.Vec

	startTime time.Time
	timeWindow bool

	text *text.Text
}

func (s *score) setMultiplier(multiplier int) {
	s.multiplier = multiplier
}

func (s *score) draw() {
	atlas := text.NewAtlas(
		basicfont.Face7x13,
		[]rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', 'x'},
	)

	s.text = text.New(s.pos, atlas)

	// @TODO colours for scores, perhaps a little animated multi tone stuff for big ones
	switch s.multiplier {
	case 1:
		s.text.Color = colornames.Aqua
	case 2:
		s.text.Color = colornames.Coral

	}

	_, err := fmt.Fprintf(s.text, "%dx", s.multiplier)

	if err != nil {
		panic(err)
	}
}

func (s *score) init() {
	s.multiplier = 0
	s.pos = pixel.V(20, 20)

	s.startTime = time.Now()
}

func (s *score) update() {
	// If time is within 10ms of the bpm (from start time)
	if time.Now().UnixNano() - s.startTime.UnixNano() % (int64(60000000000/bpm)) <= 10000000 {
		s.timeWindow = true
	} else {
		s.timeWindow = false
	}


	if win.JustPressed(pixelgl.MouseButtonLeft) {

		if s.timeWindow {
			s.increment--

			if s.increment <= 0 {
				s.multiplier--
				s.increment = 8
			}

			if s.multiplier < 0 {
				s.multiplier = 0
			}

		} else {
			s.increment++

			if s.increment >= 8 {
				s.multiplier++
				s.increment = 0
			}

			if s.multiplier > 8 {
				s.multiplier = 8
			}
		}

	}

	println(s.multiplier)
}
