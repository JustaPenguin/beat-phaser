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
		s.decoded.Seek(0)
	})))
}

func loadAudio() {
	f1, err := os.Open("audio/tracks/Kevin_MacLeod_Backed_Vibes_Clean.mp3")

	if err != nil {
		panic(err)
	}

	s1, format, err := mp3.Decode(f1)

	if err != nil {
		panic(err)
	}

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	playing := make(chan struct{})

	songVolume := effects.Volume{
		Streamer: s1,
		Base:     2,
		Volume:   -3,
		Silent:   true,
	}

	speaker.Play(beep.Seq(&songVolume, beep.Callback(func() {
		close(playing)
	})))
	<-playing
}
