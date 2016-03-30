security:
	pandoc --smart security.md -B title-security.tex -H header.tex --filter pandoc-fignos --filter pandoc-tablenos -N -o security.tex -V lang=de-DE --chapters
	latexmk -pdf security.tex

p2p:
	pandoc --smart p2p.md -B title-p2p.tex -H header.tex --filter pandoc-fignos --filter pandoc-tablenos -N -o p2p.tex -V lang=de-DE --chapters
	latexmk -pdf p2p.tex
