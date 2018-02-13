package libeth

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

type Log struct {
	Address Address
	Topics  []Hash
	Data    []byte
}

type LogForStorage Log
type Logs []*Log

func (log *Log) String() string {
	topics := make([]string, len(log.Topics))
	for i, t := range log.Topics {
		topics[i] = t.Hex()
	}
	return fmt.Sprintf("Log{Address:%s, Topics:%s, Data:%s}",
		log.Address.Hex(),
		strings.Join(topics, ","),
		common.Bytes2Hex(log.Data))
}

func (log *Log) Clone() *Log {
	topics := make([]Hash, len(log.Topics))
	copy(topics, log.Topics)
	data := make([]byte, len(log.Data))
	copy(data, log.Data)
	return &Log{log.Address, topics, data}
}
