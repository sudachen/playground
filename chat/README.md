# P2P Chat Board

The chat is an go-ethereum service implementing decentralized chat board over ethereum p2p network. 

The P2P Chat copying and temporaly store board messages over p2p network on every node running the Chat service. All messages is anonimouse, however a message can be signed to identify sender. User can choose any nikname to represent his identity with or without signature.

Every message has a room selector, it can be used by interface applications and bots to grouping massages to chating spaceses.

The P2P Chat has public API in cht namespace available by RPC on every node starting the Chat nd the RPC services.
