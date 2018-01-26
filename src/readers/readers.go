package readers

import (
    "os"
    "net/http"
    "log"
    "io"
)

func GetHttpReader(uri string) io.Reader{
    get, err := http.Get(uri); if err != nil{
        log.Fatal(err)
    }

    return get.Body
}

func GetFilesystemReader(uri string) io.Reader{
    log.Println(uri)
    f, err := os.Open(uri); if err != nil{
        log.Fatal(err)
    }

    return f
}
