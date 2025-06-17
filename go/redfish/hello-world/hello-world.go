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

	bmcConfig := gofish.ClientConfig{}
	bmcConfig.Username = BMC_USER
	bmcConfig.Password = BMC_PASSWORD
	bmcConfig.Endpoint = BMC_FQDN
	bmcConfig.Insecure = true

	client, err := gofish.Connect(bmcConfig)
	if err != nil {
		panic(err)
	}
	defer client.Logout()

	service := client.GetService()
	fmt.Printf("%s Redfish Version %s ", bmcConfig.Endpoint, service.RedfishVersion)
	systems, err := service.Systems()
	if err != nil {
		panic(err)
	}
	fmt.Printf("-- %d systems\n found", len(systems))
}
