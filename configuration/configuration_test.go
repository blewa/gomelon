// Copyright 2015 Quoc-Viet Nguyen. All rights reserved.
// This software may be modified and distributed under the terms
// of the BSD license. See the LICENSE file for details.

package configuration

import (
	"testing"

	"github.com/goburrow/gomelon"
)

func TestLoadJSON(t *testing.T) {
	bootstrap := gomelon.Bootstrap{
		Arguments: []string{"server", "configuration_test.json"},
	}
	testFactory(t, &bootstrap)
}

func testFactory(t *testing.T, bootstrap *gomelon.Bootstrap) {
	factory := Factory{}
	config, err := factory.BuildConfiguration(bootstrap)
	if err != nil {
		t.Fatal(err)
	}
	appConnector1 := gomelon.ConnectorConfiguration{
		Type: "http",
		Addr: ":8080",
	}
	appConnector2 := gomelon.ConnectorConfiguration{
		Type:     "https",
		Addr:     ":8048",
		CertFile: "/tmp/cert",
		KeyFile:  "/tmp/key",
	}
	if len(config.Server.ApplicationConnectors) != 2 ||
		config.Server.ApplicationConnectors[0] != appConnector1 ||
		config.Server.ApplicationConnectors[1] != appConnector2 {
		t.Fatalf("Invalid ApplicationConnectors: %+v", config.Server.ApplicationConnectors)
	}
	adminConnector1 := gomelon.ConnectorConfiguration{
		Type: "http",
		Addr: ":8081",
	}
	if len(config.Server.AdminConnectors) != 1 ||
		config.Server.AdminConnectors[0] != adminConnector1 {
		t.Fatalf("Invalid AdminConnectors: %+v", config.Server.AdminConnectors)
	}
	if config.Logging.Level != "INFO" ||
		config.Logging.Loggers["gomelon.server"] != "DEBUG" ||
		config.Logging.Loggers["gomelon.configuration"] != "WARN" {
		t.Fatalf("Invalid Logging: %+v", config.Logging)
	}
	if config.Metrics.Frequency != "1s" {
		t.Fatalf("Invalid Metrics: %+v", config.Metrics)
	}
}