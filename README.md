// generate by protoc

# Install protoc

brew install protobuf

# Install protoc-gen-go and protoc-gen-go-grpc

go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# bash

echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.bash_profile
source ~/.bash_profile

# zsh

echo 'export PATH=$PATH:$(go env GOPATH)/bin' >> ~/.zshrc
source ~/.zshrc

# Check

which protoc-gen-go
which protoc-gen-go-grpc

protoc --go_out=. --go-grpc_out=. inventory.proto
