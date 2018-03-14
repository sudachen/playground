package v1

const DefaultMaxChatMessageSize int = 1024
const DefaultMaxP2pMessageSize = DefaultMaxChatMessageSize * 10

type Config struct {
	MaxP2pMessageSize  int `toml:",omitempty"`
	MaxChatMessageSize int `toml:",omitempty"`
}

var DefaultConfig = Config{
	MaxP2pMessageSize:  DefaultMaxP2pMessageSize,
	MaxChatMessageSize: DefaultMaxChatMessageSize,
}
