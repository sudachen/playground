package v1

import (
	"context"
	"sync"
	"time"
)

const apiExpireTimeout = time.Minute

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
	timeout := time.NewTicker(apiExpireTimeout)
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

func (api *ChatAPI) Post(ctx context.Context, m *Message) error {
	if err := api.c.Send(m); err != nil {
		return err
	}
	return nil
}

func (api *ChatAPI) Poll(ctx context.Context, room string) (m []Message, err error) {
	return
}
