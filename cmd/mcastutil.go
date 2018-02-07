package main

// All credits to: https://github.com/MovistarTV/tv_grab_es_movistartv/blob/master/tv_grab_es_movistartv.py#L610

import (
    "net"
    "flag"
    "os"
    "fmt"
    "io"
    //"io/ioutil"
    "log"
    "time"
    "encoding/binary"
    //"encoding/xml"
    "github.com/ziutek/utils/netaddr"
    "github.com/cheggaaa/pb"
)

var bar *pb.ProgressBar
var barenabled bool
var timeout time.Duration

func newbar(size int){

    if !barenabled{ return}

    if bar != nil{
        bar.Finish()
    }
    bar = pb.New(size)
    bar.ShowSpeed = true
    bar.SetWidth(120)
    //log.Println("BAR STARTS")
    bar.Start()
}

func endbar(){

    if !barenabled{ return}

    if bar != nil{
        bar.Finish()
    }
    //log.Println("BAR FINISH")
}
func writebar(b []byte){

    if !barenabled{ return}

    if bar != nil{
        bar.Write(b)
    }
}

func dumpraw(addrinfo string, limit int, w io.WriteCloser){
    addr, err := net.ResolveUDPAddr("udp", addrinfo); if err != nil {
        log.Fatal(err)
    }
    r, _ := net.ListenMulticastUDP("udp", nil, addr)
    defer r.Close()

    log.Println("Reading from", addrinfo, "limit", limit, "w", w)

    tot := 0
    newbar(limit)
    for {
        if limit > 0 && tot > limit{
            break
        }
        data := make([]byte, 1500)
        r.SetReadDeadline(time.Now().Add(timeout * time.Second))
        n, addr, err := r.ReadFromUDP(data); if err != nil{
            endbar()
            log.Println("Error reading from", addrinfo, err)
            return
        }
        log.Println(addr, addrinfo)
        w.Write(data)
        writebar(data)

        tot += n
    }
    endbar()
}

func scansave(netinfo string, port int){
    _, ipnet, err := net.ParseCIDR(netinfo); if err != nil{
        log.Fatal(err)
    }

    log.Printf("%+v", ipnet)
    log.Println(ipnet.IP)
    first := netaddr.IPAdd(ipnet.IP, 1)
    num := ^binary.BigEndian.Uint32(ipnet.Mask) - 1
    last := netaddr.IPAdd(ipnet.IP, int(num))

    for ip := first; ! ip.Equal(last); ip = netaddr.IPAdd(ip, 1){
        addrinfo := fmt.Sprintf("%s:%d", ip, port ) //3937, 8208
        path := "scan/" + addrinfo
        f, err := os.OpenFile(path, os.O_WRONLY | os.O_CREATE | os.O_TRUNC, 0664); if err != nil{
            log.Fatal(err)
        }
        dumpraw(addrinfo, 1500 * 10, f)
        f.Close()
    }

}

func main(){

    log.SetFlags(log.LstdFlags | log.Lshortfile)

    opt_scansave := flag.String("scansave", "", "Scan network range and save samples in scan/")
    opt_dumpraw := flag.String("dumpraw", "", "Dump traffic destinated to address")
    opt_dumpone := flag.String("dumpone", "", "Dump one packet destinated to address")
    opt_dumprtp := flag.String("dumprtp", "", "Dump RTP traffic destinated to address")
    opt_dumpsds := flag.String("dumpsds", "", "Dump SDS traffic destinated to address")
    opt_dumpepg := flag.String("dumpepg", "", "Dump EPG traffic destinated to address")
    opt_timeout := flag.Int("t", 1, "Set UDP read timeout in milliseconds")
    opt_port := flag.Int("port", 0, "Port")

    flag.Parse()

    timeout = time.Duration(*opt_timeout) * 100000000

    if *opt_scansave != ""{
        scansave(*opt_scansave, *opt_port)
    }

    if *opt_dumpraw != ""{
        dumpraw(*opt_dumpraw, 0, os.Stdout)
        return
    }

    if *opt_dumpone != ""{
        barenabled = false
        dumpraw(*opt_dumpone, 1, os.Stdout)
    }

    if *opt_dumprtp != ""{
    }
    if *opt_dumpsds != ""{
    }
    if *opt_dumpepg != ""{
    }


}

/*
func main1(){

    files := make(map[uint16]*File, 0)
    xmltvf := NewXMLTVFile()

    addr, err := net.ResolveUDPAddr("udp", os.Args[1]); if err != nil {
        log.Fatal(err)
    }

    log.Println(os.Args[1])

    r, _ := net.ListenMulticastUDP("udp", nil, addr)
    //r.(net.UDPConn).SetReadBuffer(4 * 1024 * 1024)

    for{
        data := make([]byte, 1500)
        n, _ := r.Read(data)
        fmt.Printf("% 5d ", n)
        chunk := ParseChunk(data[:n])
        log.Println("SKIP!")

        if chunk.End{ break }
    }

    filedata := make([]byte, 0)
    for len(files) < 500{
        data := make([]byte, 1500)
        n, _ := r.Read(data)
        fmt.Printf("% 5d ", n)
        chunk := ParseChunk(data[:n])
        filedata = append(filedata, chunk.Data...)
        if chunk.End{
            file := ParseFile(filedata)
            //ioutil.WriteFile("filedata", filedata, 0644)
            if file != nil && file.Size > 0 && len(file.Data) != 0{
                //fmt.Printf("FILE! %s", file.ServiceURL)
                _, ok := files[file.ServiceId]; if ok{
                    break
                }
                files[file.ServiceId] = file

                xmltvc := &XMLTVChannel{}
                xmltvc.Id = fmt.Sprintf("%d", file.ServiceId)
                xmltvc.Name = file.ServiceURL //FIXME
                xmltvf.Channel = append(xmltvf.Channel, xmltvc)

                programs := ParsePrograms(file.Data)
                for _, program := range programs{
                    //log.Printf("PROGRAM %+v\n", program)
                    xmltvp := &XMLTVProgramme{}
                    xmltvp.Start = program.Start.Format("20060102150405 -0700")
                    xmltvp.Stop = program.Start.Add(program.Duration).Format("20060102150405 -0700")
                    xmltvp.Channel = fmt.Sprintf("%d", file.ServiceId)
                    xmltvp.Title = program.Title
                    xmltvp.Date = program.Start.Format("20060102")
                    xmltvf.Programme = append(xmltvf.Programme, xmltvp)
                }
            }
            filedata = make([]byte, 0)
        }
    }

    xmltvdata, _ := xml.Marshal(xmltvf)
    err = ioutil.WriteFile("movi.xmltv", xmltvdata, 0644); if err != nil{
        log.Fatal(err)
    }
}
*/
