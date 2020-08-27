package main

import (
	"fmt"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/empty"
	"github.com/google/go-containerregistry/pkg/v1/mutate"
	"github.com/google/go-containerregistry/pkg/v1/random"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/types"
	"io/ioutil"
	"log"
	"os"
)

func run(imageOutputURL, layerOutputFilePath string) error {
	// rand layer results in a random image ever time
	randLayer, err := random.Layer(512, types.DockerForeignLayer)

	// generate image with unreachable URL
	image, err := mutate.Append(empty.Image, mutate.Addendum{
		Layer: randLayer,
		URLs:  []string{"http://localhost/never-gonna-find-it"},
	})
	if err != nil {
		return err
	}

	// write layer data to disk for later upload
	layerReader, err := randLayer.Compressed()
	if err != nil {
		return err
	}
	layerBytes, err := ioutil.ReadAll(layerReader)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(layerOutputFilePath, layerBytes, 0666); err != nil {
		return err
	}

	// write image to registry, without layer data
	imageRef, err := name.ParseReference(imageOutputURL, name.WeakValidation)
	if err != nil {
		return err
	}
	if err := remote.Write(imageRef, image); err != nil {
		return err
	}

	return nil
}

func main() {
	imageOutputURL := os.Args[1]
	layerOutputFilePath := os.Args[2]
	if imageOutputURL == "" || layerOutputFilePath == "" {
		fmt.Printf("usage: %s <localhost:5000/test> <layer.tar>")
		os.Exit(1)
	}

	if err := run(imageOutputURL, layerOutputFilePath); err != nil {
		log.Fatal(err)
	}

	fmt.Println("image and layer written")
}
