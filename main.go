package main

import (
    "fmt"
    "log"
//    "sort"
    //"flag"
    "io"
    "os"
    "movigo/readers"
    "movigo/movi"
    "movigo/output"
    "rtpx"
    "rtpx/routers"
    "golang.org/x/net/webdav"
    "net/http"
    //"net/url"
    //"github.com/alexflint/go-arg"
    "time"
)

var(
    m *movi.Movi
    GetReader func(string) io.Reader
    fromprefix string
    opts *Opts
    err error
    area movi.Area
)

func setup_reader(){
    fromprefix = opts.readfrom.Raw
    if opts.readfrom.Scheme == "udp"{
        fromprefix = ""
        GetReader = readers.GetMulticastReader
    }else if opts.readfrom.Scheme == "udpxy"{
        fromprefix = fmt.Sprintf("http://%s/udp/", opts.readfrom.Host)
        GetReader = readers.GetHttpReader
    }else if opts.readfrom.Scheme != ""{
        log.Fatal("Unknown scheme", opts.readfrom.Raw)
    }else if opts.readfrom.Raw == ""{
        log.Fatal("No input specified")
    }else{
        fromprefix += "/"
        GetReader = readers.GetFilesystemReader
    }
}

func now(){
    chprogram := m.FindCurrent()
    for chname, program := range chprogram{
        if program == nil{
            log.Printf("%-30s %s", chname, "Unknown")
        }else{
            log.Printf("%-30s %s", chname, program.String())
        }
    }
}

func searchepg(){
    chprograms := m.FindProgram(opts.searchepg, opts.season, opts.episode, opts.title, false)
    for chname, programs := range chprograms{
        log.Println(chname)
        for _, program := range programs{
            log.Println(" - " + program.String())
        }
    }
}

func listpackages(){
    packages := m.GetPackages()
    for name, channels := range packages{
        log.Printf("\n-> Package: %s Channels %d\n", name, len(channels))
        for _, channel := range channels{
            printChannel(channel)
        }
    }
}

func listchannels(){
    channels := m.GetChannelList(nil, true, 1000)
    for _, channel := range channels{
        printChannel(channel)
    }
}

func savem3u(){
    channels := m.GetChannelList(nil, true, 1000)

    streamprefix := opts.streamprefix.Raw
    if opts.streamprefix.Scheme == "udpxy"{
        streamprefix = fmt.Sprintf("http://%s/udp/", opts.streamprefix.Host)
    }
    //else keep untouched

    data := output.DumpIPTVSimple(channels, streamprefix)

    var m3uwriter io.Writer

    if opts.savem3u.Raw == "stdout"{
        m3uwriter = os.Stdout
    }else if opts.savem3u.Raw != ""{
        m3uwriter, err = os.OpenFile(opts.savem3u.Raw, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0777); if err != nil{
            log.Fatal(err)
        }

        defer m3uwriter.(*os.File).Close()
    }

    m3uwriter.Write(data)
    log.Printf("Channels written to %+v %s", m3uwriter, opts.savem3u)
}

func savexmltv(){
    channels := m.GetChannelList(nil, true, 1000)
    data := output.DumpXMLTVEPG(channels)

    var xmltvwriter io.Writer

    if opts.savexmltv.Raw == "stdout"{
        xmltvwriter = os.Stdout
    }else if opts.savexmltv.Raw != ""{
        xmltvwriter, err = os.OpenFile(opts.savexmltv.Raw, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0777); if err != nil{
            log.Fatal(err)
        }

        defer xmltvwriter.(*os.File).Close()
    }

    xmltvwriter.Write(data)
    log.Printf("XMLTV written to %+v %s", xmltvwriter, opts.savexmltv)
}

func setupproxy(){
    proxy := rtpx.NewProxy()
    go proxy.Loop()
    routers.SetProxy(proxy)
    routers.SetAutorec(true)
    //proxy.SetMovi(m)

    dav := webdav.Handler{Prefix: "/dav/"}
    dav.Logger = func(r *http.Request, err error){
        log.Println(err, r)
    }
    dav.FileSystem = webdav.Dir("rec/")
    dav.LockSystem = webdav.NewMemLS()

    go routers.NewHTTPServer(opts.proxy, map[string]func(http.ResponseWriter, *http.Request){
        "/udp/": routers.RTPToHTTP,
        "/rtp/": routers.RTPToHTTP,
        "/rec/": routers.HTTPRec,
        "/dav/": dav.ServeHTTP,
        "/rtpdav/": routers.RTPToHTTPViaDAV,
    })
    log.Println("Proxy ready", opts.proxy)
}

func loop(){
    for{
        time.Sleep(60 * time.Minute)
        ok := m.Scan(GetReader, fromprefix, 0, true); if !ok{
            log.Fatal("Something went wrong scanning %s", area)
        }
        log.Println("Scan finished")
        savem3u()
        savexmltv()
    }
}

func main(){

    opts = parseCommandLine()
    log.Printf("%+v", opts)
    log.Println(opts.readfrom.Scheme, "H", opts.readfrom.Host, "P", opts.readfrom.Port)

    if opts.verbose{
        log.SetFlags(log.LstdFlags | log.Lshortfile)
    }

    //packages := map[string]string{
    //    "UTX32": "TDT",
    //    "UTX64": "Extra",
    //}

    area = movi.Area(opts.area)

    setup_reader()

    m = movi.NewMovi(area, opts.cachedays)
    if opts.proxy == ""{
        ok := m.Scan(GetReader, fromprefix, 0, false); if !ok{
            log.Fatal("Something went wrong scanning %s", area)
        }
    }else{
        m.LoadCaches()
        setupproxy()
        loop() //update periodically
        return
    }

    if opts.now{
        now()
    }

    if opts.searchepg != ""{
        searchepg()
    }

    if opts.listpackages{
        listpackages()
    }

    if opts.listchannels{
        listchannels()
    }

    if opts.savem3u.Raw != ""{ 
        savem3u()
    }

    if opts.savexmltv.Raw != ""{
        savexmltv()
    }

}


func printChannel(channel *movi.LogicalChannel){
    hasepg := channel.EPG != nil
    log.Printf("% 5d (% 5d) % -28s %t\n", channel.Number, channel.Id, channel.Name, hasepg)
}


