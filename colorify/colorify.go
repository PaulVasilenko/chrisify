package colorify

import (
	"github.com/lucasb-eyer/go-colorful"
	"image"
	"image/color"
	"math"
)

type Lab struct {
	Pix []float64
}

type LabStat struct {
	LStat Stat
	AStat Stat
	BStat Stat
}

type Stat struct {
	Mean   float64
	StdDev float64
}

func Transfer(src *image.RGBA, target *image.NRGBA) *image.NRGBA {
	srcLab := RGBAToLab(src)
	targetLab := NRGBAToLab(target)

	srcLabStat := srcLab.Stat()
	targetLabStat := targetLab.Stat()

	forEachLABCounter(targetLab, func(l, a, b float64, counter int) {
		targetLab.Pix[counter]   = (l-targetLabStat.LStat.Mean) * (srcLabStat.LStat.StdDev / targetLabStat.LStat.StdDev) + srcLabStat.LStat.Mean
		targetLab.Pix[counter+1] = (a-targetLabStat.AStat.Mean) * (srcLabStat.AStat.StdDev / targetLabStat.AStat.StdDev) + srcLabStat.AStat.Mean
		targetLab.Pix[counter+2] = (b-targetLabStat.BStat.Mean) * (srcLabStat.BStat.StdDev / targetLabStat.BStat.StdDev) + srcLabStat.BStat.Mean
	})

	scaleLab(targetLab)

	w := target.Rect.Dx()

	newTargetRGBA := image.NewNRGBA(target.Rect)
	for x := target.Bounds().Min.X; x < target.Bounds().Max.X; x++ {
		for y := target.Bounds().Min.Y; y < target.Bounds().Max.Y; y++ {
			point := src.RGBAAt(x, y)

			ind := (w * x + y) * 3

			c := colorful.Lab(targetLab.Pix[ind], targetLab.Pix[ind+1], targetLab.Pix[ind+2])
			R, G, B, _ := c.RGBA()

			point.R = uint8(R)
			point.B = uint8(B)
			point.G = uint8(G)

			newTargetRGBA.SetNRGBA(x, y, color.NRGBA(point))
		}
	}

	return newTargetRGBA
}

func forEachNRGBA(src *image.NRGBA, f func(r, g, b uint32)) {
	for x := src.Bounds().Min.X; x < src.Bounds().Max.X; x++ {
		for y := src.Bounds().Min.Y; y < src.Bounds().Max.Y; y++ {
			point := src.At(x, y)
			r, g, b, _ := point.RGBA()
			f(r, g, b)
		}
	}
}

func forEachRGBA(src *image.RGBA, f func(r, g, b uint32)) {
	var r, g, b uint32

	for i, pix := range src.Pix {
		switch i % 4 {
		case 0:
			r = uint32(pix)
		case 1:
			g = uint32(pix)
		case 2:
			b = uint32(pix)
			f(r, g, b)
		case 3:
			continue
		}
	}
}

func forEachLAB(src *Lab, f func(l, a, b float64)) {
	var l, a, b float64

	for i, pix := range src.Pix {
		switch i % 3 {
		case 0:
			l = pix
		case 1:
			a = pix
		case 2:
			b = pix
			f(l, a, b)
		}
	}
}

func NRGBAToLab(src *image.NRGBA) *Lab {
	lab := &Lab{}
	forEachNRGBA(src, func(R, G, B uint32){
		l, a, b := rgbToLab(R, G, B)
		lab.Pix = append(lab.Pix, l, a, b)
	})

	return lab
}

func rgbToLab(R, G, B uint32) (l, a, b float64) {
	c := colorful.LinearRgb(float64(R), float64(G), float64(B))
	l, a, b = c.Lab()
	//fR := math.Max(1.0/255.0, float64(R) / 255.0)
	//fG := math.Max(1.0/255.0, float64(G) / 255.0)
	//fB := math.Max(1.0/255.0, float64(B) / 255.0)
	//
	//L := 0.3811*float64(fR) + 0.5783*float64(fG) + 0.0402*float64(fB)
	//M := 0.1967*float64(fR) + 0.7244*float64(fG) + 0.0782*float64(fB)
	//S := 0.0241*float64(fR) + 0.1288*float64(fG) + 0.8444*float64(fB)
	//
	//L = math.Log(L)
	//M = math.Log(M)
	//S = math.Log(S)
	//
	//l = 1.0 / math.Sqrt(3) * L + 1.0 / math.Sqrt(3) * M + 1.0 / math.Sqrt(3) * S
	//a = 1.0 / math.Sqrt(6) * L + 1.0 / math.Sqrt(6) * M - 2.0 / math.Sqrt(6) * S
	//b = 1.0 / math.Sqrt(2) * L - 1.0 / math.Sqrt(2) * M + 0 * S
	return
}

func RGBAToLab(src *image.RGBA) *Lab {
	lab := &Lab{}
	forEachRGBA(src, func(R, G, B uint32){
		l, a, b := rgbToLab(R, G, B)
		lab.Pix = append(lab.Pix, l, a, b)
	})

	return lab
}

func scaleLab(src *Lab) {
	var minL, maxL, minA, maxA, minB, maxB float64
	forEachLAB(src, func(l, a, b float64) {
		if l < minL {
			minL = l
		}
		if a < minA {
			minL = a
		}
		if b < minB {
			minB = b
		}

		if l > maxA {
			maxL = l
		}
		if a > maxA {
			maxA = a
		}
		if b > maxA {
			maxB = b
		}
	})

	if minL < 0 || maxL > 255 || minA < 0 || maxA > 255 || minB < 0 || maxB > 255 {
		forEachLABCounter(src, func(l, a, b float64, counter int) {
			src.Pix[counter] = 255 * (l - minL) / (maxL - minL)
			src.Pix[counter+1] = 255 * (a - minA) / (maxA - minA)
			src.Pix[counter+2] = 255 * (b - minB) / (maxB - minB)
		})
	}
}

func (src *Lab) Stat() *LabStat {
	var lMean, aMean, bMean float64
	forEachLAB(src, func(l, a, b float64) {
		lMean += l
		aMean += a
		bMean += b
	})

	amount := float64(len(src.Pix)) / 3

	lMean /= amount
	aMean /= amount
	bMean /= amount

	var lStd, aStd, bStd float64
	forEachLAB(src, func(l, a, b float64) {
		lStd += math.Pow(l - lMean, 2)
		aStd += math.Pow(a - aMean, 2)
		bStd += math.Pow(b - bMean, 2)
	})

	lStd = math.Sqrt(lStd/amount-1)
	aStd = math.Sqrt(aStd/amount-1)
	bStd = math.Sqrt(bStd/amount-1)

	return &LabStat{
		LStat: Stat{
			Mean:   lMean,
			StdDev: lStd,
		},
		AStat: Stat{
			Mean:   lMean,
			StdDev: lStd,
		},
		BStat: Stat{
			Mean:   bMean,
			StdDev: bStd,
		},
	}
}

func forEachLABCounter(src *Lab, f func(l, a, b float64, counter int)) {
	counter := 0
	forEachLAB(src, func(l, a, b float64) {
		f(l, a, b, counter)
		counter += 3
	})
}