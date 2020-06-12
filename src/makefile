.PHONY: all rall test visual
all:
	go install cli/...

rall:
	go build -a cli/...

test:
	go test chord/test/...

visual:
	./visual/setup.sh