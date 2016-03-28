all:
	pandoc --smart --bibliography expose.bib --csl ieee.csl expose.md -B title.tex -H header.tex --filter pandoc-fignos --filter pandoc-tablenos -N -o expose.tex -V lang=de-DE --chapters
	latexmk -pdf expose.tex

iuk:
	pandoc --smart --bibliography expose.bib --csl ieee.csl iuk.md -B iuk-title.tex -H header.tex --filter pandoc-fignos --filter pandoc-tablenos -N -o iuk.tex -V lang=de-DE --chapters
	latexmk -pdf iuk.tex
	pdfunite iuk.pdf images/gantt.pdf iuk_final.pdf

