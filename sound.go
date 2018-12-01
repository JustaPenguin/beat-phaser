package main

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/effects"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/vorbis"
)

var (
	acidJazzAudio = &audio{
		filepath: filepath.Join("audio", "tracks", "Kevin_MacLeod_-_AcidJazz.mp3"),
		loop:     -1,
		bpm:      110.724,
		buffered: true,
	}

	backedVibesAudio = &audio{
		filepath: filepath.Join("audio", "tracks", "Kevin_MacLeod_Backed_Vibes_Clean.mp3"),
		loop:     -1,
		bpm:      102.230,
	}

	nightOnTheDocksAudio = &audio{
		filepath: filepath.Join("audio", "tracks", "Kevin_MacLeod_-_Night_on_the_Docks_-_Sax.mp3"),
		loop:     -1,
		bpm:      139.658,
	}
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

type audio struct {
	filepath string
	loop     int
	bpm      float64
	buffered bool

	buf      *beep.Buffer
	streamer beep.StreamSeeker
	format   beep.Format
	file     *os.File

	ctx context.Context
	cfn func()
}

func (a *audio) load() error {
	var err error
	a.file, err = os.Open(a.filepath)

	if err != nil {
		return err
	}

	s1, format, err := mp3.Decode(a.file)

	if err != nil {
		return err
	}

	if a.buffered {
		buf := beep.NewBuffer(format)
		buf.Append(s1)

		a.buf = buf
		a.streamer = a.buf.Streamer(0, a.buf.Len())
		a.format = a.buf.Format()
	} else {
		a.streamer = s1
		a.format = format
	}

	a.ctx, a.cfn = context.WithCancel(context.Background())

	return nil
}

func (a *audio) play(ch chan struct{}) {
	err := speaker.Init(a.format.SampleRate, a.format.SampleRate.N(time.Second/10))

	if err != nil {
		panic(err)
	}
loop:
	playing := make(chan struct{})

	err = a.streamer.Seek(0)

	if err != nil {
		panic(err)
	}

	ctrl := &effects.Volume{
		Streamer: beep.Loop(-1, a.streamer),
		Base:     2,
		Volume:   -3,
		Silent:   false,
	}

	speaker.Play(beep.Seq(ctrl, beep.Callback(func() {
		close(playing)
	})))
	go func() {
		ch <- struct{}{}
	}()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-playing:
			if a.loop != 0 {
				a.loop--
				goto loop
			}
		}
	}

}

func (a *audio) close() error {

	return a.file.Close()
}
