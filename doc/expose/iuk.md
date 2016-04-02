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

# Projektkonzept

## Gesamtziel des Vorhabens

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
- Flexible Benutzerverwaltung, die sich in existierende Systeme einfügt.
- Versionsverwaltung großer Dateien mit definierbarer Tiefe.

## Aufgaben der Projektpartner

\textcolor{Red}{TODO}

Wie ist die Aufgabenverteilung im Vorhaben?

- ... noch zu klären ...
- Entwicklungsarbeit hauptsächlich bei uns.
- Beratende Tätigkeit von Secomba
- Anwendung durch/in einem weiteren Unternehmen?

## Kompetenzen der beteiligten Partner

### Secomba

Secomba entwickelt mit ``Boxcryptor`` eine Verschlüsselungssoftware für
Cloudservices wie Dropbox. Aufgrund dieser Vorerfahrungen sind die dortigen
Entwickler bestens mit der Problematik der sicheren Dateisynchronisation
vertraut.

## Beschreibung der bayerischen Standorte der Projektpartner

Wie viele Mitarbeiter sind an den bayerischen Standorten beschäftigt?
Welche Umsätze werden an den bayerischen Standorten erzielt?
Welche Fertigungs- und Entwicklungsressourcen existieren an den bayerischen Standorten?

... von Partnern auszufüllen, siehe excel tabelle? ...

\textcolor{Red}{TODO}

# Stand der Wissenschaft und Technik

## Stand der Technik

Die Innovation bei unserem Projekt  besteht daher hauptsächlich darin, bekannte
Technologien »neu zusammen zu stecken«, woraus sich viele neue Möglichkeiten
ergeben.  Wie im nächsten Kapitel beleuchtet wird, ist ``brig`` der Versuch
viele gute, bestehende und in der Praxis geprüfte Ideen in einem konsistenten
Programm zu vereinen.

## Stand der Wissenschaft

Zwar ist das Projekt stark anwendungsorientiert, doch basiert es auf gut
erforschten Technologien wie Peer--to--Peer-Netzwerken (kurz *P2P*, siehe auch
[@peer2peer_arch]), von der NIST[^NIST] zertifizierten kryptografischen
Standard-Algorithmen[@everyday_crypto] und verteilten Systemen im Allgemeinen.
P2P--Netzwerke wurden in den letzten
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





Was ist der internationale Stand der Technik auf diesem Gebiet?
Was ist der interne Stand der Technik der jeweiligen Partner?
Welche Vorarbeiten zu dem Vorhaben existieren ggf.?
Auf welchen bestehenden Verfahren/Produkte baut das Vorhaben auf?
Welche für das Vorhaben relevanten FuE-Projekte werden zurzeit durchgeführt oder wurden in den letzten drei Jahren abgeschlossen?
Welche alternativen Verfahren/Produkte existieren bereits?
Wie hebt sich das geplante Vorhaben vom Stand der Technik ab?
Welche Vorteile existieren gegenüber bestehenden Lösungen?

## Markt und Wettbewerber

\textcolor{Red}{TODO}

Welche Zielmärkte werden durch das Vorhaben adressiert?
Welches wirtschaftliche Volumen haben diese Märkte?
Welche Wettbewerber sind auf diesen Märkten präsent?

|                      | **FOSS**            | **Dezentral**       | **Kein SPoF**                 | **Versionierung**                    | **Einfach nutzbar** | **P2P**         |  
| -------------------- | ------------------- | ------------------- | --------------------------- | -------------------------------------- | ------------------- |------------------|
| *Dropbox/Boxcryptor* | \xmark              | \xmark              | \xmark                      | \textcolor{YellowOrange}{Rudimentär}   | \cmark              | \xmark           |
| *ownCloud*           | \cmark              | \xmark              | \xmark                      | \textcolor{YellowOrange}{Rudimentär}   | \cmark              | \xmark           |
| *Syncthing*          | \cmark              | \cmark              | \cmark                      | \textcolor{YellowOrange}{Archivordner} | \cmark              | \xmark           |
| *BitTorrent Sync*    | \xmark              | \cmark              | \cmark                      | \textcolor{YellowOrange}{Archivordner} | \cmark              | \xmark           |
| ``git-annex``        | \cmark              | \cmark              | \cmark                      | \cmark                                 | \xmark              | \xmark           |
| ``brig``             | \cmark              | \cmark              | \cmark                      | \cmark                                 | \cmark              | \cmark           |


## Schutzrechtslage

Es sind uns keine Schutzrechte Dritter bekannt. Bei der Entwicklung wird auf
freie Software gesetzt, die eine libertäre Nutzung der Software ermöglicht und
selbst zudem so gut wie immer frei von Schutzrechten ist.

# Ausführliche Beschreibung des Vorhabens

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
* *Benutzerverwaltung* mittels dezentralen Logins.
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
Benutzer--Account der Form ``nutzer@server.de/ressource`` geknüpft.
Ein Überblick über die sicherheitsrelevanten Zusammenhänge findet sich
in Abbildung {@fig:security}.

Alle Änderungen an einem Repository werden in einer Metadatendatenbank
gespeichert. Diese kann dann mit anderen Teilnehmern über einen
authentifizierten und verschlüsselten Steuerkanal ausgetauscht werden. Jeder
Teilnehmer hat dadurch den gesamten Dateiindex. Die eigentlichen Dateien können
aber »irgendwo« im Teilnehmernetz sein. Sollte eine Datei lokal benötigt werden,
so kann man sie »pinnen«, um sie lokal zu speichern. Ansonsten werden nur selbst
erstellte Dateien gespeichert und andere Dateien maximal solange vorgehalten,
bis die Speicherquote erreicht ist.

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

# Wirtschaftliches Verwertungskonzept

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
``brig``--kompatiblen Benutzer--Account anlegt.

### Supportverträge

Normalerweise werden Fehler bei Open--Source--Projekten auf einen dafür
eingerichteten Bugtracker gemeldet. Die Entwickler können dann, nach einiger
Diskussion und Zeit, den Fehler reparieren. Unternehmen haben aber für
gewöhnlich kurze Deadlines bis etwas funktionieren muss.

Unternehmen mit Supportverträgen würden daher von folgenden Vorteilen profitieren:

- *Installation* der Software.
- *Priorisierung* bei Bug--Reports.
- Persönlicher *Kontakt* zu den Entwicklern.
- *Wartung* von nicht--öffentlichen *Spezialfeatures*.
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

# Hebelwirkung für den Wirtschaftsstandort Bayern

\textcolor{Red}{TODO: Ausformulieren}

- Innovation im Bereich *Industrie 4.0* und *Internet of Things*.
- Erweiterung des Geschäftsfeldes von Secomba,
  dadurch mehr Arbeitsplätze?
- Allgemein: Sichere Dateiübertragung und Transparenz im Unternehmens- und Behördenumfeld.

# Beschreibung des Arbeitsplans 

## Arbeitsschritte, Beiträge der Partner

Welche Arbeitspakte (ggf. Unterpakete bei >6 PM) sind geplant?
Was sind Inhalt und Ziele der Arbeitspakete?
Welcher Partner übernimmt welche Aufgaben pro Arbeitspaket?

Wie viele Personenmonate (PM) entfallen auf welchen Partner?

\textcolor{Red}{TODO: Aufgabenteilung mit Secomba/anderen?}

## Zeit- und Personalplan, Planungshilfen

Übersicht: Balkenplan (Zeitverlauf) der Arbeitspakete
Übersicht: Personenmonate pro Partner, pro Arbeitspaket

(siehe Anlage, Gantt Diagramm)

\textcolor{Red}{TODO: Personenmonate?}

## Meilensteinplanung

Die Meilensteine sind im Gantt--Diagramm ersichtlich und entsprechen jeweils
dem Ende der Projektphasen: 

- Abschluss Prototyp I
- Abschluss Prototyp II
- Ende der Stabilisierungsphase.
- Ende der Erweiterungsphase.

# Finanzierung des Vorhabens 

## Kostenplan

\textcolor{Red}{TODO: Minimaler/Komfortabler Kostenplan}

- Gesamtvolumen bestimmen
- Welche Pauschalen für E13
- 2x Halbtagsjob
- Kostenaufteilung auf Secomba und andere Partner? 

Welche Kosten pro Partner entstehen für Personal (Pauschalen beachten!), Material, Fremdleistungen und Sondereinzelkosten?

\newpage

# Literaturverzeichnis
