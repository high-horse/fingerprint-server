package main

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/jtejido/sourceafis"
	"github.com/jtejido/sourceafis/config"

	"log"
)

type TransparencyContents struct {
}

func (c *TransparencyContents) Accepts(key string) bool {
	return true
}

func (c *TransparencyContents) Accept(key, mime string, data []byte) error {
	//fmt.Printf("%d B  %s %s \n", len(data), mime, key)
	return nil
}

func compareFingerprint(probeImage, candidatImage string) (float64, error) {
	now := time.Now()
	config.LoadDefaultConfig()
	config.Config.Workers = runtime.NumCPU()

	probeImg, err := sourceafis.LoadImage(probeImage)
	if err != nil {
		return 0, fmt.Errorf("failed to load probe image: %w", err)
	}

	l := sourceafis.NewTransparencyLogger(new(TransparencyContents))
	tc := sourceafis.NewTemplateCreator(l)
	probe, err := tc.Template(probeImg)
	if err != nil {
		return 0, fmt.Errorf("failed to create template for probe image: %w", err)
	}

	candidateImg, err := sourceafis.LoadImage(candidatImage)
	if err != nil {
		return 0, fmt.Errorf("failed to load candidate image: %w", err)
	}
	candidate, err := tc.Template(candidateImg)
	if err != nil {
		return 0, fmt.Errorf("failed to create template for candidate image: %w", err)
	}

	matcher, err := sourceafis.NewMatcher(l, probe)
	if err != nil {
		return 0, fmt.Errorf("failed to create matcher: %w", err)
	}

	ctx := context.Background()
	score := matcher.Match(ctx, candidate)

	fmt.Println("elapsed: ", time.Since(now))
	return score, nil
}

func example() {
	now := time.Now()
	config.LoadDefaultConfig()
	config.Config.Workers = runtime.NumCPU()
	probeImg, err := sourceafis.LoadImage("probe.png")
	if err != nil {
		log.Fatal(err.Error())
	}
	l := sourceafis.NewTransparencyLogger(new(TransparencyContents))
	tc := sourceafis.NewTemplateCreator(l)
	probe, err := tc.Template(probeImg)
	if err != nil {
		log.Fatal(err.Error())
	}
	candidateImg, err := sourceafis.LoadImage("matching.png")
	if err != nil {
		log.Fatal(err.Error())
	}
	candidate, err := tc.Template(candidateImg)
	if err != nil {
		log.Fatal(err.Error())
	}

	candidateImg2, err := sourceafis.LoadImage("nonmatching.png")
	if err != nil {
		log.Fatal(err.Error())
	}
	candidate2, err := tc.Template(candidateImg2)
	if err != nil {
		log.Fatal(err.Error())
	}

	matcher, err := sourceafis.NewMatcher(l, probe)
	if err != nil {
		log.Fatal(err.Error())
	}
	ctx := context.Background()
	fmt.Println("matching score ===> ", matcher.Match(ctx, candidate))
	fmt.Println("non-matching score ===> ", matcher.Match(ctx, candidate2))
	fmt.Println("elapsed: ", time.Since(now))
}