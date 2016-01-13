/*
Copied from https://github.com/flapjack/flapjack/blob/master/src/flapjack/transport.go
(to ease local debugging/testing)
*/

package flapjack_icinga2

import (
  "encoding/json"
  "github.com/garyburd/redigo/redis"
)

// FlapjackTransport is a representation of a Redis connection.
type FlapjackTransport struct {
  Address    string
  Database   int
  Connection redis.Conn
}

// Dial establishes a connection to Redis, wrapped in a FlapjackTransport.
func FlapjackDial(address string, database int) (FlapjackTransport, error) {
  // Connect to Redis
  conn, err := redis.Dial("tcp", address)
  if err != nil {
    return FlapjackTransport{}, err
  }

  // Switch database
  conn.Do("SELECT", database)

  transport := FlapjackTransport{
    Address:    address,
    Database:   database,
    Connection: conn,
  }
  return transport, nil
}

// Send takes an event and sends it over a transport.
func (t FlapjackTransport) Send(event FlapjackEvent) (interface{}, error) {
  err := event.IsValid()
  if err == nil {
    data, _ := json.Marshal(event)
    reply, err := t.Connection.Do("LPUSH", "events", data)
    if err != nil {
      return nil, err
    }

    return reply, nil
  } else {
    return nil, err
  }
}

func (t FlapjackTransport) Close() (interface{}, error) {
  reply, err := t.Connection.Do("QUIT")
  if err != nil {
    return nil, err
  }

  return reply, nil
}
