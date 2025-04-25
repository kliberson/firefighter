package main

import (
	"firefighter/internal/sniffer"
	"log"
)

func main() {
	sniffer.Start("enp0s3")
	log.Println("Firefighter is running...")
}
