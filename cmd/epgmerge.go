package main

import (
    "time"
    "os"
    "log"
    "io/ioutil"
    "github.com/juanmasg/xmltvtool/xmltv"
    "../movistarxml"
)

var chmap map[string]string

// http://comunicacion.movistarplus.es/index.php/guiaProgramacion/exportar
func main(){

    export, err := movistarxml.ReadFile(os.Args[1]); if err != nil{
        log.Fatal(err)
    }

    log.Println("Have", len(export.Pases), "pases")

    programme := make([]*xmltv.Programme, 0)
    channels := make([]*xmltv.Channel, 0)
    chs_done := make(map[string]bool)
    for i, pase := range export.Pases{
        p := &xmltv.Programme{}
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
            c := &xmltv.Channel{}
            c.Id = p.Channel
            c.Name = p.Channel
            channels = append(channels, c)
            chs_done[p.Channel] = true
        }
    }

    xmltvf := xmltv.NewXMLTVFile()
    xmltvf.Channel = channels
    xmltvf.Programme = programme

    xmltvdata, _ := xmltv.Marshal(xmltvf)
    err = ioutil.WriteFile(os.Args[1] + ".xmltv", xmltvdata, 0644); if err != nil{
        log.Fatal(err)
    }

    //log.Println(export)
}
