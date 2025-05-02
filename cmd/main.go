package main

import (
	"github.com/mariajdab/flight-price/config"
	"log"
)

func main() {
	_, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config variables: ", err)
	}

}
