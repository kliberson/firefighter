package main

import (
	"firefighter/internal/sniffer"
	"fmt"
)

func main() {
	sniffer.Start("enp0s3")
	fmt.Println("Firefighter is running...")
}
