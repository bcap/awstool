package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := loadConfig(ctx)
	if err != nil {
		log.Fatalf("Failed to load config, cannot continue: %v", err)
		os.Exit(1)
	}

	if err := run(ctx, cfg); err != nil {
		os.Exit(1)
	}

	log.Println("Done")
}

func loadConfig(ctx context.Context) (aws.Config, error) {
	log.Println("Loading default config")

	opts := []func(*config.LoadOptions) error{}
	//TODO gate this behind a flag
	opts = append(opts, config.WithSharedConfigProfile("vtex"))

	return config.LoadDefaultConfig(ctx, opts...)
}
