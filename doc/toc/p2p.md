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

# Abstract

Zusammenfassung in wenigen Worten.

# Danksagung

...

# Abbildungsverzeichnis

...

# Abkürzungsverzeichnis

...

# Einleitung

## Motivation und Problemstellung

Zentral vs Dezentral
NSA, Snowden
Die meiste Software zu schwer zu benutzen

## Projektziel

Vorstellung von brig

## Der Name

Siehe Expose

## Zielgruppe

- Endanwender
- Unternehmen
- Behörden

## Einsatzszenarien

- Dateitransfer
- Dateisynchronisation
- Datentresor
- Plattform für weitere Anwendungen

(siehe auch Expose)

## Lizensierung

AGPL, siehe Expose

# Stand der Technik

Viele Teillösungen, manche mehr oder weniger
gut, manche propetiär (siehe Konkurrenzanalyse)

brig versucht Sicherheit und Usability zu vereinen, da (Rob Pike) "Usability >
Sicherheit" oder "Geringe Absicherung ist trotzdem viel besser als gar keine"
Trotzdem Einsatz bewährter kryptografischer Protokolle und Primitiven, dieses
soll aber möglichst vom Benutzer versteckt werden.

## Wissenschaftlicher Stand

Eigenschaften eines P2P Netzwerkes

(syncthing kann zB keine daten routen)

### P2P-Netzwerke

### Ähnliche Arbeiten: Bazil

## Konkurrenzanalyse

### Dropbox/Boxcryptor

### Syncthing

### MooseFS

### LizardFS

## Problemstellung

## Wahl der Sprache

# Dezentrale Netzwerke

## IPFS

### Dezentrales Routing

### Merkle-Tree

### Pinning

### Speicherquoten

### Service Discovery

## Metadatenübertragung

### XMPP (historisch)

### MQTT

### Benutzerverwaltung

# Architektur von brig

## Client/Server Aufteilung

## Metadatenindex

### Datenstrukturen

### BoltDB

## Serialisierung

## Streaming Architektur

## Sonstiges

### Logging

### Konfiguration

# Storage

## Verschlüsselungslayer

## Kompressionslayer

## Deduplizierung

## FUSE Mount

# Usability

## Frontends

### Kommandozeile

### FUSE Mount

### Grafische Oberflächen

# Ausblick

## Selbstkritik

## Portierbarkeit

## Weitere Entwicklung

## Wirtschaftliche Verwertung

## Beiträge zu anderen Open Source Projekten

... minilock, goxmpp, ipfs...

# Anhänge























