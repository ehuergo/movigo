package main

import (
    "flag"
    "net/url"
    "movigo/movi"
)

type Opts struct{
    readfrom        URL             // udpxy://IP:PORT
                                    // udp://
                                    // path/to/file

    savem3u         URL             // stdout
                                    // path/to/file

    savexmltv       URL             // stdout
                                    // path/to/file

    streamprefix    URL             // udpxy | udp | rtp

    area            int             // see movistartv.go for area details
    cachedays       int
    listpackages    bool
    listchannels    bool
    searchepg       string
    season          string
    episode         string
    title           string
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
    flag.Var(&opts.streamprefix, "streamprefix", "udpxy://IP:PORT, udp:// or rtp://")
    flag.IntVar(&opts.area, "area", int(movi.MADRID), "Area code")
    flag.IntVar(&opts.cachedays, "cachedays", 1, "Ignore cache if older than N days (default: 1)")
    flag.BoolVar(&opts.listpackages, "packages", false, "")
    flag.BoolVar(&opts.listchannels, "channels", false, "")
    flag.BoolVar(&opts.verbose, "v", false, "")
    flag.StringVar(&opts.searchepg, "searchepg", "", "Search a program")
    flag.StringVar(&opts.season, "season", "", "filter season for -searchepg")
    flag.StringVar(&opts.episode, "episode", "", "filter episode for -searchepg")
    flag.StringVar(&opts.title, "title", "", "filter title for -searchepg")

    //opts.savem3u.Raw = "stdout"

    flag.Parse()

    return opts
}
