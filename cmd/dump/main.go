package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/retry"
	"github.com/aws/aws-sdk-go-v2/config"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := configure(ctx)
	if err != nil {
		log.Fatalf("Failed to load config, cannot continue: %v", err)
		os.Exit(1)
	}

	result, err := DumpAWS(ctx, cfg)
	if err != nil {
		log.Printf("Execution failed: %v", err)
		os.Exit(1)
	}

	log.Print("Data fully loaded, encoding to json")

	jsonBytes, err := json.Marshal(result)
	if err != nil {
		log.Fatalf("Error while encoding result to json: %v", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonBytes))

	log.Println("Done")
}

func configure(ctx context.Context) (aws.Config, error) {
	createRetryer := func() aws.Retryer {
		return retry.NewStandard(
			func(opts *retry.StandardOptions) {
				opts.MaxAttempts = 5
				opts.MaxBackoff = 10 * time.Second
			},
		)
	}

	return config.LoadDefaultConfig(
		ctx,
		config.WithSharedConfigProfile("vtex"),
		config.WithRetryer(createRetryer),
	)
}
