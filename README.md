# protodump

Protodump is a utility to dump all Protobuf file descriptors from a given binary as *.proto files:

![Demo](https://raw.githubusercontent.com/arkadiyt/protodump/main/demo/demo.gif)

## Usage

```
git clone https://github.com/arkadiyt/protodump
cd protodump
go build -o protodump cmd/main.go
./protodump -file <file to extract from> -output <output directory>
```

This was thrown together in a day and still has bugs :)

## Getting in touch

Feel free to contact me on twitter: https://twitter.com/arkadiyt
