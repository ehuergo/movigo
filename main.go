package main

import (
    "fmt"
    "log"
    "sort"
    //"flag"
    "net/url"
    "github.com/alexflint/go-arg"
)

type Options struct{
    read_from       string          // udpxy://IP:PORT
                                    // udp://
                                    // path/to/file

    save_to         string          // stdout
                                    // path/to/file

    playlist_format string          // m3u
    stream_access   string          // udpxy | udp | rtp

    area            int             // see movistartv.go for area details
}


var opts struct{
    ReadFrom        string  `arg:"-r" help:"Access method. udp:// reads straight from the network. udpxy://IP:PORT reads via udpxy proxy. Otherwise it's considered a file"`
    SaveTo          string  `arg:"-s" help:"stdout Dumps the file to stdout. Otherwise it's considered a path in the filesystem. Defaults to stdout"`
    PlaylistFmt     string  `arg:"-f" help:"Only m3u is supported"`
    StreamAccess    string  `arg:"-x" help:"udpxy://IP:PORT, udp:// or rtp://"`
    Area            Area    `arg:"-a" help:"Area code"`
}

func main(){

    opts.Area = MADRID

    arg.MustParse(&opts)
    log.Printf("%+v\n", opts)

    readfrom, err := url.Parse(opts.ReadFrom); if err != nil{
        log.Fatal(err)
    }
    saveto, err := url.Parse(opts.SaveTo); if err != nil{
        log.Fatal(err)
    }
    streamaccess, err := url.Parse(opts.StreamAccess); if err != nil{
        log.Fatal(err)
    }

    fromprefix := ""    //file

    if readfrom.Scheme == ""{
        fromprefix = ""
    }else if readfrom.Scheme == "udp"{
        fromprefix = "udp://"
    }else if readfrom.Scheme == "udpxy"{
        fromprefix = fmt.Sprintf("http://%s:%d/udp/", readfrom.Host, readfrom.Port)
    }

    movi := NewMovi(MADRID)
    ok := movi.Scan(fromprefix); if !ok{
        log.Fatal("Something went wrong scanning %s", MADRID)
    }

    packages := map[string]string{
        "UTX32": "TDT",
        "UTX64": "Extra",
    }

    //movi.ListPackages()

    groups := movi.GetChannelGroups(nil)

    var keys []int
    for k := range groups{
        keys = append(keys, k)
    }
    sort.Ints(keys)

    //for _, group := range groups{
    //for _, k := range keys{
    //    group := groups[k]
    //    log.Println(group.Number, len(group.SD), len(group.HD))
    //}
    //log.Println("GROUPS", groups)
    //channels := movi.GetChannelList(nil) //packages)
    //DumpIPTVSimple(channels, "172.16.10.9", 9998)
    DumpGroupsAsIPTVSimple(groups, "172.16.10.9", 9998)
}
