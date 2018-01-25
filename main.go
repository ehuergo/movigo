package main

import (
    "fmt"
    "log"
    "sort"
    //"flag"
    "io"
    "os"
    //"net/url"
    //"github.com/alexflint/go-arg"
)

func main(){

    var err error

    opts := parseCommandLine()
    log.Printf("%+v", opts)
    log.Println(opts.readfrom.Scheme, "H", opts.readfrom.Host, "P", opts.readfrom.Port)

    if opts.verbose{
        log.SetFlags(log.LstdFlags | log.Lshortfile)
    }

    packages := map[string]string{
        "UTX32": "TDT",
        "UTX64": "Extra",
    }

    area := Area(opts.area)

    //areadfrom
    fromprefix := opts.readfrom.Raw

    if opts.readfrom.Scheme == "udp"{
        fromprefix = "udp://"
    }else if opts.readfrom.Scheme == "udpxy"{
        fromprefix = fmt.Sprintf("http://%s/udp/", opts.readfrom.Host)
    }else if opts.readfrom.Scheme != ""{
        log.Fatal("Unknown scheme", opts.readfrom.Raw)
    }else if opts.readfrom.Raw == ""{
        log.Fatal("No input specified")
    }else{
        fromprefix += "/"
    }    

    //saveto
    var writer io.Writer

    if opts.saveto.Raw == "stdout"{
        writer = os.Stdout
    }else{ 
        writer, err = os.OpenFile(opts.saveto.Raw, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0777); if err != nil{
            log.Fatal(err)
        }

        defer writer.(*os.File).Close()
    }

    //streamacceess
    streamprefix := opts.streamaccess.Raw
    if opts.streamaccess.Scheme == "udpxy"{
        streamprefix = fmt.Sprintf("http://%s/udp/", opts.streamaccess.Host)
    }
    //else keep untouched

    movi := NewMovi(area)
    ok := movi.Scan(fromprefix); if !ok{
        log.Fatal("Something went wrong scanning %s", area)
    }

    if opts.listpackages{
        movi.ListPackages()
        return
    }

    groups := movi.GetChannelGroups(packages)

    var keys []int
    for k := range groups{
        keys = append(keys, k)
    }
    sort.Ints(keys)

    //channels := movi.GetChannelList(nil) //packages)
    //DumpIPTVSimple(channels, "172.16.10.9", 9998)
    data := DumpGroupsAsIPTVSimple(groups, streamprefix)
    writer.Write(data)
    log.Printf("Channels written to %+v %s", writer, opts.saveto)
}

