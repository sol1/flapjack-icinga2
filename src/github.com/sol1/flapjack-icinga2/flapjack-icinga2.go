package main

import (
	"fmt"
	"github.com/sol1/flapjack-icinga2/flapjack"
	"gopkg.in/alecthomas/kingpin.v2"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	app = kingpin.New("flapjack-icinga2", "Transfers Icinga 2 events to Flapjack")

	icinga_env_api_user     = os.Getenv("ICINGA2_API_USER")
	icinga_env_api_password = os.Getenv("ICINGA2_API_PASSWORD")

	icinga_server    = app.Flag("icinga-url", "Icinga 2 API endpoint to connect to (default localhost:5665)").Default("localhost:5665").String()
	icinga_certfile  = app.Flag("icinga-certfile", "Path to Icinga 2 API TLS certfile").String()
	icinga_user      = app.Flag("icinga-user", "Icinga 2 basic auth user (required, also checks ICINGA2_API_USER env var)").Default(icinga_env_api_user).String()
	icinga_password  = app.Flag("icinga-password", "Icinga 2 basic auth password (required, also checks ICINGA2_API_PASSWORD env var)").Default(icinga_env_api_password).String()
	icinga_queue     = app.Flag("icinga-queue", "Icinga 2 event queue name to use (default flapjack)").Default("flapjack").String()
	icinga_timeout   = app.Flag("icinga-timeout", "Icinga 2 API connection timeout, in milliseconds (default 30_000)").Default("30000").Int()
	icinga_keepalive = app.Flag("icinga-keepalive", "Icinga 2 API frequency of keepalive traffic, in milliseconds (default 30_000)").Default("30000").Int()

	// default Redis port is 6380 rather than 6379 as the Flapjack packages ship
	// with an Omnibus-packaged Redis running on a different port to the
	// distro-packaged one
	redis_server   = app.Flag("redis-url", "Redis server to connect to (default localhost:6380)").Default("localhost:6380").String()
	redis_database = app.Flag("redis-db", "Redis database to connect to (default 0)").Int()

	flapjack_version = app.Flag("flapjack-version", "Flapjack version being delivered to (1 or 2) (default 1)").Default("1").Int()
	flapjack_events  = app.Flag("flapjack-events", "Flapjack event queue name to use (default events)").Default("events").String()

	debug = app.Flag("debug", "Enable verbose output (default false)").Bool()
)

func main() {
	app.Version("0.1.0")
	app.Writer(os.Stdout) // direct help to stdout
	kingpin.MustParse(app.Parse(os.Args[1:]))
	app.Writer(os.Stderr) // ... but ensure errors go to stderr

	config := Config{
		IcingaServer:      *icinga_server,
		IcingaCertfile:    *icinga_certfile,
		IcingaUser:        *icinga_user,
		IcingaPassword:    *icinga_password,
		IcingaQueue:       *icinga_queue,
		IcingaTimeoutMS:   *icinga_timeout,
		IcingaKeepAliveMS: *icinga_keepalive,
		RedisServer:       *redis_server,
		RedisDatabase:     *redis_database,
		FlapjackVersion:   *flapjack_version,
		FlapjackEvents:    *flapjack_events,
		Debug:             *debug,
	}

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

	// shutdown signal handler
	sigs := make(chan os.Signal, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	transport, err := flapjack.Dial(config.RedisServer, config.RedisDatabase)
	if err != nil {
		log.Fatalf("Couldn't establish Redis connection: %s\n", err)
	}

	finished := make(chan error, 1)

	api_client := ApiClient{config: config, redis: transport}
	api_client.Connect(finished)

	select {
	case <-sigs:
		log.Println("Interrupted, cancelling request")
		// TODO determine if request not currently active...
		api_client.Cancel()
	case err := <-finished:
		if config.Debug {
			fmt.Println("Finished with error // ", err)
		}
	}

	// close redis connection
	transport.Close()

	// TODO output some stats on events handled etc.
}
