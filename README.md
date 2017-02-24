# Airlock
Airlock is a secure P2P chat and collaboration platform.

## Executive Summary
Airlock is designed to bring friends, colleagues, and peers together so that they can share experiences, relax, or work on projects together. Airlock is based off of a completely peer to peer design, in which users will need to directly connect to each other to chat, share files, or collaborate on documents. There are no servers involved and chat rooms may be hosted by anybody. In large chat rooms users can send each other personal messages or work on documents with subsets of the connected peers. Airlock also incorporates security from the ground up. Every chat room is encrypted with proven and open source algorithms.  

Airlock provides a safe way for users to communicate and collaborate. It is designed to protect private matters with encryption between every user in every chat room. Airlock does not collect any information for analysis, and only users who have shared keys can decipher messages from each other. Users can rest assured that their private communications stay private.  

#### Problem Statement
Privacy is very important to users. Personal communications and personally identifying information should be protected. If personal information lands in the wrong hands, a person may be affected in a number of ways. For example, medical information may prevent a person from getting insurance, or conversations with a friend could offend a person's employer. People deserve to be able to communicate without the worry that their personal information is at risk.  

#### Merit
Airlock is intended to provide another avenue for private communication without making the process uncomfortable or difficult. The end goal of the Airlock project is to make it easy for users to maintain their privacy while participating in the important aspects of their lives.  

#### Project Goals
1. Allow users to communicate
2. Encrypt communications between users and groups of users separately
3. Allow users to share files, media, and links in a secure manner
4. Allow users to collaborate on documents or source code without 3rd party sites
5. Allow users have a secure and modern chat experience on any platform (Web, Desktop, Mobile)

## Overview of Development
High-level planned functionality is as follows:
- Chatrooms
- Links
- Document and code collaboration
- File sharing
- Pictures?
- GIFs/WEBMs?

### Outlined Tasks
1. Connect to peers from direct IP (Ready for development)
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
  
2. Establish encryption from the start (Partially ready)
  - Each peer must create an identity key and associate it with their username
  - Each peer creates a public-private keypair
    - Each peer must establish a strong AES public-private pre-keypair
    - Each peer must allow other peers to collect their public pre-key
    - X3DH? 
    - Double Ratchet?
  
3. Send messages (Partially ready)
  - Personal message
    - Peers must send all individual messages to another peer encrypted with the receivers public pre-key
    - The participating peers must establish a shared key
    - The participating peers must then communicate with the agreed upon shared key
  - Message a subset of peers in a chat room (Design validation)
    - Peers will establish a shared key using subsequent key exchanges e.g. 2 peers would start with an exchange and add a 3rd peer by completing an echange between the initial 2 peers and the new peer.
      - To complete the exchange to add the nth peer the joining peer will complete an exchange with only a single peer using the current shared key. That contacted peer will broadcast the new shared key, established for the newly added peer, to the remainder of the old shared key users encrypted with the old shared key.
  - Message entire circle (Still requires design)
    - It turns out that efficient group key exchanges are a big problem in cryptography. This may cause performance issues with large circles.
    - The idea here is to keep doing key exchanges, as with the 'message a subset' task, until everybody has a shared key for the global channel. This may be difficult as users will have to constantly generate new keys or update their keys.
  
4. Collaboration (Still requires design)
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

Do not try to implement a feature without specifying testable and verifiable requirements! I will not accept the pull request if you cannot tell me what it is supposed to accomplish and how I am supposed to test and verify it.  

## Terms
These terms are not set in stone, but for the moment I have chosen to use bitTorrent-like terminology to keep things simple

Peer: A participant in the peer to peer network  
Circle: A group of connected peers aka swarm in bitTorrent terms.

## Contributing
If you would like to contribute please send me a pull request.  

Ensure that your code is re-usable if at all possible and that you include tests for all functionality!  
Remember that security is the heart of this project, so please make sure you are using secure code practices and have the users' safety in mind.

Finally, make sure that you document any changes you've made clearly and concisely in the [change log](http://keepachangelog.com/en/0.3.0/).
