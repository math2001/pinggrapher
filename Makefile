.PHONY: run test clean
.SILENT:

test: pinggrapher
	./pinggrapher 192.168.1.1


run: pinggrapher
	./pinggrapher

pinggrapher: *.go
	go build -i -o pinggrapher

clean:
	mv .pings /tmp # safer than rm
