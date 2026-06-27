package config

import "os"

const Name = "vault"
const Version = "1.4.1"

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

var SentinelURL = os.Getenv("SENTINEL_URL")
var SentinelClientID = os.Getenv("SENTINEL_CLIENT_ID")
var SentinelClientSecret = os.Getenv("SENTINEL_CLIENT_SECRET")
var SentinelSAToken = os.Getenv("SENTINEL_SA_TOKEN")
var SentinelRedirectURI = os.Getenv("SENTINEL_REDIRECT_URI")

var GitHubActionsOIDCIssuer = os.Getenv("GITHUB_ACTIONS_OIDC_ISSUER")
var GitHubActionsOIDCAudience = os.Getenv("GITHUB_ACTIONS_OIDC_AUDIENCE")

var VaultMasterKey = os.Getenv("VAULT_MASTER_KEY")
var VaultMasterKeyID = os.Getenv("VAULT_MASTER_KEY_ID")

func IsProduction() bool {
	return Env == "PROD"
}
