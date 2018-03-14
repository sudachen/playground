package v1

const DefaultMaxMessageSize = uint32(1024)

type Config struct {
	MaxMessageSize     uint32  `toml:",omitempty"`
}

var DefaultConfig = Config{
	MaxMessageSize:     DefaultMaxMessageSize,
}

