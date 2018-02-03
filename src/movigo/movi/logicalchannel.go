package movi

import (
    "log"
    "fmt"
    "strings"
    "movigo/epg"
    "movigo/dvbstp"
)

type ChannelGroup struct{
    Number          int
    HD              []*LogicalChannel
    SD              []*LogicalChannel
}

type LogicalChannel struct{
    Name            string
    Number          int
    Id              int
    HD              bool
    FromPackage     string
    Url             StreamURL
    Address         string
    Port            int
    Description     string
    Genre           string
    EPG             *epg.EPGFile
}

func NewLogicalChannel(packagename string, pkgservice *dvbstp.Service, service *dvbstp.SingleService, epgfile *epg.EPGFile) *LogicalChannel{

    var url StreamURL

    if len(service.ServiceLocation) > 0{
        url = &MulticastStreamURL{
            Address:        service.ServiceLocation[0].Address,
            Port:           service.ServiceLocation[0].Port,
        }
    }else if len(service.ServiceLocationHTTP) > 0{
        url = &HTTPStreamURL{
            Url:            service.ServiceLocationHTTP[0],
        }
    }

    if url == nil{
        log.Println("No service location for", service.SI.Name, pkgservice.LogicalChannelNumber)
    }

    return &LogicalChannel{
        Name:           strings.Replace(service.SI.Name, "\n", "", -1),
        Number:         pkgservice.LogicalChannelNumber,
        Id:             pkgservice.TextualID.ServiceName,
        HD:             strings.Contains(service.SI.Name, "HD"), //service.SI.Name[len(service.SI.Name)-2:] == "HD",
        FromPackage:    packagename,
        Url:            url,
        Description:    service.SI.Description,
        Genre:          service.SI.Genre,
        EPG:            epgfile,
    }
}

func (c *LogicalChannel) GetLogoPath() (path string){
    path = strings.Replace(c.Name, " ", "", -1)
    path = strings.ToLower(path)
    return
}

/* StreamURL */

type StreamURL interface{
    AsRTP() string
    AsUDP() string
    AsUDPXY(xyaddr string, xyport int) string
    Raw() string
}

type HTTPStreamURL struct{
    Url     string
}
func (url *HTTPStreamURL) Raw() string{
    return url.Url
}
func (url *HTTPStreamURL) AsRTP() string{
    return ""
}
func (url *HTTPStreamURL) AsUDP() string{
    return ""
}
func (url *HTTPStreamURL) AsUDPXY(xyaddr string, xyport int) string{
    return url.Url
}

type MulticastStreamURL struct{
    Address     string
    Port        int
}

func (url *MulticastStreamURL) Raw() string{
    return fmt.Sprintf("%s:%d", url.Address, url.Port)
}

func (url *MulticastStreamURL) AsRTP() string{
    return fmt.Sprintf("rtp://%s:%d", url.Address, url.Port)
}
func (url *MulticastStreamURL) AsUDP() string{
    return fmt.Sprintf("udp://%s:%d", url.Address, url.Port)
}
func (url *MulticastStreamURL) AsUDPXY(xyaddr string, xyport int) string{
    return fmt.Sprintf("http://%s:%d/rtp/%s:%d", xyaddr, xyport, url.Address, url.Port)
}



