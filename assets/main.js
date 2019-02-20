const CHUNK_SIZE = 8 * 1024 * 1024; // 8Mb

$("#chunkedForm").submit(function(e) {
    
    var fileName = e.target.file.value;
    console.log("Chunked upload begins for:", fileName)
    var keyAESMAC = CryptoJS.enc.Hex.parse(e.target.keyAESMAC.value);
    var keyAESCBC = CryptoJS.enc.Hex.parse(e.target.keyAESCBC.value);
    console.log("AES CMAC-CBC 128-bit Key (hex-encoded):", keyAESMAC);
    console.log("AES-CBC 256-bit Key (hex-encoded):", keyAESCBC);
    var iv = CryptoJS.enc.Hex.parse("00000000000000000000000000000000");
    console.log("IV", iv.toString());

    chunkUpload = function(chunk) {
        console.log("chunk length:", chunk.length);
        var encrypted = CryptoJS.AES.encrypt(chunk, keyAESCBC, { iv: iv, format: CryptoJS.format.OpenSSL });
        var mac = CryptoJS.CMAC(keyAESMAC, encrypted.ciphertext);
        console.log("encrypted CMAC:", mac.toString());

        var formData = new FormData();
        var encryptedBlob = new Blob([encrypted], { type: "application/octet-stream" });
        formData.append("chunkFilename", fileName)
        formData.append("chunkData", encryptedBlob)
        formData.append("chunkMac", mac.toString())
        
        return $.post({
            url: "/chunks/upload",
            method: "POST",
            async: false,
            data: formData,
            processData: false,
            contentType: false
        })

        // TODO: upload chunk while encrypting the next one.
        // Actually it's a job for a JS developer with more advanced libs.
    }

    parseFile(e.target.file.files[0], chunkUpload, CHUNK_SIZE);

    return false;
});

// parseFile is a chunker function from https://stackoverflow.com/a/28318964
function parseFile(file, readCallback, chunkSize) {
    var fileSize = file.size;
    var offset = 0;
    var self = this; // we need a reference to the current object
    var chunkReaderBlock = null;

    var readEventHandler = function(e) {
        if (e.target.error == null) {
            offset += e.target.result.length;
            readCallback(e.target.result);
        } else {
            console.log("Read error: " + e.target.error);
            return;
        }
        if (offset >= fileSize) {
            return;
        }
        chunkReaderBlock(offset, chunkSize, file);
    }

    chunkReaderBlock = function(_offset, length, _file) {
        var r = new FileReader();
        var blob = _file.slice(_offset, _offset + length);
        r.onload = readEventHandler;
        r.readAsBinaryString(blob);
    }
    chunkReaderBlock(offset, chunkSize, file);
}