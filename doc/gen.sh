#!/bin/sh

docker run -it --rm -v /Users/asm/Dev/pandemic-projects/personal-website/backend/doc:/doc quay.io/bukalapak/snowboard html -o spec-doc spec.apib
