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

![](timeline.png?raw=true)

#### Risk List
| Risk Name | Value | Impact | Likelihood | Description |
|---|---:|---:|---:|---|
| Wrong Language | 8 | 8 | 1 | Go may turn out to be the wrong language for the job. |
| Outatime | 36 | 9 | 4 | I could have bitten off more than I could chew in the alotted time frame. |
| Wrong Skillset | 15 | 5 | 3 | I may not have the skills to complete parts of the project within the given parameters. |
| Sickness | 10 | 5 | 2 | I am a single point of failure since I am the only person developing this project. |
| Bugs | 9 | 1 | 9 | Bugs in the code are somewhat expected even after testing. |

1. I have already mitigated the possibility of Go being an incorrect language by learning it and using it for several other projects.  
2. I plan to monitor my time tables and I have partially mitigated time issues by creating a tentative schedule.  
3. I have the correct skillset for the project barring UI/UX design. I can mitigate this by learning what I can and utilizing well known frameworks.  
4. I accept the risk of illness, I typically wouldn't be able to do anything to mitigate it.  
5. I can mitigate the risk of bugs by testing thoroughly and completely.  

## Application Requirements

Below is an outline of in-progress user stories and a link to a use/misuse case diagram. Both of these items will evolve throughout the development of the project.  

If something isn't hashed out to the point of testable and verifiable requirements please feel free to submit suggestions in the issues section!  

Do not try to implement a feature without specifying testable and verifiable requirements! I will not accept the pull request if you cannot tell me what it is supposed to accomplish and how I am supposed to test and verify it.

#### User Stories
1. Connect to peers from direct IP (Roughly implemented)
  - Connect to a circle ([#1](https://github.com/Phosphoresce/airlock/issues/1))
    - If a peer is the first in the circle wait for incoming connections from other peers
    - If a peer is not the first in the circle request to join the circle
    - A peer must ask its initial peer for the list of other peers
  - Record participating peers ([#2](https://github.com/Phosphoresce/airlock/issues/2))
    - Each peer should create and maintain a list of other peers in a circle
    - After obtaining a list of peers each peer should announce their connection to the circle they are connecting to
    - Each peer should send their list of peers to another peer upon request 
  - Usernames ([#3](https://github.com/Phosphoresce/airlock/issues/3))
    - Each peer will request a username from its initial peer
    - The initial peer must only allow a new peer to connect if a requested username is not in use
  - Remove peers after timeout/disconnect ([#4](https://github.com/Phosphoresce/airlock/issues/4))
    - Each peer will send a keep-alive every specified interval
    - Each peer must remove other inactive peers
    - Each peer will broadcast a disconnect voluntarily leaving
  - ~~Detect peers on your network ([#5](https://github.com/Phosphoresce/airlock/issues/5))~~
    - ~~If a peer hasn't specified a target, try to detect peers on their network~~
  
2. Establish encryption from the start (Partially ready)
  - Each peer must create an identity key and associate it with their username
  - Each peer creates a public-private keypair
    - Each peer must establish a strong AES public-private pre-keypair
    - Each peer must allow other peers to collect their public pre-key
    - X3DH? 
    - Double Ratchet?
  
3. Send messages (Partially implemented)
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
5. UI/UX (Roughly Implemented)
  - QT5 for cross compiling
  - Terminal
  - GUI
  - web app  
  
#### Use Case Diagram
Go [here](https://www.lucidchart.com/invitations/accept/10e89bdf-b0f9-4b3d-81b2-e387545c307b) to view the use case diagram.  

#### Architectural Diagram
Go [here](https://www.lucidchart.com/invitations/accept/fb951839-3a90-4375-aa8c-ccd46285cc09) to view the architectural diagram.

The components depicted in the architectural diagram are split into two main systems.  

The first system is the core functionality. The second system is the UI or presentation layer which gathers relevant information via an API exposed by the core.  

The following components make up these two systems:  

**Core**  
Peer-to-peer connection manager:  
This component contains the main logic for providing low level networking and connection management for peer-to-peer communication. It keeps track of other peers in order to send them messages and it listens on a specified port for messages and connections.  

Encryption Manager:  
This component provides functions for encrypting data passed to it. It manages identity keys, and message keys for the chat room and private messages.  

File Sharing Manager:  
This component provides functionality to send files, links, and share locally hosted files.  

User Interface API:  
This component exposes an API to allow a suitable front-end application to display data and messages generated by the core components.  

**UI/UX**  
Front-End UI:  
This component comsumes the user interface API to display messages and data generated by the core components.  

#### Activity Diagrams
Go [here](https://www.lucidchart.com/invitations/accept/6e66a31a-8f6d-4d0d-875f-bdd26d6a6692) to view the activity diagrams.  

## Resources Required
| Resource | Dr. Hale needed? | Investigating Team member | Description |
|---|---|---|---|
| Distributed encryption research | no | Me | Need some whitepapers and research for distributed peer2peer encryption |
| UI/UX Experience | no | Me | Will need to look into UI frameworks, and decide on ways to present this application to users |

## Installation
To use Airlock you must build it from source.  

To build from source without the GUI follow the basic instructions in the [contributing document](CONTRIBUTING.md).
The basic steps are:
1. Install Golang
2. Clone the repository
3. `cd airlock`
4. `go install` or `go build`

To build from source with the GUI start with the basic instructions for installing without a GUI but then continue below:

1. Make sure you have Docker installed and make sure your user is in the Docker group:
  `sudo pacman -S docker && sudo usermod -G docker <your user>`
2. Pull a docker container with the compile time requirements:
  `docker pull therecipe/qt:windows_32_shared` or `docker pull therecipe/qt:linux` or `docker pull therecipe/qt:android`
3. Follow the instructions posted [here](https://github.com/therecipe/qt/blob/master/README.md#minimal-setup).
3. Use the makefile to build the project for your operating system:
  `make qt`
4. Alternatively, build the project for another operating system (you must pull the appropriate docker container):
  ${GOPATH}/bin/qtdeploy -docker build android
5. The resulting built project will be located in the deploy directory under the target operating system name. For example: `airlock/deploy/linux/...`
  
## Getting Started
To run Airlock gui:
- Microsoft Windows: Double click the `airlock.exe`
- Linux: run airlock.sh
- Android: `abd install build-debug.apk` then Tap the 'go' app icon on your phone.
  Note: You must have the android developer toolkit installed to install apks to an android device. Your phone must also be plugged into the computer.  

By default, the application will launch a chat Window and listen on port `9001` for messages from other peers.  

To connect to a peer you must specify a target and port to listen on from the command line: `airlock.sh -t localhost:9001 -p 9002`

```
Usage: airlock [-tpg]
-t --target Connects to the specified domain or ip address and port pair
-p --port   Specifies a custom port to listen for messages on
-g --gui    A boolean flag to enable or disable GUI
```

## Terms
These terms are not set in stone, but for the moment I have chosen to use bitTorrent-like terminology to keep things simple

Peer: A participant in the peer to peer network  
Circle: A group of connected peers aka swarm in bitTorrent terms.

## Contributing
If you would like to contribute please send me a pull request.  

Ensure that your code is re-usable if at all possible and that you include tests for all functionality!  
Remember that security is the heart of this project, so please make sure you are using secure code practices and have the users' safety in mind.

Finally, make sure that you document any changes you've made clearly and concisely in the [change log](http://keepachangelog.com/en/0.3.0/).
