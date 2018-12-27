package devdashpb

// Run "go generate" in this directory to update. You need to:
//
// - have the protoc binary in your $PATH
// - go get github.com/golang/protobuf
//
// See https://github.com/golang/protobuf#installation for how to install
// the protoc binary.

//go:generate protoc --proto_path=$GOPATH/src:. --go_out=. devdash.proto
