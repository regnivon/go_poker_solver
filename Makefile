build:
	go build -o bin/solver driver/main.go

run: build
	./bin/solver

profile: build
	./bin/solver -cpuprofile cpu.txt