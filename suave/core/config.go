package suave

type Config struct {
	SuaveEthRemoteBackendEndpoint string // deprecated
	RedisStorePubsubUri           string
	RedisStoreUri                 string
	PebbleDbPath                  string
	EthBundleSigningKeyHex        string
	EthBlockSigningKeyHex         string
	ExternalWhitelist             []string
	DnsRegistry                   map[string]string
	LocalRelayListenAddress       string // OP PoC
}

var DefaultConfig = Config{}
