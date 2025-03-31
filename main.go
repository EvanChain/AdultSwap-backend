package main

import (
	"awesomeProject2/config"
	"awesomeProject2/swapevent"
)

func init() {
	config.SetupConfig()
}
func main() {
	swapevent.GetSwapEventInfo()
}
