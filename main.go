package main

import (
	"fmt"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"log"
	"os"
)

var logger *log.Logger

func run(imageName string, writeLayerBlob bool) error {
	imageRef, err := name.ParseReference(imageName, name.WeakValidation)

	// rand layer results in a random foreignLayerImage ever time
	randLayer, err := random.Layer(512, types.DockerForeignLayer)

	if writeLayerBlob {
		logger.Printf("writing layer data first %s", imageName)
		if err := remote.WriteLayer(imageRef.Context(), randLayer, remote.WithAuthFromKeychain(authn.DefaultKeychain)); err != nil {
			return err
		}
	} else {
		logger.Print("ommitting layer blob. Image will be unpullable")
	}

	logger.Printf("writing actual image with foreign layer references in tact %s", imageName)
	foreignLayerImage, err := mutate.Append(empty.Image, mutate.Addendum{
		Layer: randLayer,
		URLs:  []string{"http://localhost/never-gonna-find-it"},
	})
	if err != nil {
		return err
	}
	if err := remote.Write(imageRef, foreignLayerImage, remote.WithAuthFromKeychain(authn.DefaultKeychain)); err != nil {
		return err
	}

	return nil
}

func main() {
	logger = log.New(os.Stderr, "", log.Lshortfile)

	imageName := os.Args[1]
	if imageName == "" {
		fmt.Printf("usage: %s <localhost:5000/test> <optional: 'false' to write image without layer data>")
		os.Exit(1)
	}

	writeLayerBlob := true
	if len(os.Args) == 3 && os.Args[2] == "false" {
		writeLayerBlob = false
	}

	if err := run(imageName, writeLayerBlob); err != nil {
		log.Fatal(err)
	}

	fmt.Println("image and layer written")
}
