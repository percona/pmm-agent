# gocoverutil

[![Release](https://github-release-version.herokuapp.com/github/AlekSi/gocoverutil/release.svg?style=flat)](https://github.com/AlekSi/gocoverutil/releases/latest)
[![Travis CI](https://travis-ci.org/AlekSi/gocoverutil.svg?branch=master)](https://travis-ci.org/AlekSi/gocoverutil)
[![AppVeyor](https://ci.appveyor.com/api/projects/status/bxcbywwapyvsprju/branch/master?svg=true)](https://ci.appveyor.com/project/AlekSi/gocoverutil)
[![Codecov](https://codecov.io/gh/AlekSi/gocoverutil/branch/master/graph/badge.svg)](https://codecov.io/gh/AlekSi/gocoverutil)
[![Coveralls](https://coveralls.io/repos/github/AlekSi/gocoverutil/badge.svg?branch=master)](https://coveralls.io/github/AlekSi/gocoverutil)
[![Go Report Card](https://goreportcard.com/badge/AlekSi/gocoverutil)](https://goreportcard.com/report/AlekSi/gocoverutil)


Install it with `go get`:
```
go get -u github.com/AlekSi/gocoverutil
```

gocoverutil contains two commands: merge and test.

Merge command merges several go coverage profiles into a single file.
Run `gocoverutil merge -h` for usage information. Example:
```
gocoverutil -coverprofile=cover.out merge internal/test/package1/package1.out internal/test/package2/package2.out
```

Test command runs `go test -cover` with correct flags and merges profiles.
Packages list is passed as arguments; they may contain `...` patterns.
The list is expanded, sorted and duplicates and ignored packages are removed.
`go test -coverpkg` flag is set automatically to the same list.
Only a single package is passed at once to `go test`, so it always acts as if `-p 1` is passed.
If tests are failing, gocoverutil exits with a correct exit code.
Run `gocoverutil test -h` for usage information. Example:
```
gocoverutil -coverprofile=cover.out test -v -covermode=count github.com/AlekSi/gocoverutil/internal/test/...
```
