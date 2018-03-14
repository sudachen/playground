package v1

import (
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/ethereum/go-ethereum/log"
	"runtime"
	"time"
	"container/list"
	"github.com/ethereumproject/go-ethereum/whisper"
	"errors"
)

const (
	NumberOfMessageCodes = 128
)

const (
	ProtocolVersion    = uint64(1) // Protocol version number
	ProtocolVersionStr = "1.0"     // The same, as a string
	ProtocolName       = "cht"     // Nickname of the protocol in geth
	updateClockTimeout = time.Second
	messageQueueLimit  = 1024
)

type Watcher interface {
	Watch(*Message)
}

type Chat struct {
	protocol p2p.Protocol
	rooms map[string][]Watcher
	queue chan *Message
	quit chan struct{}
}

func New(cfg *Config) *Chat {
	if cfg == nil {
		cfg = &DefaultConfig
	}

	c := &Chat{
		queue:  make(chan *Message, messageQueueLimit),
		quit:   make(chan struct{}),
		rooms:  make(map[string][]Watcher),
	}

	c.protocol = p2p.Protocol{
		Name:     ProtocolName,
		Version:  uint(ProtocolVersion),
		Length:   NumberOfMessageCodes,
		Run:      c.handlePeer,
		NodeInfo: func() interface{} {
			return map[string]interface{}{
				"version":        ProtocolVersionStr,
				"maxMessageSize": cfg.MaxMessageSize,
			}
		},
	}

	return c
}

func (c *Chat) Protocols() []p2p.Protocol {
	return []p2p.Protocol{c.protocol}
}

func (c *Chat) APIs() []rpc.API {
	return []rpc.API{
		{
			Namespace: ProtocolName,
			Version:   ProtocolVersionStr,
			Service:   NewChatAPI(c),
			Public:    true,
		},
	}
}

func (c *Chat) clock() {
	// do delayed actions here
}

func (c *Chat) update() {
	clock := time.NewTicker(updateClockTimeout)
	for {
		select {
		case <-clock.C:
			c.clock()

		case <-c.quit:
			return
		}
	}
}

func (c *Chat) dequeue() {
	for {
		select {
		case <-c.quit:
			return

		case m := <-c.queue:
			if r, ok := c.rooms[m.Room]; ok {
				for _, w := range r {
					w.Watch(m)
				}
			}
		}
	}
}

func (c *Chat) handlePeer(peer *p2p.Peer, rw p2p.MsgReadWriter) error {

	// handle peer messages here

	return nil
}

func (c *Chat) Start(server *p2p.Server) error {
	log.Info("started whisper v." + ProtocolVersionStr)
	go c.update()
	go c.dequeue()
	return nil
}

func (c *Chat) Stop() error {
	close(c.quit)
	return nil
}

var AlreadySubscribedError = errors.New("already subscribed")
var NotSubscribedError = errors.New("not subscribed")

func (c *Chat) Subscribe(room string, w Watcher) error {
	if ws, ok := c.rooms[room]; ok {
		for _, x := range ws {
			if w == x {
				return AlreadySubscribedError
			}
		}
		ws = append(ws,w)
	} else {
		ws = []Watcher{w}
		c.rooms[room] = ws
	}
	return nil
}

func (c *Chat) Unsubscribe(room string, w Watcher) error {
	if ws, ok := c.rooms[room]; ok {
		for i, x := range ws {
			if w == x {
				L := len(ws) - 1
				if L > 0 && i != L {
					ws[i] = ws[L]
				}
				c.rooms[room] = ws[:L]
				return nil
			}
		}
	}
	return NotSubscribedError
}