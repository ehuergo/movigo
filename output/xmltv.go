package output

import (
    "../movi"
    "fmt"
    "encoding/xml"
    "github.com/juanmasg/xmltvtool/xmltv"
)

func DumpXMLTVEPG(channels []*movi.LogicalChannel) []byte{
    xmltvf := xmltv.NewXMLTVFile()
    //xmltvf.XMLName.Local = "tv"
    for _, channel := range channels{
        xmltvc := &xmltv.Channel{}
        xmltvc.Id = fmt.Sprintf("%d", channel.Id)
        xmltvc.Name = channel.Name
        xmltvf.Channel = append(xmltvf.Channel, xmltvc)

        if channel.EPG == nil || len(channel.EPG.Programs) == 0{ continue }

        for _, program := range channel.EPG.Programs{
             xmltvp := &xmltv.Programme{}
             xmltvp.Start = program.Start.Format("20060102150405 -0700")
             xmltvp.Stop = program.Start.Add(program.Duration).Format("20060102150405 -0700")
             xmltvp.Channel = fmt.Sprintf("%d", channel.Id) //fmt.Sprintf("%d", channel.EPG.File.ServiceId)
             xmltvp.Title = program.Title
             xmltvp.Date = program.Start.Format("20060102")
             xmltvf.Programme = append(xmltvf.Programme, xmltvp)
        }
    }
    //xmltvdata, _ := xml.MarshalIndent(xmltvf, "", "  ")
    xmltvdata, _ := xml.Marshal(xmltvf)
    return xmltvdata
}

