zig-installer:
	go build -o zig-installer main.go
clean:
	rm zig-installer || true
