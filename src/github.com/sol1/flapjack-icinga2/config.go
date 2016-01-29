package main

import (
	"errors"
	"strings"
)

type Config struct {
	IcingaServer      string
	IcingaCertfile    string
	IcingaUser        string
	IcingaPassword    string
	IcingaQueue       string
	IcingaTimeoutMS   int
	IcingaKeepAliveMS int
	RedisServer       string
	RedisDatabase     int
	FlapjackVersion   int
	FlapjackEvents    string
	Debug             bool
}

func (c Config) Errors() []error {
	var errs []error

	icinga_addr := strings.Split(c.IcingaServer, ":")
	if len(icinga_addr) != 2 {
		errs = append(errs, errors.New("Icinga server should be in format `host:port` (e.g. 127.0.0.1:5665)"))
	}

	redis_addr := strings.Split(c.RedisServer, ":")
	if len(redis_addr) != 2 {
		errs = append(errs, errors.New("Redis server should be in format `host:port` (e.g. 127.0.0.1:6380)"))
	}

	if c.IcingaUser == "" {
		errs = append(errs, errors.New("No Icinga2 API user specified in ICINGA2_API_USER env variable or --user option"))
	}

	if c.IcingaPassword == "" {
		errs = append(errs, errors.New("No Icinga2 API password specified in ICINGA2_API_PASSWORD env variable or --password option"))
	}

	return errs
}
