package main

import (
	"flag"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"

	"github.com/lestrrat-go/dither"
	"github.com/pkg/errors"
)

var (
	vraptorAPIURL = "http://vraptor.local:5000/api"
	vraptorUser   = "vraptor"
	vraptorPass   = "vraptor"

	flagURL          string
	flagUser         string
	flagPass         string
	flagDitherMethod string
)

func main() {
	flag.StringVar(&flagURL, "url", vraptorAPIURL, "vraptor api url")
	flag.StringVar(&flagUser, "u", vraptorUser, "vraptor api user")
	flag.StringVar(&flagPass, "p", vraptorPass, "vraptor api password")
	flag.StringVar(&flagDitherMethod, "d", "burkes", "dither mehtods: none, burkes, floydsteinberg, sierra2, sierra3, sierra_lite, stucki, atkinson")
	flag.Parse()

	vr := newVRaptor(flagURL, flagUser, flagPass)
	if vr == nil {
		log.Fatal("failed to create vraptor client")
	}

	imgFN := flag.Arg(0)
	file, err := os.Open(imgFN)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		log.Fatal(err)
	}

	// check dimensions
	bounds := img.Bounds()
	if bounds.Dx() != 256 || bounds.Dy() != 64 {
		log.Fatalf("image size should be 256x64, but got %dx%d", bounds.Dx(), bounds.Dy())
	}

	img, err = ditherImage(img, flagDitherMethod)
	if err != nil {
		log.Fatal(err)
	}

	err = vr.SetImage(img)
	if err != nil {
		log.Fatal(err)
	}

	err = vr.ImageMode(true)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("success to set image: %s", imgFN)
}

func ditherImage(img image.Image, filterName string) (image.Image, error) {
	var df *dither.Filter
	switch filterName {
	case "none":
		return img, nil
	case "burkes":
		df = dither.Burkes
	case "floydsteinberg":
		df = dither.FloydSteinberg
	case "sierra2":
		df = dither.Sierra2
	case "sierra3":
		df = dither.Sierra3
	case "sierra_lite":
		df = dither.SierraLite
	case "stucki":
		df = dither.Stucki
	case "atkinson":
		df = dither.Atkinson
	default:
		return nil, errors.Errorf("unknown dither method: %s", filterName)
	}

	ditheredImg := dither.Monochrome(df, img, 1.18)
	return ditheredImg, nil
}
