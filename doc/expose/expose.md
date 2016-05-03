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
``AGPLv3`` Lizenz entwickelt.

Nutzbar soll das resultierende Produkt, neben dem Standardanwendungsfall der
Dateisynchronisation, auch als Backup- bzw. Archivierungs--Lösung sein. Des
Weiteren kann es auch als verschlüsselter Daten--Safe oder als Plattform für
andere, verteilte Anwendungen dienen -- wie beispielsweise aus dem Industrie 4.0 Umfeld.

Von anderen Softwarelösungen soll es sich stichpunkthaft durch folgende Merkmale
abgrenzen:

- Verschlüsselte Übertragung *und* Speicherung.
- Unkomplizierte Installation und einfache Nutzung durch simplen Ordner im Dateimanager.
- Transparenz, Anpassbarkeit und Sicherheit durch *Free Open Source Software (FOSS)*.
- Kein *Single Point of Failure* (*SPoF*), wie bei zentralen Diensten.
- Dezentrales Peer--to--Peer--Netzwerk auf Basis von ``ipfs``.
- Globales Benutzermanagement auf Basis von ``ipfs`` (Anbindung an existierende
  Systeme sowie Single--Sign--On technisch möglich).
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
abregnen. Tools wie Boxcryptor[^Boxcryptor] lindern diese Problematik zwar etwas
indem sie die Dateien verschlüsseln, heilen aber nur die Symptome und nicht das
zugrunde liegende Problem. Ein weiteres Problem bei zentralen Diensten ist die
Abhängigkeit von der Verfügbarkeit des Dienstes. Daten können nur ausgetauscht
werden solange der Dienst existiert, online ist und die Kosten regelmäßig
beglichen werden.

[^Boxcryptor]: Krypto-Layer für Cloud-Dienste, siehe \url{https://www.boxcryptor.com/de}

Dropbox ist leider kein Einzelfall --- beinahe alle Cloud--Dienste haben, oder
hatten, architekturbedingt ähnliche Sicherheitslecks. Für ein Unternehmen wäre
es vorzuziehen ihre Daten auf Servern zu speichern, die sie selbst
kontrollieren. Dazu gibt es bereits einige Werkzeuge wie *ownCloud*[^OWNCLOUD]
oder Netzwerkdienste wie *Samba*, doch technisch bilden diese nur die zentrale
Architektur von Cloud--Diensten innerhalb eines Unternehmens ab. 

[^OWNCLOUD]: *ownCloud*--Homepage: \url{https://owncloud.org/}

## Ziele und Einsatzmöglichkeiten

Ziel ist daher die Entwicklung einer sicheren, dezentralen und unternehmenstauglichen
Dateisynchronisationssoftware namens ``brig``. Die »Tauglichkeit« für ein
Unternehmen ist natürlich sehr individuell. Wir meinen damit im Folgenden diese Punkte:

- *Balance zwischen Benutzbarkeit und Sicherheit:* Sichtbar soll nach der
  Einrichtung nur ein Ordner im Dateimanager sein, alle Daten werden 
  auf Basis von einem einzigen Passwort verschlüsselt, welches beispielsweise
  aus einem bestehendem System abgeleitet werden kann.
- *Effiziente Übertragung von Dateien:* Intelligentes Routing vom Speicherort zum Nutzer.
- *Speicherquoten:* Nur relevante Dateien müssen synchronisiert werden.
- *Automatische Backups:* Versionsverwaltung auf Knoten mit großem Speicherplatz.
- *Schnelle Auffindbarkeit:* Kategorisierung durch optionale Verschlagwortung.
- *Kein Vendor Lock-In dank freier Software:* Herstellerunabhängigkeit
  gewährleistet volle Kontrolle über die Software und gewährleistet den Fortbestand.

Um eine solche Software zu entwickeln, wollen wir auf bestehende Komponenten
aufsetzen. Die grundlegende Basis bildet dabei das *InterPlanetaryFileSystem*
(kurz ``ipfs``, ein verteiltes »Dateisystem« [@peer2peer]). Dies macht die
Entwicklung eines Prototypen mit vertretbaren Aufwand möglich.

Von einem Prototypen zu einer marktreifen Software ist es allerdings stets ein
sehr weiter Weg. Daher wollen in den Folgejahren, neben der Weiterentwicklung,
einen großen Teil der Zeit damit verbringen, die Software bezüglich Sicherheit,
Performance und Benutzerfreundlichkeit zu optimieren. Da es dafür keinen
standardisierten Weg gibt, ist hier auch ein dementsprechend hoher
Forschungsaufwand nötig.

``brig`` soll letzendlich deutlich flexibler nutzbar sein als zentrale Dienste
und vergleichbare Software. Nutzbar soll es sein als…

- *Synchronisationslösung*: Spiegelung von zwei oder mehr Ordnern.
- *Transferlösung*: »Veröffentlichen« von Dateien nach Außen mittels Hyperlinks.
- *Versionsverwaltung*: Bis zu einer konfigurierbaren Tiefe können Dateien wiederhergestellt werden.
- *Backup- und Archivierungslösung*: Verschiedene »Knoten--Typen« möglich.
- *Verschlüsselter Safe*: ein »Repository«[^REPO] kann »verschlossen« und wieder »geöffnet« werden.
- *Semantisch durchsuchbares* Tag-basiertes Dateisystem[^TAG].
- *Plattform* für verteilte und sicherheitskritische Anwendungen.
- …einer beliebigen Kombination der oberen Punkte.

[^TAG]: Mit einem ähnlichen Ansatz wie \url{https://en.wikipedia.org/wiki/Tagsistant}
[^REPO]: *Repository:* Hier ein »magischer« Ordner in denen alle Dateien im Netzwerk angezeigt werden.

## Zielgruppen

Auch wenn ``brig`` extrem flexibel einsetzbar ist, sind die primären Zielgruppen
Unternehmen und Heimanwender. Aufgrund der starken Ende-zu-Ende Verschlüsselung
ist ``brig`` auch für Berufsgruppen, bei denen eine hohe Diskretion bezüglich
Datenschutz gewahrt werden muss, attraktiv. Hier wären in erster Linie
Journalisten, Anwälte, Ärzte mit Schweigepflicht auch Aktivisten und politisch
verfolgte Minderheiten, zu nennen.

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
vorhalten. Endanwender würden alle Daten sehen, aber auf ihrem Gerät nur die
Daten tatsächlich speichern, die sie auch benötigen. Hat beispielsweise ein
Kollege im selben Büro die Datei bereits vorliegen, kann ``brig`` diese dann
direkt transparent vom Endgerät des Kollegen holen. Das »intelligente Routing«
erlaubt den Einsatz von ``brig`` auf Smartphones, Tablets und anderen
speicherplatz-limitierten Geräten. Nutzer, die eine physikalische Kopie der Datei
auf ihrem Gerät haben wollen, können das entsprechende Dokument »pinnen«. Ist
ein Außendienstmitarbeiter beispielsweise im Zug unterwegs, kann er vorher eine
benötigtes Dokument pinnen, damit ``brig`` die Datei persistent verfügbar macht.

Indirekt sorgt auch die einfache Benutzbarkeit von ``brig`` für höhere
Sicherheit, da Mitarbeiter sich weniger durch die Sicherheitsrichtlinien ihres
Unternehmens gegängelt fühlen und nicht die Notwenigkeit sehen, wichtige
Dokumente auf private Geräte oder Speicher zu kopieren. Dies wirkt ebenfalls
Gefahren wie Industriespionage entgegen.

Da ``brig`` auch das Teilen von Dateien mittels Hyperlinks über ein »Gateway«
erlaubt, ist beispielsweise ein Kunde eines Ingenieurbüros nicht genötigt
``brig`` ebenso installieren zu müssen.


### Privatpersonen / Heimanwender

Heimanwender können ``brig`` für ihren Datenbestand aus Fotos, Filmen, Musik und
sonstigen Dokumenten nutzen. Ein typischer Anwendungsfall wäre dabei ein
NAS--Server, der alle Dateien mit niedriger Versionierung speichert. Endgeräte,
wie Laptops und Smartphones, würden dann ebenfalls ``brig`` nutzen, aber mit
deutlich geringeren Speicherquotas (maximales Speicherlimit), so dass nur die
aktuell benötigten Dateien physikalisch auf dem Gerät vorhanden sind. Die
anderen Dateien lagern »im Netz« und können transparent von ``brig`` von anderen
verfügbaren Knoten geholt werden. 

### Plattform für industrielle Anwendungen

Da ``brig`` auch komplett automatisiert und ohne Interaktion nutzbar sein soll,
kann es auch als Plattform für jede andere Anwendung genutzt werden, die Dateien
sicher austauschen und synchronisieren müssen. Eine Anwendung in der Industrie 4.0 
wäre beispielsweise die Synchronisierung von Konfigurationsdateien im gesamten Netzwerk.


### Einsatz im öffentlichen Bereich

Aufgrund der Ende-zu-Ende Verschlüsselung und einfachen Benutzbarkeit ist eine
Nutzung an Schulen, Universitäten sowie auch in Behörden zum Dokumentenaustausch
denkbar. Vorteilhaft wäre für die jeweiligen Institutionen hierbei vor allem,
dass man sich aufgrund des Open--Source Modells an keinen Hersteller bindet
(Stichwort: *Vendor Lock--In*) und keine behördlichen Daten in der »Cloud«
landen. Eine praktische Anwendung im universitären Bereich wäre die Verteilung
von Studienunterlagen an die Studenten. Mangels einer »Standardlösung« ist es
heutzutage schwierig Dokumente sicher mit Behörden auszutauschen. ``brig``
könnte hier einen »Standard« etablieren und in Zukunft als eine »Plattform«
dienen, um beispielsweise medizinische Unterlagen mit dem Hospital auszutauschen.

### Berufsgruppen mit hohen Sicherheitsanforderungen 

Hier wären in erster Line Berufsgruppen mit Schweigepflicht zu nennen wie Ärzte,
Notare und Anwälte aber auch Journalisten und politisch verfolgte Aktivisten.
Leider ist zum jetzigen Zeitpunkt keine zusätzliche Anonymisierung vorgesehen,
die es erlauben würde auch die Quelle der Daten unkenntlich zu machen. Dies
könnte allerdings später mit Hilfe des Tor Netzwerks (Tor Onion Routing Projekt)
realisiert werden.

# Stand der Technik

Die Innovation bei unserem Projekt besteht hauptsächlich darin, bekannte
Technologien »neu zusammen zu stecken«, woraus sich viele neue Möglichkeiten
ergeben.  Wie im nächsten Kapitel beleuchtet wird, ist ``brig`` der Versuch
viele gute, bestehende und praxisgeprüfte Ideen in einem konsistenten Konzept zu
vereinen.

## Stand der Wissenschaft

Zwar ist das Projekt stark anwendungsorientiert, doch basiert es auf gut
erforschten Technologien wie verteilten Netzwerken und von u.a. der NIST[^NIST]
zertifizierten kryptografischen Standard-Algorithmen[@everyday_crypto].
Verteilte Netzwerke wurden in den letzten Jahren gut erforscht und haben sich auch in der Praxis
bewährt: Skype ist vermutlich das bekannteste, kommerzielle »distributed Network«.

Allerdings ist uns keine für breite Massen nutzbare Software bekannt, die es
Nutzern ermöglicht selbst ein verteiltes Netzwerk aufzuspannen, um Dateien
auszutauschen. Am nähsten kommen dem die beiden Softwareprojekte
»``Syncthing``« (OpenSource, [^SYNCTHING]) und »``BitTorrent Sync``«
(proprietär, [^BITSYNC]). 

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
Ein näherer Blick zeigt, dass die Softwareprojekte dort oft nur in Teilaspekten
gut funktionieren oder mit anderen unlösbaren Problemen behaftet sind. Manch
andere Software wie ``bazil``[^BAZIL] oder ``infinit``[^INFINIT] ist
vielversprechender, allerdings ebenfalls noch im Entstehen und im Falle von
``infinit`` auch nur teilweise quelloffen.

[^BAZIL]: \url{https://bazil.org}
[^INFINIT]: \url{http://infinit.sh}


## Verschiedene Alternativen

Im Folgenden geben wir eine unvollständige Übersicht über bekannte
Dateisynchronisations--Programme. Davon stehen nicht alle in Konkurrenz zu
``brig``, sind aber zumindest aus Anwendersicht ähnlich. ``brig`` hat sich zum
Ziel gesetzt, die Vorteile der unterschiedlichen Werkzeuge in Punkto Sicherheit
und Benutzerfreundlichkeit zu vereinen, mit dem Versuch die Probleme der
einzelnen Alternative zu minimieren.

#### Dropbox + Boxcryptor

Der vermutlich bekannteste und am weitesten verbreitete zentrale Dienst zur
Dateisynchronisation. Verschlüsselung kann man mit Tools wie ``encfs``
(Open--Source, siehe auch [^ENCFS]) oder dem etwas umfangreicheren, proprietären
*Boxcryptor* nachrüsten. Was das Backend genau tut ist leider das Geheimnis von
Dropbox --- es ist nicht Open--Source. 

[^ENCFS]: Mehr Informationen unter \url{https://de.wikipedia.org/wiki/EncFS}

Die Server von Dropbox stehen in den Vereinigten Staaten, was spätestens seit
den Snowden--Enthüllungen für ein mulmiges Gefühl sorgen sollte. Wie oben
erwähnt, kann diese Problematik durch die Verschlüsselungssoftware *Boxcryptor*
abgemildet werden. Diese kostet aber zusätzlich und benötigt noch einen
zusätzlichen zentralen Keyserver[^KEYSERVER]. Ein weiterer Nachteil ist hier die
Abhängigkeit von der Verfügbarkeit des Dienstes.

[^KEYSERVER]: Mehr Informationen zum Keyserver unter \url{https://www.boxcryptor.com/de/technischer-\%C3\%BCberblick\#anc09}

Technisch nachteilhaft ist vor allem, dass jede Datei »über den Pazifik« hinweg
synchronisiert werden muss, nur um schließlich auf dem Arbeitsrechner 
»nebenan« anzukommen.

#### ownCloud

Aus dieser Problemstellung heraus entstand die Open--Source Lösung *ownCloud*.
Nutzer hosten auf ihren Servern selbst eine ownCloud--Instanz und stellen
ausreichend Speicherplatz bereit. Vorteilhaft ist also, dass die Daten auf den
eigenen Servern liegen. Nachteilig hingegen, dass das zentrale Modell von Dropbox
lediglich auf eigene Server übertragen wird. Einerseits ist ownCloud nicht so
stark wie ``brig`` auf Sicherheit fokusiert, andererseits ist die Installation
eines Serversystems für viele Nutzer eine »große« Hürde und somit zumindest für
den Heimanwender nicht praktikabel.


#### Syncthing

Das 2013 veröffentliche quelloffene *Syncthing* versucht diese zentrale Instanz
zu vermeiden, indem die Daten jeweils von Peer zu Peer übertragen werden. Es ist
allerdings kein vollständiges Peer--to--peer--Netzwerk: Geteilte Dateien liegen
immer als vollständige Kopie bei allen Teilnehmern, welche die Datei haben.
Alternativ ist selektives Synchronisieren von Dateien möglich.

*Syncthing* besitzt bereits eine Art »intelligentes Routing«, d.h. Dateien werden
vom nächstgelegenen Peer mit der höchsten Bandbreite übertragen. Praktisch ist
auch, dass *Syncthing* Instanzen mittels eines zentralen Discovery--Servers
entdeckt werden können. Nachteilig hingegen ist die fehlende
Benutzerverwaltung: Man kann nicht festlegen von welchen Nutzern man Änderungen
empfangen will und von welchen nicht. 

#### BitTorrent Sync

In bestimmten Kreisen scheint auch das kommerzielle und proprietäre *BitTorrent
Sync* beliebt zu sein. Hier wird das bekannte und freie BitTorrent Protokoll zur
Übertragung genutzt. Vom Feature--Umfang ist es in etwa vergleichbar mit
*Syncthing*. Die Dateien werden allerdings noch zusätzlich AES--verschlüsselt
abgespeichert.

Genauere Aussagen kann man leider aufgrund der geschlossenen Natur des Programms
und der eher vagen Werbeprosa nicht treffen. Ähnlich zu *Syncthing* ist
allerdings, dass eine Versionsverwaltung nur mittels eines »Archivordners«
vorhanden ist. Gelöschte Dateien werden schlicht in diesen Ordner verschoben und
können von dort wiederhergestellt werden. 

#### ``git-annex``

Das 2010 erstmals veröffentlichte ``git-annex``[^ANNEX] geht in vielerlei Hinsicht
einen anderen Weg. Einerseits ist es in der funktionalen Programmiersprache
Haskell geschrieben, andererseits nutzt es das Versionsverwaltungssystem ``git``[@git],
um die Metadaten zu den Dateien abzuspeichern, die es verwaltet. Auch werden
Dateien standardmäßig nicht automatisch synchronisiert, hier ist die Grundidee
die Dateien selbst zu »pushen«, beziehungsweise zu »pullen«.

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

*Zusammenfassung:* Obwohl ``brig`` eine gewisse Ähnlichkeit mit verteilten
Dateisystemen, wie *GlusterFS* hat, wurden diese in der Übersicht weggelassen
--- einerseits aus Gründen der Übersicht, andererseits weil diese andere Ziele
verfolgen und von Heimanwendern kaum genutzt werden. Zudem ist der
Vollständigkeit halber auch OpenPGP zu nennen, was viele Nutzer zum
Verschlüsseln von E-Mails benutzen. Aber auch hier ist der größte Nachteil die
für den Ottonormalbenutzer schwierige Einrichtung und Benutzung.

\newpage

Zusammengefasst findet sich hier noch eine tabellarische Übersicht mit den aus
unserer Sicht wichtigsten Eigenschaften: 

**Technische Aspekte:**

|                      | **Dezentral**       | **Verschlüsselung (Client)**     | **Versionierung**                      |  **Quotas**       | **N-Kopien**    |  
| -------------------- | ------------------- | -------------------------------- | -------------------------------------- | ------------------|------------------|
| *Dropbox/Boxcryptor* | \xmark              | \xmark                           | \textcolor{YellowOrange}{Rudimentär}   |  \xmark           | \xmark          |
| *ownCloud*           | \xmark              | \xmark                           | \textcolor{YellowOrange}{Rudimentär}   |  \xmark           | \xmark          |
| *Syncthing*          | \cmark              | \cmark                           | \textcolor{YellowOrange}{Archivordner} |  \xmark           | \xmark          |
| *BitTorrent Sync*    | \cmark              | \cmark                           | \textcolor{YellowOrange}{Archivordner} |  \xmark           | \xmark          |
| ``git-annex``        | \cmark              | \cmark                           | \cmark                                 |  \xmark           |  \cmark         |
| ``brig``             | \cmark              | \cmark                           | \cmark                                 |  \cmark           |  \cmark         |


**Praktische Aspekte:**

|                      | **FOSS**            | **Einfach nutzbar** | **Einfache Installation**  | **Intelligentes Routing** | **Kompression** |
| -------------------- | ------------------- | ------------------- |--------------------------  | ------------------------- |-----------------|
| *Dropbox/Boxcryptor* | \xmark              | \cmark              | \cmark                     |  \xmark                   | \xmark          |
| *ownCloud*           | \cmark              | \cmark              | \xmark                     |  \xmark                   | \xmark          |
| *Syncthing*          | \cmark              | \cmark              | \cmark                     |  \cmark                   | \xmark          |
| *BitTorrent Sync*    | \xmark              | \cmark              | \cmark                     |  \cmark                   | \xmark          |
| ``git-annex``        | \cmark              | \xmark              | \xmark                     |  \xmark                   | \xmark          |
| ``brig``             | \cmark              | \cmark              | \cmark                     |  \cmark                   | \cmark          |

# Das Projekt »``brig``« 

Optimal wäre also eine Kombination aus den Vorzügen von *Syncthing*, *BitTorrent
Sync* und ``git-annex``. Wie wir die technischen Vorzüge der genannten Lösungen
in einem Produkt vereinen wollen, beleuchten wir in den nächsten Abschnitten.

## Wissenschaftliche und technische Arbeitsziele

Um die oben genannten Ziele zu realisieren ist eine sorgfältige Auswahl der
Technologien wichtig. Der Einsatz eines Content--Adressable--Network (CAN[^CAN])
ermöglicht eine vergleichsweise leichte Umsetzung der oben genannten
Features. Jeder Teilnehmer, der ein Dokument aus dem Netzwerk empfangen will,
muss nur die Prüfsumme des Dokumentes kennen. ``brig`` kann basierend darauf für
alle Dateien, die es kennt eine Historie mit allen Prüfsummen des Dokumentes
speichern. Sobald der Zugriff auf den aktuellen Stand gefordert wird, kann
``brig`` die aktuelle Prüfsumme für den Dateinamen nachschlagen und sie aus dem
*CAN* holen. Alte Stände können, sofern noch im Netzwerk vorhanden, ebenfalls von ``brig``
wiederhergestellt werden, indem die Historie des Dokumentes betrachtet wird.

[^CAN]: Näheres dazu hier: \url{https://en.wikipedia.org/wiki/Content_addressable_network}

Im Vergleich zu zentralen Ansätzen (bei dem der zentrale Server einen *Single
Point of Failure* darstellt) können Dateien intelligent geroutet werden und
müssen nicht physikalisch auf allen Geräten verfügbar sein. Wird beispielsweise
ein großes Festplattenimage (~8GB) in einem Vorlesungssaal von jedem Teilnehmer
heruntergeladen, so muss bei zentralen Diensten die Datei vielmals über das
vermutlich bereits ausgelastete Netzwerk der Hochschule gezogen werden. In einem
*CAN*, kann die Datei in Blöcke unterteilt werden, die von jedem Teilnehmer
gleich wieder verteilt werden können, sobald sie heruntergeladen wurden. Der
Nutzer sieht dabei ganz normal die Datei, ``brig``, bzw. das *CAN* erledigt
dabei das Routing transparent im Hintergrund.

Zudem reicht es prinzipiell wenn eine Datei nur einmal im Netz vorhanden ist.
Ein Rechenzentrum mit mehr Speicherplatz könnte alle Dateien zwischenhalten,
während ein *Thin--Client* nur die Dateien vorhalten muss mit denen gerade
gearbeitet wird.

Da eine gute Balance zwischen Usability und Sicherheit hergestellt werden soll,
muss der Nutzer nach der Einrichtung nur beim Entsperren eines ``brig``--Ordners
(»Repository« genannt) ein Passwort eingeben. Das dahinterliegende
Sicherheitsmodell soll möglichst vom Nutzer versteckt werden. 

## Lösungsansätze

Als *CAN* werden wir das freie *InterPlanetary FileSystem*[^ipfs] nutzen.
Dieses implementiert für uns bereits den Dateiaustausch zwischen den einzelnen
``ipfs``--Knoten. Damit die Dateien nicht nur verschlüsselt übertragen sondern
auch abgespeichert werden, werden sie vor dem Hinzufügen zu ``ipfs`` mittels
AES--128 im GCM--Modus von ``brig`` verschlüsselt und zuvor optional
komprimiert. Zur Nutzerseite hin bietet ``brig`` eine Kommandozeilenanwendung
und ein FUSE-Dateisystem[^FUSE], welches alle Daten in einem ``brig`` Repository
wie normale Dateien in einem Ordner aussehen lässt. Beim »Klick« auf eine Datei
wird diese von ``brig`` dann, für den Nutzer transparent, im Netzwerk
lokalisiert, empfangen, entschlüsselt und als Dateistrom nach außen gegeben. 

[^ipfs]: Mehr Informationen unter \url{http://ipfs.io/}
[^FUSE]: FUSE: *Filesystem in Userspace*, siehe auch \url{https://de.wikipedia.org/wiki/Filesystem_in_Userspace}
[^WEBDAV]: Siehe dazu auch: \url{https://de.wikipedia.org/wiki/WebDAV}

![Übersicht über die Kommunikation zwischen zwei Partnern/Repositories, mit den relevanten Sicherheits--Protokollen](images/security.png){#fig:security}

Jedes Repository besitzt nach der Einrichtung über einen öffentlichen und
privaten Schlüssel. Das Schlüsselpaar ist verfügbar sobald der Nutzer sein
Passwort eingegeben hat und ``brig`` eines mittels ``scrypt`` abgeleiteten 
Schlüssels das Repository »öffnet«.

Im Netzwerk identifiziert sich ein Knoten dabei über den Hash--Wert des
öffentlichen Schlüssels und weist seine Identität über ein
Challenge--Response--Verfahren nach. Da dieser Hash--Wert natürlich 
nur schwer vom Benutzer zu merken ist, vergibt dieser beim Anlegen eines
Repositories einen Benutzernamen im Jabber--ähnlichen Schema 
``nutzer@gruppe.domain/ressource``. Dabei sind alle Teile nach dem ``@``
optional. Die Gruppe kann dabei von Unternehmen genutzt werden andere
``brig``--Teilnehmer automatisch zu entdecken.

Alle Änderungen an einem Repository werden in einer Metadatendatenbank
gespeichert. Diese wird dann mit anderen Teilnehmern über einen separaten,
verschlüsselten und authentifizierten Seitenkanal ausgetauscht. Sowohl der
Seitenkanal als auch die eigentliche Dateiübertragung kann dabei NAT--Grenzen
mittels Hole--Punching[^HOLE] überwinden und benötigt normalerweise keine
gesonderte Konfiguration.

Jeder Teilnehmer hat ähnliche wie bei ``git`` den gesamten Dateiindex jedes
anderen Teilnehmers. Die eigentlichen Dateien können aber »irgendwo« im
Teilnehmernetz sein. Sollte eine Datei lokal benötigt werden, so kann man sie
»pinnen«, um sie lokal zu speichern. Ansonsten werden nur selbst erstellte
Dateien gespeichert und andere Dateien maximal solange vorgehalten, bis die
Speicherquote erreicht ist.

[^HOLE]: Siehe dazu auch \url{https://de.wikipedia.org/wiki/Hole_Punching}

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
den Einsatz eines anderen *CANs* zu ermöglichen.

Erfahrungsgemäß nimmt auch die Portierung und Wartung auf anderen Plattformen
sehr viel Zeit in Anspruch. Durch die Wahl der hochportablen Programmiersprache
Go minimieren wir dies drastisch.

Wie für jede sicherheitsrelevante Software ist die Zeit natürlich ein Risiko.
Ein Durchbruch im Bereich der Quantencomputer könnte daher in absehbarer
Zeit zu einem Sicherheitsrisiko werden.

# Wirtschaftliche Verwertung

## Open--Source--Lizenz und Monetarisierung

Als Lizenz für ``brig`` soll die Copyleft--Lizenz ``AGPLv3`` zum Einsatz kommen.
Diese stellt sicher, dass Verbesserungen am Projekt auch wieder in dieses
zurückfließen müssen.

Dass die Software quelloffen ist, ist kein Widerspruch zur wirtschaftlichen
Verwertung. Statt auf Softwareverkäufe zu setzen lässt sich mit dem Einsatz und
der Anpassung der Software Geld verdienen.  Das Open--Source Modell bietet aus
unserer Sicht hierbei sogar einige grundlegende Vorteile:

- Schnellere Verbreitung durch fehlende Kostenbarriere auf Nutzerseite.
- Kann von Nutzern und Unternehmen ihren Bedürfnissen angepasst werden.
- Transparenz in Punkto Sicherheit (keine offensichtlichen Backdoors möglich).
- Fehlerkorrekturen, Weiterentwicklung und Testing aus der Community.

## Verwertungskonzepte

Es folgen einige konkrete Verwertungs--Strategien, die teilweise auch in
Partnerschaft mit dazu passenden Unternehmen oder Institutionen ausgeführt
werden könnten. Prinzipiell soll die Nutzung für private und gewerbliche Nutzer
kostenfrei sein, weitergehende Dienstleistungen aber nicht.

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
machen. Ein Beispiel wäre ein Skript, dass für jeden Unternehmens--Login ein
``brig``--Repository lokal mit passenden Nutzernamen und Passwort anlegt
(beispielsweise mittels LDAP--Anbindung).

### Supportverträge und Consulting

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

Für Unternehmen, die unsere Software als Teil ihres eigenen Angebots nutzen
wollen, kann die Erteilung einer anderen Lizenz in Frage kommen:

- Eine Consulting Firma könnte eine Lizenz bei uns erwerben, um selbst
  Speziallösungen zu entwickeln, die sie dann nicht als *Open--Source*
  veröffentlichen müssen.

- Ein Hosting Anbieter der ``brig`` nutzen möchte, müsste wegen der ``AGPLv3``
  dazu erst die Erlaubnis bei uns einholen.  Je nach Fall könnte dann ein
  entsprechender Vertrag ausgehandelt werden.

### Zertifizierte NAS-Server

Besonders für Privatpersonen oder kleine Unternehmen wie Ingenieurbüros wäre
ein vorgefertigter Rechner mit vorinstallierter Software interessant. Das
Software- und Hardware--Zusammenspiel könnte dann vorher von uns 
getestet werden und mit der Zeit auch dem technischen Fortschritt angepasst
werden.

# Beschreibung des Arbeitsplans

## Technische Arbeitsschritte

Im Rahmen unserer Masterarbeiten werden wir ein Proof--of--Concept entwickeln, der
bereits in Grundzügen die oben beschriebenen Technologien demonstriert.
Gute Performanz, Portabilität und Anwenderfreundlichkeit sind zu diesem Zeitpunkt aus
Zeitmangel allerdings noch keine harten Anforderungen.
In diesem ersten Prototypen sollen bereits die grundlegenden Features von
``brig`` vorhanden sein:  die verschlüsselte Speicherung in ``ipfs``, der
Zugriff über FUSE und grundlegende Synchronisationsfähigkeiten.

Die dabei gewonnen Erkenntnisse wollen wir dazu nutzen, einen Prototypen für
technische Nutzer zu entwickeln und dabei nötigenfalls eine »Kurskorrektur«
durchzuführen, wobei nach Möglichkeit die Codebasis bereits vereinfacht und
stabilisiert werden sollte. Zu diesem Zeitpunkt soll ``brig`` bereits ähnlich
wie ``git`` als »Toolbox« für Dateisynchronisation benutzbar sein.

In einem zweiten Endnutzer--orientierten  Prototypen werden dann in kleinen
Iterationen Features hinzugefügt. Jedes dieser Feature sollte für sich alleine
stehen, daher sollte zu diesem Zeitpunkt bereits die grundlegende Architektur
relativ stabil sein.

Nach einer gewissen Stabilisierungsphase wollen wir ein erstes öffentliches
Release in der Open--Source--Community anstrebgen. Dies hätte bereits eine
gewisse Verbreitung zur Folge und die in ``brig`` eingesetzten
Sicherheitstechnologien könnten von externen Sicherheitsexperten auditiert
werden.

## Meilensteinplanung

Der oben stehende Zeitplan ist nochmal in Abbildung {@fig:milestones} als
Gantt--Diagramm auf drei Jahre gerechnet zu sehen. Dort sind auch die jeweiligen
Features eingetragen, die wir uns für den jeweiligen Entwicklungsstand wünschen.
Bestimmte Aufgaben wie Tests und Benchmarks werden iterativ bei jedem
Meilenstein wiederholt und werden nicht explizit aufgeführt.

![Grobe Meilensteinplanung von 2016 bis 2019.](images/gantt.png){#fig:milestones}

# Über uns

Wir sind zwei Master--Studenten an der Hochschule Augsburg, die von freier
Software begeistert sind und mit ihr die Welt ein bisschen besser machen wollen.
Momentan entwickeln wir ``brig`` im Rahmen unserer Masterarbeiten bei Prof
Dr.-Ing. Thorsten Schöler in der Distributed--Systems--Group[^DSG]. 
Wir haben beide Erfahrung darin Open--Source--Software zu entwickeln und zu
betreuen, weswegen wir das nun auch gerne hauptberuflich fortführen würden.

Unsere momentanen sonstigen Projekte finden sich auf GitHub:

* \url{https://github.com/sahib} (Projekte von Christopher Pahl)
* \url{https://github.com/qitta} (Projekte von Christop Piechula)
* \url{https://github.com/studentkittens} (gemeinsame Projekte und Studienarbeiten)

[^DSG]: Siehe auch: \url{http://dsg.hs-augsburg.de/}

## Der Name

Eine »Brigg« (englisch »brig«) ist ein kleines und wendiges
Zweimaster--Segelschiff aus dem 18-ten Jahrhundert. Passend erschien uns der
Name einerseits, weil wir flexibel »Güter« (in Form von Dateien) in der ganzen
Welt verteilen, andererseits weil ``brig`` auf (Datei-)Strömen operiert.

Dass der Name ähnlich klingt und kurz ist wie ``git``, ist kein Zufall. Das
Versionsverwaltungssystem (version control system, kurz VCS) hat durch seine
sehr flexible und dezentrale Arbeitsweise bestehende zentrale Alternativen wie
``svn`` oder ``cvs`` fast vollständig abgelöst. Zusätzlich ist der Gesamteinsatz
von Versionsverwaltungssystemen durch die verhältnismäßige einfache Anwendung
gestiegen. Wir hoffen mit ``brig`` eine ähnlich flexible Lösung für große
Dateien etablieren zu können. 

## Aktueller Projektstatus

Wir sind bereits auf gutem Wege das Proof--of--Concept fertig zustellen. Die
momentane Codebasis unterstützt bereits Verschlüsselung, Streaming--Kompression,
ein Daemon--Server Modus mit Kommandozeilen--Anwendung, ein FUSE--Dateisystem
und einen authentifizierten und verschlüsselten Seitenkanal, um Metadaten
auszutauschen.

Die aktuelle Entwicklung ist öffentlich und kann auf GitHub verfolgt werden:

* \url{https://github.com/disorganizer/brig}

## Was ``brig`` *nicht* ist

Auch wenn ``brig`` sehr flexibel einsetzbar ist, ist und soll es keineswegs die
beste Alternative in allen Bereichen sein. Keine Software kann eine »eierlegende
Wollmilchsau« sein und sollte auch nicht als solche benutzt werden.

Besonders im Bereich Effizienz kann es nicht mit hochperformanten
Cluster--Dateisystemen wie Ceph[^CEPH] oder GlusterFS[^GLUSTER] mithalten.  Das
liegt besonders an der sicheren Ausrichtung von ``brig``, welche oft
Rechenleistung zugunsten von Sicherheit eintauscht. Auch kann ``brig`` keine
Echtzeit--Garantien geben. 

Auch wenn ein ``brig``--Repository in der geschlossenen Form als sicherer
»Datensafe« einsetzbar ist, so bietet ``brig`` nicht die Eigenschaft der
»glaubhaften Abstreitbarkeit«[^ABSTREIT], die Werkzeuge wie Veracrypt bieten. 

Im Gegensatz zu Versionsverwaltungssystemen wie ``git``, kann ``brig`` keine
Differenzen zwischen zwei Ständen anzeigen, da es nur auf den Metadaten von
Dateien arbeitet. Auch muss auf der Gegenseite ein ``brig``--Daemon--Prozess
laufen, um mit der Gegenseite zu kommunizieren.

[^CEPH]: \url{http://ceph.com}
[^GLUSTER]:  \url{https://www.gluster.org}
[^ABSTREIT]: \url{https://de.wikipedia.org/wiki/VeraCrypt\#Glaubhafte_Abstreitbarkeit}

# Finanzierung des Vorhabens

Die Entwicklung von ``brig`` ist sehr zeitintensiv, daher ist eine solide
Finanzierung unerlässlich. Um eine freie und kontinuierliche Entwicklung in einem
akademischen Umfeld zu gewährleisten, streben wir an als wissenschaftliche
Mitarbeiter im Hochschulbereich angestellt zu werden. Da die Hochschule Augsburg
allerdings nicht über die Mittel verfügt zwei neue Stellen von Grund auf zu
finanzieren, benötigen wir ein oder mehrere Sponsoren. Dabei sind wir für alle
Optionen offen, im Folgenden stellen wir aber eine auf Unternehmen
zugeschnittene Kooperationsmöglichkeit vor. Der gewünschte Förderungsbeginn wäre
im jedem Fall Sept/Okt. 2016 und würde optimalerweise über drei Jahre gehen.

## Mittels IuK--Bayern

Eine mögliche Finanzierungstrategie bietet das IuK--Programm[^IUK] des
Freistaates Bayern. Dabei werden Kooperation zwischen Fachhochschulen und
Unternehmen mit bis zu 50% des Fördervolumens vom Freistaat Bayern gefördert.
Vom IuK--Programm gern gesehen ist dabei ein Großunternehmen, welches zusammen
mit einem kleinen bis mittleren Unternehmen (``KMU``) das Fördervolumen
aufbringt. Aus diesen Mitteln könnte die Hochschule Augsburg dann bis zu zwei
volle Stellen für wissenschaftliche Mitarbeiter über eine gewisse Dauer finanzieren.

Konkret berechnet sich das dabei folgedermaßen: Ein oder mehr Unternehmen
bringen ein gewissen Betrag auf mit denen sie interne Arbeitskräfte bezahlen
die an dem gemeinsamen Kooperationsprojekt arbeiten. Dieser Betrag wird dann
vom Freistaat Bayern verdoppelt. Von der zweiten Hälfte werden dann
wissenschaftliche Mitarbeiter an der Hochschule Augsburg bezahlt. Bleibt ein
Überschuss übrig, so fließt dieser zurück an die Unternehmen.

Steigt ein einzelnes Unternehmen in das IuK--Programm mit ein, so
ergeben sich beispielsweise folgende grob gerechnete Möglichkeiten:

* Ein wissenschaftlicher Mitarbeiter in Halbzeit an der HSA und ein Mitarbeiter
  in Vollzeit beim Unternehmen plus 10.000 Euro Sachmittel benötigt ein Betrag
  von etwa 105.000 Euro auf Unternehmensseite. An das Unternehmen fließen keine
  Fördergelder zurück.
* Ein wissenschaftlicher Mitarbeiter in Vollzeit an der HSA und zwei Mitarbeiter
  in Vollzeit beim Unternehmen plus 10.000 Euro Sachmittel benötigt ein Betrag 
  von etwa 210.000 Euro auf Unternehmensseite. Das Unternehmen erhält eine etwa
  30%-ige Förderquote.
* Zwei wissenschaftliche Mitarbeiter in Vollzeit an der HSA und zwei Mitarbeiter
  in Vollzeit beim Unternehmen plus 20.000 Euro Sachmittel benötigt ein Betrag 
  von etwa 420.000 Euro auf Unternehmensseite. Das Unternehmen erhält eine etwa 
  20%-ige Förderquote.

Dieses Exposé enthält bereits alle Informationen und Textbausteine zur
Formulierung eines IuK--Antrags. 

[^IUK]: Mehr Informationen unter \url{http://www.iuk-bayern.de/}

\newpage

# Literaturverzeichnis
