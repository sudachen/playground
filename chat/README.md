# P2P Chat Board

The chat is a go-ethereum service implementing decentralized chat board over ethereum p2p network.

The P2P Chat copies and store board messages temporarily over p2p network on every node running the Chat service. All messages are anonymous, however, a message can be signed to identify the sender. A user can choose any nickname to represent his identity with or without a signature.

Every message has a room selector, it can be used by interface applications and bots to grouping messages into chatting spaces.

The P2P Chat has public API in cht namespace available by RPC on every node starting the Chat nd the RPC services.
