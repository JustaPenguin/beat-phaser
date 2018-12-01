package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"image/color"
)

type colorLight struct {
	color  color.Color
	point  pixel.Vec
	angle  float64
	radius float64

	spread float64

	imd *imdraw.IMDraw
}

func (cl *colorLight) init() {
	// create the light arc if not created already
	if cl.imd == nil {
		imd := imdraw.New(nil)

		cl.imd = imd
	}
}

func (cl *colorLight) draw(dst pixel.Target) {
	cl.imd.Clear()

	if playerScore.timeWindow {
		cl.imd.Intensity = 0
	} else {
		cl.imd.Intensity = 0.1
	}

	cl.imd.Color = cl.color
	cl.imd.SetMatrix(pixel.IM.Scaled(pixel.ZV, cl.radius).Rotated(pixel.ZV, cl.angle).Moved(cl.point))
	cl.imd.Push(pixel.ZV)
	cl.imd.Color = pixel.Alpha(0)
	for angle := -cl.spread / 2; angle <= cl.spread/2; angle += cl.spread / 64 {
		cl.imd.Push(pixel.V(1, 0).Rotated(angle))
	}
	cl.imd.Polygon(0)

	cl.imd.Draw(dst)
}
