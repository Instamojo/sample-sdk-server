package config

import (
	"flag"
	"log"
)

type config struct {
	ProdClientID string

	ProdClientSecret string

	TestClientID string

	TestClientSecret string

	ProdURL string

	TestURL string
}

// Config stores the configs
var Config config

func init() {
	prodClientID := flag.String("production-client-id", "", "Production Client ID")
	prodClientSecret := flag.String("production-client-secret", "", "Production Client Secret")
	testClientID := flag.String("test-client-id", "", "Test Client ID")
	testClientSecret := flag.String("test-client-secret", "", "Test Client Secret")
	flag.Parse()

	if *prodClientID == "" {
		log.Fatal("Production Client ID is missing")
	}

	if *prodClientSecret == "" {
		log.Fatal("Production Client secret is missing")
	}

	if *testClientID == "" {
		log.Fatal("Test Client ID is missing")
	}

	if *testClientSecret == "" {
		log.Fatal("Test Client Secret is missing")
	}

	Config = config{
		ProdClientID:     *prodClientID,
		ProdClientSecret: *prodClientSecret,
		TestClientID:     *testClientID,
		TestClientSecret: *testClientSecret,
		ProdURL:          "https://api.instamojo.com",
		TestURL:          "https://test.instamojo.com",
	}
}
