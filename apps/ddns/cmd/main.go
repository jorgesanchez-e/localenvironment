package main

import (
	"fmt"
	"log"

	"github.com/jorgesanchez-e/localenvironment/config"
)

func main() {
	cnf, err := config.New()
	if err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	simpleDDNS, err := cnf.GetSimpleDDNSConfig()
	if err != nil {
		log.Fatalf("Failed to get simple DDNS config: %v", err)
	}

	fmt.Println(simpleDDNS)
}
