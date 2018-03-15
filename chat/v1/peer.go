package v1

import (
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/rlp"
)

const batchLength = 10
const broadcastTimeout = 100 * time.Millisecond

type peer struct {
	c    *Chat
	p2   *p2p.Peer
	rw   p2p.MsgReadWriter
	quit chan struct{}
}

func newPeer(c *Chat, p2 *p2p.Peer, rw p2p.MsgReadWriter) *peer {
	p := &peer{
		c,
		p2,
		rw,
		make(chan struct{}),
	}
	return p
}

func (p *peer) ID() []byte {
	id := p.p2.ID()
	return id[:]
}

func (p *peer) start() {
	go p.broadcast()
	p.c.attach(p)
}

func (p *peer) stop() {
	p.c.detach(p)
	close(p.quit)
}

func (p *peer) broadcast() {
	t := time.NewTicker(broadcastTimeout)
	var index uint64
	batch := make([][]byte, batchLength)
	for {
		if done2(p.quit, t.C) {
			return
		}
		now := time.Now().Unix()
		n := 0
		i, m := p.c.get(index)
		for m != nil && n < batchLength {
			if m.deathTime() > now {
				batch[n] = m.body
				n++
			}
			if done(p.quit) {
				return
			}
			i, m = p.c.get(i)
		}

		if n != 0 {
			if err := p2p.Send(p.rw, messagesCode, batch[:n]); err != nil {
				// log?
			} else {
				index = i
			}
		}

	}
}

func (p *peer) handshake() error {
	log.Trace("starting handshake with peer", p.ID())

	ec := make(chan error, 1)
	go func() {
		ec <- p2p.SendItems(p.rw, statusCode, ProtocolVersion)
	}()
	pkt, err := p.rw.ReadMsg()
	if err != nil {
		return err
	}
	if pkt.Code != statusCode {
		return fmt.Errorf("peer [%x] sent packet %x before status packet", p.ID(), pkt.Code)
	}
	s := rlp.NewStream(pkt.Payload, uint64(pkt.Size))
	_, err = s.List()
	if err != nil {
		return fmt.Errorf("peer [%x] sent bad status message: %v", p.ID(), err)
	}
	peerVersion, err := s.Uint()
	if err != nil {
		return fmt.Errorf("peer [%x] sent bad status message (unable to decode version): %v", p.ID(), err)
	}
	if peerVersion != ProtocolVersion {
		return fmt.Errorf("peer [%x]: protocol version mismatch %d != %d", p.ID(), peerVersion, ProtocolVersion)
	}
	if err := <-ec; err != nil {
		return fmt.Errorf("peer [%x] failed to send status packet: %v", p.ID(), err)
	}
	log.Trace("succeeded handshake with pear", p.ID())
	return nil
}

func (p *peer) loop() error {

	if err := p.handshake(); err != nil {
		return nil
	}

	p.start()
	defer p.stop()

	for {
		pkt, err := p.rw.ReadMsg()
		if err != nil {
			log.Warn("message loop", "peer", p.ID(), "err", err)
			return err
		}
		if pkt.Size > p.c.MaxP2pMessageSize() {
			log.Warn("oversized packet received", "peer", p.ID())
			return errors.New("oversized packet received")
		}

		switch pkt.Code {
		case statusCode:
			log.Warn("unxepected status packet received", "peer", p.ID())

		case messagesCode:
			var bs [][]byte

			if err := pkt.Decode(&bs); err != nil {
				log.Warn("failed to decode messages, peer will be disconnected", "peer", p.ID(), "err", err)
				return errors.New("invalid messages")
			}

			for _, b := range bs {
				m := &message{body: b}
				if err := m.validate(); err != nil {
					log.Error("bad message received, peer will be disconnected", "peer", p.ID(), "err", err)
					return errors.New("invalid message")
				}
				p.c.enqueue(m)
			}
		}
	}
}
