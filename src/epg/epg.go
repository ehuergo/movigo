package epg

// All credits to: https://github.com/MovistarTV/tv_grab_es_movistartv/blob/master/tv_grab_es_movistartv.py#L610

import (
    "fmt"
    "log"
    "io"
)

type EPG struct{
    Files   map[uint16]*EPGFile
}

type EPGFile struct{
    File        *File
    Programs    []*Program
}

func newEPG() *EPG{
    epg := &EPG{}
    epg.Files = make(map[uint16]*EPGFile)
    return epg
}

func ReadEPG(r io.Reader) *EPG{
    epg := newEPG()

    for{
        data := make([]byte, 1500)
        n, _ := r.Read(data)
        fmt.Printf("% 5d ", n)
        chunk := ParseChunk(data[:n])
        log.Println("SKIP!")

        if chunk.End{ break }
    }

    filedata := make([]byte, 0)
    for len(epg.Files) < 500{ // Force finish
        data := make([]byte, 1500)
        n, _ := r.Read(data)
        fmt.Printf("% 5d ", n)
        chunk := ParseChunk(data[:n])
        filedata = append(filedata, chunk.Data...)
        if chunk.End{
            file := ParseFile(filedata)
            if file != nil && file.Size > 0 && len(file.Data) != 0{
                //fmt.Printf("FILE! %s", file.ServiceURL)
                _, ok := epg.Files[file.ServiceId]; if ok{
                    break
                }
                //epg.Files[file.ServiceId] = file

                //xmltvc := &XMLTVChannel{}
                //xmltvc.Id = fmt.Sprintf("%d", file.ServiceId)
                //xmltvc.Name = file.ServiceURL //FIXME
                //xmltvf.Channel = append(xmltvf.Channel, xmltvc)

                programs := ParsePrograms(file.Data)
                //for _, program := range programs{
                //    //log.Printf("PROGRAM %+v\n", program)
                //    xmltvp := &XMLTVProgramme{}
                //    xmltvp.Start = program.Start.Format("20060102150405 -0700")
                //    xmltvp.Stop = program.Start.Add(program.Duration).Format("20060102150405 -0700")
                //    xmltvp.Channel = fmt.Sprintf("%d", file.ServiceId)
                //    xmltvp.Title = program.Title
                //    xmltvp.Date = program.Start.Format("20060102")
                //    xmltvf.Programme = append(xmltvf.Programme, xmltvp)
                //}
                epg.Files[file.ServiceId] = &EPGFile{file,programs}
            }
            filedata = make([]byte, 0)
        }
    }
    return epg
}

/*

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
*/
