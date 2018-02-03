package main

import (
    "flag"
    "net/url"
)

type Opts struct{
    readfrom        URL             // udpxy://IP:PORT
                                    // udp://
                                    // path/to/file

    savem3u         URL             // stdout
                                    // path/to/file

    savexmltv       URL             // stdout
                                    // path/to/file

    streamaccess    URL             // udpxy | udp | rtp

    area            int             // see movistartv.go for area details
    listpackages    bool
    verbose         bool
}

type URL struct{
    url.URL
    Raw     string
}
func (u *URL) String() string{
    return u.Raw
}
func (u *URL) Set(s string) (err error){
    parsed, err := url.Parse(s)
    u.URL = *parsed
    u.Raw = s
    return
}

func parseCommandLine() *Opts{

    opts := &Opts{}

    flag.Var(&opts.readfrom, "readfrom", "Access method. udp:// reads straight from the network. udpxy://IP:PORT reads via udpxy proxy. Otherwise it's considered a file")
    flag.Var(&opts.savem3u, "savem3u", "stdout Dumps the file to stdout. Otherwise it's considered a path in the filesystem. Defaults to stdout")
    flag.Var(&opts.savexmltv, "savexmltv", "stdout Dumps the file to stdout. Otherwise it's considered a path in the filesystem. Defaults to stdout")
    flag.Var(&opts.streamaccess, "streamaccess", "udpxy://IP:PORT, udp:// or rtp://")
    flag.IntVar(&opts.area, "area", int(MADRID), "Area code")
    flag.BoolVar(&opts.listpackages, "l", false, "")
    flag.BoolVar(&opts.verbose, "v", false, "")

    //opts.savem3u.Raw = "stdout"

    flag.Parse()

    return opts
}
