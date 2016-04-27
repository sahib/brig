# Basic Information

## What is your project name?  

brig - Research and development on a distributed and secure file synchronisation toolbox

(TODO @Schöler: Ist das thema okay so?)

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

Our idea is to create a tool which allows individuals to safely exchange
documents and files without the need of a centralized or company controlled
service. The focus is on finding a good balance between security and usability.
We are developing the software as free open source software under the terms of
the AGPLv3 licence.

The goal is to offer an alternative to centralized cloud services like Dropbox
and additionally provide strong end-to-end encryption to protect data being
accessed by third parties.

To replace the centralized infrastructure we are building on the distributed
Interplanetary Filesystem (ipfs), which allows us to implement a whole range
of unique features:

**No single point of failure:** ipfs works as a distributed filesystem, only
requiring a handful of bootstrap nodes in order to connect to the network.
It is therefore hard to block or throttle the network access, whereas 
centralized cloud services depend on the availability of the service itself. 

**Version control system for large files:** Since ipfs uses content addressed
storage (i.e. files are addressed by their checksum), it is easy to build a
version control system for binary files.

**No vendor lock-in:** Both ipfs and brig is free software, everybody can
download, redistribute or modify the software. Even if ipfs or brig development
stalls, users will still be able to access their files or fork the original
project.

**Storage quotas:** On devices with limited storage space, a user will be able
to access files directly in the ipfs network, without having a full copy of all
the files. File deduplication and compression further reduces the storage
needed, making it suitable for low-end hardware and mobile devices.

Since ipfs is not easily usable by itself and has no focus on security, brig
attempts to fix this by making these features accessible through a simple
interface. This enables the following use cases:

* Secure file synchronisation and transfer.
* Encrypted backup or archive possibilities.
* Usable as encrypted offline container.
* Platform for other security focused applications.

## What are hoped for goals or longer term effects of the project?

Our goal is to develop a software product which does not scare off users by
complicated or additional security related options. The focus is to make brig as
straightforward to use as e.g. Dropbox, having all security related requirements
already included in a way which does not interfere with usability. Many existing
systems for secure file transfer (also including OpenPGP) are too complicated
for the majority of users and therefore not being used at all.

Another goal is to introduce a stable, always available and secure 'sharing
platform' for journalists, activists but also other people in countries with an
oppressive government or generally anybody in the need of end-to-end secured
file transfer.

On the other hand all the listed benefits can be introduced in our everyday's life
while sharing sensible data with your doctor, your lawyer or just applying for a job.

The licence and distributed infrastructure ensures that, like the world wide
web, the service is always available and cannot just get 'turned off' even if a
government decides to pull the plug. It is however possible that an oppressive
regime blocks or filters access to brig. This could be fixed or mitigated in the
future by offering to use brig in conjunction with the tor project. 

In a world where brig would be used instead of Dropbox, mass surveillance would
get tremendously harder since the data is no longer in the cloud (with companies
behind that may be potentially gagged by institutions like the NSA), but on
peers in almost every home and company.

In a nutshell: More protection for human rights and democracy by lowering the
hurdle to share documents securely.

<!-- We want to know how you think the world could be, what larger purpose this
project is a part of, and/or the bigger target you aiming for. Bulleted lists
are good. --> 

## Focus *

* Awareness of privacy and security threats
* Privacy enhancement
* Security from danger or threat online

## Status *

* It Exists! (Alpha/Beta)

## Technology attributes *

* User interface/experience
* Application deployment
* Server daemon
* Cryptography
* Desktop client
* Sensitive data
* Networking
* Other

<!-- If the proposed project is working very closely with technology such as
developing software or hardware, select any of the following that could describe
the technology. -->

# How will you do it?

## Describe how

<!-- Briefly and clearly list key milestones, objectives, and/or activities
briefly. These should be specific, measurable, attainable, realistic, and
time-relatable. Bulleted lists are ideal. -->

We are two computer science master students at the university of applied
sciences Augsburg, Germany. Currently we are working on a proof of concept code base
for brig, which should be available at the end of our master thesis.

Generally, our primary goal is to work on the topic of secure and distributed file
synchronization as research fellows, possibly also as PhD students. (TODO:
@schöler: passt der ausdruck dafür?)
To finance these positions a sound funding is required.  In turn we would be
able to steadily continue the research and development of brig. Sadly, funding
is hard to get on such a general topic that is additionally open source and
therefore often not seen as an attractive solution for businesses in this area.

We're planning these milestones:

* *Proof-of-concept:* A first draft of brig. Should include the most basic
  features (file encryption, basic synchronization capabilities and working fuse filesystem).
  Planned to be finished at the end of the master thesis (around Sept/Oct.
  2016).
* *Technical prototype:* Working prototype for technically affine users.
  Should include a solid technology preview on the following features: 
  Streaming compression, sound key management, version control, service
  discovery and solid user management.
  At this point brig should be usable as "git"-like toolbox for file
  synchronization. Planned to be finished half a year after the proof of
  concept (March/April 2017).
* *Test and stabilization phase:* The earlier prototype should be hardened,
  benchmarked and audited (also by external developers) in order to increase our
  confidence in the correctness of our software. At the end of this phase we
  plan to release brig to the open source community in the hope of feedback.
  Additionally we attempt to take first steps to make the software more usable
  to non-technical users. Planned to be finished half a year after the technical
  prototype (Sept/Oct. 2017).

We hope to get funding beginning at Sept/Oct. 2016.

## Objective(s)

* Research
* Technology development
* Deploying technology Software or hardware development
* Testing

## How long will it take?

(TODO: @Schöler: Summe und Zeitraum wurde gewählt weil projekte mit weniger als
300.000$ und weniger als 12 Monaten Zeitraum bevorzugt werden)

12 months

## How much do you want?

285.500 USD (about 250.000 in Euro)

# Who is the project for?

## Describe them

<!-- In other words, who are the people benefiting or affected most by this
effort and how well do you know them?
-->

TODO: Noch etwas weiter ausformulieren? 
TODO: @Schöler: Oder noch andere Usecases?

File synchronization is useful for everyone, but secure file sharing is
particularly useful for the following groups:

* *Individuals with interest in privacy*: Alice and Bob may share files without
  being afraid of surveillance by oppressive governments. If brig ever could be
  established as a general "standard", people like lawyers, physicians,
  journalists and activists would clearly benefit from it. 
* *Technical individuals*: Flexible toolbox to manage and share large amount of data.
  Additionally it enables sophisticated and script-able scenarios for
  non-standard use cases.
* *Companies*: Flexible toolbox which can be adapted to a company's needs. A company
  is able to build a private company controlled network for sharing sensitive
  data using brig. Since brig also has the notion of a user and a group of
  users, it is possible to map between existing user management systems and brig.
* *Academia and government organizations*: In academia and government organizations
  it might be used as internal storage for documents (similar to companies) or as
  exchange platform between students and a lecturer or between government
  offices and citizens.
* *Usage in Industry 4.0 or Smart Home:* brig could be used as flexible network
  mount to safely exchange e.g. log data between several distributed instances
  in this area.

## What community currently exists around this project?

<!-- Define the community as you see it. If your answer is none, please explain
how you plan to cultivate community around the proposed effort, including
mechanisms to receive feedback and get others involved. -->

Since brig is in an early development stage, no official community has been established
yet. However, we are in close contact with the ipfs community. Additionally the
scientific environment allows us to keep in close contact with professionals
addressing security, usability and other brig related topics like the distributed
systems group of Prof. Dr. Thorsten Schöler (http://dsg.hs-augsburg.de).

As we are developing free software since several years now by ourselves, we are
maintaining contact to fellow developers and people behind various linux
distributions all around the world.

## Beneficiaries

* General public
* Activists
* Journalists
* Advocacy groups/NGOs
* Academia
* Technologists
* Entrepreneurs
* Government

## Region

* Global

# Why is this project needed?

## Describe why
<!-- Describe one or more of the following: the specific needs of the group(s)
being met, how it uniquely solves a known issue or improve upon existing
solutions, and/or what knowledge, research, technology, or community gap the
proposed effort is intending to fill. If the effort targets a specific group of
people, note any research or analysis you have done to ensure the effort serves
the target population. -->

Journalists, lawyers, physicians and generally all people that are interested
(or have to!) in sharing documents in a secure fashion can currently choose from
a large amount of technologies to transmit documents. Sadly there is no
"default" way for this.

The most well known variant is the usage of cloud storage services like Dropbox,
Google Drive or iCloud. Although those services advertise with using
strong encryption, a user has to trust the company behind the service. Even if
e.g. Dropbox encrypts all files in the cloud and on the transfer, they're still
in possession of the encryption keys. Ever since Snowden's NSA leak, it should
be clear that a trust relationship is hard to maintain with a company that hosts
their servers in the USA. Apart from that, the proprietary
backend and client software of such services might contain backdoors or
bugs that leak user data. 

Privacy interested individuals try to partly solve this issue by encrypting
their files before uploading it to the cloud by using an encryption layer like
BoxCryptor or encfs. Sadly, this does not fix the actual problem, but just a few
of the symptoms: It might be still possible for the proprietary client software
to read the data before encryption, it introduces the need for additional,
potentially complicated, software for all people sharing a document, it depends
on the proprietary infrastructure of companies and sharing is only possible as
long one can afford to pay the service or as long the service still is online.

E-mail encryption software like OpenPGP is available, but the hard setup scares
off many potential non-technical users. Other FOSS products like Syncthing or 
git-annex provide distributed file sharing, but don't focus on security or are
too complicated.

By developing brig we hope to address most of these issues and find a good
balance between usability and security, while being completely free software
that is suitable to serve as secure standard document exchange software.

## Addressed problems

* Restrictive Internet filtering by technical methods (IP blocking, DNS filtering, TCP RST, DPI, etc.)
* Blocking, filtering, or modification of political, social, and/or religious content (including apps)
* Technical attacks against government critics, journalists, and/or human rights organizations (Cyberattacks)
* Localized or nationwide communications shut down or throttling (Blackouts)
* Physical intimidation, arrest, violence (including device seizure or destruction), and death for political or social reasons
* Repressive surveillance or monitoring of communication
* Policies, laws, or directives that increase surveillance, censorship, and punishment
* Government practices that hold intermediaries (social networks or ISPs) liable for user content
* Other

# Other information

brig is currently hosted on GitHub and under heavy development:

* https://github.com/disorganizer/brig

The public open source software of the two developers can be viewed here:

* https://github.com/sahib (christopher.pahl@hs-augsburg.de)
* https://github.com/qitta (christoph.piechula@hs-augsburg.de)

Research and development on brig is done in cooperation with our university, the
university of applied sciences Augsburg (https://www.hs-augsburg.de/).
The current code is capable of storing encrypted files and serving them again
via a fuse layer. Network synchronisation will be available in the very near
future.

As we also might want to graduate as PhD students, we have a time frame of around
3 years that has been estimated by our university where we can fully concentrate on
working on brig. Since funding is required to get into such a position we hope
that Open Technology Fund would make this dream possible for us by helping us
through the first year of our work.
