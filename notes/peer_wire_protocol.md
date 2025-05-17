#  Introduction

- Used for communication between peers in a swarm.
## Terms
- A swarm is a collection of peers in the network that are serving or downloading the files at a specific point of time.
- Choke: Action of denying to give a peer file pieces.
- Unchoke: Action of accepting a peer to download file pieces.
- Interested: If a peer is 'interested', it means that peer wants to download pieces from another peer.
- Not interested: opposite of Interested
# Specification

## Flow
Here is the protocol flow:
1. Handshake: two peers exchange torrent's SHA1 infohash to start the communication.
2. Two peers exchange messages
## Handshake phase
Two peers initiate the communication by exchanging handshake messages:
```
<pstrlen><pstr><reserved><info_hash><peer_id>
```

- `pstrlen`: 1 byte — length of `pstr` (usually 19)
- `pstr`: "BitTorrent protocol"
- `reserved`: 8 bytes for future use (e.g., DHT, extensions)
- `info_hash`: 20-byte SHA-1 hash of the torrent metadata
- `peer_id`: 20-byte identifier of the peer

## Message exchanging phase
- Two peers exchange messages to each other on top of a long-lived TCP/uTP connection.
- Message format: `<length prefix><message ID><payload>`
	- length prefix (4 bytes): total length of the message, **excluding** itself.
    - message ID (1 byte): indicates the type of message.
	- payload: message-specific content.
### Message Types

| ID  | Name           | Payload Format                        | Purpose                                                                                                                        |
| --- | -------------- | ------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| -   | Keep-alive     | None (length = 0)                     | Prevent timeout                                                                                                                |
| 0   | Choke          | None                                  | Refuse to upload to the peer                                                                                                   |
| 1   | Unchoke        | None                                  | Allow uploads to peer                                                                                                          |
| 2   | Interested     | None                                  | Want to download from peer                                                                                                     |
| 3   | Not Interested | None                                  | Not interested in downloading                                                                                                  |
| 4   | Have           | `index` (4 bytes)                     | Notify that a piece is available                                                                                               |
| 5   | Bitfield       | Bitfield (X bytes)                    | Send initial piece availability. Used to tell the other peers which pieces the client already has.                             |
| 6   | Request        | `index`, `begin`, `length` (12 bytes) | Ask for a block (of a piece). `index` is the piece's index, `begin` and `length` are offset and length of the requested block. |
| 7   | Piece          | `index`, `begin`, `block`             | Send a block of a piece                                                                                                        |
| 8   | Cancel         | Same as Request                       | Cancel a request                                                                                                               |
| 9   | Port           | `listen-port` (2 bytes)               | Used in DHT                                                                                                                    |
| 20  | Extended       | Extended protocol ID + payload        | Used for extensions (e.g., metadata exchange)                                                                                  |
## Choking algorithm

### Problem
- Some of the peers contribute nothing, but only download.
- A mechanism for peer selection is needed. The more a peer contributes, the more it is prioritized (unchoke) by other peers.
- Starvation problem may happen: Newcomers to the network may not have a chance to show their contribution.
### Solution
Different clients have different solutions for choking algorithm. There is no fixed solution.
E.g, we can use multi-level priority queues for peer selection, with priority rotation to prevent starvation problem.
