go-bindata -o asset/main.go -pkg asset -ignore=main.go asset/

go build -o bin/bitcoin-alert
