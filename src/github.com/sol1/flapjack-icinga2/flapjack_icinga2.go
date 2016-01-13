package flapjack_icinga2

// TODO clean up, split into multiple files

// TODO tests

// NB: all completely WIP, not running yet

import (
  "bytes"
  "encoding/json"
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
  "syscall"
)

var (
	app = kingpin.New("flapjack-icinga2", "")

	icinga_server = app.Flag("icinga", "Icinga 2 API endpoint to connect to (default localhost:5665)").Default("localhost:5665").String()
	icinga_queue  = app.Flag("queue", "Icinga 2 event queue name to use (default flapjack)").Default("flapjack").String()

	// default Redis port is 6380 rather than 6379 as the Flapjack packages ship
	// with an Omnibus-packaged Redis running on a different port to the
	// distro-packaged one
	redis_server   = app.Flag("redis", "Redis server to connect to (default localhost:6380)").Default("localhost:6380").String()
	redis_database = app.Flag("db", "Redis database to connect to (default 0)").Int()

	debug = app.Flag("debug", "Enable verbose output (default false)").Bool()
)

type Config struct {
	IcingaServer  string
	IcingaQueue   string
	RedisServer   string
	RedisDatabase int
	Debug         bool
}

func main() {
	app.Version("0.0.1")
	app.Writer(os.Stdout) // direct help to stdout
	kingpin.MustParse(app.Parse(os.Args[1:]))
	app.Writer(os.Stderr) // ... but ensure errors go to stderr

	icinga_addr := strings.Split(*icinga_server, ":")
	if len(icinga_addr) != 2 {
		fmt.Println("Error: invalid icinga_server specified:", *icinga_server)
		fmt.Println("Should be in format `host:port` (e.g. 127.0.0.1:5665)")
		os.Exit(1)
	}

	redis_addr := strings.Split(*redis_server, ":")
	if len(redis_addr) != 2 {
		fmt.Println("Error: invalid redis_server specified:", *redis_server)
		fmt.Println("Should be in format `host:port` (e.g. 127.0.0.1:6380)")
		os.Exit(1)
	}

	config := Config{
		IcingaServer:  *icinga_server,
		IcingaQueue:   *icinga_queue,
		RedisServer:   *redis_server,
		RedisDatabase: *redis_database,
		Debug:         *debug,
	}

	if config.Debug {
		log.Printf("Booting with config: %+v\n", config)
	}

	// shutdown signal handler
	sigs := make(chan os.Signal, 1)
	done := false

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

  icinga_url_parts := []string{
    "http://", config.IcingaServer, "events?queue=", config.IcingaQueue,
    "&types=CheckResult", // &types=StateChange&types=CommentAdded&types=CommentRemoved",
  }
  var icinga_url bytes.Buffer
  for i := range icinga_url_parts {
    icinga_url.WriteString(icinga_url_parts[i])
  }

  transport, err := FlapjackDial(config.RedisServer, config.RedisDatabase)
  if err != nil {
    fmt.Println("Couldn't establish Redis connection: %s", err)
    os.Exit(1)
  }

	req, _ := http.NewRequest("GET", icinga_url.String(), nil)
	tr := &http.Transport{} // TODO settings from DefaultTransport
	client := &http.Client{Transport: tr}
	c := make(chan error, 1)

	for done == false {

		go func() {
			resp, h_err := client.Do(req)

			if h_err == nil {
				defer resp.Body.Close()

        decoder := json.NewDecoder(resp.Body)
        var data interface{}
        json_err := decoder.Decode(&data)

        if json_err != nil {
          fmt.Printf("%T\n%s\n%#v\n", err, err, err)
        } else {
          m := data.(map[string]interface{})

          switch m["type"] {
            case "CheckResult":
              check_result := m["check_result"].(map[string]interface{})
              vars_before  := check_result["vars_before"].(map[string]interface{})
              vars_after   := check_result["vars_after"].(map[string]interface{})

              timestamp    := m["timestamp"].(float64)

              // TODO determine Flapjack state from changes in vars_before/vars_after
              _ = vars_before
              _ = vars_after

              // build and submit Flapjack redis event
              event := FlapjackEvent{
                Entity:  m["host"].(string),
                Check:   m["service"].(string),
                Time:    int64(timestamp),
                // State:   "ok",
                Summary: check_result["output"].(string),
              }

              _ = event

              // _, t_err = transport.Send(event)
            default:
              fmt.Println(m["type"], "is of a type I don't know how to handle")
          }
			 }
      }

			c <- h_err
		}()

		select {
		case <-sigs:
			log.Println("Cancelling request")
			tr.CancelRequest(req)
			done = true
		case err := <-c:
      _ = err
			// log.Println("Client finished:", err)
		}
	}

  // close redis connection
  transport.Close()
}
