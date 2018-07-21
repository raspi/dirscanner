integration-test:
	GOCACHE=off
	go test -v -run TestIntegration

coverage:
	go test -v -cover

test:
	go test -v
