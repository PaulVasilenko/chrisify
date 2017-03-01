package main

import (
	"cloud.google.com/go/vision"
	"flag"
	"github.com/disintegration/imaging"
	"golang.org/x/net/context"
	"image"
	"image/draw"
	"image/jpeg"
	_ "image/png"
	"log"
	"log/syslog"
	"math/rand"
	"os"
	"path/filepath"
)

var haarCascade = flag.String("haar", "haarcascade_frontalface_alt.xml", "The location of the Haar Cascade XML configuration to be provided to OpenCV.")
var facesDir = flag.String("faces", "faces", "The directory to search for faces.")

func main() {
	log.SetFlags(0)

	syslogWriter, err := syslog.New(syslog.LOG_INFO, "chrisify")

	if err == nil {
		log.SetOutput(syslogWriter)
	}

	flag.Parse()

	var chrisFaces FaceList
	var facesPath string

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
	log.Println("Processing file: ", file)
	baseImage := loadImage(file)

	ctx := context.Background()

	client, err := vision.NewClient(ctx)
	if err != nil {
		panic(err)
	}

	f, err := os.Open(file)

	if err != nil {
		panic(err)
	}

	defer f.Close()

	imageToDetect, err := vision.NewImageFromReader(f)

	faces, err := client.DetectFaces(ctx, imageToDetect, 10)

	if err != nil {
		panic(err)
	}

	bounds := baseImage.Bounds()

	canvas := canvasFromImage(baseImage)

	numberList := rand.Perm(len(faces))

	for i, face := range faces {
		rect := image.Rect(
			face.BoundingPoly[0].X,
			face.BoundingPoly[0].Y,
			face.BoundingPoly[2].X,
			face.BoundingPoly[2].Y)
		newFace := chrisFaces[numberList[i]%len(chrisFaces)]
		if newFace == nil {
			panic("nil face")
		}
		chrisFace := imaging.Resize(newFace, rect.Dx(), rect.Dy(), imaging.Lanczos)

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
