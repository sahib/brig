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
lof: yes
lot: yes
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
Transparenz wird die Software mit dem Namen ``brig`` dabei quelloffen unter der AGPLv3 Lizenz
entwickelt. 

Nutzbar soll das resultierende Produkt, neben dem Standardanwendungsfall der
Dateisynchronisation, auch als Backup- bzw. Archivierungs-lösung sein
beziehungsweise auch als verschlüsselter Daten--Safe oder als Plattform für
andere, verteilte Anwendungen aus dem Industrie 4.0 Umfeld.

# Projektsteckbrief

## Einleitung

Unternehmen (TODO: Wie VW?) haben hohe Ansprüche an die Sicherheit, welche
zentrale Alternativen wie Dropbox nicht bieten können. Zwar wird die Übertragung
von Daten zu den zentralen Dropbox-Servern verschlüsselt, was allerdings danach
mit den Daten ,,in der Cloud'' passiert liegt nicht mehr in der Kontrolle der
Nutzer. Dort sind die Daten schonmal für andere Nutzer wegen Bugs einsehbar oder
werden gar von Dropbox an amerikanische Geheimdienste weitergegeben. 
(TODO: Footnote links)

[Sprichwörtlich, regnen die Daten irgendwo anders aus der Cloud ab.]
Tools wie Boxcryptor lindern diese Problematik zwar etwas, heilen aber nur die
Symptome, nicht das zugrunde liegende Problem.

Dropbox ist leider kein Einzelfall -- beinahe alle Cloud--Dienste haben, oder
hatten, architektur-bedingt ähnliche Sicherheitslecks. Für ein Unternehmen wäre
es vorzuziehen ihre Daten auf Servern zu speichern, die sie selbst
kontrollieren. Dazu gibt es bereits einige Werkzeuge wie ownCloud oder ein
Netzwerkverzeichnis wie Samba, doch technisch bilden diese nur die zentrale
Architektur von Cloud--Diensten innerhalb eines Unternehmens ab. 

## Ziele

Ziel ist die Entwicklung einer sicheren, dezentralen und unternehmenstauglichen
Dateisynchronisationssoftware names ``brig``. Die ,,Tauglichkeit" für ein
Unternehmen ist sehr variabel. Wir meinen damit im Folgenden diese Punkte:

- Einfach Benutzbarkeit für nicht-technische Angestellte: Sichtbar soll nach der
  Einrichtung nur ein simpler Ordner im Dateimanager sein.
- Schnelle Auffindbarkeit der Dateien durch Verschlagwortung.
- Effiziente Übertragung von Dateien: Intelligentes Routing vom Speicherort zum Nutzer.
- Speicherquoten: Nicht alle Dateien müssen synchronisiert werden.
- Automatische Backups: Versionsverwaltung auf Knoten mit großem Speicherplatz.

Um eine solche Software entwickeln, wollen wir auf bestehende Komponenten wie
InterPlanetaryFileSystem (ein konfigurierbares P2P Netzwerk) und XMPP (ein
Messenging Protokoll und Infrastruktur) aufsetzen. Dies erleichtert unsere
Arbeit und macht einen Prototyp der Software erst möglich. 

Von einem Prototypen zu einer marktreifen Software ist es allerdings stets ein
weiter Weg. Daher wollen wir einen großen Teil der darauf folgenden Iterationen
damit verbringen, die Software bezüglich Sicherheit, Performance und einfacher
Benutzerfreundlichkeit zu optimieren. Da es dafür keinen standardisierten Weg
gibt, ist dafür ein großes Maß an Forschung nötig.

## Use cases

``brig`` soll deutlich flexibler nutzbar sein als beispielsweise zentrale
Dienste. Nutzbar soll es unter anderem sein als…

- …**Synchronisationslösung**: Spiegelung von 2-n Ordnern.
- …**Transferlösung**: "Veröffentlichen" von Dateien nach Außen mittels Hyperlinks.
- …**Versionsverwaltung**: 
  Alle Zugriffe an eine Datei werden aufgezeichnet.
  Bis zu einer bestimmten Tiefe können alte Dateien abgespeichert werden.
- …**Backup- und Archivierungslösung**: Verschiedene Knoten Typen möglich.
- …**verschlüsselten Safe**: ein ,,Repository'' kann ,,geschlossen'' werden.
- …**Semantisch durchsuchbares** tag basiertes Dateisystem[^TAG].
- …als **Plattform** für andere Anwendungen.
- …einer beliebigen Kombination der oberen Punkte.

[^TAG]: Ähnlich zu https://en.wikipedia.org/wiki/Tagsistant

## Zielgruppen

``brig`` zielt hauptsächlich auf Unternehmenskunden und Heimanwender.
Daneben sind aber auch noch andere Zielgruppen denkbar.

### Unternehmen

Großunternehmen können ``brig`` nutzen, um ihre Daten und Dokumente intern zu
verwalten. Besonders sicherheitskritische Dateien entgehen so der Lagerung in
Cloud Services oder der Gefahr von zig Kopien auf Mitarbeiter-Endgeräten.
Größere Unternehmen verwalten dabei meist ein Rechenzentrum auf den
firmeninterne Dokumente gespeichert werden. Diese werden dann meist mittels
ownCloud, Samba o.ä. von den Nutzern "manuell" heruntergeladen. 

In diesem Fall könnte man ``brig`` im Rechenzentrum und allen Endgeräten installieren.
Das Rechenzentrum würde die Datei mit tiefer Versionierung und vollem Caching vorhalten.
Endanwender würden alle Daten sehen, aber auf ihren Gerät nur die Daten tatsächlich
speichern, die sie auch benutzen. Hat ein Kollege im selben Büro beispielsweise die
Datei bereits kann ``brig`` sie dann auch teilweise von ihm holen.

Kleinunternehmen wie Ingenieurbüros können ``brig`` dazu nutzen Dokumente nach
außen freizugeben, ohne dass sie dazu vorher irgendwo "hochgeladen" werden
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

Eine Anwendung in der Industrie 4.0 wäre beispielweise...
TODO: Beispiel.

### Einsatz im öffentlichen Bereich

Aufgrund seiner Transparenz und einfachen Benutzbarkeit wäre ebenfalls eine
Nutzung an Schulen, Universitäten oder auch in Behörden zum Dokumentenaustausch
denkbar. Vorteilhaft wäre hierbei vor allem, dass man sich aufgrund des
Open--Source Modells an keinen Hersteller bindet (Stichwort: Vendor Lock) und
keine behördlichen Daten in der ,,Cloud" landen.

## Innovation

Die Innovation bei unserem Projekt  besteht daher darin bekannte Technologien
neu ,,neu zusammen zu stecken'', woraus sich viele neue Möglichkeiten ergeben.

TODO?

# Stand der Technik

## Stand der Wissenschaft

Zwar ist das Projekt stark anwendungsorientiert, doch basiert es auf gut
erforschten Technologien wie Peer-to-Peer-Netzwerken (P2P), von der NIST
zertifizierten kryptografischen Standard-Algorithmen und verteilten Systemen im
Allgemeinen (TODO: XMPP). Peer to Peer Netzwerke wurden in den letzten Jahren gut erforscht
und haben sich auch in der Praxis bewährt (Skype ist ein Beispiel). 

Allerdings ist uns keine für breite Massen nutzbare Software bekannt, die es
Nutzern ermöglicht selbst ein P2P Netzwerk aufzuspannen und darin Dateien
auszutauschen. Am nähsten kommen dabei die beiden Softwareprojekte
``Syncthing`` (OpenSource) und ``BitTorrent Sync`` (propritär).

Der wissenschaftliche Beitrag unserer Arbeit wäre daher die Entwicklung einer
freien Alternative, die von allen eingesehen, auditiert und studiert werden
kann. Diese freie Herangehensweise ist insbesondere für sicherheitskritische
Software relevant, da keine offensichtlichen ,,Exploits" in die Software
eingebaut werden können.

## Markt und Wettbewerber

Bereits ein Blick auf Wikipedia zeigt, dass der momentane Mark an
Dateisynchronisationssoftware (im weitesten Sinne) sehr unübersichtlich ist.

https://en.wikipedia.org/wiki/Comparison_of_file_synchronization_software

Bei einem näheren Blick stellt sich oft heraus, dass die Software dort oft nur
in Teilaspekten gut funktioniert oder andere nicht lösbare Probleme besitzt.
Im Folgenden wird eine kleine Übersicht gegeben welche aktuelle Software
Alternativen zu machen Usecases von ``brig`` darstellen:

### Verschiedene Alternativen:x

#### Dropbox + Boxcryptor

- Zentrale Lösung

#### Owncloud

+ Daten liegen auf eigenen Servern.

- Zentrale Lösung.
