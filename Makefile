.PHONY: setup add clean build auth

setup:
	go install github.com/spf13/cobra-cli@latest
	cobra-cli help

add:
	cobra-cli add $(COMMAND)

clean:
	rm ./gtx

build:
	go build -o gtx .
