version=0.1.0

compile:
	GOOS=windows GOARCH=amd64 go build -o bin/frinkiac-bot-$(version)-windows-amd64.exe ./cmd/frinkiac
	GOOS=windows GOARCH=arm64 go build -o bin/frinkiac-bot-$(version)-windows-arm64.exe ./cmd/frinkiac
	GOOS=darwin GOARCH=amd64 go build -o bin/frinkiac-bot-$(version)-darwin-amd64 ./cmd/frinkiac
	GOOS=darwin GOARCH=arm64 go build -o bin/frinkiac-bot-$(version)-darwin-arm64 ./cmd/frinkiac

	zip frinkiac-bot-$(version)-windows-amd64.zip bin/frinkiac-bot-$(version)-windows-amd64.exe
	zip frinkiac-bot-$(version)-windows-arm64.zip bin/frinkiac-bot-$(version)-windows-arm64.exe
	zip frinkiac-bot-$(version)-darwin-amd64.zip bin/frinkiac-bot-$(version)-darwin-amd64
	zip frinkiac-bot-$(version)-darwin-arm64.zip bin/frinkiac-bot-$(version)-darwin-arm64