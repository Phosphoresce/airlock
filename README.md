# Airlock
Airlock is a secure P2P chat and collaboration platform.

## Design

Planned functionality is as follows:
- Chatrooms
- Links
- Code collaboration
- File sharing
- Pictures?
- GIFs/WEBMs?

## Overview of Development

1. Connect to peers from direct IP
  - Connect to a circle
    - If a peer is the first in the circle wait for incoming connections from other peers
    - If a peer is not the first in the circle request to join the circle
    - A peer must ask its initial peer for the list of other peers
  - Record participating peers
    - Each peer should create and maintain a list of other peers in a circle
    - After obtaining a list of peers each peer should announce their connection to the circle they are connecting to
    - Each peer should send their list of peers to another peer upon request 
  - Usernames
    - Each peer will request a username from its initial peer
    - The initial peer must only allow a new peer to connect if a requested username is not in use
  - Remove peers after timeout/disconnect
    - Each peer will send a keep-alive every specified interval
    - Each peer must remove other inactive peers
    - Each peer will broadcast a disconnect voluntarily leaving
  - Detect peers on your network
    - If a peer hasn't specified a target, try to detect peers on their network
  
2. Establish encryption from the start
  - Each peer creates a public-private keypair
    - X3DH? 
    - Double Ratchet?
  - Different keys for each chatroom a user participates in
  
3. Send messages
  - Message entire group
  - Message a subset (chat room)
  - Personal message
  
4. Collaboration
  - Links
    - Open in browser
    - Anonymous proxy links
  - Images
    - Show in-line if possible
    - Allow download (danger zone: people could use this maliciously)
  - Documents
    - Peer hosts the file
    - Both peers access hosted file via API ( read file, append, insert )
5. UI/UX
  - Terminal
  - GUI
  - web app
  
If something isn't hashed out to the point of testable and verifiable requirements please feel free to submit suggestions in the issues section!  

Do not try to implement a feature without specifying testable and verifiable requirements! I will not except the pull request if you cannot tell me what it is supposed to accomplish and how I am supposed to test and verify it.  

## Terms
These terms are not set in stone, but for the moment I have chosen to use bitTorrent-like terminology to keep things simple

Peer: A participant in the peer to peer network
Circle: A group of connected peers aka swarm in bitTorrent terms.

## Contributing
If you would like to contribute please send me a pull request.  

Ensure that your code is re-usable if at all possible and that you include tests for all functionality!  
Remember that security is the heart of this project, so please make sure you are using secure code practices and have the users' safety in mind.

Finally, make sure that you document any changes you've made clearly and concisely in the change log.
