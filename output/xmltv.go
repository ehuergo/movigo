package output

import (
    "../movi"
    "fmt"
    "encoding/xml"
)

type XMLTVFile struct{
    XMLName     xml.Name
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

func DumpXMLTVEPG(channels []*movi.LogicalChannel) []byte{
    xmltvf := NewXMLTVFile()
    //xmltvf.XMLName.Local = "tv"
    for _, channel := range channels{
        xmltvc := &XMLTVChannel{}
        xmltvc.Id = fmt.Sprintf("%d", channel.Id)
        xmltvc.Name = channel.Name
        xmltvf.Channel = append(xmltvf.Channel, xmltvc)

        if channel.EPG == nil || len(channel.EPG.Programs) == 0{ continue }

        for _, program := range channel.EPG.Programs{
             xmltvp := &XMLTVProgramme{}
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

