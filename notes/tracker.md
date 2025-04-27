
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
      
     - Request paramters:
	     - `info_hash`: Hash of the metainfo file, used to identify the torrent file.
	     - `peer_id`: ID of the requester joinining the Bittorrent network. This is usually randomly generated. 
	     - `port`: Port on which the requester is listening. Follows Bittorrent protocol spec for ports, i.e, the port ranges from 6881 to 6889. This information is needed since other peers may want to download pieces that the requester has.
	     - `uploaded`: total of bytes uploaded
	     - `downloaded`: total of bytes downloaded
	     - `left`: total of bytes remaining to be downloaded (there are chances that some of the downloaded data fail the integrity check, and this means `left` might be different from `total - downloaded`
	     - 

