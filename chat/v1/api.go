package v1

import (
	"time"
	"sync"
	"context"
)

const apiFilterTimeout = 300

type ChatAPI struct {
	c  *Chat
	mu sync.Mutex
}

func NewChatAPI(c *Chat) *ChatAPI {
	api := &ChatAPI{c: c}
	go api.run()
	return api
}

func (api *ChatAPI) run() {
	timeout := time.NewTicker(2 * time.Minute)
	for {
		<-timeout.C

		api.mu.Lock()

		// do delayed actions here

		api.mu.Unlock()
	}
}

func (api *ChatAPI) Version(ctx context.Context) string {
	return ProtocolVersionStr
}

func (api *ChatAPI) Post(ctx context.Context, m Message) (bool, error) {
	return true, nil
}

func (api *ChatAPI) Poll(room string) (m []Message, err error) {
	return
}


