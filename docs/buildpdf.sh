#!/bin/bash

# Run pdflatex a few times
for i in {1..5}
do
	pdflatex dokumentace.tex
done

# Clean non-pdf files
rm dokumentace.aux
rm dokumentace.log
rm dokumentace.lot
rm dokumentace.out
rm dokumentace.toc
