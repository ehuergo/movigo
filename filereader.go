package main

import (
    "net/http"
    "os"
    "io/ioutil"
    "io"
    "log"
    //"strings"
)

func ReadAllAsBytes(uri string) []byte{
    r := getUriReader(uri)
    data, err := ioutil.ReadAll(r); if err != nil{
        log.Fatal(err)
        return nil
    }
    return data
}

func ReadAllAsString(uri string) string{
    return string(ReadAllAsBytes(uri))
}

func getUriReader(uri string) io.Reader{
    if len(uri) > 7 && uri[:7] == "http://" || len(uri) > 8 && uri[:8] == "https://"{ //FIXME: Improve this
        return getHttpReader(uri)
    }

    return getFileReader(uri)
}

func getHttpReader(uri string) io.Reader{
    get, err := http.Get(uri); if err != nil{
        log.Fatal(err)
    }

    return get.Body
}

func getFileReader(uri string) io.Reader{
    f, err := os.Open(uri); if err != nil{
        log.Fatal(err)
    }

    return f
}
