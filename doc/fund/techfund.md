# Basic Information

## What is your project name?  

brig - Research/developement of a distributed/secure file exchange/synchronisation tool

## Your name 

Christopher Pahl 
Christoph Piechula

## E-mail

christopher.pahl@hs-augsburg.de
christoph.piechula@hs-augsburg.de

## Have you ever applied to or received funding as an OTF project?

No.

# What is your idea?

## Describe it (2000c)
<!-- In as few sentences as possible, please describe your idea. -->

Creating a tool which allows individuals to safely exchange documents and files
without the need of a centralized or company controlled service. The focus is on
finding a good balance between security and usability. We are developing the
software as free open source software under the terms of the AGPLv3 licence.

The goal is to offer an alternative to centralized cloud services like Dropbox
and additionally provide strong end to end encryption to protect data being
accessed by third parties.

To replace the centralized infrastructure we are building on the distributed
Interplanetary Filesystem (IPFS), which allows us to implement a whole range
of unique features:

**No single point of failure:** IPFS works as a P2P filesystem, only requiring a
handful of bootstrap nodes in order to connect to the network. Centralized cloud
services depend on the availability of the service itself. It is therefore hard
to block or throttle the network access.

**Version control system for large files:** Since IPFS uses content addressed
storage (ie. files are addressed by their checksum), it is easy to build a
version control system for binary files.

**No vendor lock-in:** Both IPFS and brig is free software, everybody can
download, redistribute or modify the software. Even  if IPFS or brig
developement stalls, users will still be able to access their files.

**Storage Quotas:**: On devices with limited storage space, a user will be able
to access files directly in the IPFS network, without having a full copy of all
the files. File deduplication and compression further reduces the storage
needed.

Use cases:

    * File synchronisation 
    * Encrypted/secure file transfer
    * Backup or archive possibilities
    * Data safe usage 
    * Plattform for other applications

-- Usability
-- Security -> Journalists, Activists et cetera


## What are hoped for goals or longer term effects of the project? (2000)

A world without dropbox and corporations that hoard our data and benefit from it.

<!-- We want to know how you think the world could be, what larger purpose this
project is a part of, and/or the bigger target you aiming for. Bulleted lists
are good. --> 

## Focus *
Access to the Internet
Awareness of privacy and security threats
Privacy enhancement
Security from danger or threat online
Choose the options that most applies to the proposed effort.
## Status *
Just an Idea (Pre-alpha)
It Exists! (Alpha/Beta)
It's basically done. (Release)
People Use It. (Production)
Choose the option that most applies to the proposed effort.
## Technology attributes *
Browser extension
Browser plugin
Unmanaged language
User interface/experience
Anonymity
Application deployment
Web application
Server daemon
Web API/Mobile application (serverside)
Mobile application (clientside)
Cryptography
Desktop client
Desktop App
Dependency integration
Software as a Service (SaaS)
Platform as a service (PaaS)
Infrastructure as a service (IaaS)
Sensitive data
Networking
Wireless Communication
Hardware/Embedded device(s)
Reverse Engineering
Other
Not applicable
<!-- If the proposed project is working very closely with technology such as
developing software or hardware, select any of the following that could describe
the technology. -->

