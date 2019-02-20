package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	cli "github.com/jawher/mow.cli"
	"github.com/xlab/closer"
)

var app = cli.App("chunked", "A PoC service for chunked and encrypted file uploads.")
var uploadsDir = app.StringOpt("d uploads-dir", "uploads/", "Specify chunk uploads directory.")
var listenAddr = app.StringOpt("l listen-addr", "127.0.0.1:2019", "Specify server listen address.")
var assetsPath = app.StringOpt("w web-assets", "assets/", "Sepcify the web assets path to serve.")
var cmacKeyHex = app.StringOpt("k cmac-key", "d2d2e0e43a87abd12baba39df25edc3f", "An AES-128 key for CMAC-CBC, generated using gen-keys command.")

func init() {
	gin.SetMode(gin.ReleaseMode)
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	app.Command("gen-keys", "Generates AES keys for use in data encoding and integity checks.", genCmd)
	app.Action = mainCmd
}

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Fatalln("[ERR]", err)
	}
}

func mainCmd() {
	defer closer.Close()
	closer.Bind(func() {
		log.Println("Bye!")
	})

	// ensure that cmacKey is specified and correct.
	var cmacKey, err = hex.DecodeString(*cmacKeyHex)
	if err != nil {
		closer.Fatalln("[ERR]", err)
	} else if keyLen := len(cmacKey); keyLen != 16 {
		closer.Fatalf("[ERR] CMAC-CBC key must be a 128-bit AES key. Current length = %d bytes.", keyLen)
	}

	// Ensure the uploads directory exists
	if err := os.MkdirAll(*uploadsDir, 0700); err != nil {
		closer.Fatalln("[ERR]", err)
	}

	r := gin.Default()
	r.POST("/chunks/upload", uploadHandler(cmacKey))
	r.StaticFile("/", filepath.Join(*assetsPath, "index.html"))
	r.Static("/assets", *assetsPath)

	go func() {
		if err := r.Run(*listenAddr); err != nil {
			closer.Fatalln(err)
		}
	}()

	fmt.Println("Shared CMAC-CBC AES Key:", *cmacKeyHex)
	fmt.Printf("Open your browser at http://%s\n", *listenAddr)
	closer.Hold()
}

func genCmd(c *cli.Cmd) {
	c.Action = func() {
		fmt.Println("AES-128 for CMAC-CBC:", hex.EncodeToString(randKey(16)))
		fmt.Println("AES-256 for AES-CBC:", hex.EncodeToString(randKey(32)))
	}
}

func randKey(bytes int) []byte {
	key := make([]byte, bytes)

	if _, err := rand.Read(key); err != nil {
		panic(err)
	}
	return key
}
