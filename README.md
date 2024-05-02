# go-replace-undo
Utility to undo Go replace local copies for Veracode Static Analysis

## Download
On Linux with go get:

```
export GOPATH=`go env GOPATH` &&
export PATH="$GOPATH/bin:$PATH" &&
go install github.com/relaxnow/go-replace-undo@latest
```

## Run

```
go-replace-undo path/to/project/go.mod
```
