GOC=GOOS=windows go build
GOFLAGS=-a -ldflags '-s'
CGOR=CGO_ENABLED=0

all: build

build:
	$(GOC)

run:
	go run main.go

stat:
	$(CGOR) $(GOC) $(GOFLAGS) 

noterm:
	$(GOC) -ldflags="-H windowsgui"

qt:
	${GOPATH}/bin/qtdeploy -docker test desktop

qtfast:
	${GOPATH}/bin/qtdeploy -docker -fast test desktop

fmt:
	gofmt -w .

clean:
	rm airlock.exe

qtclean:
	sudo rm -r deploy qml
	sudo rm rcc*
