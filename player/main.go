package main

import "github.com/reinodovo/hanabi-bot/lib/api"

func main() {
	cred := api.Credentials{
		User: "reinodovopp6969",
		Pass: "Max",
	}

	api.Connect(cred, api.Table{ Id: 16297 })
}
