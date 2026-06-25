package config

import "os"

const Name = "vault"
const Version = "0.1.0"

func FormattedNameWithVersion() string {
	return Name + ":v" + Version
}

var Env = os.Getenv("ENV")
var Port = os.Getenv("PORT")

var DatabaseHost = os.Getenv("DATABASE_HOST")
var DatabasePort = os.Getenv("DATABASE_PORT")
var DatabaseUser = os.Getenv("DATABASE_USER")
var DatabasePassword = os.Getenv("DATABASE_PASSWORD")
var DatabaseName = os.Getenv("DATABASE_NAME")

var VaultMasterKey = os.Getenv("VAULT_MASTER_KEY")
var VaultMasterKeyID = os.Getenv("VAULT_MASTER_KEY_ID")

func IsProduction() bool {
	return Env == "PROD"
}
