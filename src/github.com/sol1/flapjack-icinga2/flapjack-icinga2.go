package main

import (
	"fmt"
	flapjack "github.com/flapjack/flapjack/src/flapjack"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cli := CLI{}

	config := cli.ParseArgs()
	errs := config.Errors()

	if len(errs) > 0 {
		for _, e := range errs {
			fmt.Println(e)
		}
		os.Exit(1)
	}

	if config.Debug {
		log.Printf("Starting with config: %+v\n", config)
	}

	// shutdown signal handler; must be buffered, or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	transport, err := flapjack.Dial(config.RedisServer, config.RedisDatabase)
	if err != nil {
		log.Fatalf("Couldn't establish Redis connection: %s\n", err)
	}

	finished := make(chan error)

	api_client := ApiClient{config: config, redis: transport}
	api_client.Connect(finished)

	select {
	case err := <-finished:
		if config.Debug {
			log.Printf("Finished with error // %s\n", err)
		}

		signal.Stop(sigs)
	case <-sigs:
		log.Println("Interrupted, cancelling request")
		signal.Stop(sigs)

		// TODO determine if request not currently active...
		api_client.Cancel()
	}

	// close redis connection
	transport.Close()
}
