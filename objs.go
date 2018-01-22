package main

import (
//    "log"
    "fmt"
    "strings"
)

/* SPD SDS */

type ServiceProviderOffering struct{
    Address         string      `xml:"Address,attr"`
    Port            int         `xml:"Port,attr"`
}

type ServiceProvider struct{
    DomainName      string      `xml:"DomainName,attr"`
    Version         int         `xml:"Version,attr"`
    Offering        []ServiceProviderOffering `xml:"Offering>Push"`
}

type ServiceProviderDiscovery struct{
    ServiceProviders    []*ServiceProvider `xml:"ServiceProvider"`
}


/* Service Discovery */

type IPMulticastAddress struct{
    Port        int             `xml:"Port,attr"`
    Address     string          `xml:"Address,attr"`
    Type        string          `xml:"Type,attr"`
}

type TextualIdentifier struct{
    ServiceName     int         `xml:"ServiceName,attr"`
    LogoURI         string      `xml:"logoURI,attr"`
}

type SIName struct{
    Language        string      `xml:"Language,attr"`
    Name            string      `xml:"Name"`
}

type SI struct{
    ServiceType     int         `xml:"ServiceType"`
    ServiceInfo     int         `xml:"ServiceInfo"`
    Name            string      `xml:"Name"`
    ShortName       string      `xml:"ShortName"`
    Description     string      `xml:"Description"`
    Genre           string      `xml:"Genre>Name"`
}

type SingleService struct{
    ServiceLocation     []IPMulticastAddress    `xml:"ServiceLocation>IPMulticastAddress"`
    TextualIdentifier   TextualIdentifier       `xml:"TextualIdentifier"`
    SI                  SI                      `xml:"SI"`
}

type BroadcastDiscovery struct{
    DomainName          string                  `xml:"DomainName,attr"`
    Version             int                     `xml:"Version,attr"`
    ServiceList         []*SingleService         `xml:"ServiceList>SingleService"`
}

func (bd *BroadcastDiscovery) GetServiceByTextualID(id int) *SingleService{
    for _, s := range bd.ServiceList{
        if s.TextualIdentifier.ServiceName == id{
            return s
        }
    }

    return nil
}

type ServiceDiscovery struct{
    Xmlns               string              `xml:"xmlns,attr"`
    Xmlnsurn            string              `xml:"urn,attr"`
    Xmlnsmpeg7          string              `xml:"mpeg7,attr"`
    BroadcastDiscovery  BroadcastDiscovery
    PackageDiscovery    PackageDiscovery
    ServiceProviderDiscovery ServiceProviderDiscovery
}

/* Package Discovery */

type PackageDiscovery struct{
    DomainName          string              `xml:"DomainName,attr"`
    Version             int                 `xml:"Version,attr"`
    PackageList         []*Package          `xml:"Package"`
}

type TextualID  struct{
    ServiceName         int                 `xml:"ServiceName,attr"`
}

type Service struct{
    TextualID               TextualID       `xml:"TextualID"`
    LogicalChannelNumber    int             `xml:"LogicalChannelNumber"`
}

type Package struct{
    PackageName         string              `xml:"PackageName"`
    Services            []*Service          `xml:"Service"`
}

/* Logical Channel */
type ChannelGroup struct{
    Number          int
    HD              []*LogicalChannel
    SD              []*LogicalChannel
}

type LogicalChannel struct{
    Name            string
    Number          int
    HD              bool
    FromPackage     string
    Url             StreamURL
    Address         string
    Port            int
    Description     string
    Genre           string
}

func NewLogicalChannel(packagename string, pkgservice *Service, service *SingleService) *LogicalChannel{
    return &LogicalChannel{
        Name:           service.SI.Name,
        Number:         pkgservice.LogicalChannelNumber,
        HD:             service.SI.Name[len(service.SI.Name)-2:] == "HD",
        FromPackage:    packagename,
        Url:            StreamURL{
            Address:        service.ServiceLocation[0].Address,
            Port:           service.ServiceLocation[0].Port,
        },
        Description:    service.SI.Description,
        Genre:          service.SI.Genre,
    }
}

func (c *LogicalChannel) GetLogoPath() (path string){
    path = strings.Replace(c.Name, " ", "", -1)
    path = strings.ToLower(path)
    return
}

func (c *LogicalChannel) GetRTPString() (uri string){
    return fmt.Sprintf("rtp://%s:%d", c.Address, c.Port)
}

func (c *LogicalChannel) GetUDPXYString(ip string, port int) (uri string){
    return fmt.Sprintf("http://%s:%d/rtp/%s:%d", ip, port, c.Address, c.Port)
}

/* StreamURL */

type StreamURL struct{
    Address     string
    Port        int
}

func (url *StreamURL) AsRTP() string{
    return fmt.Sprintf("rtp://%s:%d", url.Address, url.Port)
}
func (url *StreamURL) AsUDP() string{
    return fmt.Sprintf("udp://%s:%d", url.Address, url.Port)
}
func (url *StreamURL) AsUDPXY(xyaddr string, xyport int) string{
    return fmt.Sprintf("http://%s:%d/rtp/%s:%d", xyaddr, xyport, url.Address, url.Port)
}

