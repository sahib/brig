---
documentclass: scrreprt
classoption: toc=listof,index=totoc
include-headers:
    - \usepackage{url} 
    - \usepackage[ngerman]{babel}
    - \usepackage{csquotes}
    - \usepackage[babel, german=quotes]{csquotes}
fontsize: 11pt
sections: yes
toc: yes
lof: no
lot: no
date: \today
---

\newpage
\pagenumbering{arabic} 
\setcounter{page}{1}

# Zusammenfassung der Projektziele

Ziel des Projektes ist die Entwicklung einer sicheren und dezentralen
Alternative zu Cloud--Storage Lösungen wie Dropbox, die sowohl für Unternehmen,
als auch für Heimanwender nutzbar ist. Trotz der Prämisse, einfache Nutzbarkeit
zu gewährleisten, wird auf Sicherheit sehr großen Wert gelegt.  Aus Gründen der
Transparenz wird die Software mit dem Namen »``brig``« dabei quelloffen unter der
``AGPL`` Lizenz entwickelt.

Nutzbar soll das resultierende Produkt, neben dem Standardanwendungsfall der
Dateisynchronisation, auch als Backup- bzw. Archivierungs--Lösung sein.
Des Weiteren kann es auch als verschlüsselter Daten--Safe oder als Plattform für
andere, verteilte Anwendungen (wie beispielsweise aus dem Industrie 4.0 Umfeld)
dienen.

Von anderen Softwarelösungen soll es sich stichpunkthaft durch folgende Merkmale
abgrenzen:

- Verschlüsselte Übertragung *und* Speicherung.
- Unkomplizierte Installation und einfache Nutzung durch simplen Ordner im
  Dateimanager.
- Transparenz, Anpassbarkeit und Sicherheit durch *Free Open Source Software (FOSS)*.
- Kein *Single Point of Failure* (*SPoF*), wie bei zentralen Diensten.
- Dezentrales Peer--to--Peer--Netzwerk auf Basis von ``ipfs``.
- Benutzerverwaltung auf Basis der ``XMPP``--Infrastruktur.
- Versionsverwaltung großer Dateien mit definierbarer Tiefe.

# Steckbrief 

## Einleitung

Viele Unternehmen haben sehr hohe Ansprüche an die Sicherheit, welche zentrale
Alternativen wie beispielsweise Dropbox[^Dropbox] nicht bieten können. Zwar wird
die Übertragung von Daten zu den zentralen Dropbox-Servern verschlüsselt, was
allerdings danach mit den Daten »in der Cloud« passiert liegt nicht mehr unter der
Kontrolle der Nutzer. Dort sind die Daten schon manchmal für andere Nutzer wegen Bugs
einsehbar[^BUGS] oder müssen gar von Dropbox an amerikanische Geheimdienste[^NSA]
weitergegeben werden.

[^NSA]: Siehe auch \url{http://www.spiegel.de/netzwelt/web/dropbox-edward-snowden-warnt-vor-cloud-speicher-a-981740.html}
[^BUGS]: Siehe dazu auch \url{http://www.cnet.com/news/dropbox-confirms-security-glitch-no-password-required/}
[^Dropbox]: Mehr Informationen unter \url{https://www.dropbox.com/}

Sprichwörtlich gesagt, kann man nicht kontrollieren wo die Daten aus der Cloud
abregnen. Tools wie Boxcryptor[^Boxcryptor] lindern diese Problematik zwar
etwas indem sie die Dateien verschlüsseln, heilen aber nur die Symptome und nicht
das zugrunde liegende Problem.

[^Boxcryptor]: Krypto-Layer für Cloud-Dienste, siehe \url{https://www.boxcryptor.com/de}

Dropbox ist leider kein Einzelfall --- beinahe alle Cloud--Dienste haben, oder
hatten, architekturbedingt ähnliche Sicherheitslecks. Für ein Unternehmen wäre
es vorzuziehen ihre Daten auf Servern zu speichern, die sie selbst
kontrollieren. Dazu gibt es bereits einige Werkzeuge wie *ownCloud*[^OWNCLOUD]
oder Netzwerkdienste wie *Samba*, doch technisch bilden diese nur die zentrale
Architektur von Cloud--Diensten innerhalb eines Unternehmens ab. 

[^OWNCLOUD]: *ownCloud*--Homepage: \url{https://owncloud.org/}

## Ziele

Ziel ist daher die Entwicklung einer sicheren, dezentralen und unternehmenstauglichen
Dateisynchronisationssoftware namens ``brig``. Die »Tauglichkeit« für ein
Unternehmen ist natürlich sehr individuell. Wir meinen damit im Folgenden diese Punkte:

- *Einfache Benutzbarkeit:* Sichtbar soll nach der
  Einrichtung nur ein Ordner im Dateimanager sein.
- *Effiziente Übertragung von Dateien:* Intelligentes Routing vom Speicherort zum Nutzer.
- *Speicherquoten:* Nur relevante Dateien müssen synchronisiert werden.
- *Automatische Backups:* Versionsverwaltung auf Knoten mit großem Speicherplatz.
- *Schnelle Auffindbarkeit:* Kategorisierung durch optionale Verschlagwortung.

Um eine solche Software zu entwickeln, wollen wir auf bestehende Komponenten wie
dem *InterPlanetaryFileSystem* (kurz ``ipfs``, ein flexibles P2P
Netzwerk[@peer2peer]) und *XMPP* (ein Messanging Protokoll und Infrastruktur,
siehe [@xmpp]) aufsetzen. Dies macht die Entwicklung eines Prototypen mit
vertretbaren Aufwand möglich.

Von einem Prototypen zu einer marktreifen Software ist es allerdings stets ein
weiter Weg. Daher wollen wir einen großen Teil der Zeit nach dem Prototyp damit
verbringen, die Software bezüglich Sicherheit, Performance und
Benutzerfreundlichkeit zu optimieren. Da es dafür nun mal keinen
standardisierten Weg gibt, ist dafür ein gewisses Maß an Forschung nötig.

## Einsatzmöglichkeiten

``brig`` soll deutlich flexibler nutzbar sein als zentrale Dienste. Nutzbar soll
es sein als…

- *Synchronisationslösung*: Spiegelung von zwei oder mehr Ordnern.
- *Transferlösung*: »Veröffentlichen« von Dateien nach Außen mittels Hyperlinks.
- *Versionsverwaltung*: 
  Bis zu einer bestimmten Tiefe können alte Dateien wiederhergestellt werden.
- *Backup- und Archivierungslösung*: Verschiedene »Knoten--Typen« möglich.
- *Verschlüsselter Safe*: ein »Repository«[^REPO] kann »verschlossen« und 
  wieder »geöffnet« werden.
- *Semantisch durchsuchbares* Tag-basiertes Dateisystem[^TAG].
- *Plattform* für verteilte Anwendungen.
- einer beliebigen Kombination der oberen Punkte.

[^TAG]: Mit einem ähnlichen Ansatz wie \url{https://en.wikipedia.org/wiki/Tagsistant}
[^REPO]: *Repository:* Hier ein »magischer« Ordner in denen alle Dateien im Netzwerk angezeigt werden.

## Zielgruppen

Die primäre Zielgruppe von ``brig`` sind Unternehmenskunden und Heimanwender.
Wie man unten sehen kann, sind noch weitere sekundäre Zielgruppen denkbar.

### Unternehmen

Unternehmen können ``brig`` nutzen, um ihre Daten und Dokumente intern zu
verwalten. Besonders sicherheitskritische Dateien entgehen so der Lagerung in
Cloud--Services oder der Gefahr von Kopien auf unsicheren
Mitarbeiter--Endgeräten. Größere Unternehmen verwalten dabei meist ein
Rechenzentrum in dem firmeninterne Dokumente gespeichert werden. Von den
Nutzern werden diese dann meist mittels Diensten wie *ownCloud* oder *Samba*
»händisch« heruntergeladen.

In diesem Fall könnte man ``brig`` im Rechenzentrum und auf allen Endgeräten
installieren. Das Rechenzentrum würde die Datei mit tiefer Versionierung
vorhalten. Endanwender würden alle Daten sehen, aber auf ihren Gerät nur die
Daten tatsächlich speichern, die sie auch benutzen. Hat beispielsweise ein
Kollege im selben Büro die Datei bereits vorliegen, kann brig diese dann direkt
blockweise vom Endgerät des Kollegen holen.


Kleinere Unternehmen, wie Ingenieurbüros, können ``brig`` dazu nutzen Dokumente nach
Außen freizugeben, ohne dass sie dazu vorher irgendwo »hochgeladen« werden
müssen. Dies wird dadurch möglich gemacht, dass Dateien mittels eines
*Hyperlinks* nach außen publik gemacht werden können. So muss die Gegenseite
``brig`` nicht installiert haben.

### Privatpersonen / Heimanwender

Heimanwender können ``brig`` für ihren Datenbestand aus Fotos, Filmen, Musik und
sonstigen Dokumenten nutzen. Ein typischer Anwendungsfall wäre dabei auf einem
NAS Server, der alle Dateien mit niedriger Versionierung speichert. Die
Endgeräte, wie Laptops und Smartphones, würden dann ebenfalls ``brig`` nutzen,
aber mit deutlich geringeren Speicherquotas (maximales Speicherlimit), so dass
nur die aktuell benötigten Dateien physikalisch auf dem Gerät vorhanden sind.
Die anderen Dateien lagern »im Netz« und können transparent von ``brig`` von
anderen verfügbaren Knoten geholt werden. Sollte der Nutzer, beispielsweise auf
einer längeren Zugfahrt, offline sein, so kann er benötigte Dateien vorher
»pinnen«, um sie lokal zwischenzuspeichern.

### Plattform für industrielle Anwendungen

Da ``brig`` auch komplett automatisiert und ohne Interaktion nutzbar sein soll, kann
es auch als Plattform für jede andere Anwendungen genutzt werden, die Dateien
austauschen und synchronisieren müssen. Eine Anwendung in der Industrie 4.0 wäre
beispielweise die Synchronisierung von Konfigurationsdateien im gesamten
Netzwerk.

### Einsatz im öffentlichen Bereich

Aufgrund seiner Transparenz und einfachen Benutzbarkeit wäre ebenfalls eine
Nutzung an Schulen, Universitäten oder auch in Behörden zum Dokumentenaustausch
denkbar. Vorteilhaft wäre für die jeweiligen Institutionen hierbei vor allem,
dass man sich aufgrund des Open--Source Modells an keinen Hersteller bindet
(Stichwort: *Vendor Lock*) und keine behördlichen Daten in der »Cloud« landen.
Eine praktische Anwendung im universitären Bereich wäre die Verteilung von
Studienunterlagen an die Studenten.

# Stand der Technik

Die Innovation bei unserem Projekt  besteht daher hauptsächlich darin, bekannte
Technologien »neu zusammen zu stecken«, woraus sich viele neue Möglichkeiten
ergeben.  Wie im nächsten Kapitel beleuchtet wird, ist ``brig`` der Versuch
viele gute, bestehende und praxisgeprüfte Ideen in einem konsistenten Programm
zu vereinen.

## Stand der Wissenschaft

Zwar ist das Projekt stark anwendungsorientiert, doch basiert es auf gut
erforschten Technologien wie Peer--to--Peer-Netzwerken (kurz *P2P*, siehe auch
[@peer2peer_arch]), von der NIST[^NIST] zertifizierten kryptografischen
Standard-Algorithmen[@everyday_crypto] und verteilten Systemen im Allgemeinen
(wie der freie XMPP Standard). P2P--Netzwerke wurden in den letzten
Jahren gut erforscht und haben sich auch in der Praxis bewährt: Skype ist
vermutlich das bekannteste, kommerzielle P2P Netzwerk (siehe auch @peer2peer, S.2). 

Allerdings ist uns keine für breite Massen nutzbare Software bekannt, die es
Nutzern ermöglicht selbst ein P2P Netzwerk aufzuspannen und darin Dateien
auszutauschen. Am nähsten kommen dem die beiden Softwareprojekte
»``Syncthing``« (OpenSource, [^SYNCTHING]) und »``BitTorrent Sync``«
(proprietär, [^BITSYNC]). Beide nutzen zwar P2P--Technologie zum Austausch der
Dateien, modellieren aber kein »echtes« P2P--Netzwerk, bei dem nicht jeder
Teilnehmer eine volle Kopie sämtlicher Daten haben muss.

[^SYNCTHING]: Siehe auch dazu: \url{https://syncthing.net/}
[^BITSYNC]: Siehe Hersteller--Webpräsenz: \url{https://www.getsync.com/}

Der wissenschaftliche Beitrag unserer Arbeit wäre daher die Entwicklung einer
freien Alternative, die von allen eingesehen, auditiert und studiert werden
kann. Diese freie Herangehensweise ist insbesondere für sicherheitskritische
Software relevant, da keine (offensichtlichen) »Backdoors« in die Software
eingebaut werden können.

[^NIST]: NIST: *National Institute of Standards and Technology*

## Markt und Wettbewerber

Bereits ein Blick auf Wikipedia[@wiki_filesync] zeigt, dass der momentane Markt
an Dateisynchronisationssoftware (im weitesten Sinne) sehr unübersichtlich ist.
Ein näherer Blick zeigt, dass die Software dort oft nur
in Teilaspekten gut funktioniert oder mit anderen unlösbaren Problemen
behaftet sind.

### Verschiedene Alternativen

Im Folgenden geben wir eine unvollständige Übersicht über bekannte
Dateisynchronisations--Programmen. Davon stehen nicht alle in Konkurrenz zu
``brig``, sind aber aus Anwendersicht ähnlich. 

#### Dropbox + Boxcryptor

Der vermutlich bekannteste und am weitesten verbreitete zentrale Dienst zur
Dateisynchronisation. Verschlüsselung kann man mit Tools wie ``encfs``
(Open--Source, siehe auch [^ENCFS]) oder dem ähnlich umfangreichen, proprietären
*Boxcryptor* nachrüsten. Was das Backend genau tut ist leider das
Geheimnis von Dropbox --- es ist nicht Open--Source. 

[^ENCFS]: Mehr Informationen unter \url{https://de.wikipedia.org/wiki/EncFS}

Die Server von Dropbox stehen in den Vereinigten Staaten, was spätestens
seit den Snowden--Enthüllungen für ein mulmiges Gefühl sorgen sollte. Wie oben
erwähnt, kann diese Problematik durch die Verschlüsselungssoftware *Boxcryptor*
abgemildet werden. Diese kostet aber zusätzlich und benötigt noch einen
zusätzlichen zentralen Keyserver[^KEYSERVER]. 

[^KEYSERVER]: Mehr Informationen zum Keyserver unter \url{https://www.boxcryptor.com/de/technischer-\%C3\%BCberblick\#anc09}

Technisch nachteilhaft ist vor allem, dass jede Datei »über den Pazifik« hinweg
synchronisiert werden muss, nur um eventuell auf dem Arbeitsrechner nebenan 
anzukommen.

#### ownCloud

Aus dieser Problemstellung heraus entstand die Open--Source Lösung *ownCloud*.
Nutzer hosten auf ihren Servern selbst eine ownCloud--Instanz und stellen
ausreichend Speicherplatz bereit. Vorteilhaft ist also, dass die Daten auf den
eigenen Servern liegen. Nachteilig hingegen, dass das zentrale Modell von Dropbox
lediglich auf eigene Server übertragen wird. Die Daten müssen zudem von einer
Weboberfläche geholt werden und liegen nicht in einem »magischen«,
selbst--synchronisierenden Ordner.

#### Syncthing

Das 2013 veröffentliche quelloffene *Syncthing* versucht diese zentrale
Instanz zu vermeiden, indem die Daten jeweils von Peer zu Peer übertragen
werden. Es ist allerdings kein vollständiges Peer--to--peer--Netzwerk: Geteilte
Dateien liegen immer als Kopie bei allen Teilnehmern, die die Datei haben.
Alternativ ist aber auch selektives Synchronisieren von Dateien möglich.

Besser als bei ownCloud ist hingegen gelöst, dass ein »magischer« Ordner
existiert in dem man einfach Dateien legen kann, um sie zu teilen. Zudem wird die
Datei vom nächstgelegenen Knoten übertragen. Praktisch ist auch, dass
*Syncthing* Instanzen mittels eines zentralen Discovery--Servers entdeckt werden
können.  Nachteilig hingegen ist die fehlende Benutzerverwaltung: Man kann nicht
festlegen von welchen Nutzern man Änderungen empfangen will und von welchen
nicht. 

#### BitTorrent Sync

In bestimmten Kreisen scheint auch das kommerzielle und proprietäre 
*BitTorrent Sync* beliebt zu sein. Hier wird das bekannte und freie BitTorrent
Protokoll zur Übertragung genutzt. Vom Feature--Umfang ist es in etwa
vergleichbar mit *Syncthing*. Die Dateien werden allerdings noch zusätzlich
AES--verschlüsselt abgespeichert.

Genauere Aussagen kann man leider aufgrund der geschlossenen Natur des Programms
und der eher vagen Werbeprosa nicht treffen. Ähnlich zu *Syncthing* ist
allerdings, dass eine Versionsverwaltung nur mittels eines »Archivordners«
vorhanden ist. Gelöschte Dateien werden schlicht in diesen Ordner verschoben und
können von dort wiederhergestellt werden. Die meisten anderen Vor- und Nachteile
von *Syncthing* treffen auch hier zu.

#### ``git-annex``

Das 2010 erstmals veröffentlichte ``git-annex``[^ANNEX] geht in vielerlei Hinsicht
einen anderen Weg. Einerseits ist es in der funktionalen Programmiersprache
Haskell geschrieben, andererseits nutzt es das Versionsverwaltungssystem ``git``[@git],
um die Metadaten zu den Dateien abzuspeichern, die es verwaltet. Auch werden
Dateien standardmäßig nicht automatisch synchronisiert, man muss Dateien selbst
»pushen«, beziehungsweise »pullen«.

[^ANNEX]: Webpräsenz: \url{https://git-annex.branchable.com/}

Dieser »Do-it-yourself« Ansatz ist sehr nützlich, um ``git-annex`` als Teil der
eigenen Anwendung einzusetzen. Für den alltäglichen Gebrauch ist es aber selbst
für erfahrene Anwender zu kompliziert, um es praktikabel einzusetzen.

Trotzdem sollen zwei interessante Features nicht verschwiegen werden, welche wir
langfristig gesehen auch in ``brig`` realisieren wollen:

* *Special Remotes:* »Datenablagen« bei denen ``git-annex`` nicht installiert sein muss.
                      Damit können beliebige Cloud--Dienste als Speicher genutzt werden.
+ *N-Copies:* Von wichtigen Dateien kann ``git-annex`` bis zu ``N`` Kopien speichern.
              Versucht man eine Kopie zu löschen, so verweigert ``git-annex`` dies.

### Zusammenfassung

Obwohl ``brig`` eine gewisse Ähnlichkeit mit verteilten Dateisystemen, wie
*GlusterFS* hat, wurden diese in der Übersicht weggelassen --- einerseits aus
Gründen der Übersicht, andererseits weil diese andere Ziele verfolgen und von
Heimanwendern kaum genutzt werden.

Zusammengefasst findet sich hier noch eine tabellarische Übersicht mit den aus
unserer Sicht wichtigsten Eigenschaften: 

|                      | **FOSS**            | **Dezentral**       | **Kein SPoF**                 | **Versionierung**                    | **Einfach nutzbar** | **P2P**         |  
| -------------------- | ------------------- | ------------------- | --------------------------- | -------------------------------------- | ------------------- |------------------|
| *Dropbox/Boxcryptor* | \xmark              | \xmark              | \xmark                      | \textcolor{YellowOrange}{Rudimentär}   | \cmark              | \xmark           |
| *ownCloud*           | \cmark              | \xmark              | \xmark                      | \textcolor{YellowOrange}{Rudimentär}   | \cmark              | \xmark           |
| *Syncthing*          | \cmark              | \cmark              | \cmark                      | \textcolor{YellowOrange}{Archivordner} | \cmark              | \xmark           |
| *BitTorrent Sync*    | \xmark              | \cmark              | \cmark                      | \textcolor{YellowOrange}{Archivordner} | \cmark              | \xmark           |
| ``git-annex``        | \cmark              | \cmark              | \cmark                      | \cmark                                 | \xmark              | \xmark           |
| ``brig``             | \cmark              | \cmark              | \cmark                      | \cmark                                 | \cmark              | \cmark           |

# Das Projekt ``brig``

Optimal wäre also eine Kombination aus den Vorzügen von *Syncthing*, *BitTorrent
Sync* und ``git-annex``. Wie wir diese technichen Vorzüge ohne große Nachteile
erreichen wollen, wird im Folgenden beleuchtet.

## Der Name

Eine »Brigg« (englisch »brig«) ist ein kleines und wendiges
Zweimaster--Segelschiff aus dem 18-ten Jahrhundert. Passend erschien uns der Name
einerseits, weil wir flexibel »Güter« (in Form von Dateien) in der ganzen Welt
verteilen, andererseits weil ``brig`` auf (Datei-)Strömen operiert.

Dass der Name ähnlich klingt und kurz ist wie ``git``, ist kein Zufall. Das
Versionsverwaltungssystem (kurz VCS) hat durch seine sehr flexible und dezentrale
Arbeitsweise bestehende zentrale Alternativen wie ``svn`` oder ``cvs`` fast
vollständig abgelöst. Zusätzlich ist der Gesamteinsatz von
Versionsverwaltungssystemen durch die verhältnismäßige einfache Anwendung
gestiegen.
Wir hoffen mit ``brig`` eine ähnlich flexible Lösung für große Dateien
etablieren zu können. 

## Wissenschaftliche und technische Arbeitsziele

Um die oben genannten Ziele zu realisieren ist eine sorgfältige Auswahl der
Technologien wichtig. Der Einsatz eines Peer--to--Peer Netzwerk zum Dateiaustausch
ermöglicht interessante neue Möglichkeiten. Bei zentralen Ansätzen müssen
Dateien immer vom zentralen Server (der einen *Single Point of Failure*
darstelle) geholt werden. Dies ist relativ ineffizient, besonders wenn viele
Teilnehmer im selben Netz die selbe große Videodatei empfangen wollen. Bei ``brig``
würde der Fortschritt beim Ziehen der Datei unter den Teilnehmern aufgeteilt
werden. Hat ein Teilnehmer bereits einen Block einer Datei, so kann er sie mit
anderen direkt ohne Umweg über den Zentralserver teilen.

Zudem reicht es prinzipiell wenn eine Datei nur einmal im Netz vorhanden ist.
Ein Rechenzentrum mit mehr Speicherplatz könnte alle Dateien zwischenhalten,
während ein *Thin--Client* nur die Dateien vorhalten muss mit denen gerade
gearbeitet wird.
Zu den bereits genannten allgemeinen Zielen kommen also noch folgende technischen Ziele:

* Verschlüsselte Übertragung *und* Speicherung.
* *Deduplizierung*: Gleiche Dateien werden nur einmal im Netz gespeichert.
* *Benutzerverwaltung* mittels XMPP--Logins.
* *Speicherquoten* & Pinning (Dateien werden lokal »festgehalten«)
* Kein offensichtlicher *Single Point of Failure*.
* Optionale *Kompression* mittels der Algorithmen ``snappy`` oder ``brotli``.
* *Zweifaktor-Authentifizierung* und *paranoide* Sicherheit--Standards »Made in Germany«.

## Lösungsansätze

Als Peer--to--Peer Filesystem werden wir das InterPlanetaryFileSystem[^IPFS]
nutzen.  Dieses implementiert für uns bereits den Dateiaustausch zwischen den
einzelnen ``ipfs``--Knoten. Damit die Dateien nicht nur verschlüsselt übertragen
sondern auch abgespeichert werden, werden sie vor dem Hinzufügen zu IPFS mittels
AES im GCM--Modus von ``brig`` verschlüsselt und optional komprimiert. Zur
Nutzerseite hin bietet ``brig`` dann eine Kommandozeilenanwendung und ein
FUSE-Dateisystem[^FUSE], welches alle Daten in einem ``brig`` Repository wie normale
Dateien in einem Ordner aussehen lässt. Beim »Klick« auf eine Datei wird diese
von ``brig`` dann, für den Nutzer unsichtbar, im Netzwerk lokalisiert,
empfangen, entschlüsselt und als Dateistrom nach außen gegeben.

[^IPFS]: Mehr Informationen unter \url{http://ipfs.io/}
[^FUSE]: FUSE: *Filesystem in Userspace*, siehe auch \url{https://de.wikipedia.org/wiki/Filesystem_in_Userspace}

![Übersicht über die Kommunikation zwischen zwei Partnern/Repositories, mit den relevanten Sicherheits--Protokollen](images/security.png){#fig:security}

Der AES--Schlüssel wird dabei an ein Passwort geknüpft, welches der Nutzer beim
Anlegen des Repositories angibt. Das Passwort wiederum ist an einen
XMPP--Account der Form ``nutzer@server.de/ressource`` geknüpft.
Ein Überblick über die sicherheitsrelevanten Zusammenhänge findet sich
in Abbildung {@fig:security}.

Alle Änderungen an einem Repository werden in einer Metadatendatenbank
gespeichert. Diese kann dann mit anderen Teilnehmern über XMPP, und
verschlüsselt via OTR[^OTR], ausgetauscht werden. Jeder Teilnehmer hat dadurch
den gesamten Dateiindex. Die eigentlichen Dateien können aber »irgendwo« im
Teilnehmernetz sein. Sollte eine Datei lokal benötigt werden, so kann man sie
»pinnen«, um sie lokal zu speichern. Ansonsten werden nur selbst erstellte
Dateien gespeichert und andere Dateien maximal solange vorgehalten, bis die
Speicherquote erreicht ist.

[^OTR]: *Off--the--Record--Messaging:* Mehr Informationen unter \url{https://de.wikipedia.org/wiki/Off-the-Record_Messaging}

Nutzer die ``brig`` nicht installiert haben, oder mit denen man aus
Sicherheitsgründen nicht das gesamte Repository teilen möchte, können einzelne
Dateien ganz normal aus ihrem Browser heraus herunterladen. Dazu muss die Datei
vorher »publik« gemacht werden. Der außenstehende Nutzer kann dann die Datei
über ein von ``brig`` bereitgestelltes »Gateway« von einem öffentlich
erreichbaren Rechner mittels einer ``URL`` herunterladen.

Um Portabilität zu gewährleisten wird die Software in der Programmiersprache
``Go``[@go_programming_language] geschrieben sein. Der Vorteil hierbei ist, dass am
Ende eine einzige sehr portable, statisch gelinkte Binärdatei erzeugt wird.
Weitere Vorteile sind die hohe Grundperformanz und die sehr angenehmen
Werkzeuge, die mit der Sprache mitgeliefert werden. Die Installation von
``brig`` ist beispielsweise unter Unix nur ein einzelner Befehl:

```bash
$ go get github.com/disorganizer/brig
```

## Technische Risiken 

Der Aufwand für ein Softwareprojekt dieser Größe ist schwer einzuschätzen. Da
wir auf relativ junge Technologien wie ``ipfs`` setzen, ist zu erwarten, dass
sich in Details noch Änderungen ergeben. Auch die Tauglichkeit bezüglich
Performance ist momentan noch schwer einzuschätzen. Aus diesen Gründen werden
wir zwischen ``brig`` und ``ipfs`` eine Abstraktionsschicht bauen, um notfalls
den Einsatz anderer Backends zu ermöglichen.

Erfahrungsgemäß nimmt auch die Portierung und Wartung auf anderen Plattformen
sehr viel Zeit in Anspruch. Durch die Wahl der hochportablen Programmiersprache
Go minimieren wir dies drastisch.

Wie für jede sicherheitsrelevante Software ist die Zeit natürlich ein Risiko.
Ein Durchbruch im Bereich der Quantencomputer könnte daher in absehbarer
Zeit zu einem Sicherheitsrisiko werden.

# Wirtschaftliche Verwertung

## Open--Source--Lizenz und Monetarisierung

Als Lizenz für ``brig`` soll die Copyleft--Lizenz ``AGPL`` zum Einsatz kommen.
Diese stellt sicher, dass Verbesserungen am Projekt auch wieder in dieses
zurückfließen müssen.

Dass die Software quelloffen ist, ist kein Widerspruch zur wirtschaftlichen
Verwertung. Statt auf Softwareverkäufe zu setzen lässt sich mit dem Einsatz und
der Anpassung der Software Geld verdienen.  Das Open--Source Modell bietet aus
unserer Sicht hierbei sogar einige Vorteile:

- Schnellere Verbreitung durch fehlende Kostenbarriere auf Nutzerseite.
- Kann von Nutzern und Unternehmen ihren Bedürfnissen angepasst werden.
- Transparenz in Punkto Sicherheit (keine offensichtlichen Backdoors möglich).
- Fehlerkorrekturen, Weiterentwicklung und Testing aus der Community.

## Verwertungskonzepte

Es folgen einige konkrete Verwertungs--Strategien, die teilweise auch in
Partnerschaft mit dazu passenden Unternehmen ausgeführt werden könnten.
Prinzipiell soll die Nutzung für private und gewerbliche Nutzer kostenfrei sein,
weitergehende Dienstleistungen aber nicht.

### Bezahlte Entwicklung spezieller Features

Für sehr spezielle Anwendungsfälle wird auch ``brig`` nie alle Features
anbieten können, die der Nutzer sich wünscht. Das ist auch gut so, da es die
Programmkomplexität niedriger hält. Für Nutzer, die bereit sind für Features zu
zahlen, wären zwei Szenarien denkbar:

*Allgemein nützliche Änderungen:* Diese werden direkt in ``brig`` integriert und
sind daher als Open--Source für andere nutzbar. Dies bietet Unternehmen die
Möglichkeit, die weitere Entwicklung von ``brig`` mittels finanziellen Mitteln
zu steuern.

*Spezielle Lösungen:* Lösungen die nur für spezifische Anwendungsfälle Sinn
machen. Ein Beispiel wäre ein Skript, dass für jeden Unternehmens--Login einen
XMPP--Account anlegt.

### Supportverträge

Normalerweise werden Fehler bei Open--Source--Projekten auf einen dafür
eingerichteten Bugtracker gemeldet. Die Entwickler können dann, nach einiger
Diskussion und Zeit, den Fehler reparieren. Unternehmen haben aber für
gewöhnlich kurze Deadlines bis etwas funktionieren muss.

Unternehmen mit Supportverträgen würden daher von folgenden Vorteilen profitieren:

- *Installation* der Software.
- *Priorisierung* bei Bug--Reports.
- Persönlicher *Kontakt* zu den Entwicklern.
- *Wartung* von nicht--öffentlichen Spezialfeatures
- Installation von *YubiKeys*[^YUBI] oder anderer Zwei--Faktor--Authentifizierung.

[^YUBI]: Ein flexibles 2FA-Token. Mehr Informationen unter \url{https://www.yubico.com/faq/yubikey}

### Mehrfachlizensierung

Für Unternehmen, die unsere Software als Teil ihres eigenen Angebots nutzen
wollen, kann die Erteilung einer anderen Lizenz in Frage kommen:

- Eine Consulting Firma könnte eine Lizenz bei uns erwerben, um selbst
  Speziallösungen zu entwickeln, die sie dann nicht als *Open--Source*
  veröffentlichen müssen.

- Ein Hosting Anbieter der ``brig`` nutzen möchte, müsste wegen der ``AGPL``
  dazu erst die Erlaubnis bei uns einholen.  Je nach Fall könnte dann ein
  entsprechender Vertrag ausgehandelt werden.

### Zertifizierte NAS-Server

Besonders für Privatpersonen oder kleine Unternehmen wie Ingenieurbüros wäre
ein vorgefertigter Rechner mit vorinstallierter Software interessant. Das
Software- und Hardware--Zusammenspiel könnte dann vorher von uns 
getestet werden und mit der Zeit auch dem technischen Fortschritt angepasst
werden.

### Lehrmaterial und Consulting.

Auf lange Sicht wären auch Lehrmaterial, Schulungen und Consulting im
Allgemeinen als Eingabequelle denkbar. 
Respektable Einnahmen könnte man auch mit Merchandise, wie beispielsweise
Flaschenschiffen, erzielen. \smiley{}

# Beschreibung des Arbeitsplans

## Technische Arbeitsschritte

Im Rahmen unserer Masterarbeiten werden wir einen Prototypen entwickeln, der
bereits in Grundzügen die oben beschriebenen Technologien demonstriert.
Gute Performanz, Portabilität und Anwenderfreundlichkeit sind zu diesem Zeitpunkt aus
Zeitmangel allerdings noch keine harten Anforderungen.

Die im ersten Prototypen gewonnen Erkenntnisse wollen wir dazu nutzen,
nötigenfalls eine technische »Kurskorrektur« durchzuführen und den ersten
Prototypen nach Möglichkeit zu vereinfachen und zu stabilisieren.

Zu diesem zweiten Prototypen werden dann in kleinen Iterationen Features
hinzugefügt. Jedes dieser Feature sollte für sich alleine stehen, daher sollte
zu diesem Zeitpunkt bereits die grundlegende Architektur relativ stabil sein.

Nachdem ein gewisses Mindestmaß an nützlichen Features hinzugekommen ist, wäre
ein erstes öffentliches Release anzustreben. Dies hätte bereits eine gewisse
Verbreitung zur Folge und die in ``brig`` eingesetzten Sicherheitstechnologien
könnten von externen Sicherheitsexperten auditiert werden.

## Meilensteinplanung

Der oben stehende Zeitplan ist nochmal in Abbildung {@fig:milestones} auf drei Jahre
gerechnet zu sehen.

![Grobe Meilensteinplanung von 2016 bis 2019.](images/milestones.png){#fig:milestones}

Dabei sollen Prototyp I & II mindestens folgende Features beinhalten:

*Prototyp I:*s

- Grundlegende Dateiübertragung.
- Verschlüsselte Speicherung.
- FUSE Layer zum Anzeigen der Dateien in einem »magischen« Ordner.


*Prototyp II:*

- Sichere XMPP--Benutzerverwaltung.
- Erste Effizienzsteigerungen.
- Tag--basierte Ansicht im FUSE Layer.
- Verlässliche Benutzung auf der Kommandozeile. (ähnlich ``git``)

Weitere Features kommen dann in kleinen, stärker abgekapselten, Iterationen hinzu.

# Finanzierung des Vorhabens

Eine mögliche Finanzierungstrategie bietet das IuK--Programm[^IUK] des
Freistaates Bayern. Dabei werden Kooperation zwischen Fachhochschulen und
Unternehmen mit bis zu 50% des Fördervolumens gefördert. Gern gesehen ist dabei
ein Großunternehmen, welches zusammen mit einem kleinen bis mittleren
Unternehmen (``KMU``) das Fördervolumen aufbringt. Aus diesen Mitteln 
könnte die Hochschule Augsburg dann zwei Stellen für wissenschaftliche Mitarbeiter
über eine gewisse Dauer finanzieren.

Die Höhe des Fördervolumens richtet sich primär nach der Dauer der Förderung und
dem jeweiligen akademischen Abschluss. Die Dauer würden wir dabei auf mindestens
zwei, optimalerweise drei Jahre ansetzen. Sehr grob überschlagen kommen wir
dabei für das nötige Fördervolumen auf folgende Summe:

```python
>>> gehalt = 3500 + 2000                # Bruttogehalt + Arbeitgeberanteil 
>>> spesen = 30000                      # Anschaffungen, Büro, etc.
>>> pro_mann = 12 * gehalt              # =  66000 Euro
>>> pro_jahr = 2 * pro_mann + spesen    # = 162000 Euro
>>> budget = 3 * pro_jahr               # = 486000 Euro ~ 500.000 Euro
```

Für einen erfolgreichen Projektstart sollten daher zwei Unternehmen bereit sein,
diese Summe gemeinsam aufzubringen. Die Gegenleistung bestünde dann einerseits
natürlich aus der fertigen Software, andererseits aus möglichen weiteren daraus
resultierenden Kooperationen.

[^IUK]: Mehr Informationen unter \url{http://www.iuk-bayern.de/}

\newpage

# Literaturverzeichnis
