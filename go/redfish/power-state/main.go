package main

import (
	"fmt"
	"os"

	"github.com/iskylite/nodeset"
	"github.com/stmcginnis/gofish"
)

type SystemState struct {
	URI          string
	Manufacturer string
	PowerState   string
}

func main() {
	BMC_PASSWORD := os.Getenv("BMC_PASSWORD")
	BMC_USER := os.Getenv("BMC_USER")
	BMC_FQDN := os.Getenv("BMC_FQDN")

	nodeset, err := nodeset.Expand(BMC_FQDN)
	if err != nil {
		panic(err)
	}
	for _, node := range nodeset {
		bmcConfig := gofish.ClientConfig{}
		bmcConfig.Username = BMC_USER
		bmcConfig.Password = BMC_PASSWORD
		bmcConfig.Endpoint = "https://" + node
		bmcConfig.Insecure = true

		client, err := gofish.Connect(bmcConfig)
		if err != nil {
			panic(err)
		}
		defer client.Logout()
		service := client.GetService()
		systems, err := service.Systems()
		if err != nil {
			panic(err)
		}
		for _, system := range systems {
			state := SystemState{}
			state.URI = bmcConfig.Endpoint + "/redfish/v1/Systems/" + system.ID
			state.Manufacturer = system.Manufacturer
			state.PowerState = string(system.PowerState)
			fmt.Printf("%s power state %s\n", state.URI, system.PowerState)
		}
	}
}
