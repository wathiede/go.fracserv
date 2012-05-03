// Copyright 2012 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package fractal

import (
	"log"
	"image"
	"image/color"
	"math"
	"net/url"
	"strconv"
)

type Options struct {
	url.Values
}

// Converts value with key 'k' to int, in absence or failure 'd' is returned
func (o Options) GetIntDefault(k string, d int) int {
	v, err := strconv.Atoi(o.Get(k))
	if err != nil {
		log.Printf("Failed to parse %s: %s", k, o.Get(k), err)
		return d
	}
	return v
}

func (o Options) GetFloat64Default(k string, d float64) float64 {
	v, err := strconv.ParseFloat(o.Get(k), 64)
	if err != nil {
		log.Printf("Failed to parse %s: %s", k, o.Get(k), err)
		return d
	}
	return v
}

type Fractal interface {
	image.Image
}

type Navigator interface {
	// Convert pixel space to fractal space
	Transform(p image.Point) (float64, float64)
	// Set offset in pixel space
	Translate(offset image.Point)
	// Set the zoom depth for this transformation
	Zoom(z float64)
}

type DefaultNavigator struct {
	z float64
	offset image.Point
}

func NewDefaultNavigator(z float64, xoff, yoff int) DefaultNavigator {
	return DefaultNavigator{z, image.Pt(xoff, yoff)}
}

func (n *DefaultNavigator) Transform(p image.Point) (float64, float64) {
	x := p.X + n.offset.X
	y := p.Y + n.offset.Y
	z := math.Pow(2, n.z)
	return float64(x)/z, float64(y)/z
}

func (n *DefaultNavigator) GetTranslate() image.Point {
	return n.offset
}

func (n *DefaultNavigator) Translate(offset image.Point) {
	n.offset = offset
}

func (n *DefaultNavigator) Zoom(z float64) {
	n.z = z
}

func (n *DefaultNavigator) GetZoom() float64 {
	return n.z
}

func HSVToRGBA(h, s, v float64) color.RGBA {
	hi := int(math.Mod(math.Floor(h / 60), 6))
	f := (h / 60) - math.Floor(h / 60)
	p := v * (1 - s)
	q := v * (1 - (f*s))
	t := v * (1 - ((1 - f) * s))

	RGB := func(r, g, b float64) color.RGBA {
		return color.RGBA{uint8(r * 255), uint8(g * 255), uint8(b * 255), 255}
	}

	var c color.RGBA
	switch hi {
		case 0: c = RGB(v, t, p)
		case 1: c = RGB(q, v, p)
		case 2: c = RGB(p, v, t)
		case 3: c = RGB(p, q, v)
		case 4: c = RGB(t, p, v)
		case 5: c = RGB(v, p, q)
	}
	return c
}
