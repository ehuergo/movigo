package epg

import (
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



