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
	"image"
	"testing"
)

type P image.Point
type F struct {
	X float64
	Y float64
}

func TestNavigator(t *testing.T) {
	var datum = []struct {
		z        float64
		pix, off P
		frac     F
	}{
		{1, P{0, 0}, P{0, 0}, F{0, 0}},
		{1, P{0, 0}, P{4, 4}, F{4, 4}},
		{4, P{0, 0}, P{0, 0}, F{0, 0}},
		{4, P{0, 0}, P{4, 4}, F{1, 1}},
		{4, P{4, 4}, P{0, 0}, F{1, 1}},
		{4, P{4, 4}, P{4, 4}, F{2, 2}},
	}

	for _, d := range datum {
		nav := DefaultNavigator{d.z, image.Pt(d.off.X, d.off.Y)}
		transX, transY := nav.Transform(image.Pt(d.pix.X, d.pix.Y))
		if transX != d.frac.X {
			t.Errorf("Transform failed for %v: expected X of %f, got %f", d,
				d.frac.X, transX)
		}
		if transY != d.frac.Y {
			t.Errorf("Transform failed for %v: expected Y of %f, got %f", d,
				d.frac.Y, transY)
		}
	}
}
