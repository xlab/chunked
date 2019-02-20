## chunked

This is a PoC of a chunked upload manager. The chunking algorithm uses JS on the frontend side to split large files in chunks, each chunk is then encoded using AES-CBC-256 with provided private key (or randomly generated) and uploaded to the server. The server doesn't know the key to decode, so in order to guarantee the integrity of data we use Encrypt-then-MAC with AES-CMAC as the integrity verification algorithm. We use two different keys: keyAESMAC is a 128-bit AES key that is shared with server, but keyAESCBC is a 256-bit AES key and is used only to encrypt data on the frontend.

### Installation

```
$ mkdir -p $GOPATH/github.com/xlab/chunked
$ cd $GOPATH/github.com/xlab/chunked
$ git clone git@github.com:xlab/chunked.git .
$ go build
```

### Usage

```
$ ./chunked -h

Usage: chunked [OPTIONS] COMMAND [arg...]

A service for chunked and encrypted file uploads.

Options:
  -d, --uploads-dir   Specify chunk uploads directory. (default "uploads/")
  -l, --listen-addr   Specify server listen address. (default "127.0.0.1:2019")
  -w, --web-assets    Sepcify the web assets path to serve. (default "assets/")
  -k, --cmac-key      An AES-128 key for CMAC-CBC, generated using gen-keys command. (default "d2d2e0e43a87abd12baba39df25edc3f")

Commands:
  gen-keys            Generates AES keys for use in data encoding and integity checks.

Run 'chunked COMMAND --help' for more information on a command.
```

### Example

```
$ ./chunked gen-keys

AES-128 for CMAC-CBC: 389d69688f50776b9ab943eaa056ab47
AES-256 for AES-CBC: 99b253e281d069f059105bb1a97d4936d5b530464778fed8c03dc62a1c41c65a

$ ./chunked -k 389d69688f50776b9ab943eaa056ab47

Shared CMAC-CBC AES Key: 389d69688f50776b9ab943eaa056ab47
Open your browser at http://127.0.0.1:2019
```

Then use Web UI to upload the file, also check the JavaScript console for debug messages. The algorithm on the frontend is not effective for production IMO, it's totally synchronous.

### License

Not for sharing.
