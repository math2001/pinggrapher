.PHONY: run test clean
.SILENT:

test: pinggrapher
	ping 192.168.1.1 \
		| awk  '/from/ { split($$7, resArr, "="); print resArr[2]; fflush() }'\
		| ./pinggrapher -delay 5000


run: pinggrapher
	./pinggrapher

pinggrapher: *.go
	go build -i -o pinggrapher

clean:
	mv .pings /tmp # safer than rm
