package config

import "os"

const Name = "vault"
const Version = "0.1.0"

func FormattedNameWithVersion() string {
	return Name + ":v" + Version
}

var Env = os.Getenv("ENV")
var Port = os.Getenv("PORT")

func IsProduction() bool {
	return Env == "PROD"
}
