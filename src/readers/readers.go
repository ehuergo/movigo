package readers

import (
    "os"
    "net"
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

func GetMulticastReader(uri string) io.Reader{
    addr, err := net.ResolveUDPAddr("udp", uri); if err != nil {
        log.Fatal(err)
    }
    r, _ := net.ListenMulticastUDP("udp", nil, addr)

    return r

}
