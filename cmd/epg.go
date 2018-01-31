package main

import (
    "encoding/xml"
    "time"
    "os"
    "log"
    "io"
    "sort"
    "io/ioutil"
)

type Export struct{
    Pases       []Pase  `xml:"pase"`
}

type Pase struct{
    Cadena      string  `xml:"cadena,attr"`
    Fecha       string  `xml:"fecha,attr"`
    Hora        string  `xml:"hora"`
    Time        time.Time
    DescCorta   string  `xml:"descripcion_corta"`
    Titulo      string  `xml:"titulo"`
    TipoFicha   string  `xml:"tipo_ficha"`
    SinopCorta  string  `xml:"sinopsis_corta"`
    SinopLarga  string  `xml:"sinopsis_larga"`
    Web         string  `xml:"web"`
}

type XMLTV struct{
    Channel     []*Channel  `xml:"channel"`
    Programme   []*Programme `xml:"programme"`
}

type Channel struct{
    Id      string  `xml:"id,attr"`
    Name    string  `xml:"display-name"`
}

type Programme struct{
    Start       string  `xml:"start,attr"`
    Stop        string  `xml:"stop,attr"`
    Channel     string  `xml:"channel,attr"`
    Title       string  `xml:"title"`
    SubTitle    string  `xml:"sub-title"`
    Desc        string  `xml:"desc"`
    Date        string  `xml:"date"`
}

// http://comunicacion.movistarplus.es/index.php/guiaProgramacion/exportar
func main(){
    f, err := os.Open(os.Args[1]); if err != nil{
        log.Fatal(err)
    }

    fi, _ := f.Stat()

    export := &Export{}
    data := make([]byte, fi.Size())
    io.ReadFull(f, data)
    xml.Unmarshal(data, &export)

    log.Println("Have", len(export.Pases), "pases")

    sort.Slice(export.Pases, func(i, j int) bool{ 
            //return export.Pases[i].Fecha + " " + export.Pases[i].Hora + " " + export.Pases[i].Cadena < export.Pases[j].Fecha + " " + export.Pases[j].Hora + " " + export.Pases[j].Cadena
            return export.Pases[i].Cadena + export.Pases[i].Fecha + export.Pases[i].Hora < export.Pases[j].Cadena + export.Pases[j].Fecha + export.Pases[j].Hora
    })
    //sort.Slice(export.Pases, func(i, j int) bool{ 
    //    return export.Pases[i].Cadena < export.Pases[j].Cadena
    //})

    programme := make([]*Programme, 0)
    channels := make([]*Channel, 0)
    chs_done := make(map[string]bool)
    for i, pase := range export.Pases{
        p := &Programme{}
        t, _ := time.Parse("2006-01-02 15:04:05", pase.Fecha + " " + pase.Hora)
        p.Start = t.Format("20060102150405 -0700")

        //log.Println(i)
        if i + 1 < len(export.Pases) && export.Pases[i].Cadena == export.Pases[i+1].Cadena{
            t2, _ := time.Parse("2006-01-02 15:04:05", export.Pases[i+1].Fecha + " " + export.Pases[i+1].Hora)
            p.Stop = t2.Format("20060102150405 -0700")
        }else{
            p.Stop = t.Add(30 * time.Minute).Format("20060102150405 -0700")
        }

        p.Channel = pase.Cadena
        p.Title = pase.DescCorta
        p.SubTitle = pase.Titulo
        p.Desc = pase.SinopCorta
        p.Date = t.Format("20060102")
        programme = append(programme, p)
        _, ok := chs_done[p.Channel]; if !ok{
            c := &Channel{}
            c.Id = p.Channel
            c.Name = p.Channel
            channels = append(channels, c)
            chs_done[p.Channel] = true
        }
    }

    xmltv := &XMLTV{channels,programme}

    xmltvdata, _ := xml.Marshal(xmltv)
    err = ioutil.WriteFile("movi.xmltv", xmltvdata, 0644); if err != nil{
        log.Fatal(err)
    }

    //log.Println(export)
}
