# flapjack-icinga2

[![Build Status](https://travis-ci.org/sol1/flapjack-icinga2.png)](https://travis-ci.org/sol1/flapjack-icinga2)

** Work in Progress **

A client for Icinga 2's Event Streams API feature (first added in Icinga 2.4.0), which connects to Redis and places events on Flapjack's events queue.

This is loosely based on Flapjack's httpbroker, although that acts as a HTTP server running to receive callbacks while Icinga 2 expects a long-polling HTTP client to receive event data streamed over a long-lived HTTP 1.1 connection.