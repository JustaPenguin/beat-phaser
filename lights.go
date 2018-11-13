package main

import (
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
)

type colorlight struct {
	color  pixel.RGBA
	point  pixel.Vec
	angle  float64
	radius float64
	dust   float64

	spread float64

	imd *imdraw.IMDraw
}

func (cl *colorlight) apply(dst pixel.ComposeTarget, center pixel.Vec) {
	// create the light arc if not created already
	if cl.imd == nil {
		imd := imdraw.New(nil)
		imd.Color = pixel.Alpha(1)
		imd.Push(pixel.ZV)
		imd.Color = pixel.Alpha(0)
		for angle := -cl.spread / 2; angle <= cl.spread/2; angle += cl.spread / 64 {
			imd.Push(pixel.V(1, 0).Rotated(angle))
		}
		imd.Polygon(0)
		cl.imd = imd
	}

	// draw the light arc
	dst.SetMatrix(pixel.IM.Scaled(pixel.ZV, cl.radius).Rotated(pixel.ZV, cl.angle).Moved(cl.point))
	dst.SetColorMask(pixel.Alpha(1))
	dst.SetComposeMethod(pixel.ComposePlus)
	cl.imd.Draw(dst)
}
