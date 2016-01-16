all:
	pandoc --smart --bibliography expose.bib --csl ieee.csl expose.md -B title.tex -H header.tex --filter pandoc-fignos --filter pandoc-tablenos -N -o expose.tex -V lang=de-DE --chapters
	latexmk -pdf expose.tex

