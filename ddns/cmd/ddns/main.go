package main

import (
	"context"
	"fmt"

	"github.com/jorgesanchez-e/localenvironment/ddns/internal/adapters/publicip"
)

func main() {
	ctx := context.Background()

	ddnsClient := publicip.New()
	ipv4 := ddnsClient.IPv4(ctx)
	ipv6 := ddnsClient.IPv6(ctx)

	if ipv4 != nil {
		fmt.Printf("ipv4:%s\n", ipv4.String())
	}

	if ipv6 != nil {
		fmt.Printf("ipv6:%s\n", ipv6.String())
	}
}
