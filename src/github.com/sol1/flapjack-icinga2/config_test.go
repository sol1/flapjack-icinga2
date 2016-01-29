package main

import "testing"

func TestEmptyConfigFails(t *testing.T) {
	config := Config{}
	errs := config.Errors()

	if len(errs) != 4 {
		t.Error("Expected validation to fail.")
	}
}

func TestIcingaServerIsAddressAndPort(t *testing.T) {
	config := Config{}
	errs := config.Errors()

	var found bool

	for _, e := range errs {
		if e.Error() == "Icinga server should be in format `host:port` (e.g. 127.0.0.1:5665)" {
			found = true
		}
	}

	if !found {
		t.Error("Expected Icinga server address:port to be validated.")
	}
}

func TestRedisServerIsAddressAndPort(t *testing.T) {
	config := Config{}
	errs := config.Errors()

	var found bool

	for _, e := range errs {
		if e.Error() == "Redis server should be in format `host:port` (e.g. 127.0.0.1:6380)" {
			found = true
		}
	}

	if !found {
		t.Error("Expected Redis server address:port to be validated.")
	}
}

func TestIcingaAPIUserIsSet(t *testing.T) {
	config := Config{}
	errs := config.Errors()

	var found bool

	for _, e := range errs {
		if e.Error() == "No Icinga2 API user specified in ICINGA2_API_USER env variable or --user option" {
			found = true
		}
	}

	if !found {
		t.Error("Expected Icinga API user presence to be checked.")
	}
}

func TestIcingaAPIPasswordIsSet(t *testing.T) {
	config := Config{}
	errs := config.Errors()

	var found bool

	for _, e := range errs {
		if e.Error() == "No Icinga2 API password specified in ICINGA2_API_PASSWORD env variable or --password option" {
			found = true
		}
	}

	if !found {
		t.Error("Expected Icinga API password presence to be checked.")
	}
}
