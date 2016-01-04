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
Transparenz wird die Software dabei quelloffen unter der AGPLv3 Lizenz
entwickelt. 

Nutzbar soll das resultierende Produkt, neben dem Standardanwendungsfall der
Dateisynchronisation, auch als Backup- bzw. Archivierungs-lösung sein
beziehungsweise auch als verschlüsselter Daten--Safe oder als Plattform für
andere, verteilte Anwendungen aus dem Industrie 4.0 Umfeld.

5-6 Sätze.

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
Dateisynchronisationssoftware. Die Tauglichkeit für ein Unternehmen ist sehr
variabel. Wir meinen damit im Folgenden diese Punkte:

- Einfach Benutzbarkeit für nicht-technische Angestellte.
  (ein einfacher Ordner im Dateimanager)
- Durchsuchbarkeit.
- Zentrale Speicherung möglich, aber nicht zwingend. 
  Von Angestellten unbenutzte Dateien sollen dessen Speicherplatz nicht
  belasten. (TODO: Drück dich besser aus, Junge.)
- Kein Vendorlock (-> Open Source)
- TODO

TODO: Git für große Dateien irgendwo reinbringen. (für technisch versierte)

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

Privatpersonen: Schutz der Privatsphäre. 

Plattform für industrielle Anwendungen. (I4.0)

Einsatz im öffentlichen Bereich, Schulen, Universitäten...

## Innovation

Wie bereits oben angedeutet, gibt es bereits zahlreiche Möglichkeiten Dateien in
einem Netzwerk auszutauschen. Diese erfüllen aber stets nur Teilaspekte unserer
obigen Ziele.

Die Innovation bei unserem Projekt (TODO: brig als Name einführen?) besteht
daher darin bekannte Technologien neu ,,neu zusammen zu stecken'', woraus sich
neue Möglichkeiten (siehe oben) ergeben.

-- TODO: Bekannte Technologien, neu zusammengewürfelt, neue Möglichkeiten.
-- 
-- Unternehmenstaugliche (benutzerfreundliche) dezentrale und sichere
-- Dateisynchronisation  mit Ende zu Ende Verschlüsselung. Unabhängigkeit von
-- Hersteller und Cloudservices.

## Lizenz

TODO: Warum freie Lizenz. Vorteile -> Sicherheit und Verbreitung. Transparenz.
Weiterentwicklung/mehr Kontrolle durch Unternehmen. Sichergestellt dass
Verbesserungen wieder ins Projekt zurückfließen.


# Stand der Wissenschaft und Technik

## Stand der Wissenschaft

TODO: Zentral, Dezentral, P2P. Technologien wie Cloud, XMPP. Single Point of
Failure. Propritäre Lösungen -> Sicherheit unklar, Freiheit auch.

## Markt und Wettbewerber

Es gibt viele verschiedene Lösungen die jeweils immer in Teilaspekten gut
funktionieren. Beispielhafte Konkurrenten:

* Syncthing -> Heimanwender. Immer physikalische Kopie?
* Git--annex
* Owncloud -> Zentrale Lösung.
* Dropbox und Konsorten + Boxcryptor


# Ausführliche Beschreibung des Vorhabens

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

## Lösungsansätze

* IPFS, XMPP, GO, Crypostandards (sym, asym.)
* Idea!

## Technische Risiken 

Der Aufwand für ein Softwareprojekt dieser Größe ist schwer einzuschätzen.
Da wird auf relativ junge Technologien wie ``ipfs`` setzen.

* IPFS ist eine junge Software, optimale Tauglichkeit noch zu erforschen.
* Problematische Entwicklung bzgl. Kryptographischen Verfahren.
* Aufwand zur Entwicklung von Brig ist schwer einschätzbar.

TODO: Risikominimierung

* IPFS austauschbar machen. 

# Wirtschaftliches Verwertungskonzept

## Wirtschaftliche Verwertung 

Mögliche Einnahmequellen, durch…

…bezahlte Entwicklung spezieller Features.
…Supportverträge.
…Mehrfachlizensierung.
…Utility Bereitstellung (LDAP, yubikeys, …)
…zertifizierte NAS-Server.
…Schulungen, Lehrmaterial und Consulting.

TODO: Made in Germany

# Beschreibung des Arbeitsplans

## Arbeitsschritte

* Prototyp als Masterarbeit, grundlegende Features.
* Erforschung erweiterter verwertbarer Technologien, zweiter Prototyp mit
  erweiterten Technologischen Möglichkeiten.
* Iterative weitere Prototypen und Features bis zur stabilen ersten Version,
  welche unabhängig bezüglich Sicherhetistechnologien zertifiziert werden soll.

## Meilensteinplanung

Zusammensetzen und Meilensteine definieren.

# Finanzierung des Vorhabens (Grobskizze)

TODO: IuK erklären. Grobekostenplanung ~500.000,-
