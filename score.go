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
	timeWindow, onBeat bool

	atlas *text.Atlas
	text *text.Text
}

func (s *score) setMultiplier(multiplier int) {
	s.multiplier = multiplier
}

func (s *score) draw() {
	s.text = text.New(s.pos, s.atlas)

	// @TODO colours for scores, perhaps a little animated multi tone stuff for big ones
	switch s.multiplier {
	case 0:
		s.text.Color = colornames.White
	case 1:
		s.text.Color = colornames.Aqua
	case 2:
		s.text.Color = colornames.Blue
	case 3:
		s.text.Color = colornames.Blueviolet
	case 4:
		s.text.Color = colornames.Purple
	case 5:
		s.text.Color = colornames.Deeppink
	case 6:
		s.text.Color = colornames.Coral
	case 7:
		s.text.Color = colornames.Orangered
	case 8:
		s.text.Color = colornames.Red


	}

	_, err := fmt.Fprintf(s.text, "%dx", s.multiplier)

	if err != nil {
		panic(err)
	}
}

func (s *score) init() {
	s.multiplier = 0
	s.pos = pixel.V(20, 20)

	s.atlas = text.NewAtlas(
		basicfont.Face7x13,
		[]rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', 'x'},
	)

	s.startTime = time.Now()
}

func (s *score) update() {
	// If time is within 10ms of the bpm (from start time)
	timeSince := (time.Now().UnixNano() - s.startTime.UnixNano()) % (int64(60000000000/bpm))
	if timeSince <= 100000000 || timeSince >= 400000000 {
		s.timeWindow = true
	} else {
		s.timeWindow = false
	}


	if win.JustPressed(pixelgl.MouseButtonLeft) {

		if s.timeWindow {
			s.onBeat = true
			s.increment++

			if s.increment >= 8 {
				s.multiplier++
				s.increment = 0
			}

			if s.multiplier > 8 {
				s.multiplier = 8
			}

		} else {
			s.onBeat = false
			s.increment = s.increment - 2

			if s.increment <= 0 {
				s.multiplier--
				s.increment = 8
			}

			if s.multiplier < 0 {
				s.multiplier = 0
			}
		}

	}

}