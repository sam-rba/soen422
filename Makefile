SRC = server.go record.go

server: ${SRC}
	go build $^

fmt:
	gofmt -l -s -w ${SRC}

clean:
	rm -f server
