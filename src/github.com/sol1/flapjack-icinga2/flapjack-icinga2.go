package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"github.com/sol1/flapjack-icinga2/flapjack"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

var (
	app = kingpin.New("flapjack-icinga2", "Transfers Icinga 2 events to Flapjack")

	icinga_env_api_user     = os.Getenv("ICINGA2_API_USER")
	icinga_env_api_password = os.Getenv("ICINGA2_API_PASSWORD")

	icinga_server   = app.Flag("icinga", "Icinga 2 API endpoint to connect to (default localhost:5665)").Default("localhost:5665").String()
	icinga_certfile = app.Flag("certfile", "Path to Icinga 2 API TLS certfile").String()
	icinga_user     = app.Flag("user", "Icinga 2 basic auth user (required, also checks ICINGA2_API_USER env var)").Default(icinga_env_api_user).String()
	icinga_password = app.Flag("password", "Icinga 2 basic auth password (required, also checks ICINGA2_API_PASSWORD env var)").Default(icinga_env_api_password).String()
	icinga_queue    = app.Flag("queue", "Icinga 2 event queue name to use (default flapjack)").Default("flapjack").String()
	icinga_timeout  = app.Flag("timeout", "Icinga 2 API connection timeout, in milliseconds (default 60_000)").Default("60000").Int()

	// default Redis port is 6380 rather than 6379 as the Flapjack packages ship
	// with an Omnibus-packaged Redis running on a different port to the
	// distro-packaged one
	redis_server   = app.Flag("redis", "Redis server to connect to (default localhost:6380)").Default("localhost:6380").String()
	redis_database = app.Flag("db", "Redis database to connect to (default 0)").Int()

	debug = app.Flag("debug", "Enable verbose output (default false)").Bool()
)

type Config struct {
	IcingaServer    string
	IcingaCertfile  string
	IcingaUser      string
	IcingaPassword  string
	IcingaQueue     string
	IcingaTimeoutMS int
	RedisServer     string
	RedisDatabase   int
	Debug           bool
}

func main() {
	app.Version("0.1.0")
	app.Writer(os.Stdout) // direct help to stdout
	kingpin.MustParse(app.Parse(os.Args[1:]))
	app.Writer(os.Stderr) // ... but ensure errors go to stderr

	icinga_addr := strings.Split(*icinga_server, ":")
	if len(icinga_addr) != 2 {
		log.Printf("Error: invalid icinga_server specified: %s\n", *icinga_server)
		log.Println("Should be in format `host:port` (e.g. 127.0.0.1:5665)")
		os.Exit(1)
	}

	redis_addr := strings.Split(*redis_server, ":")
	if len(redis_addr) != 2 {
		log.Printf("Error: invalid redis_server specified: %s\n", *redis_server)
		log.Println("Should be in format `host:port` (e.g. 127.0.0.1:6380)")
		os.Exit(1)
	}

	if *icinga_user == "" {
		log.Println("No Icinga2 API user specified in ICINGA2_API_USER env variable or --user option")
		os.Exit(1)
	}

	if *icinga_password == "" {
		log.Println("No Icinga2 API password specified in ICINGA2_API_PASSWORD env variable or --password option")
		os.Exit(1)
	}

	config := Config{
		IcingaServer:    *icinga_server,
		IcingaCertfile:  *icinga_certfile,
		IcingaUser:      *icinga_user,
		IcingaPassword:  *icinga_password,
		IcingaQueue:     *icinga_queue,
		IcingaTimeoutMS: *icinga_timeout,
		RedisServer:     *redis_server,
		RedisDatabase:   *redis_database,
		Debug:           *debug,
	}

	if config.Debug {
		log.Printf("Starting with config: %+v\n", config)
	}

	// shutdown signal handler
	sigs := make(chan os.Signal, 1)
	done := false

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	icinga_url_parts := []string{
		"https://", config.IcingaServer, "/v1/events?queue=", config.IcingaQueue,
		"&types=CheckResult", // &types=StateChange&types=CommentAdded&types=CommentRemoved",
	}
	var icinga_url bytes.Buffer
	for i := range icinga_url_parts {
		icinga_url.WriteString(icinga_url_parts[i])
	}

	transport, err := flapjack.Dial(config.RedisServer, config.RedisDatabase)
	if err != nil {
		log.Fatalf("Couldn't establish Redis connection: %s\n", err)
	}

	var tls_config *tls.Config

	if config.IcingaCertfile != "" {
		// assuming self-signed server cert -- /etc/icinga2/ca.crt
		// TODO check behaviour for using system cert store (valid public cert)
		CA_Pool := x509.NewCertPool()
		serverCert, err := ioutil.ReadFile(config.IcingaCertfile)
		if err != nil {
			log.Fatalln("Could not load server certificate")
		}
		CA_Pool.AppendCertsFromPEM(serverCert)

		tls_config = &tls.Config{RootCAs: CA_Pool}
	}

	req, _ := http.NewRequest("POST", icinga_url.String(), nil)
	req.Header.Add("Accept", "application/json")
	req.SetBasicAuth(config.IcingaUser, config.IcingaPassword)
	var tr *http.Transport
	if tls_config == nil {
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		} // TODO settings from DefaultTransport
		log.Println("Skipping verification of server TLS certificate")
	} else {
		tr = &http.Transport{
			TLSClientConfig: tls_config,
		} // TODO settings from DefaultTransport
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Duration(config.IcingaTimeoutMS) * time.Millisecond,
	}
	finished := make(chan error, 1)

	go func() {
		for done == false {
			resp, err := client.Do(req)
			if config.Debug {
				fmt.Println("post-req err", err)
			}
			if err == nil {
				if config.Debug {
					fmt.Printf("URL: %+v\n", icinga_url.String())
					fmt.Printf("Response: %+v\n", resp.Status)
				}
				err = processResponse(config, resp, transport)
				if config.Debug {
					fmt.Println("post-process err", err)
				}
			}

			if err != nil {
				if config.Debug {
					fmt.Println("finishing, found err", err)
				}
				finished <- err
				done = true
			}
		}
	}()

	select {
	case <-sigs:
		log.Println("Interrupted, cancelling request")
		// TODO determine if request not currently active...
		tr.CancelRequest(req)
	case err := <-finished:
		if config.Debug {
			fmt.Println("Finished with error", err)
		}
	}

	// close redis connection
	transport.Close()

	// TODO output some stats on events handled etc.
}

func processResponse(config Config, resp *http.Response, transport flapjack.Transport) error {
	defer func() {
		// this makes sure that the HTTP connection will be re-used properly -- exhaust
		// stream and close the handle
		io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	decoder := json.NewDecoder(resp.Body)
	var data interface{}
	err := decoder.Decode(&data)

	if err != nil {
		return err
	}

	m := data.(map[string]interface{})

	if config.Debug {
		fmt.Printf("Decoded Response: %+v\n", data)
	}

	switch m["type"] {
	case "CheckResult":
		check_result := m["check_result"].(map[string]interface{})
		timestamp := m["timestamp"].(float64)

		// https://github.com/Icinga/icinga2/blob/master/lib/icinga/checkresult.ti#L37-L48
		var state string
		switch check_result["state"].(float64) {
		case 0.0:
			state = "ok"
		case 1.0:
			state = "warning"
		case 2.0:
			state = "critical"
		case 3.0:
			state = "unknown"
		default:
			return fmt.Errorf("Unknown check result state %f", check_result["state"].(float64))
		}

		// build and submit Flapjack redis event

		var service string
		if serv, ok := m["service"]; ok {
			service = serv.(string)
		} else {
			service = "HOST"
		}

		event := flapjack.Event{
			Entity:  m["host"].(string),
			Check:   service,
			Type:    "service",
			Time:    int64(timestamp),
			State:   state,
			Summary: check_result["output"].(string),
		}

		_, err := transport.Send(event)
		return err
	default:
		return fmt.Errorf("Unknown type %s", m["type"])
	}
}
