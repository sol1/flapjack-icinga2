# flapjack-icinga2

[![Build Status](https://travis-ci.org/sol1/flapjack-icinga2.png)](https://travis-ci.org/sol1/flapjack-icinga2)

A client for Icinga 2's [Event Streams API](http://docs.icinga.org/icinga2/latest/doc/module/icinga2/chapter/icinga2-api#icinga2-api-event-streams) feature (first added in Icinga v2.4.0), which connects to Redis and places events on Flapjack's events queue.

This is loosely based on Flapjack's [httpbroker](https://github.com/flapjack/flapjack/blob/master/libexec/httpbroker.go), although that acts as a HTTP server running to receive callbacks while Icinga 2 expects a long-polling HTTP client to receive event data streamed over a long-lived HTTP 1.1 connection.

It only triggers on 'CheckResult' and 'StateChange' event types; Flapjack is built to expect a regular heartbeat of events, and display the changed result outputs. The other events aren't really useful in a distributed notification environment, as Flapjack handles downtime, notification etc.

## INSTALLATION

This API client is written in Go; you'll need Go 1.5+ installed on the machine you are building on. The `build.sh` [script](https://github.com/sol1/flapjack-icinga2/blob/master/build.sh) compiles a standalone binary (in `bin/flapjack-icinga2`) which can be run on other machines without external dependencies.

You'll also need to [set up API access for Icinga 2](http://docs.icinga.org/icinga2/latest/doc/module/icinga2/chapter/icinga2-api#icinga2-api-setup) if you haven't already. The API user you will be using for this client will need to have (at a minimum) permission to access the correct event types:

```
/*
 /etc/icinga2/conf.d/api-users.conf
*/

object ApiUser "username" {
  password = "password"
  permissions = [
    { permission = "events/checkresult" },
    { permission = "events/statechange" }
  ]
}
```

These permissions may be filtered if you want to limit the full range of events that would otherwise be received.

## USAGE

```
usage: flapjack-icinga2 [<flags>]

Transfers Icinga 2 events to Flapjack

Flags:
  --help                         Show context-sensitive help (also try
                                 --help-long and --help-man).
  --icinga-url="localhost:5665"  Icinga 2 API endpoint to connect to
  --icinga-certfile=ICINGA-CERTFILE
                                 Path to Icinga 2 API TLS certfile
  --icinga-user=ICINGA-USER      Icinga 2 basic auth user (required, also checks
                                 ICINGA2_API_USER env var)
  --icinga-password=ICINGA-PASSWORD
                                 Icinga 2 basic auth password (required, also
                                 checks ICINGA2_API_PASSWORD env var)
  --icinga-queue="flapjack"      Icinga 2 event queue name to use
  --icinga-timeout=30000         Icinga 2 API connection timeout, in
                                 milliseconds
  --icinga-keepalive=30000       Icinga 2 API frequency of keepalive traffic, in
                                 milliseconds
  --redis-url="localhost:6380"   Redis server to connect to
  --redis-db=0                   Redis database to connect to
  --flapjack-version=1           Flapjack version being delivered to (1 or 2)
  --flapjack-events="events"     Flapjack event queue name to use
  --debug                        Enable verbose output (default false)
  --version                      Show application version.
```

## HISTORY

#### 0.2.0 ????-??-??

* Bugfixes

#### 0.1.0 2016-01-20

* Initial release

## TODO

* More tests.

## CREDITS

Written by [Ali Graham](https://github.com/ali-graham). Development sponsored by [Solutions First](http://sol1.com.au/).
