.PHONY: clean run

clean:
	go clean
	go mod tidy
	rm *.out
	rm coverage.html

run-core:
	chmod +x scripts/run-core.sh
	./scripts/run-core.sh

run-discord:
	chmod +x scripts/run-discord.sh
	./scripts/run-discord.sh

keygen:
	chmod +x scripts/keygen.sh
	./scripts/keygen.sh