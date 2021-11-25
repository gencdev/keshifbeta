package keshif

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Route struct {
	Vhost string      `json:"vhost"`
	Ip    string      `json:"ip"`
	Port  json.Number `json:"port"`
}

const routeConfigFilePath = "/.keshif/routes.json"

func exists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func create(pth string, dirname string) bool {
	err := os.MkdirAll(dirname, os.ModePerm)

	if err != nil {
		panic(fmt.Sprintf("Unable to create directory: %v", err))
	}

	err = ioutil.WriteFile(pth, []byte("{}"), 0755)
	if err != nil {
		panic(fmt.Sprintf("Unable to create route config: %v", err))
	}

	return exists(routeConfigFilePath)
}

func GetRoutes() map[string]Route {
	dirname, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	pth := dirname + routeConfigFilePath
	if !exists(pth) {
		create(pth, dirname + "/.keshif")
	}

	plan, _ := ioutil.ReadFile(pth)
	routes := make(map[string]Route)
	err = json.Unmarshal(plan, &routes)

	if err != nil {
		panic(err)
	}

	return routes
}
