package main

import (
	"flag"
	"image"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"

	"github.com/disintegration/imaging"
	"github.com/harrydb/go/img/grayscale"
	"github.com/paulvasilenko/chrisify/facefinder"
)

var haarCascade = flag.String("haar", "haarcascade_frontalface_alt.xml", "The location of the Haar Cascade XML configuration to be provided to OpenCV.")
var facesDir = flag.String("faces", "faces", "The directory to search for faces.")

func main() {
	flag.Parse()

	var chrisFaces FaceList

	var facesPath string
	var err error

	if *facesDir != "" {
		facesPath, err = filepath.Abs(*facesDir)
		if err != nil {
			panic(err)
		}
	}

	err = chrisFaces.Load(facesPath)
	if err != nil {
		panic(err)
	}
	if len(chrisFaces) == 0 {
		panic("no faces found")
	}

	file := flag.Arg(0)

	finder := facefinder.NewFinder(*haarCascade)

	baseImage := loadImage(file)

	imageToDetect := imaging.Resize(
		baseImage,
		640,
		0,
		imaging.Lanczos)

	blurred := imaging.Resize(
		imageToDetect,
		640,
		0,
		imaging.Gaussian)

	grayImg := grayscale.Convert(blurred, grayscale.ToGrayLuminance)

	faces := finder.Detect(grayImg)

	bounds := imageToDetect.Bounds()

	canvas := canvasFromImage(imageToDetect)

	for _, face := range faces {
		rect := rectMargin(5.0, face)
		newRect := image.Rect(
			rect.Min.X,
			rect.Min.Y,
			rect.Max.X+rect.Min.X/6,
			rect.Max.Y+rect.Min.X/6)

		newFace := chrisFaces.Random()
		if newFace == nil {
			panic("nil face")
		}
		chrisFace := imaging.Fit(newFace, newRect.Dx(), newRect.Dy(), imaging.Lanczos)

		draw.Draw(
			canvas,
			newRect,
			chrisFace,
			bounds.Min,
			draw.Over,
		)
	}

	if len(faces) == 0 {
		face := imaging.Resize(
			chrisFaces[0],
			bounds.Dx()/3,
			0,
			imaging.Lanczos,
		)
		face_bounds := face.Bounds()
		draw.Draw(
			canvas,
			bounds,
			face,
			bounds.Min.Add(image.Pt(-bounds.Max.X/2+face_bounds.Max.X/2, -bounds.Max.Y+int(float64(face_bounds.Max.Y)/1.9))),
			draw.Over,
		)
	}

	jpeg.Encode(os.Stdout, canvas, &jpeg.Options{jpeg.DefaultQuality})
}
