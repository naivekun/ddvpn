package main

import (
	"ddvpn/client"
	"ddvpn/common"
	"ddvpn/conf"
	"ddvpn/server"
	"flag"
	"fmt"
	"log"
)

var configFileName = "config.json"

func main() {
	fmt.Println(`
 _______   _______  ____    ____ .______   .__   __. 
|       \ |       \ \   \  /   / |   _  \  |  \ |  | 
|  .--.  ||  .--.  | \   \/   /  |  |_)  | |   \|  | 
|  |  |  ||  |  |  |  \      /   |   ___/  |  . '  | 
|  '--'  ||  '--'  |   \    /    |  |      |  |\   | 
|_______/ |_______/     \__/     | _|      |__| \__| 

@naivekun`)

	if !common.FileExists(configFileName) {
		log.Println("create default config to " + configFileName)
		conf.CreateNewConfigFile(configFileName)
		return
	}

	config := &conf.Config{}
	config.ReadConfigFile(configFileName)
	if config.Mode == "client" {
		c, err := client.Config(*config)
		common.Must(err)
		common.Must(client.Run(c))
	} else if config.Mode == "server" {
		s, err := server.Config(*config)
		common.Must(err)
		common.Must(server.Run(s))
	} else {
		log.Fatalln("invalid mode: " + config.Mode)
	}
}

func init() {
	flag.StringVar(&configFileName, "config", "config.json", "specify config file")
	flag.Parse()
}
