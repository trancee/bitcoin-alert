go-bindata -o asset/main.go -pkg asset -ignore=main.go asset/

# Build for Windows
env GOOS=windows GOARCH=amd64 go build -o bin/bitcoin-alert.exe

go build -o bin/bitcoin-alert
