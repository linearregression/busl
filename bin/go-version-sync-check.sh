#!/bin/bash

travis=$(cat .travis.yml | grep '^go:' | sed 's/^go: \(.*\)$/\1/')
godep=$(cat Godeps/Godeps.json | grep 'GoVersion' | sed 's/.*"go\(.*\)".*$/\1/')
diff <(echo $travis) <(echo $godep)
