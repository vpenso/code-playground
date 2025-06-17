package main

import (
    "fmt"
    "os"

    "github.com/stmcginnis/gofish"
)

func main() {

	BMC_PASSWORD := os.Getenv("BMC_PASSWORD")
	BMC_USER := os.Getenv("BMC_USER")
	BMC_FQDN := "https://" + os.Getenv("BMC_FQDN")
    fmt.Printf("Connect to: %#v\n\n", BMC_FQDN)

	bmcConfig := gofish.ClientConfig{}
	bmcConfig.Username = BMC_USER
	bmcConfig.Password = BMC_PASSWORD
	bmcConfig.Endpoint = BMC_FQDN
	bmcConfig.Insecure = true

    c, err := gofish.Connect(bmcConfig)
    if err != nil {
        panic(err)
    }

    service := c.Service
    chassis, err := service.Chassis()
    if err != nil {
        panic(err)
    }

    for _, chass := range chassis {
        fmt.Printf("Chassis: %#v\n\n", chass)
    }
}
