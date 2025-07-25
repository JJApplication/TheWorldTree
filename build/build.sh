go env GOPROXY
go env GOOS
go env GOARCH
echo "start to build"
go build -o twt -trimpath