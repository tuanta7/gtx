.PHONY: build dev clean

build:
	go build -o tig .

auth:
	go build -o tig .
	./tig auth

clean:
	rm ./tig