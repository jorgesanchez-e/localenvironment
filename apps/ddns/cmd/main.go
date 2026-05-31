package main

import (
	"fmt"
	"log"

	"github.com/jorgesanchez-e/localenvironment/config"
)

type ddns struct {
	Storage string `yaml:"storage"`
}

func main() {
	var err error

	cnf, err := config.New()
	if err != nil {
		log.Fatalf("Failed to create config: %v", err)
	}

	ddns := &ddns{}
	err = cnf.Decode("ddns", ddns)
	if err != nil {
		log.Fatalf("Failed to decode config: %v", err)
	}
	fmt.Println(ddns)
}
