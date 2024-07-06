docker run --rm -v "$PWD":/app -w /app -e GOOS=linux -e GOARCH=arm64 -e CGO_ENABLED=1 golang:1.22 go build -o ./finaltask ./cmd/final-project/
docker build --no-cache --tag finaltask:v1 .
rm ./finaltask