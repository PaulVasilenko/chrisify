package colorify

import (
	"image"
	"math"
)

type Lab struct {
	Pix []float64
}

func ApplyToRGBA(src *image.RGBA, dst *image.RGBA) {

}

func forEachRGBA(src *image.RGBA, f func(r, g, b uint8)) {
	var r, g, b uint8

	for i, pix := range src.Pix {
		switch i % 4 {
		case 0:
			r = pix
		case 1:
			g = pix
		case 2:
			b = pix
			f(r, g, b)
		case 3:
			continue
		}
	}
}

func RGBAToLab(src *image.RGBA) *Lab {
	lab := &Lab{}
	forEachRGBA(src, func(R, G, B uint8){
		fR := math.Max(1.0/255.0, float64(R) / 255.0)
		fG := math.Max(1.0/255.0, float64(G) / 255.0)
		fB := math.Max(1.0/255.0, float64(B) / 255.0)

		L := 0.3811*float64(fR) + 0.5783*float64(fG) + 0.0402*float64(fB)
		M := 0.1967*float64(fR) + 0.7244*float64(fG) + 0.0782*float64(fB)
		S := 0.0241*float64(fR) + 0.1288*float64(fG) + 0.8444*float64(fB)

		L = math.Log(L)
		M = math.Log(M)
		S = math.Log(S)

		l := 1.0 / math.Sqrt(3) * L + 1.0 / math.Sqrt(3) * M + 1.0 / math.Sqrt(3) * S
		a := 1.0 / math.Sqrt(6) * L + 1.0 / math.Sqrt(6) * M - 2.0 / math.Sqrt(6) * S
		b := 1.0 / math.Sqrt(2) * L - 1.0 / math.Sqrt(2) * M + 0 * S

		lab.Pix = append(lab.Pix, l, a, b)
	})

	return lab
}
