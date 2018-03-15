package v1

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/console"
	ethn "github.com/ethereum/go-ethereum/node"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/rpc"
)

func flagset(a ...string) *flag.FlagSet {
	fst := flag.NewFlagSet("ppf", flag.ContinueOnError)
	fst.Parse(a)
	return fst
}

func startGethChat(t *testing.T, p2pPort int, rpcPort int, a ...string) (stk *ethn.Node, c *Chat, con *conn) {
	var err error

	//app := cli.NewApp()
	//ctx := cli.NewContext(app, flagset(a...), nil)

	ncfg := ethn.DefaultConfig
	ncfg.Name = "test-cht"
	ncfg.Version = params.VersionWithCommit("")
	ncfg.HTTPModules = append(ncfg.HTTPModules, "cht")
	ncfg.WSModules = append(ncfg.WSModules, "cht")
	ncfg.IPCPath = "test-cht.ipc"
	ncfg.HTTPPort = rpcPort
	ncfg.P2P.ListenAddr = fmt.Sprintf("127.0.0.1:%d", p2pPort)

	stk, err = ethn.New(&ncfg)
	if err != nil {
		t.Fatalf("failed to register the Whisper service: %v", err)
	}

	c = New(nil)
	err = stk.Register(func(n *ethn.ServiceContext) (ethn.Service, error) {
		return c, nil
	})
	if err != nil {
		t.Fatalf("failed to register the Whisper service: %v", err)
	}

	err = stk.Start()
	if err != nil {
		t.Fatalf("error starting protocol stack: %v", err)
	}

	clnt, err := stk.Attach()
	if err != nil {
		t.Fatalf("failed to attach to self: %v", err)
	}

	con = &conn{inout: make(prompter), clnt: clnt}
	cons, err := console.New(console.Config{
		DataDir:  stk.InstanceDir(),
		DocRoot:  "testdata",
		Client:   clnt,
		Prompter: con.inout,
		Printer:  &con.prnt,
		Preload:  nil,
	})
	if err != nil {
		t.Fatalf("failed to create JavaScript console: %v", err)
	}
	con.cons = cons
	con.Eval(Console_JS)
	con.prnt.Reset()
	return
}

type prompter chan string

func (p prompter) PromptInput(prompt string) (string, error) {
	// Send the prompt to the tester
	select {
	case p <- prompt:
	case <-time.After(time.Second):
		return "", errors.New("prompt timeout")
	}
	// Retrieve the response and feed to the console
	select {
	case input := <-p:
		return input, nil
	case <-time.After(time.Second):
		return "", errors.New("input timeout")
	}
}

func (p prompter) PromptPassword(prompt string) (string, error) {
	return "", errors.New("not implemented")
}
func (p prompter) PromptConfirm(prompt string) (bool, error) {
	return false, errors.New("not implemented")
}
func (p prompter) SetHistory(history []string)              {}
func (p prompter) AppendHistory(command string)             {}
func (p prompter) ClearHistory()                            {}
func (p prompter) SetWordCompleter(c console.WordCompleter) {}

type conn struct {
	inout prompter
	prnt  bytes.Buffer
	cons  *console.Console
	clnt  *rpc.Client
}

func (c *conn) Eval(js string) (string, error) {
	c.prnt.Reset()
	c.cons.Evaluate(js)
	output := c.prnt.String()
	return strings.TrimSpace(output), nil
}

func (c *conn) Post(mesg *Message) error {
	b, err := json.Marshal(mesg)
	if err != nil {
		return err
	}
	r, err := c.Eval("cht.post(" + string(b) + ")")
	if err != nil {
		return err
	}
	if r != "true" {
		return errors.New("unexpected output: " + r)
	}
	return nil
}

func (c *conn) Poll(room string) (ms []*Message, err error) {
	r, err := c.Eval("cht.pollStr(\"" + room + "\");")
	if err != nil {
		return
	}
	if r != "null" && r != "" {
		var s string
		err = json.Unmarshal([]byte(r), &s)
		err = json.Unmarshal([]byte(s), &ms)
	}
	return
}
