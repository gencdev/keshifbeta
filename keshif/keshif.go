package keshif

import (
	"fmt"
	"os"
	"os/signal"
)

func Start(envoy bool, clear bool) {
	routes := GetRoutes()
	AddHost(routes)

	if envoy {
		GenerateEnvoyConfig(routes)
		return
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	StartProxy(routes, "", "")

	<-stop

	if clear {
		ClearHosts(routes)
		fmt.Println("Vhosts cleared.")
	}
}

func Clear() {
	routes := GetRoutes()
	ClearHosts(routes)
	fmt.Println("Vhosts cleared.")
}
