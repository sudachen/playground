package v1

import (
	"testing"
)

const chatNodesCount = 5

func TestDiscovering(t *testing.T) {
	ns := initialize(chatNodesCount, t)

	// do some work here

	ns.stop()
}
