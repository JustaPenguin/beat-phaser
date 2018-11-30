package main

import (
	"fmt"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
	"image/color"
	"time"
)

type score struct {
	score float64
	multiplier    int
	increment     int
	multiplierPos pixel.Vec
	scorePos      pixel.Vec

	startTime          time.Time
	timeWindow, onBeat bool

	audio   *audio
	audioCh chan time.Time

	atlas *text.Atlas

	color color.Color
}

func (s *score) incrementScore(by float64) {
	s.score += by
}

func (s *score) setMultiplier(multiplier int) {
	if multiplier < 1 {
		return
	}

	s.multiplier = multiplier
}

func (s *score) draw(win *pixelgl.Window, canvas *pixelgl.Canvas) {
	multiplierText := text.New(s.multiplierPos, s.atlas)
	multiplierText.Color = s.color

	_, err := fmt.Fprintf(multiplierText, "%dx", s.multiplier)

	if err != nil {
		panic(err)
	}

	multiplierText.Draw(win, pixel.IM.Moved(canvas.Bounds().Min))

	scoreText := text.New(s.scorePos, s.atlas)
	scoreText.Color = s.color

	_, err = fmt.Fprintf(scoreText, "%.0f", s.score)

	if err != nil {
		panic(err)
	}

	scorePos := canvas.Bounds().Min
	scorePos.X = canvas.Bounds().Max.X - 40

	scoreText.Draw(win, pixel.IM.Moved(scorePos))
}

func (s *score) init() {

	s.multiplier = 1

	// change audio track here
	track := acidJazzAudio

	err := track.load()

	if err != nil {
		panic(err)
	}

	s.audio = track
	s.audioCh = make(chan time.Time)

	s.multiplierPos = pixel.V(20, 20)
	s.scorePos = pixel.V(0, 20)

	s.atlas = text.NewAtlas(
		basicfont.Face7x13,
		[]rune{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'x', '-'},
	)

	go func() {
		for {
			select {
			case t := <-s.audioCh:
				s.startTime = t.Add(time.Nanosecond * 27000000)
			}
		}
	}()

	// start the current track
	go s.audio.play(s.audioCh)
}

func (s *score) update() {
	// If time is within 10ms of the bpm (from start time)
	timeSince := (time.Now().UnixNano() - s.startTime.UnixNano()) % (int64(60000000000 / s.audio.bpm))
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

			if s.multiplier < 1 {
				s.multiplier = 1
			}
		}
	}

	// @TODO colours for scores, perhaps a little animated multi tone stuff for big ones
	switch s.multiplier {
	case 0:
		s.multiplier = 1
	case 1:
		s.color = colornames.Aqua
	case 2:
		s.color = colornames.Blue
	case 3:
		s.color = colornames.Blueviolet
	case 4:
		s.color = colornames.Purple
	case 5:
		s.color = colornames.Deeppink
	case 6:
		s.color = colornames.Coral
	case 7:
		s.color = colornames.Orangered
	case 8:
		s.color = colornames.Red
	}
}
