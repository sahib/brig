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
zu gewährleisten, wird auf Sicherheit sehr großer Wert gelegt.  Aus Gründen der
Transparenz wird die Software mit dem Namen ``brig`` dabei quelloffen unter der AGPLv3 Lizenz
entwickelt. 

Nutzbar soll das resultierende Produkt, neben dem Standardanwendungsfall der
Dateisynchronisation, auch als Backup- bzw. Archivierungs-lösung sein
beziehungsweise auch als verschlüsselter Daten--Safe oder als Plattform für
andere, verteilte Anwendungen aus dem Industrie 4.0 Umfeld.

TODO: 5-6 Sätze. Kurz genug?

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
TODO: Boxcryptor erwähenn?

Dropbox ist leider kein Einzelfall -- beinahe alle Cloud--Dienste haben, oder
hatten, architektur-bedingt ähnliche Sicherheitslecks. Für ein Unternehmen wäre
es vorzuziehen ihre Daten auf Servern zu speichern, die sie selbst
kontrollieren. Dazu gibt es bereits einige Werkzeuge wie ownCloud oder ein
Netzwerkverzeichnis wie Samba, doch technisch bilden diese nur die zentrale
Architektur von Cloud--Diensten innerhalb eines Unternehmens ab. 

## Ziele

Ziel ist die Entwicklung einer sicheren, dezentralen und unternehmenstauglichen
Dateisynchronisationssoftware names ``brig``. Die Tauglichkeit für ein
Unternehmen ist sehr variabel. Wir meinen damit im Folgenden diese Punkte:

- Einfach Benutzbarkeit für nicht-technische Angestellte.
  (ein einfacher Ordner im Dateimanager)
- Durchsuchbarkeit.
- Zentrale Speicherung möglich, aber nicht zwingend. 
  Von Angestellten unbenutzte Dateien sollen dessen Speicherplatz nicht
  belasten. (TODO: Drück dich besser aus, Junge.)
- Kein Vendorlock (-> Open Source)
- TODO

Um eine solche Software entwickeln, wollen wir auf bestehende Komponenten wie
IPFS (ein p2p Netzwerk) und XMPP (ein Messanging Protokoll und Infrastruktur)
aufsetzen. Dies erleichtert unsere Arbeit und macht einen Prototyp der Software
erst möglich. 

Von einem Prototypen zu einer marktreifen Software ist es allerdings stets ein
weiter Weg. Daher wollen wir einen großen Teil der darauf folgenden Iterationen
damit verbringen, die Software bezüglich Sicherheit, Performance und einfacher
Benutzerfreundlichkeit zu optimieren. Da es dafür keinen standardisierten Weg
gibt, ist dafür ein großes Maß an Forschung nötig.

## Use cases

Nutzbar als…

…Transferlösung (Hyperlinks möglich um einzelne Dateien nach Außen zu sharen).
…Synchronisationslösung.
…Backup- oder Archivierungslösung.
…Versionsverwaltung.
…verschlüsselten Safe.
…semantisch durchsuchbares tag basiertes datei systemkkg=
…als Plattform für andere Anwendungen.

## Zielgruppen

TODO: Unternehmen. Groß- und Kleinunternehmen. Kundenaustausch. Große
Unternehmen zur internen Datenverwaltung. Einsatz im Sicherheitskritischen
Bereichen -> Made In Germany.

### Unternehmen

- Großunternehmen können ``brig`` nutzen, um ihre Daten und Dokumente intern
  zu verwalten.
- Kleinunternehmen wie Ingenieurbüros können brig dazu nutzen Dokumente nach
  außen freizugeben, ohne dass sie dazu vorher irgendwo "hochgeladen" werden
  müssen. (TODO: Gateway beschreiben?)

### Privatpersonen

Privatpersonen: Schutz der Privatsphäre. 

### Plattform für industrielle Anwendungen

Da ``brig`` auch komplett automatisiert ohne Interaktion nutzbar sein soll,
kann es auch als Plattform für andere Anwendungen genutzt werden, die Dateien
austauschen und synchronisieren müssen.

TODO: Beispiel.

### Einsatz im öffentlichen Bereich

Aufgrund seiner Transparenz und einfachen Benutzbarkeit wäre ebenfalls eine
Nutzung an Schulen, Universitäten oder auch in Behörden zum Dokumentenaustausch
denkbar. Vorteilhaft wäre hierbei vor allem, dass man sich aufgrund des
Open--Source Modells an keinen Hersteller bindet (Stichwort: Vendor Lock) und
keine behördlichen Daten in der ,,Cloud" landen.

## Innovation

Wie bereits oben angedeutet, gibt es bereits zahlreiche Möglichkeiten Dateien in
einem Netzwerk auszutauschen. Diese erfüllen aber stets nur Teilaspekte unserer
obigen Ziele.

Die Innovation bei unserem Projekt (TODO: brig als Name einführen?) besteht
daher darin bekannte Technologien neu ,,neu zusammen zu stecken'', woraus sich
neue Möglichkeiten (siehe oben) ergeben.

# Stand der Wissenschaft und Technik

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

Bereits ein Blick auf Wikipedia zeigt, dass es 

https://en.wikipedia.org/wiki/Comparison_of_file_synchronization_software

Bei einem näheren Blick stellt sich oft heraus, dass die Software dort oft nur
in Teilaspekten gut funktioniert oder andere nicht lösbare Probleme besitzt.
Im Folgenden wird eine 

* Syncthing -> Heimanwender. Immer physikalische Kopie?

    + Open Source
    + Einfacher Ordner auf Dateisystemebene.

    - Keine Benutzerverwaltung
    - kein p2p netzwerk
    - zentraler key server

* BitSync - Unternehmensanwender

    + p2p netzwerk
    + verschlüsselte Speicherung
    
    - propritär und kommerziell
    - Keine Benutzerverwaltung
    - Versionsverwaltung nur als "Archiv-Folder"

    BitSync


* Git--annex

    + sehr featurereich 
    + special remotes
    + Open Source

    - kein p2p netzwerk
    - Selbst für erfahrene Benutzer nur schwierig zu benutzen

* Owncloud -> Zentrale Lösung.

    - zentral
    - Zugriff über Weboberfläche

* Dropbox und Konsorten + Boxcryptor

* GlusterFS
    
    + Hochperformant.

    - Nicht portabel.

Zusammengefasst findet sich hier noch eine tabellarische Übersicht:

# Ausführliche Beschreibung des Vorhabens

Optimal wäre also eine Kombination aus den Vorzügen von Syncthing, BitTorrent
Sync und git-annex. Unser Versuch diese Balance hinzubekommen heißt ``brig``.

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

Finden von Technologien die gut geeignet sind um die oben gelisteten Ziele
möglichst gut zu erreichen bzw die Probleme möglichst gut zu lösen.

Ziepparameter: Kein SPOF, Datei nur ein Mal im Netz, Bandbreitenrouting,
Benutzerverwaltung, Einfache Softwareinstallation und Benutzung, Sicherheit Made
in Germany.

* Verschlüsselte Übertragung und Speicherung.
* Kompression & Deduplizierung (optional; mittels brotli)
* Speicherquoten & Pinning (Thin-Client vs. Storage Server)
* Versionierung mit definierbarer Tiefe.
* Benutzerverwaltung mittels XMPP.
* 2F Authentifizierung und paranoide Sicherheit.

Portabilität und hohe Grundperformanz durch Go.

## Lösungsansätze

* IPFS, XMPP, GO, Crypostandards (sym, asym.)
* Portabilität durch Go.
* Idea!

## Technische Risiken 

Der Aufwand für ein Softwareprojekt dieser Größe ist schwer einzuschätzen.
Da wird auf relativ junge Technologien wie ``ipfs`` setzen.

* IPFS ist eine junge Software, optimale Tauglichkeit noch zu erforschen.
* Problematische Entwicklung bzgl. Kryptographischen Verfahren.
* Aufwand zur Entwicklung von Brig ist schwer einschätzbar.
* Performanz-Probleme
* Firewalls u.ä.

TODO: Risikominimierung

* IPFS austauschbar machen. 
* Go gegen portabilitätsprobleme und performanz

# Wirtschaftliches Verwertungskonzept

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

TODO: Made in Germany (*seufz* really?)

Es folgen einige konkrete Verwertung Strategien, die auch in Partnerschaft mit
Unternehmen ausgeführt werden könnten.

### Bezahle Entwicklung spezieller Features

Die Open-Source-Entwickler Erfahrung der Autoren hat gezeigt, dass sich Nutzer
oft ungewöhnliche Features wünschen, die sie oft zu einem bestimmten Termin
brauchen. (TODO: blabla)

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

## Arbeitsschritte

Im Rahmen unserer Masterarbeiten werden wir einen Prototypen entwickeln der
bereits in Gründzügen die oben beschriebene Technologie demonstriert. Performanz
und Portabilität sind zu diesem Zeitpunkt aus Zeitmangel allerdings noch keine
harten Anforderungen.

* Prototyp als Masterarbeit, grundlegende Features.
* Erforschung erweiterter verwertbarer Technologien, zweiter Prototyp mit
  erweiterten Technologischen Möglichkeiten.
* Iterative weitere Prototypen und Features bis zur stabilen ersten Version,
  welche unabhängig bezüglich Sicherhetistechnologien zertifiziert werden soll.

## Meilensteinplanung

Zusammensetzen und Meilensteine definieren.

# Finanzierung des Vorhabens (Grobskizze)

Eine mögliche Finanzierungstrategie bietet das IuK--Programm des Freistaates
Bayern. Dabei werden Kooperation zwischen Fachhochschulen und Unternehmen mit
bis zu 50% gefördert. Gern gesehen ist dabei beispielsweise ein Großunternehmen
und ein kleines bis mittleres Unternehmen (KMU). 

Beide zusammen würden dann das Fördervolumen stellen, womit die Hochschule dann
zwei Stellen für wissenschaftliche Arbeiter finanzieren könnte.

http://www.iuk-bayern.de/

Die Höhe des Fördervolumens richtet sich primär nach der Dauer der Förderung und
dem jeweiligen akademischen Abschluss. Die Dauer würden wir dabei auf mindestens
zwei, optimalerweise drei Jahre ansetzen. 

TODO: Grobekostenplanung ~500.000,-
