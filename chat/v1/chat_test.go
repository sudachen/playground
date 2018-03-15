package v1

import (
	"fmt"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/sudachen/misc/out"
	_ "github.com/sudachen/playground/log/ethereum"
)

const chatNodesCount = 6
const chatMesgsCount = 2

func init() {
	out.Warn.SetCurrent()
}

func TestPropagation(t *testing.T) {
	ns := initialize(chatNodesCount, t)
	out.Info.Printf("%d nodes started", len(ns))

	hs := make(map[common.Hash]string)

	<-time.After(5 * time.Second)
	out.Info.Print("SENDING MESSAGES ...")

	// sending unique messages via every node
	for i := 0; i < chatMesgsCount; i++ {
		for j, n := range ns {
			s := fmt.Sprintf("test message %d via node %d", i, j)
			h, err := n.send("", s)
			if err != nil {
				t.Fatal(err)
			}
			hs[h] = s
		}
	}

	<-time.After(5 * time.Second)
	out.Info.Print("SENDING FINALIZATION MESSAGE ...")

	finmesg, err := ns[0].send("", "END")
	if err != nil {
		t.Fatal(err)
	}

	// waiting until all nodes recive END message
	for j, n := range ns {
		succeeded := false
	WaitForEndMesg:
		for i := 0; i < 20; i++ { // 2 Seconds
			if _, ok := n.hashes[finmesg]; ok {
				succeeded = true
				break WaitForEndMesg
			}
			<-time.After(100 * time.Millisecond)
		}
		if !succeeded {
			t.Fatalf("node %d has not recived END message", j)
		}
	}

	out.Info.Print("CHECKING PROPAGATION FOR MESSAGES")
	for h, s := range hs {
		out.Info.Printf("%s\n\t%s", h.Hex(), s)
	}

	// check all nodes recived all test messages
	for j, n := range ns {
		count := 0
		for v := range hs {
			if _, ok := n.hashes[v]; !ok {
				count++
			}
		}
		if count != 0 {
			t.Errorf("node %d lost %d of %d messages", j, count, len(hs))
		}
	}

	ns.stop()
}

func TestOneMessage(t *testing.T) {
	var err error
	var ms []*Message
	stk, _, con := startGethChat(t, 29999, 0)
	out.Info.Print("READY TO WORK")

	// subscribe on the room '.'
	ms, err = con.Poll(".")
	if err != nil {
		t.Fatal(err)
	}
	if len(ms) > 0 {
		t.Fatal("unempty message list on start")
	}

	mesg := &Message{Room: ".", Text: "HELLO", TTL: 10000}

	// send message to the room '.'
	err = con.Post(mesg)
	if err != nil {
		t.Fatal(err)
	}

	out.Info.Print("WAITING FOR MESSAGE")

WaitingLoop:
	for i := 0; i < 10; i++ { // 1 second
		ms, err = con.Poll(".")
		if err != nil {
			t.Fatal(err)
		}
		if len(ms) > 0 {
			break WaitingLoop
		}
		<-time.After(100 * time.Millisecond)
	}

	if len(ms) != 1 {
		t.Fatal("message does not arrived")
	}

	if !ms[0].EqualNoStamp(mesg) {
		t.Fatal("message is broken")
	}

	stk.Stop()
}
