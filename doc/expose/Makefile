all:
	pandoc --smart --bibliography expose.bib --csl ieee.csl expose.md -B title.tex -H header.tex -N --filter pandoc-fignos -o expose.tex -V lang=ngerman
	latexmk -pdf -pdflatex=pdflatex  -interaction=batchmode expose.tex
