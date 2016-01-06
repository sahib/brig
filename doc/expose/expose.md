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
Alternative zu Cloud-Storage Lösungen wie Dropbox, die sowohl für Unternehmen
als auch für Heimanwender nutzbar ist. Trotz der Prämisse, einfache Nutzbarkeit
zu gewährleisten, wird auf Sicherheit sehr großen Wert gelegt.  Aus Gründen der
Transparenz wird die Software mit dem Namen ``brig`` dabei quelloffen unter der
AGPLv3 Lizenz entwickelt.

Nutzbar soll das resultierende Produkt, neben dem Standardanwendungsfall der
Dateisynchronisation, auch als Backup- bzw. Archivierungs-lösung sein
beziehungsweise auch als verschlüsselter Daten--Safe oder als Plattform für
andere, verteilte Anwendungen aus dem Industrie 4.0 Umfeld.

# Projektsteckbrief

## Einleitung

Viele Unternehmen haben sehr hohe Ansprüche an die Sicherheit, welche zentrale
Alternativen wie Dropbox[^Dropbox] nicht bieten können. Zwar wird die
Übertragung von Daten zu den zentralen Dropbox-Servern verschlüsselt, was
allerdings danach mit den Daten »in der Cloud« passiert liegt nicht mehr in
der Kontrolle der Nutzer. Dort sind die Daten schonmal für andere Nutzer wegen
Bugs einsehbar oder werden gar von Dropbox an amerikanische Geheimdienste
weitergegeben. 

[^Dropbox]: Mehr Informationen unter \url{https://www.dropbox.com/}

[Sprichwörtlich, regnen die Daten irgendwo anders aus der Cloud ab.] Tools wie
Boxcryptor[^Boxcryptor] lindern diese Problematik zwar etwas, heilen aber nur
die Symptome, nicht das zugrunde liegende Problem.

[^Boxcryptor]: Krypto-Layer für Cloud-Dienste, siehe \url{https://www.boxcryptor.com/de}

Dropbox ist leider kein Einzelfall -- beinahe alle Cloud--Dienste haben, oder
hatten, architektur-bedingt ähnliche Sicherheitslecks. Für ein Unternehmen wäre
es vorzuziehen ihre Daten auf Servern zu speichern, die sie selbst
kontrollieren. Dazu gibt es bereits einige Werkzeuge wie ownCloud oder ein
Netzwerkverzeichnis wie Samba, doch technisch bilden diese nur die zentrale
Architektur von Cloud--Diensten innerhalb eines Unternehmens ab. 

## Ziele

Ziel ist die Entwicklung einer sicheren, dezentralen und unternehmenstauglichen
Dateisynchronisationssoftware names ``brig``. Die »Tauglichkeit« für ein
Unternehmen ist sehr variabel. Wir meinen damit im Folgenden diese Punkte:

- Einfach Benutzbarkeit für nicht-technische Angestellte: Sichtbar soll nach der
  Einrichtung nur ein simpler Ordner im Dateimanager sein.
- Schnelle Auffindbarkeit der Dateien durch Verschlagwortung.
- Effiziente Übertragung von Dateien: Intelligentes Routing vom Speicherort zum Nutzer.
- Speicherquoten: Nicht alle Dateien müssen synchronisiert werden.
- Automatische Backups: Versionsverwaltung auf Knoten mit großem Speicherplatz.

Um eine solche Software entwickeln, wollen wir auf bestehende Komponenten wie
InterPlanetaryFileSystem (ein konfigurierbares P2P Netzwerk[@peer2peer]) und
XMPP (ein Messenging Protokoll und Infrastruktur, siehe [@xmpp]) aufsetzen. Dies
erleichtert unsere Arbeit und macht einen Prototyp der Software erst möglich. 

Von einem Prototypen zu einer marktreifen Software ist es allerdings stets ein
weiter Weg. Daher wollen wir einen großen Teil der darauf folgenden Iterationen
damit verbringen, die Software bezüglich Sicherheit, Performance und einfacher
Benutzerfreundlichkeit zu optimieren. Da es dafür keinen standardisierten Weg
gibt, ist dafür ein großes Maß an Forschung nötig.

## Einsatzmöglichkeiten

``brig`` soll deutlich flexibler nutzbar sein als beispielsweise zentrale
Dienste. Nutzbar soll es unter anderem sein als…

- …**Synchronisationslösung**: Spiegelung von 2-n Ordnern.
- …**Transferlösung**: »Veröffentlichen« von Dateien nach Außen mittels Hyperlinks.
- …**Versionsverwaltung**: 
  Alle Zugriffe an eine Datei werden aufgezeichnet.
  Bis zu einer bestimmten Tiefe können alte Dateien abgespeichert werden.
- …**Backup- und Archivierungslösung**: Verschiedene Knoten Typen möglich.
- …**verschlüsselten Safe**: ein »Repository« kann »geschlossen« werden.
- …**Semantisch durchsuchbares** tag basiertes Dateisystem[^TAG].
- …als **Plattform** für andere Anwendungen.
- …einer beliebigen Kombination der oberen Punkte.

[^TAG]: Mit einem ähnlichen Ansatz wie \url{https://en.wikipedia.org/wiki/Tagsistant}

## Zielgruppen

``brig`` zielt hauptsächlich auf Unternehmenskunden und Heimanwender.
Daneben sind aber auch noch andere Zielgruppen denkbar.

### Unternehmen

Großunternehmen können ``brig`` nutzen, um ihre Daten und Dokumente intern zu
verwalten. Besonders sicherheitskritische Dateien entgehen so der Lagerung in
Cloud Services oder der Gefahr von zig Kopien auf Mitarbeiter-Endgeräten.
Größere Unternehmen verwalten dabei meist ein Rechenzentrum auf den
firmeninterne Dokumente gespeichert werden. Diese werden dann meist mittels
ownCloud, Samba o.ä. von den Nutzern »manuell« heruntergeladen. 

In diesem Fall könnte man ``brig`` im Rechenzentrum und allen Endgeräten installieren.
Das Rechenzentrum würde die Datei mit tiefer Versionierung und vollem Caching vorhalten.
Endanwender würden alle Daten sehen, aber auf ihren Gerät nur die Daten tatsächlich
speichern, die sie auch benutzen. Hat ein Kollege im selben Büro beispielsweise die
Datei bereits kann ``brig`` sie dann auch teilweise von ihm holen.

Kleinunternehmen wie Ingenieurbüros können ``brig`` dazu nutzen Dokumente nach
außen freizugeben, ohne dass sie dazu vorher irgendwo »hochgeladen« werden
müssen. 

### Privatpersonen / Heimanwender

Heimanwender können ``brig`` für ihren Datenbestand aus Fotos, Filmen, Musik und
sonstigen Dokumenten nutzen. Ein typischer Anwendungsfall wäre dabei auf einem
NAS Server, der alle Dateien mit Versionierung speichert. Endgeräte wie Laptops
und Smartphones würde dann ebenfalls ``brig`` nutzen, aber mit deutlich
geringeren Speicherquotas.

### Plattform für industrielle Anwendungen

Da ``brig`` auch komplett automatisiert ohne Interaktion nutzbar sein soll,
kann es auch als Plattform für jede andere Anwendungen genutzt werden, die Dateien
austauschen und synchronisieren müssen.

Eine Anwendung in der Industrie 4.0 wäre beispielweise die Synchronisierung von Konfigurationsdateien im gesamten Netzwerk.

### Einsatz im öffentlichen Bereich

Aufgrund seiner Transparenz und einfachen Benutzbarkeit wäre ebenfalls eine
Nutzung an Schulen, Universitäten oder auch in Behörden zum Dokumentenaustausch
denkbar. Vorteilhaft wäre hierbei vor allem, dass man sich aufgrund des
Open--Source Modells an keinen Hersteller bindet (Stichwort: Vendor Lock) und
keine behördlichen Daten in der »Cloud« landen.

## Innovation

Die Innovation bei unserem Projekt  besteht daher darin bekannte Technologien
»neu zusammen zu stecken«, woraus sich viele neue Möglichkeiten ergeben.

TODO: Zu kurz?

# Stand der Technik

## Stand der Wissenschaft

Zwar ist das Projekt stark anwendungsorientiert, doch basiert es auf gut
erforschten Technologien wie Peer-to-Peer-Netzwerken (kurz P2P, siehe auch
[@peer2peer_arch]), von der NIST[^NIST] zertifizierten kryptografischen
Standard-Algorithmen[@everyday_crypto] und verteilten Systemen im Allgemeinen
(wie der freie XMPP Standard). Peer to Peer Netzwerke wurden in den letzten
Jahren gut erforscht und haben sich auch in der Praxis bewährt (Skype ist ein
Beispiel für ein kommerzielles P2P Netzwerk, siehe
auch @peer2peer, S.2). 

Allerdings ist uns keine für breite Massen nutzbare Software bekannt, die es
Nutzern ermöglicht selbst ein P2P Netzwerk aufzuspannen und darin Dateien
auszutauschen. Am nähsten kommen dabei die beiden Softwareprojekte
``Syncthing`` (OpenSource) und ``BitTorrent Sync`` (proprietär).

Der wissenschaftliche Beitrag unserer Arbeit wäre daher die Entwicklung einer
freien Alternative, die von allen eingesehen, auditiert und studiert werden
kann. Diese freie Herangehensweise ist insbesondere für sicherheitskritische
Software relevant, da keine offensichtlichen »Exploits« in die Software
eingebaut werden können.

[^NIST]: NIST: *National Institute of Standards and Technology*

## Markt und Wettbewerber

Bereits ein Blick auf Wikipedia[@wiki_filesync] zeigt, dass der momentane Mark an
Dateisynchronisationssoftware (im weitesten Sinne) sehr unübersichtlich ist.

Bei einem näheren Blick stellt sich oft heraus, dass die Software dort oft nur
in Teilaspekten gut funktioniert oder andere nicht lösbare Probleme besitzt.

### Verschiedene Alternativen

Im Folgenden geben wir eine Auswahl von bekannten Dateisynchronisations
Softwares im weitesten Sinne. Nicht alle stehen davon in direkter Konkurrenz
zu ``brig``, aber viele Usecases überlappen sich.

#### Dropbox + Boxcryptor

+ Was ist eigentlich gut?
+ Verbreitet und bekannt.

- Zentrale Lösung
- proprietär

#### Owncloud

+ Daten liegen auf eigenen Servern.

- Zentrale Lösung.
- Zugriff über Weboberfläche

#### Syncthing

Heimanwender. Immer physikalische Kopie?

+ Open Source
+ Einfacher Ordner auf Dateisystemebene.

- Keine Benutzerverwaltung
- kein p2p netzwerk
- zentraler key server

#### BitTorrent Sync

Unternehmensanwender

+ p2p netzwerk
+ verschlüsselte Speicherung

- proprietär und kommerziell
- Keine Benutzerverwaltung
- Versionsverwaltung nur als »Archiv-Folder«

#### Git--annex

Basierend auf git[@git]

+ sehr featurereich 
+ special remotes
+ Open Source
+ n-copies

- kein p2p netzwerk
- Selbst für erfahrene Benutzer nur schwierig zu benutzen

### Zusammenfassung

Obwohl ``brig`` eine gewisse Ähnlichkeit mit verteilten Dateisystemen wie
GlusterFS hat, wurden diese oben wegelassen -- einerseits aus Gründen der
Übersicht, andererseits weil diese kaum einfach von Heimanwendern genutzt werden.

Zusammengefasst findet sich hier noch eine tabellarische Übersicht mit den aus
unserer Sicht wichtigsten Eigenschaften:

|                      | **FOSS**[^FOSS]     | **Dezentral**       | **No SPoF**[^SPOF]          | **VCS**[^VCS]                          | **Einfach nutzbar** |  
| -------------------- | ------------------- | ------------------- | --------------------------- | -------------------------------------- | ------------------- |
| *Dropbox/Boxcryptor* | \xmark              | \xmark              | \xmark                      | \textcolor{YellowOrange}{Rudimentär}   | \cmark              |
| *ownCloud*           | \cmark              | \xmark              | \xmark                      | \textcolor{YellowOrange}{Rudimentär}   | \cmark              |
| *Syncthing*          | \cmark              | \cmark              | \xmark [^syncthing_key]     | \textcolor{YellowOrange}{Archivordner} | \cmark              |
| *BitTorrent Sync*    | \xmark              | \cmark              | \cmark                      | \textcolor{YellowOrange}{Archivordner} | \cmark              |
| ``git annex``        | \cmark              | \cmark              | \cmark                      | \cmark                                 | \xmark              |
| ``brig``             | \cmark              | \cmark              | \cmark                      | \cmark                                 | \cmark              |

[^FOSS]: Free Open Source Software
[^SPOF]: Single Point of Failure
[^VCS]: Version Control System um alte Stände wiederherzustellen
[^syncthing_key]: *Syncthing* benutzt einen zentralen Keyserver.

# Das Projekt

Optimal wäre also eine Kombination aus den Vorzügen von *Syncthing*,
*BitTorrent Sync* und ``git annex``. Unser Versuch diese Balance hinzubekommen
heißt ``brig``.

## Der Name

- Brig operiert auf (Datei-)Strömen
- Eine Brig ist ein Handelsschiff dass Waren in die ganze Welt liefer kann.

Dass der Name ähnlich kurz ist und klingt wie ``git`` ist kein Zufall. Das
Versionsverwaltungssystem hat durch seine sehr flexible und dezentrale
Arbeitsweise bestehende zentrale Alternativen wie ``svn`` oder ``cvs`` fast
vollständig abgelöst. Zusätzlich ist der Gesamt-einsatz von
Versionsverwaltungssystemen durch die verhältnismäßige einfache Anwendung
gestiegen.

Wir hoffen mit ``brig`` eine ähnlich flexible Lösung für große Dateien
etablieren zu können. 

## Wissenschaftliche/technische Arbeitsziele

Um die oben genannten Ziele zu realisieren ist eine sorgfältige Auswahl der
Technologien wichtig. Der Einsatz eines Peer-to-Peer Netzwerk zum Dateiaustausch
ermöglicht interessante neue Möglichkeiten. Bei zentralen Ansätzen müssen
Dateien immer vom zentralen Server (Single Point of Failure) geholt werden. Dies
ist relativ ineffizient, besonders wenn viele Teilnehmer im selben Netz die selbe
große Datei empfangen wollen. Bei ``brig`` würde der Fortschritt beim Ziehen der
Datei unter den Teilnehmern aufgeteilt werden. Hat ein Teilnehmer bereits ein
Block einer Datei, so kann er sie mit anderen direkt ohne Umweg über den
Zentralserver teilen.

Zudem reicht es prinzipiell wenn eine Datei nur einmal im Netz vorhanden ist.
Ein Rechenzentrum mit mehr Speicherplatz könnte alle Dateien zwischenhalten,
während ein Thin-Client nur die Dateien vorhalten muss mit denen gerade
gearbeitet wird.

Unsere technischen Ziele sind daher stichpunkthaft:

* Kein Single Point of Failure
* einfache Benutzung und Installation
* Verschlüsselte Übertragung und Speicherung.
* Kompression: Optional mittels snappy.
* Deduplizierung: Eine selbe Datei wird nur einmal im Netz gespeichert.
* Speicherquoten & Pinning (Thin-Client vs. Storage Server)
* Versionierung mit definierbarer Tiefe.
* Benutzerverwaltung mittels XMPP.
* Zweifaktor-Authentifizierung und paranoide Sicherheit made in Germany.
* N-Copies? 

TODO: Ausformulieren.

## Lösungsansätze

Als Peer-to-Peer Filesystem werden wir das InterPlanetaryFileSystem[^IPFS] nutzen.
Dieses implementiert für uns bereits den Dateiaustausch zwischen
den einzelnen IPFS Knoten. Damit die Dateien nicht nur verschlüsselt übertragen
werden, werden sie vor dem Hinzufügen zu IPFS mittels AES im AEAD Modus von
``brig`` verschlüsselt und optional komprimiert. Zur Nutzerseite hin bietet
``brig`` dann eine Kommandozeilenanwendung und ein FUSE-Dateisystem, welches
alle Daten in einem ``brig`` Repository wie normale Dateien in einem Ordner
aussehen lässt. Beim »Klick« auf eine Datei wird diese von ``brig`` dann, für
den Nutzer unsichtbar, im Netzwerk lokalisiert, empfangen, entschlüsselt und
nach außen gegeben.

[^IPFS]: Mehr Informationen unter \url{http://ipfs.io/}

Der AES Schlüssel wird dabei an ein Passwort geknüpft, welches der Nutzer beim
Anlegen des Repositories angibt. Das Passwort wiederum ist an einen
XMPP-Account der Form ``nutzer@server.de/ressource`` geknüpft.

Alle Änderungen an einem Repository werden in einer Metadatendatenbank
gespeichert. Diese kann dann mit anderen Teilnehmern über XMPP+OTR ausgetauscht
werden. Jeder Teilnehmer hat dadurch den gesamten Dateiindex, die eigentlichen
Dateien können aber »irgendwo« im Teilnehmernetz sein. Sollte eine Datei lokal
benötigt werden, so kann man sie »pinnen«, um sie lokal zu speichern.
Ansonsten werden nur selbst erstellte Dateien gespeichert und andere Dateien
maximal solange vorgehalten, bis die Speicherquote erreicht ist.

Nutzer die ``brig`` nicht installiert haben, oder mit denen man aus
Sicherheitsgründen nicht das gesamte Repository teilen möchte können einzelne
Dateien ganz normal aus ihrem Browser heraus downloaden. Dazu muss die Datei
vorher »publik« gemacht werden. Der außenstehende Nutzer kann dann die Datei
über ein von ``brig`` gestelltes »Gateway« die Datei von einem Rechner ziehen,
der von außen erreichbar ist.

Um Portabilität zu gewährleisten wird die Software in
Go[@go_programming_language] geschrieben sein. Der Vorteil hierbei ist, dass am
Ende eine einzige sehr portable, statisch gelinkte Binärdatei erzeugt wird.
Weitere Vorteile sind die hohe Grundperformanz und die sehr angenehmen
Werkzeuge, die mit der Sprache mitgeliefert werden.

## Technische Risiken 

Der Aufwand für ein Softwareprojekt dieser Größe ist schwer einzuschätzen. Da
wird auf relativ junge Technologien wie ``ipfs`` setzen, ist zu erwarten dass
sich in Details noch Änderungen ergeben. Auch die Tauglichkeit bezüglich
Performance ist momentan noch schwer einzuschätzen. Aus diesen Gründen werden
wir zwischen ``brig`` und ``ipfs`` eine Abstraktionsschicht bauen, um notfalls
andere Backends einzusetzen.

Wie für jede sicherheitsrelevante Software ist die Zeit natürlich ein Risiko.
Ein Durchbruch im Bereich von Quantencomputer könnte beispielsweise ein
Sicherheitsrisiko darstellen.

Erfahrungsgemäß nimmt auch die Portierung und Wartung auf anderen Plattformen
sehr viel Zeit in Anspruch. Durch die Wahl der Programmiersprache Go minimieren
wir dies drastisch.

# Wirtschaftliche Verwertung

Als Lizenz für ``brig`` soll die Copyleft--Lizenz ``AGPL``. Diese stellt sicher,
dass Verbesserungen am Projekt auch wieder in dieses zurückfließen müssen.

Dass die Software quelloffen ist ist kein Widerspruch zu wirtschaftlicher
Verwertung. Statt auf Softwareverkäufe zu setzen lässt sich mit dem Einsatz und
der Anpassung der Software Geld verdienen.

Open--Source bietet aus unserer Sicht daher einige Vorteile:

- Schnellere Verbreitung.
- Kann von Nutzern und Unternehmen ihren Bedürfnissen angepasst werden.
- Transparenz in Punkto Sicherheit (keine Exploits)

## Wirtschaftliche Verwertung 

Es folgen einige konkrete Verwertung Strategien, die auch in Partnerschaft mit
Unternehmen ausgeführt werden könnten.

### Bezahle Entwicklung spezieller Features

Die Open-Source-Entwickler Erfahrung der Autoren hat gezeigt, dass sich Nutzer
oft ungewöhnliche Features wünschen, die sie oft zu einem bestimmten Termin
brauchen. (TODO: blabla...)

Allgemein sind zwei Szenarien denkbar:

- *Allgemein nützliche Änderungen:*
  Diese werden direkt in ``brig`` integriert und sind daher als Open--Source für
  andere nutzbar.
- *Spezielle Lösungen:* 
  Lösungen die nur für Unternehmens-Anwendungsfälle Sinn machen.
  Beispielsweise ein Skript, dass für jeden Unternehmens-Login einen XMPP
  Account anlegt. 

### Supportverträge

Normalerweise werden Fehler bei Open--Source Berichte auf einen dafür
eingerichten Issue Tracker gemeldet. Die Entwickler können dann, nach einiger
Diskussion und Zeit, den Fehler reparieren. Unternehmen haben aber für
gewöhnliche (kurze) Deadlines bis etwas funktionieren muss.

- Priorisierung bei bug reports.
- Kleinere Anpassungen.
- Persönlicher Kontakt.
- Wartung von nicht-öffentlichen Spezialfeatures
- Installation der Software
- Installation von YubiKeys oder anderer ZweiFaktor

### Mehrfachlizensierung

Beispiele wären:

- Eine Consulting Firma könnte eine Lizenz bei uns erwerben, um selbst
  Speziallösungen zu entwickeln, die sie dann nicht veröffentlichen müssen.

- Ein Hosting Anbieter der ``brig`` nutzen möchte, müsste wegen der AGPL dazu
  erst die Erlaubnis bei uns einholen. Je nach Fall könnte dann ein Vertrag
  ausgehandelt werden.

### Zertifizierte NAS-Server

Besonders für Privatpersonen oder kleine Unternehmen wie Ingenieurbüros wäre
eine vorgefertigter Rechner mit vorinstallierter Software interessant.

### Lehrmaterial und Consulting.

- Schulungen
- Gedruckte Bücher oder Manuals

Consulting (welche Zwei Faktor Authentifizierung Sinn machen würde zB)

Man könnte Flaschenschiffe und anderes Merchandise verkaufen :-)

# Beschreibung des Arbeitsplans

## Technische Arbeitsschritte

Im Rahmen unserer Masterarbeiten werden wir einen Prototypen entwickeln der
bereits in Gründzügen die oben beschriebenen Technologien demonstriert.
Performanz und Portabilität sind zu diesem Zeitpunkt aus Zeitmangel allerdings
noch keine harten Anforderungen.

Die im ersten Prototypen gewonnen Erkenntnisse wollen wir dazu nutzen
nötigenfalls eine »Kurskorrektur« durchzuführen und den ersten Prototypen nach
Möglichkeit zu vereinfachen und stabilisieren.

Zu diesem zweiten Prototypen werden dann in kleinen Iterationen Features
hinzugefügt. Jedes dieser Feature sollte für sich alleine stehen, daher sollte
zu diesem Zeitpunkt bereits die grundlegende Architektur relativ stabil sein.

Nachdem ein gewisses Mindestmaß an nützlichen Features hinzugekommen ist, wäre
ein erstes öffentliches Release anzustreben. Dies hätte bereits eine gewisse
Verbreitung zur Folge und die in ``brig`` eingesetzten Sicherheitstechnologien
könnten von Externen auditiert werden.

## Meilensteinplanung

![Sehr groben Meilensteinplanung](images/milestones.png){#fig:milestones}

Siehe Abbildung {@fig:milestones}

TODO: Featureliste für Prototyp I & II

# Finanzierung des Vorhabens

Eine mögliche Finanzierungstrategie bietet das IuK--Programm[^IUK] des Freistaates
Bayern. Dabei werden Kooperation zwischen Fachhochschulen und Unternehmen mit
bis zu 50% gefördert. Gern gesehen ist dabei beispielsweise ein Großunternehmen
und ein kleines bis mittleres Unternehmen (KMU). 

Beide zusammen würden dann das Fördervolumen stellen, womit die Hochschule dann
zwei Stellen für wissenschaftliche Arbeiter finanzieren könnte.

Die Höhe des Fördervolumens richtet sich primär nach der Dauer der Förderung und
dem jeweiligen akademischen Abschluss. Die Dauer würden wir dabei auf mindestens
zwei, optimalerweise drei Jahre ansetzen. 

TODO: Grobekostenplanung ~500.000,-

```
>>> pro_mann = 12 * 3500              # =  42000€
>>> pro_jahr = 2 * pro_mann + 30000   # = 114000€
>>> budget = 3 * pro_jahr             # = 342000€
```

[^IUK]: Mehr Informationen unter \url{http://www.iuk-bayern.de/}


# Literaturverzeichnis
