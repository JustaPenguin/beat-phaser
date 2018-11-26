package main

import (
	"os"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
)

type soundEffect struct {
	filePath string
	format   beep.Format
	decoded  beep.StreamSeekCloser
}

func (s *soundEffect) load() {
	f1, err := os.Open(s.filePath)

	if err != nil {
		panic(err)
	}

	s1, format, err := vorbis.Decode(f1)

	if err != nil {
		panic(err)
	}

	s.decoded = s1
	s.format = format
}

func (s *soundEffect) play() {

	// @TODO volume sliders with effects/music control
	// Base is 2 for human-natural, 10 would be decibels
	effectVolume := effects.Volume{
		Streamer: s.decoded,
		Base:     2,
		Volume:   -3,
		Silent:   true,
	}

	speaker.Play(beep.Seq(&effectVolume, beep.Callback(func() {
		err := s.decoded.Seek(0)

		if err != nil {
			panic(err)
		}
	})))
}

func loadAudio() {
	f1, err := os.Open("audio/tracks/Kevin_MacLeod_-_AcidJazz.mp3")

	if err != nil {
		panic(err)
	}

	s1, format, err := mp3.Decode(f1)

	if err != nil {
		panic(err)
	}

	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	if err != nil {
		panic(err)
	}

	playing := make(chan struct{})

	songVolume := effects.Volume{
		Streamer: Loop(-1, s1, func() {
			playerScore.startTime = time.Now()
		}),
		Base:   2,
		Volume: -3,
		Silent: false,
	}

	speaker.Play(beep.Seq(&songVolume, beep.Callback(func() {
		close(playing)
	})))
	<-playing
}

func Loop(count int, s beep.StreamSeeker, cb func()) beep.Streamer {
	return &loopCallback{
		s:        s,
		remains:  count,
		callback: cb,
	}
}

type loopCallback struct {
	s        beep.StreamSeeker
	remains  int
	callback func()
}

func (l *loopCallback) Stream(samples [][2]float64) (n int, ok bool) {
	if l.remains == 0 || l.s.Err() != nil {
		return 0, false
	}
	for len(samples) > 0 {
		sn, sok := l.s.Stream(samples)
		if !sok {

			if l.remains == 0 {
				break
			}
			err := l.s.Seek(0)
			if err != nil {
				return n, true
			}
			if l.remains > 0 {
				l.remains--
			}

			l.callback()
			continue
		}
		samples = samples[sn:]
		n += sn
	}
	return n, true
}

func (l *loopCallback) Err() error {
	return l.s.Err()
}
