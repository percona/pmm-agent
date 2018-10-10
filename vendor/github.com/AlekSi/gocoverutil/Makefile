all: test

install:
	go install -v ./...

test: install
	# run merge with single file to sort lines in it
	cd internal/test/package1 && \
		go test -coverprofile=package1.out -covermode=count && \
		gocoverutil -coverprofile=package1.out merge package1.out
	cd internal/test/package2 && \
		go test -coverprofile=package2.out -covermode=count && \
		gocoverutil -coverprofile=package2.out merge package2.out

	gocoverutil test -v

	gocoverutil -coverprofile=cover.out -ignore=github.com/AlekSi/gocoverutil/internal/test/ignored/... \
		test -v -covermode=count \
		github.com/AlekSi/gocoverutil/internal/test/package1 \
		github.com/AlekSi/gocoverutil/internal/test/package2 \
		github.com/AlekSi/gocoverutil/internal/test/...

	go tool cover -html=cover.out -o cover.html
