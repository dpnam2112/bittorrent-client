
# Tracker

## Purposes

- It acts as an intermediary to help Bittorrent peers to find each other.

- When a peer owns a piece of a file, it may want to announce to other peers. This can be achieved
  by announce to a tracker.

## Flows

- When a peer wants to download the file:

    - The first step is to communicate with a tracker specified in the `.torrent` file. In the
      torrent file, there is a list containing available trackers that the downloader can
      communicate with (`announce_list`).

    - The tracker will respond with a list of addresses to other peers who own pieces of the file
      that the downloader wants. 
    
    - The communication can be made via HTTP or UDP protocols. With HTTP, the request is a HTTP GET
      request, parameters are a list of key-value pairs:

      `GET {http_tracker_url}?key_1=value_1&key_2=value_2&....`

      The response body is encoded in bencode format.
      
     - HTTP GET Request paramters:
	     - `info_hash`: Hash of the metainfo file, used to identify the torrent file.
	     - `peer_id`: ID of the requester joining the Bittorrent network. This is usually randomly generated. 
	     - `port`: Port on which the requester is listening. Follows Bittorrent protocol spec for ports, i.e, the port ranges from 6881 to 6889. This information is needed since other peers may want to download pieces that the requester has.
	     - `uploaded`: total of bytes uploaded
	     - `downloaded`: total of bytes downloaded
	     - `left`: total of bytes remaining to be downloaded (there are chances that some of the downloaded data fail the integrity check, and this means `left` might be different from `total - downloaded`
	     - `event`: Available values: `started`, `completed`, `stopped`, `empty`
            - If a download begins -> the client should put `started` in this field
            - A download finishes -> the client should put `completed` in this field
            - When a download is stopped -> `stopped` should be put in this field

    - Under UDP protocol, an `announce` request contains all of the fields mentioned above.

    - After the tracker receives the request, it then replies a response. The response structure
      under HTTP is a little bit different from the one under UDP:

        - under HTTP, the bencoded resposne contains information of all peers that the requester can reach
          out to retrieve pieces of the file (which is `peers`, a list of (host, port)) and
          additionally, a field `interval` to tell the requester time interval that the requester
          should wait before making the next request (it's a little bit like 'Retry-After` field in a `Too Many Request' response of HTTP).

        - under UDP, the response has the following format (applied only for IPv4, the field `seeders` is followed by a list of (host, port), `n` is the number of peers):
            ```
                    Offset      Size            Name            Value
            0           32-bit integer  action          1 // announce
            4           32-bit integer  transaction_id
            8           32-bit integer  interval
            12          32-bit integer  leechers        // <- number of peers currently downloading the requested file (not exactly)
            16          32-bit integer  seeders         // <- number of peers having a copy of the requested file
            20 + 6 * n  32-bit integer  IP address
            24 + 6 * n  16-bit integer  TCP port
            20 + 6 * N
            ```
