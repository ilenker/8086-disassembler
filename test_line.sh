#!/bin/sh

if [ "$#" -lt 1 ]; then
	echo "no input"
	exit 1
fi

echo -e "bits 16\n$1" > temp.asm
shift
if nasm temp.asm; then
	output=$(go run . temp "$@" 2>&1)
	printf "%s\n" "${output}"
fi
