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

Since IPFS is not easily usable by itself and has no focus on security, brig
attempts to fix this by making these features accessible through a simple
interface. This enables the following use cases:

* Secure file synchronisation and transfer
* Encrypted backup or archive possibilities
* Usable as encrypted offline container
* Platform for other security focused applications

## What are hoped for goals or longer term effects of the project? (2000)

Our goal is to develop a software product which does not scare off users by
complicated or additional security related options. The focus is to make brig as
straightforward to use as Dropbox having all security related requirements
already included in a way which does not interfere with usability. Many existing
systems for secure file transfer (also including OpenPGP) are too complicated
for the majority of users and therefore not beeing used at all.

Another goal is to introduce a stable, always available and secure 'sharing
plattform' for journalists, activists but also other people in countries with an
oppressive gouvernement or generally anybody in the need of end-to-end secured
file transfer.

On the other hand all the listed benefits can be introduced in our eveydays life
while sharing sensible data with your doctor, your lawyer or just applying for a job.

The licence and distributed infrastructure ensures that, like the world wide
web, the service is always available and cannot just get 'turned off' even
if a governement decides to pull the plug. It is however possible that an
oppressive regime blocks or filters access to brig. This could be fixed in the
future by offering to use brig in conjunction with the tor onion routing
project. 

In a world where brig would be used instead of Dropbox, mass surveillance would
get tremendously harder since the data is no longer in the cloud (with companies
behing that may be potentially gagged by institutions like the NSA), but on
peers in almost every home and company.

In a nutshell: More protection for human rights and democracy.

<!-- We want to know how you think the world could be, what larger purpose this
project is a part of, and/or the bigger target you aiming for. Bulleted lists
are good. --> 

## Focus *
Awareness of privacy and security threats
Privacy enhancement
Security from danger or threat online
## Status *
It Exists! (Alpha/Beta)
## Technology attributes *
User interface/experience
Application deployment
Server daemon
Cryptography
Desktop client
Sensitive data
Networking
Other
<!-- If the proposed project is working very closely with technology such as
developing software or hardware, select any of the following that could describe
the technology. -->

# How will you do it?

## Describe how
<!-- Briefly and clearly list key milestones, objectives, and/or activities
briefly. These should be specific, measurable, attainable, realistic, and
time-relatable. Bulleted lists are ideal. -->

We are two computer science master students at the University of applied sience
in Augsburg, Germany. Currently we are working on a proof of concept code base
for brig, which should be available at the end of our master thesis.

* Our primary goal is to work on the general topic of secure file
  synchronization as a postgraduate students.  (TODO: passt das?)
  This needs funding, but that is hard to get when developing open source
  software.
* Proof-of-concept after master thesis (around Sept/Oct. 2016)
* Working prototype for technical users half an year after the master thesis.
  (git like commandline usage)
* Hardening, testing and stabilization phase aferwards. Also first work in making it
  usable for non-technical users and making it easily installable. Half a year.
* Further research and work on brig even after a potential funding. Projekt ist
  auf drei Jahre angelegt

-> Technische milestones mit groben Datum noch hinschreiben?
-> Einzelne Milestones allgemein noch etwas mehr beschreiben.

## Objective(s)
Research
Technology development
Deploying technology
Software or hardware development
Testing

## How long will it take?
12 months

## How much do you want?

285.500 USD (about 250.000 in Euro)

# Who is the project for?

## Describe them
<!-- In other words, who are the people benefiting or affected most by this
effort and how well do you know them?
-->

*Individuals*: Bob and Alice may share files in future without be afraid of surveillance.

*Technical individuals*: Flexible toolset to manage and share large amount of data. 

*Companies*: Flexible toolset which can be adapted to a company needs. A company
is able to build a private company controlled network for sharing files using
brig.

*Academia and goverment organizations*: In academia and goverment organizations
it might be used as internal storage for documents (similar to companies) or as
exchange platform between students and citizens.


## What community currently exists around this project?
<!-- Define the community as you see it. If your answer is none, please explain
how you plan to cultivate community around the proposed effort, including
mechanisms to receive feedback and get others involved. -->

Since brig is in a early developement stage, no official community has been established
yet. However, we are in close contact with the IPFS community. Additionally the
scientific environment allows us to keep in close contact with professionals
addressing security/usability and other brig related topics like the distributed
systems group of Prof. Dr. Thorsten Schoeler.

As we are developing free software since several years now, we are maintaining
contact to fellow developers and people behind various linux distributions all
around the world.

## Beneficiaries *
General public
Activists
Journalists
Advocacy groups/NGOs
Academia
Technologists
Entrepreneurs
Government

## Region

Global

# Why is this project needed?

## Describe why
<!-- Describe one or more of the following: the specific needs of the group(s)
being met, how it uniquely solves a known issue or improve upon existing
solutions, and/or what knowledge, research, technology, or community gap the
proposed effort is intending to fill. If the effort targets a specific group of
people, note any research or analysis you have done to ensure the effort serves
the target population. -->

Because there is no standard technology to share/exchange document.
Various cloud storage services.
Dropbox users which are concerned about their privacy might use
encfs/boxcryptor, this hence addresses only the symptoms but not the problem
itself.
Alternatives exists, but the set of features is different and none of them is
focused on either security and/or usability. Furthermore brig others a distinct
set of unique features not seen in other solutions.


## Addressed problems

Restrictive Internet filtering by technical methods (IP blocking, DNS filtering, TCP RST, DPI, etc.)
Blocking, filtering, or modification of political, social, and/or religious content (including apps)
Technical attacks against government critics, journalists, and/or human rights organizations (Cyberattacks)
Localized or nationwide communications shut down or throttling (Blackouts)
Physical intimidation, arrest, violence (including device seizure or destruction), and death for political or social reasons
Repressive surveillance or monitoring of communication
Policies, laws, or directives that increase surveillance, censorship, and punishment
Government practices that hold intermediaries (social networks or ISPs) liable for user content
Other

# Other information

* Aktuelle Finanzierungssituation und Github, evtl. OSS Entwickler Links wie
  Bachelorarbeiten.
