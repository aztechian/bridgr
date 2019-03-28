#!/usr/bin/env sh

STYLE=${1:-html}
[[ -n $JENKINS_URL ]] && STYLE=xunit
[[ -n $TRAVIS ]] && STYLE=cover

case $STYLE in
html)
  go test -covermode=count -coverprofile=coverage.out ./...
  go tool cover --html=coverage.out
  ;;
xunit)
  go get github.com/tebeka/go2xunit
  2>&1 go test -v -race ./... | tee xunitresults
  go2xunit -fail -input $outfile -output tests.xml
  ;;
cover)
  go test -v -race -covermode=count -coverprofile=c.out
  ;;
none)
  rm -f coverage.out
  go test ./...
  ;;
*)
  echo "Unkown coverage Style: ${STYLE}"
  ;;
esac
