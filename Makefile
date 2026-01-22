.PHONY: build clean

build:
	go build -o tig .

clean:
	rm ./tig