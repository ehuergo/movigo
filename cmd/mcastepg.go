package main

// All credits to: https://github.com/MovistarTV/tv_grab_es_movistartv/blob/master/tv_grab_es_movistartv.py#L610

import (
    "net"
    "os"
    "fmt"
    "io/ioutil"
    "log"
    "time"
    "encoding/binary"
    "encoding/xml"
)

type Chunk struct{
    End bool
    Size uint32
    Type uint8
    FileId uint16
    ChunkNumber uint16
    ChunkCount uint8
    Data []byte
}

func ParseChunk(data []byte) *Chunk{
    chunk := &Chunk{}
    chunk.End = data[0] == 0x01
    chunk.Size = Uint32(append([]byte{0}, data[1:4]...))
    chunk.Type = data[4]
    chunk.FileId = Uint16(data[5:7])
    chunk.ChunkNumber = Uint16(data[8:10]) / 0x10
    chunk.ChunkCount = data[10]

    fmt.Printf("CHUNK %+v\n", chunk)
    chunk.Data = data[12:]

    return chunk
}

type File struct{
    Size        uint16
    ServiceId   uint16
    ServiceVer  uint8
    ServiceURL  string
    DataStart   uint16
    Data        []byte
    Checksum    uint32
}

func ParseFile(data []byte) *File{
    file := &File{}
    file.Size = Uint16(data[1:3])
    if file.Size == 0{ return nil }
    file.ServiceId = Uint16(data[3:5])
    file.ServiceVer = data[5]
    file.DataStart = uint16(data[6] + 7)
    file.ServiceURL = string(data[7:file.DataStart])
    file.Checksum = Uint32(data[len(data)-4:])
    file.Data = data[file.DataStart:len(data)-4]

    return file
}

type Program struct{
    Start       time.Time
    Duration    time.Duration
    TitleEnd    uint8
    Pid         uint32
    Genre       uint8
    Age         uint8
    Title       string
    IsSerie     bool

    Serie
}

func ParsePrograms(data []byte) []*Program{
    progs := make([]*Program, 0)
    off := 0
    //log.Println(data)
    for off + 32 < len(data){
        //log.Println(data[off:off+24])
        prog := &Program{}
        prog.Pid = Uint32(data[off:off+4])
        prog.Start = time.Unix(int64(Uint32(data[off+4:off+8])), 0)
        prog.Duration = time.Duration(Uint16(data[off+8:off+10])) * 1000000000
        prog.Genre = data[off+20]
        prog.Age = data[off+24]
        prog.TitleEnd = data[off+31] + 32
        prog.Title = decodetitle(data[off+32:off+int(prog.TitleEnd)]) //string(data[32:prog.TitleEnd]) 

        off += int(prog.TitleEnd)

        prog.IsSerie = data[off] == 0xf1

        if prog.IsSerie{
            soff := 0
            prog.Serie, soff = ParseSerie(data[off+1:], off+1)

            off += soff
        }
        off += int(data[off+3]) + 4
        //fmt.Printf("PROG % 5d/% 05d %+v\n", off, len(data), prog)
        progs = append(progs, prog)
    }

    return progs
}

type Serie struct{
    SerieId         uint16
    Episode         uint8
    Year            uint16
    Season          uint8
    Title           string
    TitleEnd        uint8
}

func ParseSerie(data []byte, off int) (Serie, int){
    serie := Serie{}
    //fmt.Println("SERIEDATA", data[:16])
    title_end := int(data[11] + 13)
    serie.SerieId = Uint16(data[4:6])
    serie.Episode = data[7]
    serie.Year = Uint16(data[8:10])
    serie.Season = data[10]
    //serie.Title = decodetitle(data[12:title_end - off])
    return serie, title_end
}

type XMLTVFile struct{
    Channel     []*XMLTVChannel     `xml:"channel"`
    Programme   []*XMLTVProgramme   `xml:"programme"`
}

func NewXMLTVFile() *XMLTVFile{
    xmltvf := &XMLTVFile{}
    xmltvf.Channel = make([]*XMLTVChannel, 0)
    xmltvf.Programme = make([]*XMLTVProgramme, 0)

    return xmltvf
}

type XMLTVChannel struct{
    Id      string  `xml:"id,attr"`
    Name    string  `xml:"display-name"`
}

type XMLTVProgramme struct{
    Start       string  `xml:"start,attr"`
    Stop        string  `xml:"stop,attr"`
    Channel     string  `xml:"channel,attr"`
    Title       string  `xml:"title"`
    SubTitle    string  `xml:"sub-title"`
    Desc        string  `xml:"desc"`
    Date        string  `xml:"date"`
}


func decodetitle(b[]byte) string{
    for i, x := range b{
        b[i] = x ^0x15
    }

    return string(b)
}

func Uint16(b []byte) uint16{
    return binary.BigEndian.Uint16(b)
}

func Uint32(b []byte) uint32{
    return binary.BigEndian.Uint32(b)
}

func main(){

    files := make(map[uint16]*File, 0)
    xmltvf := NewXMLTVFile()

    addr, err := net.ResolveUDPAddr("udp", os.Args[1]); if err != nil {
        log.Fatal(err)
    }
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
