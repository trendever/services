#!/bin/sh

find . -name '*.go' -and -not -name 'generated.go' -delete
cp /usr/lib/go/src/net/http/*.go .
rm -f ./*_test.go
sed -i 's/"golang_org/"golang.org/' *.go
sed -i '/go:generate/d' ./h2_bundle.go # disable codegen
for p in ./*.patch; do patch -p0 < $p; done
