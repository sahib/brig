# An welchen Open Source Projekten hast Du bisher gearbeitet?

Unter anderem:

rmlint: Sehr schnelle Deduplizierungslösung auf Dateiebene für unixoide Betriebssysteme.
(https://github.com/sahib/rmlint)

eulenfunk: Internetradio Marke Eigenbau (mit Bauanleitung!) auf Basis freier Software.
(https://github.com/studentkittens/eulenfunk)

Snøbær: Umfangreicher Web MPD-Client auf Basis von Python, C und Coffeescript.
(https://github.com/studentkittens/snobaer)

# Projekttitel

»brig« - Ein Werkzeug zur sicheren und verteilten Dateisynchronisation.

# Welches Problem willst Du mit Deinem Projekt lösen? Was ist Deine Motivation?

1000

Es gibt keinen (für den Otto-Normal-Bürger) einfach benutzbaren Standard um
Dateien sicher zu Synchronisieren. Oftmals tauschen Menschen heutzutage ihre
Dateien über herstellerabhängige, oft unsichere und zentrale Dienste (wie
beispielsweise Dropbox) oder gar via E-Mail aus. Ohne zusätzliche, meistens
proprietäre, Software (wie Boxcryptor) werden die Daten dabei nicht auf der
Seite des Benutzers verschlüsselt. Besonders im Lichte der Snowden-Enthüllungen
ist eine Benutzung von solchen Diensten für sensible Daten als fragwürdig
einzustufen.

Wünschenswert wäre ein Programm, welches dem Benutzer die Datenhoheit zurück
gibt. Um sich als »Standard«  etablieren zu können, muss das Programm eine gute
Balance zwischen Sicherheit und Usability bieten. Unserer Meinung nach muss ein
sicheres Tool dabei auch offen und transparent für den Nutzer sein. Daher
ist unsere Motivation, ein solches, herstellerunabhängiges Werkzeug zu schaffen
- nicht zuletzt für den eigenen Gebrauch.

# Wie wird Dein Projekt dieses Problem lösen?

2000

Unsere Idee ist die Schaffung eines einfachen Werkzeuges, welches Individuen
ermöglicht sicher Dateien auszutauschen, ohne sich auf einen zentralen
Cloud-Dienst verlassen zu müssen. Unser Gegenentwurf ist eine dezentrale Lösung
auf Basis des P2P-Netzwerks IPFS (http://ipfs.io). IPFS bietet aufgrund seiner
generellen Ausrichtung nur ein rudimentäre, kommandozeilenorientierte Toolbox
zur Dateiübertragung. Unser Projekt soll daher als »Frontend«  für IPFS dienen
und es um folgende Features erweitern:

- Vollständige Ende-zu-Ende Verschlüsselung von Dateien und Metadaten
- Verschlüsselte Speicherung der Daten, Entschlüsselung nur »live«  im Arbeitsspeicher
- Benutzerauthentifizierung (Passwort oder Multifaktorbasiert)
* Sicherheitskomplexität wird hinter einem einfach benutzbaren Frontend möglichst versteckt
* Teilen von Dateien auch mit nicht »brig« -Nutzern durch »Gateways«  möglich
* Es werden nur Metadaten synchronisiert, die eigentlichen Daten können dynamisch
* Durch Integritätsprüfung kann die korrekte Übertragung und Speicherung validiert werden
* Eingebaute Dateikompression mit verschiedenen wählbaren Algorithmen
* Vorhaltung einer Mindestanzahl von Kopien: Dateien können mehrfach im Netzwerk gespeichert
  werden, um die Wiederherstellung beschädigter Dateien zu ermöglichen
  on-demand besorgt werden und bis zu einer bestimmten Quota zwischengelagert werden
- Versionskontrolle für große Dateien

Im Vergleich zu Lösungen wie Dropbox bietet es folgende Vorteile:

* Kein Single-Point-of-Failure: Die Nutzer bilden die Infrastruktur
* Transparenter Einsatz von Kryptografie (daher keine Backdoors)
- Kein Vendor-Lock-in dank FOSS

Um den Nutzer ein vertrautes Bild zu bieten, werden die synchronisierten
Dateien in einem »normalen«  Ordner nach außen präsentiert. Durch diesen Ansatz
soll möglichst gute Portabilität und Usability gewährleistet werden. Lediglich
die Einrichtung darf sich von Plattform zu Plattform unterscheiden.

# Wer ist die Zielgruppe? Wie profitiert sie vom Projekt?

2000

Prinzipiell kann »brig«  von jeder Privatpersonen, jedem Unternehmen und selbst
Behörden genutzt werden. Aufgrund seiner Eigenschaften ist »brig«  auch für
Unternehmen und Behörden sehr gut geeignet:

* Keine Kosten für die Software und kein Vendor-Lock-in.
* Daten können auf eigenen Servern innerhalb des Unternehmens gelagert werden.
* Multifaktorauthentifizierung der Nutzer mittels YubiKeys.
* Hierarchische Benutzernamen (Beispiel: alice@wonderland.de/desktop)
- Einsatz bekannter und anerkannter symmetrischer und asymmetrischer Sicherheitsstandards.

Im Speziellen profitieren beispielsweise folgende Berufsgruppen durch die
Datensparsamkeit von »brig«:

- Medizinischer Bereich: Sicherer Austausch von Patientendaten.
* Journalisten, Aktivisten, politisch Verfolgte: Austausch von Dokumenten.
* Rechtsanwälte, Notare: Austausch sensibler Dokumente mit Klienten.

Aber auch öffentliche Einrichtungen wie Hochschulen können »brig« nutzen, um
Vorlesungsmaterial untereinander zu teilen. Technisch versierte Nutzer und
Entwickler können »brig«  als flexiblen Baukasten für Dateisynchronisation
jeder Art einsetzen, da es auf unterster Ebene, ähnlich wie das
Versionsverwaltungssystem git, aus einzelnen kleinen Werkzeugen aufgebaut ist.

Die Nutzer der Software bilden dabei ein Overlay-Netzwerk über das normale Internet.
Dadurch werden diese unabhängig von der Infrastruktur von Drittanbietern und können
selbst über diese bestimmen, was auch gezielte Zensurmaßnahmen erschwert.

Anmerkung: Neben den zentralen Diensten wie Dropbox gibt es bereits einige
dezentrale Ansätze wie Syncthing, git-annex oder Bittorrent-Sync. Leider sind
diese entweder nicht auf den Unternehmenseinsatz ausgelegt, hochgradig komplex,
proprietär oder haben einen anderen Fokus bzw. Einsatzzweck.

# Hast Du schon an der Idee gearbeitet? Wenn ja, beschreibe kurz den aktuellen Stand und erkläre die Neuerung.

1000

Ja.

»brig«  wird seit ca. Anfang 2016 in Kooperation mit meinem Kommilitonen
Christoph Piechula (github.com/qitta) entwickelt. Wir beide schreiben die
Software primär als Hobby neben unseren Informatik-Studium an der Hochschule
Augsburg und momentan auch als Gegenstand unserer Masterarbeiten.
Wir wollen die Software allerdings auch nach dem Studium weiterentwickeln.

Die momentane Codebasis ist in der Programmiersprache Go verfasst und auf
GitHub unter github.com/sahib/brig verfügbar. Die Lizenz ist AGPLv3.

Der aktuelle Funktionsstand ist noch nicht der eines funktionierenden
Prototypens, auch wenn viele Einzelfunktionalitäten wie Verschlüsselung,
Kompression und Datenaustausch als Proof-of-Concept existieren. Auch das lokale
Bearbeiten von Dateien in einem FUSE-Dateisystem ist bereits rudimentär
funktionsfähig.

Die Neuerung die nun in 6 Monaten Arbeit entstehen soll, ist schlicht die
Implementierung eines funktionierenden, polierten und präsentablen Prototyps.

# Wie viele Arbeitsstunden wirst Du in einem Zeitraum von 6 Monaten vermutlich für die Umsetzung der Projektidee benötigen?

960

# Skizziere kurz die wichtigsten Meilensteine Deines Projekts.

1500

Das Ziel bis zum möglichen Förderungsbeginn (März 2017) ist die Implementierung
eines zusammenhängenden Proof-of-Concepts, der die dahinter liegende Technik
veranschaulicht und zeigt, dass die angestrebte Funktionsweise (hoffentlich)
erreichbar ist.

In den darauf folgenden sechs Monaten sind folgende Meilensteine geplant:

* Entwicklung eines rudimentären User Interfaces.
- Bereitstellung einer informativen Projekt-Website.
* Bereitstellung von Dokumentation und vorkompilierten Binaries für die meisten Plattformen.
- Implementierung einer umfangreichen Testsuite.

Das Ziel ist kurz gesagt die Entwicklung eines Prototypen, welcher der
Open-Source-Community präsentiert werden kann. Dazu ist allerdings
erfahrungsgemäß ein Mindestmaß an Qualität und Dokumentation nötig, um
sinnvolles Feedback zu bekommen. Auch Beiträge von anderen Entwicklern bekommt
man meist erst, wenn diese die Software selbst produktiv einsetzen konnten.

Die Umsetzung dieser Punkte ist natürlich sehr zeitintensiv (Paretoprinzip :).
Nach unseren Studium und vor unseren Berufseinstieg würden wir das Projekt
daher gerne »in trockene Tücher bringen«  und sind daher ohne BAföG natürlich
auch auf finanzielle Mittel angewiesen.
