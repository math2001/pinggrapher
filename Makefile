.PHONY: run
.SILENT:

run: pinggrapher
	./pinggrapher

pinggrapher: *.go
	go build -i -o pinggrapher
