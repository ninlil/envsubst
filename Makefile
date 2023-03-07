check:
	@echo "Checking...\n"
	go test .
	@echo ""
	gocyclo -over 15 . || echo -n ""
	@echo ""
	golint -min_confidence 0.21 -set_exit_status ./...
	@echo ""
	go mod verify
	@echo "\nAll ok!"
