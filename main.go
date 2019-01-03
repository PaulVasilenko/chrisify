package main

import (
	"flag"
	"image"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"log"
	"log/syslog"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"cloud.google.com/go/vision/apiv1"
	"github.com/disintegration/imaging"
	"github.com/paulvasilenko/go-transcolor"
	"golang.org/x/net/context"
)

var facesDir = flag.String("faces", "faces", "The directory to search for faces.")
var colorCorrection = flag.Bool("color-correction", false, "Whether to correct face color to image color or not")

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	log.SetFlags(0)

	syslogWriter, err := syslog.New(syslog.LOG_INFO, "chrisify")

	if err == nil {
		log.SetOutput(syslogWriter)
	}

	flag.Parse()

	var facesPath string

	if *facesDir != "" {
		facesPath, err = filepath.Abs(*facesDir)
		if err != nil {
			panic(err)
		}
	}

	chrisFaces := FaceList{}
	err = chrisFaces.Load(facesPath)
	if err != nil {
		panic(err)
	}
	if len(chrisFaces) == 0 {
		panic("no faces found")
	}
	file := flag.Arg(0)
	log.Println("Processing file: ", file)
	baseImage := loadImage(file)

	ctx := context.Background()

	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		panic(err)
	}

	f, err := os.Open(file)

	if err != nil {
		panic(err)
	}

	defer f.Close()

	imageToDetect, err := vision.NewImageFromReader(f)

	faces, err := client.DetectFaces(ctx, imageToDetect, nil, 100)

	if err != nil {
		panic(err)
	}

	bounds := baseImage.Bounds()

	canvas := canvasFromImage(baseImage)
	sourceCanvas := canvasFromImage(baseImage)

	numberList := rand.Perm(100)

	for i, face := range faces {
		rect := image.Rect(
			int(face.BoundingPoly.Vertices[0].X),
			int(face.BoundingPoly.Vertices[0].Y),
			int(face.BoundingPoly.Vertices[2].X),
			int(face.BoundingPoly.Vertices[2].Y))
		newFace := chrisFaces[numberList[i]%len(chrisFaces)]
		if newFace == nil {
			panic("nil face")
		}
		var chrisFace image.Image
		chrisFace = imaging.Resize(newFace, rect.Dx(), rect.Dy(), imaging.Lanczos)
		if *colorCorrection {
			chrisFace = transcolor.Transfer(sourceCanvas.SubImage(rect), chrisFace)
		}
		draw.Draw(
			canvas,
			rect,
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
