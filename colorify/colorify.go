package colorify

import (
	"github.com/lucasb-eyer/go-colorful"
	"image"
	"image/color"
	"math"
	"log"
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
	log.Println("start color transfer")
	srcLab := RGBAToLab(src)
	targetLab := NRGBAToLab(target)
	srcLabStat := srcLab.Stat()
	targetLabStat := targetLab.Stat()
	forEachLABCounter(targetLab, func(l, a, b float64, counter int) {
		targetLab.Pix[counter]   = (l-targetLabStat.LStat.Mean) * (srcLabStat.LStat.StdDev / targetLabStat.LStat.StdDev) + srcLabStat.LStat.Mean
		targetLab.Pix[counter+1] = (a-targetLabStat.AStat.Mean) * (srcLabStat.AStat.StdDev / targetLabStat.AStat.StdDev) + srcLabStat.AStat.Mean
		targetLab.Pix[counter+2] = (b-targetLabStat.BStat.Mean) * (srcLabStat.BStat.StdDev / targetLabStat.BStat.StdDev) + srcLabStat.BStat.Mean
	})
//	scaleLab(targetLab)

	newTargetRGBA := image.NewNRGBA(target.Rect)
	ind := 0
	for x := target.Bounds().Min.X; x < target.Bounds().Max.X; x++ {
		for y := target.Bounds().Min.Y; y < target.Bounds().Max.Y; y++ {
			point := target.NRGBAAt(x, y)
			c := colorful.Lab(targetLab.Pix[ind], targetLab.Pix[ind+1], targetLab.Pix[ind+2])
			R, G, B := c.RGB255()
			point.R = R
			point.B = B
			point.G = G

			newTargetRGBA.SetNRGBA(x, y, color.NRGBA(point))
			ind += 3
		}
	}

	log.Println("finish color transfer")
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
	for x := src.Bounds().Min.X; x < src.Bounds().Max.X; x++ {
		for y := src.Bounds().Min.Y; y < src.Bounds().Max.Y; y++ {
			point := src.At(x, y)
			r, g, b, _ := point.RGBA()
			f(r, g, b)
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
	c := colorful.Color{float64(R) / 65535.0, float64(G) / 65535.0, float64(B) / 65535.0}
	l, a, b = c.Lab()
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

	lStd = math.Sqrt(lStd/(amount-1))
	aStd = math.Sqrt(aStd/(amount-1))
	bStd = math.Sqrt(bStd/(amount-1))

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
