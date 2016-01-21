# flapjack-icinga2

[![Build Status](https://travis-ci.org/sol1/flapjack-icinga2.png)](https://travis-ci.org/sol1/flapjack-icinga2)

A client for Icinga 2's [Event Streams API](http://docs.icinga.org/icinga2/latest/doc/module/icinga2/chapter/icinga2-api#icinga2-api-event-streams) feature (first added in Icinga v2.4.0), which connects to Redis and places events on Flapjack's events queue.

This is loosely based on Flapjack's [httpbroker](https://github.com/flapjack/flapjack/blob/master/libexec/httpbroker.go), although that acts as a HTTP server running to receive callbacks while Icinga 2 expects a long-polling HTTP client to receive event data streamed over a long-lived HTTP 1.1 connection.

It only triggers on 'CheckResult' event types; 'StateChanged' ones are a subset of these, and Flapjack is built to expect a regular heartbeat of events, and display the changed result summaries. The other events aren't really useful in a distributed notification environment, as Flapjack handles downtime, notification etc.

## INSTALLATION

This API client is written in Go; you'll need Go 1.5 installed on the machine you are building on. The `build.sh` [script](https://github.com/sol1/flapjack-icinga2/blob/master/build.sh) compiles a standalone binary (in `bin/flapjack-icinga2`) which can be run on other machines without external dependencies.

You'll also need to [set up API access for Icinga 2](http://docs.icinga.org/icinga2/latest/doc/module/icinga2/chapter/icinga2-api#icinga2-api-setup) if you haven't already.

## USAGE

```
$ bin/flapjack-icinga2 --help
usage: flapjack-icinga2 [<flags>]

Transfers Icinga 2 events to Flapjack

Flags:
  --help                         Show context-sensitive help (also try
                                 --help-long and --help-man).
  --icinga-url="localhost:5665"  Icinga 2 API endpoint to connect to (default
                                 localhost:5665)
  --icinga-certfile=ICINGA-CERTFILE
                                 Path to Icinga 2 API TLS certfile
  --icinga-user=ICINGA-USER      Icinga 2 basic auth user (required, also checks
                                 ICINGA2_API_USER env var)
  --icinga-password=ICINGA-PASSWORD
                                 Icinga 2 basic auth password (required, also
                                 checks ICINGA2_API_PASSWORD env var)
  --icinga-queue="flapjack"      Icinga 2 event queue name to use (default
                                 flapjack)
  --icinga-timeout=60000         Icinga 2 API connection timeout, in
                                 milliseconds (default 60_000)
  --redis-url="localhost:6380"   Redis server to connect to (default
                                 localhost:6380)
  --redis-db=REDIS-DB            Redis database to connect to (default 0)
  --flapjack-version=1           Flapjack version being delivered to (default 1)
  --flapjack-events="events"     Flapjack event queue name to use (default
                                 events)
  --debug                        Enable verbose output (default false)
  --version                      Show application version.
```

## HISTORY

#### 0.1.0 2016-01-20

* Initial release

## TODO

* More tests.

## CREDITS

Written by [Ali Graham](https://github.com/ali-graham). Development sponsored by [Solutions First](http://sol1.com.au/).
