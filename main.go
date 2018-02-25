package main

import (
    "fmt"
    "log"
//    "sort"
    //"flag"
    "io"
    "os"
    "./readers"
    "./movi"
    "./output"
    //"github.com/juanmasg/rtpx/rtpx"
    //"github.com/juanmasg/rtpx/rtpx/routers"
    "./rtpx"
    "./rtpx/routers"
    "golang.org/x/net/webdav"
    "net/http"
    //"net/url"
    //"github.com/alexflint/go-arg"
    "time"
    "strings"
    //"./vfs"
)

var(
    m               *movi.Movi
    GetReader       func(string) io.Reader
    fromprefix      string
    opts            *Opts
    err             error
    area            movi.Area
    proxy           *rtpx.Proxy
    channels        []*movi.LogicalChannel

    ignorepkgs      map[string]bool
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

//func now(){
//    chprogram := m.FindCurrent()
//    for chname, program := range chprogram{
//        if program == nil{
//            log.Printf("%-30s %s", chname, "Unknown")
//        }else{
//            log.Printf("%-30s %s", chname, program.String())
//        }
//    }
//}
//
//func searchepg(){
//    chprograms := m.FindProgram(opts.searchepg, opts.season, opts.episode, opts.title, false)
//    for chname, programs := range chprograms{
//        log.Println(chname)
//        for _, program := range programs{
//            log.Println(" - " + program.String())
//        }
//    }
//}
//
func listpackages(){
    packages := m.GetPackages()
    for name, channels := range packages{
        log.Printf("\n-> Package: % 5s Channels % 4d SD % 3d HD % 3d\n", name, len(channels), 0 , 0)
        for _, channel := range channels{
            printChannel(name, channel)
        }
    }
}

func listchannels(){
    channels := m.GetChannelList(nil, ignorepkgs, true, 1000)
    for _, channel := range channels{
        printChannel("", channel)
    }
}

func printChannel(pname string, channel *movi.LogicalChannel){
    hasepg := channel.EPG != nil
    log.Printf("% 5s % 5d (% 5d) % -28s epg:%t\n", pname, channel.Number, channel.Id, channel.Name, hasepg)
}



func savem3u(){
    channels := m.GetChannelList(nil, ignorepkgs, true, 1000)

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
    channels := m.GetChannelList(nil, ignorepkgs, true, 1000)
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
    proxy = rtpx.NewProxy()
    go proxy.Loop()
    routers.SetProxy(proxy)
    routers.SetAutorec(true)
    //proxy.SetMovi(m)

    dav := webdav.Handler{Prefix: "/dav/"}
    dav.Logger = func(r *http.Request, err error){
        log.Println(err, r)
    }
    dav.FileSystem = webdav.Dir("/media/data/movigo")
    dav.LockSystem = webdav.NewMemLS()

    go routers.NewHTTPServer(opts.proxy, map[string]func(http.ResponseWriter, *http.Request){
        "/udp/": RTPToHTTP,
        //"/rtp/": routers.RTPToHTTP,
        //"/rec/": routers.HTTPRec,
        "/dav/": dav.ServeHTTP,
        //"/rtpdav/": routers.RTPToHTTPViaDAV,
    })
    log.Println("Proxy ready", opts.proxy)
}

func loop_updates(){

    for{
        channels = m.GetChannelList(nil, ignorepkgs, true, 1000)
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

    ignorepkgs = map[string]bool{ "UTY0H": true, "UTX0I": true, "UTX2D": true }


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
        ok := m.Scan(GetReader, fromprefix, opts.cachedays, false); if !ok{
            log.Fatal("Something went wrong scanning %s", area)
        }
    }else{
        if ! m.LoadCaches(){
            ok := m.Scan(GetReader, fromprefix, opts.cachedays, false); if !ok{
                log.Fatal("Something went wrong scanning %s", area)
            }
        }
        setupproxy()
        loop_updates()
        return
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

func HTTPToHTTP(w http.ResponseWriter, req *http.Request){
}

func RTPToHTTP(w http.ResponseWriter, req *http.Request){

    log.Println(req)

    if req.Method != "GET"{
        return
    }

    addrinfo := strings.Split(req.URL.Path, "/")[2]
    channel := m.GetChannelByUri(addrinfo)
    program := channel.EPGNow()
    log.Println("Requested to channel", channel.Name, program)

    proxy.RegisterReader(addrinfo)

    filepath := fmt.Sprintf("/media/data/movigo/live/%s", channel.Name, time.Now().Format("20060102, Mon"))
    os.MkdirAll(filepath, 0755)

    //filename := fmt.Sprintf("%s/%s.ts", filepath, program.Filename())

    wc := NewWriterChannel(w)
    defer wc.Close()

    //c := make(chan []byte, 1024)
    proxy.RegisterWriter(addrinfo, wc.c)

    started := time.Now()

    done := false
    closed := w.(http.CloseNotifier).CloseNotify()

    for{
        select{
        case b := <-wc.c:
            n, err := wc.Write(b); if err != nil{
                log.Println(err, n)
                done = true
                break
            }
        case <- closed:
            done = true
            break
        }
        if done{ break }
    }

    proxy.RemoveWriter(addrinfo, wc.c)

    //os.Rename(filename, fmt.Sprintf("%s/%s (%s).ts", filepath, program.Filename(), time.Now().Sub(started)))

    log.Println("RTPtoHTTP session for", addrinfo, "lasted", time.Now().Sub(started))
}

type WriterChannel struct{
    w       io.Writer
    c       chan []byte
    f       *os.File
}

//func NewWriterChannel(w io.Writer, backup string) *WriterChannel{
func NewWriterChannel(w io.Writer) *WriterChannel{
    wc := &WriterChannel{}
    wc.w = w
    wc.c = make(chan []byte, 1024)
    //if backup != ""{
    //    wc.f, err = os.OpenFile(backup, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0777); if err != nil{
    //        log.Println("WARNING: cannot open file for writing", savepath, err)
    //    }
    //}

    return wc
}

func (wc *WriterChannel) Write(b []byte) (int, error){
    //wc.f.Write(b)
    return wc.w.Write(b)
}

func (wc *WriterChannel) Close(){
    //wc.w.(io.WriteCloser).Close()
    //wc.f.Close()
}


