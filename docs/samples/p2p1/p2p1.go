package p2p1

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
)

const (
	broadcastTimeout = 100 * time.Millisecond
)

const (
	statusCode = iota
	messagesCode
	NumberOfMessageCodes
)

const ProtocolName = "p2p1"
const ProtocolVersion = uint64(1)
const ProtocolVersionStr = "1.0"

func protocols(state *known) []p2p.Protocol {
	return []p2p.Protocol{
		p2p.Protocol{
			Name:    ProtocolName,
			Version: uint(ProtocolVersion),
			Length:  NumberOfMessageCodes,
			Run: func(p2 *p2p.Peer, mrw p2p.MsgReadWriter) error {
				return (&peer{p2, mrw, known{}}).loop(state)
			},
		},
	}
}

type peer struct {
	p2   *p2p.Peer
	rw   p2p.MsgReadWriter
	last known
}

func (p *peer) handshake() error {
	ec := make(chan error, 2)
	go func() {
		ec <- p2p.Send(p.rw, statusCode, ProtocolVersion)
	}()

	pkt, err := p.rw.ReadMsg()
	if err != nil {
		return fmt.Errorf("peer [%x]: failed to read status packet: %v", p.p2.ID(), err)
	}
	if pkt.Code != statusCode {
		return fmt.Errorf("peer [%x]: sent packet %x before status packet", p.p2.ID(), pkt.Code)
	}
	var peerVersion uint64
	if err := pkt.Decode(&peerVersion); err != nil {
		return fmt.Errorf("peer [%x]: failed to decode status packet: %v", p.p2.ID(), err)
	}
	if peerVersion != ProtocolVersion {
		return fmt.Errorf("peer [%x]: protocol version mismatch %d != %d", p.p2.ID(), peerVersion, ProtocolVersion)
	}
	if err := <-ec; err != nil {
		return fmt.Errorf("peer [%x]: failed to send status packet: %v", p.p2.ID(), err)
	}

	return nil
}

func (p *peer) broadcast(state *known, quit chan struct{}) {
	t := time.NewTicker(broadcastTimeout)
	for {
		select {
		case <-t.C:
		case <-quit:
			return
		}

		if last, ok := p.last.pass(state.value()); ok {
			if err := p2p.Send(p.rw, messagesCode, last); err != nil {
				log.Warn("failed to send messages", "peer", p.p2.ID, "err", err)
			}
		}
	}
}

func (p *peer) loop(state *known) error {
	if err := p.handshake(); err != nil {
		return err
	}

	quit := make(chan struct{})
	go p.broadcast(state, quit)
	defer close(quit)

	for {
		pkt, err := p.rw.ReadMsg()
		if err != nil {
			if err.Error() != "EOF" {
				log.Warn("message loop", "peer", p.p2.ID(), "err", err)
			}
			return err
		}

		switch pkt.Code {
		case statusCode:
			log.Warn("unexpected status packet received", "peer", p.p2.ID())

		case messagesCode:
			var t uint64

			if err := pkt.Decode(&t); err != nil {
				return fmt.Errorf("invalid messages: %v", err)
			}

			if last, ok := p.last.pass(t); ok {
				state.pass(last)
			}
		}
	}

	return nil
}

type known struct {
	uint64
}

func (k *known) pass(t uint64) (uint64, bool) {
	for {
		last := atomic.LoadUint64(&k.uint64)
		if last < t {
			if atomic.CompareAndSwapUint64(&k.uint64, last, t) {
				return t, true
			}
		} else {
			return last, false
		}
	}
}

func (k *known) value() uint64 {
	return atomic.LoadUint64(&k.uint64)
}
