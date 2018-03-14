package v1

type Message struct {
	Room       string    `json:"room"`
	Nickname   string    `json:"nickname"`
	PublicKey  []byte    `json:"pubKey,omitempty"`
	Sig        string    `json:"sig"`
	TTL        uint32    `json:"ttl,omitempty"`
	Text       string    `json:"text"`
	Peer 	   string    `json:"peer,omitempty"`
	Timestamp  uint32    `json:"timestamp,omitempty"`
}

