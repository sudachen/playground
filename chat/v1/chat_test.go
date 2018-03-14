package v1

import (
	"testing"
)

const chatNodesCount = 16

func TestDiscovering(t *testing.T) {
	ns := initialize(chatNodesCount, t)
	ns.stop()
}
