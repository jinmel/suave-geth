package suave

type Config struct {
	SuaveEthRemoteBackendEndpoint string // deprecated
	RedisStorePubsubUri           string
	RedisStoreUri                 string
	PebbleDbPath                  string
	EthBundleSigningKeyHex        string
	EthBlockSigningKeyHex         string
	ExternalWhitelist             []string
	AliasRegistry                 map[string]string
	LocalRelayListenAddress       string // OP PoC
}

var DefaultConfig = Config{}
