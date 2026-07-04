package main

import (
	"context"
	"fmt"
	"log"

	"github.com/jorgesanchez-e/localenvironment/apps/ddns/internal/infra/ipgetter"
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

	fmt.Println("Simple DDNS config:", simpleDDNS)

	ipGetter := ipgetter.NewIPGetter()
	ipv4, err := ipGetter.GetIPV4(context.Background())
	if err != nil {
		log.Printf("ipgetter error %v", err)
	}
	ipv6, err := ipGetter.GetIPV6(context.Background())
	if err != nil {
		log.Printf("ipgetter error %v", err)
	}

	fmt.Println("IPv4:", ipv4)
	fmt.Println("IPv6:", ipv6)
}
