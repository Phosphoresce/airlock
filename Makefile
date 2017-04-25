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

fmt:
	gofmt -w .

clean:
	rm airlock.exe
