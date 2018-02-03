package main

import (
    "fmt"
    "log"
    "sort"
    "epg"
    "encoding/xml"
)

func DumpGroupsAsIPTVSimple(groups map[int]*ChannelGroup, prefix string) []byte{
    var keys []int
    data := []byte("#EXTM3U\n")

    for k := range groups{
        keys = append(keys, k)
    }
    sort.Ints(keys)

    for _, k := range keys{
        group := groups[k]
        if len(group.HD) > 0{
            data = append(data, dumpIPTVSimpleChannel(group.HD[0], prefix)...)
        }else if len(group.SD) > 0{
            data = append(data, dumpIPTVSimpleChannel(group.SD[0], prefix)...)
        }else{
            log.Println("WARNING: No SD or HD channels in group", group)
        }
    }

    return data
}

func dumpIPTVSimpleChannel(c *LogicalChannel, prefix string) []byte{

    extinf := fmt.Sprintf("#EXTINF:-1 tvg-chid=\"%d\" tvg-logo=\"%s\" tvg-chno=\"%d\" group-title=\"%s\", %s\n",
        c.Id,
        c.GetLogoPath(),
        c.Number,
        c.FromPackage,
        c.Name)

    url := fmt.Sprintf("%s%s\n", prefix, c.Url.Raw())

    return append([]byte(extinf), []byte(url)...)
}

func dumpXMLTVEPG(channels []*LogicalChannel) []byte{
    xmltvf := epg.NewXMLTVFile()
    //xmltvf.XMLName.Local = "tv"
    for _, channel := range channels{
        xmltvc := &epg.XMLTVChannel{}
        xmltvc.Id = fmt.Sprintf("%d", channel.Id)
        xmltvc.Name = channel.Name
        xmltvf.Channel = append(xmltvf.Channel, xmltvc)

        if channel.EPG == nil || len(channel.EPG.Programs) == 0{ continue }

        for _, program := range channel.EPG.Programs{
             xmltvp := &epg.XMLTVProgramme{}
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

func DumpIPTVSimple(channels []*LogicalChannel, prefix string) []byte{

    data := []byte("#EXTM3U\n")

    for _, c := range channels{
        data = append(data, dumpIPTVSimpleChannel(c, prefix)...)
    }

    return data
}
