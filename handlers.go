package main

import (
	"crypto/aes"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	cmacAES "github.com/aead/cmac/aes"
	"github.com/gin-gonic/gin"
)

const (
	maxChunkSize = 128 * 1024 * 1024
)

func uploadHandler(cmacKey []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		chunk, err := c.FormFile("chunkData")
		if err != nil {
			log.Println("[WARN] failed to read chunk data:", err)
			c.AbortWithError(400, err)
			return
		}
		log.Println("Chunk file:", chunk.Filename, chunk.Size)
		f, err := chunk.Open()
		if err != nil {
			log.Println("[WARN] failed to read chunk file:", err)
			c.AbortWithError(500, err)
			return
		}
		chunkDataDecoder := base64.NewDecoder(base64.StdEncoding, f)
		chunkData, _ := ioutil.ReadAll(io.LimitReader(chunkDataDecoder, maxChunkSize))
		f.Close()

		chunkFilename := c.PostForm("chunkFilename")
		chunkFilenameParts := strings.Split(chunkFilename, `\`)
		chunkFilename = chunkFilenameParts[len(chunkFilenameParts)-1]
		chunkFilename = filepath.Base(chunkFilename)
		log.Println("Name:", chunkFilename)

		chunkMac := c.PostForm("chunkMac")
		macData, _ := hex.DecodeString(chunkMac)
		verified := cmacAES.Verify(macData, chunkData, cmacKey, aes.BlockSize)
		log.Println("MAC:", chunkMac, "Valid:", verified)

		if !verified {
			err := errors.New("CMAC verificaiton failed")
			log.Println("[WARN]", err)
			c.AbortWithError(400, err)
			return
		}

		// TODO(xlab):
		// Below there is an extremely stupid way of saving chunks. Just for debug purposes.
		// The real logic must invoke MongoDB transaction saving a header to collection,
		// also data transfer to a remote storage, with proper timeouts, etc.

		targetDir := filepath.Join(*uploadsDir, chunkFilename)
		if err := os.MkdirAll(targetDir, 0700); err != nil {
			log.Println("[ERR] failed to create target directory:", err)
			c.AbortWithError(500, err)
			return
		}
		err = ioutil.WriteFile(filepath.Join(targetDir, chunkMac), chunkData, 0600)
		if err != nil {
			log.Println("[ERR] failed to write chunk file:", err)
			c.AbortWithError(500, err)
			return
		}
	}
}
