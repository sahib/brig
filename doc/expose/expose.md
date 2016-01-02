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

## Ziele

Entwicklung einer sicheren und unternehmenstauglichen
Dateisynchronisationssoftware. Forschung/Weiterentwicklung beziehungsweise
Erweiterung der bereits bestehender Standards und Produkte. Erstellung einer
neuartigen Software auf Basis vorhandener/erweiterter Technologien. Erforschung
unternehmenstauglicher Technologien um eine Optimierung bezüglich Sicherheit,
Performance und Benutzerfreundlichkeit (einfache Bedienung) zu ermöglichen.
TODO: Konkreteres Abstract.

## Use cases

* Welches Problem wird mit dem Vorhaben gelöst?  

Use Cases. Aktuell Cloudproblematik beschreiben. Herstellerunabhängige
dezentrale Synchronisation von Daten ohne Cloud. 

Nutzbar als…

…Transferlösung (Hyperlinks möglich).
…Synchronisationslösung.
…Backup- oder Archivierungslösung.
…Versionsverwaltung.
…verschlüsselten Safe.
…als Plattform für andere Anwendungen.

## Zielgruppen

TODO: Unternehmen. Groß- und Kleinunternehmen. Kundenaustausch. Große
Unternehmen zur internen Datenverwaltung. Einsatz im Sicherheitskritischen
Bereichen -> Made In Germany.

Privatpersonen: Schutz der Privatsphäre. 

Plattform für industrielle Anwendungen. (I4.0)

Einsatz im öffentlichen Bereich, Schulen, Universitäten...

## Innovation

* Worin besteht die Innovation des Vorhabens?

TODO: Bekannte Technologien, neu zusammengewürfelt, neue Möglichkeiten.

Unternehmenstaugliche (benutzerfreundliche) dezentrale und sichere
Dateisynchronisation  mit Ende zu Ende Verschlüsselung. Unabhängigkeit von
Hersteller und Cloudservices.

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
