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

func run(imageName string) error {
	// write foreignLayerImage to registry, without layer data
	imageRef, err := name.ParseReference(imageName, name.WeakValidation)

	// rand layer results in a random foreignLayerImage ever time
	randLayer, err := random.Layer(512, types.DockerForeignLayer)

	logger.Printf("writing actual layer data %s", imageName)
	if err := remote.WriteLayer(imageRef.Context(), randLayer, remote.WithAuthFromKeychain(authn.DefaultKeychain)); err != nil {
		return err
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
		fmt.Printf("usage: %s <localhost:5000/test>")
		os.Exit(1)
	}

	if err := run(imageName); err != nil {
		log.Fatal(err)
	}

	fmt.Println("image and layer written")
}
