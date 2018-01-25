package main

import (
    "log"
    "fmt"
    "strings"
)

/* SPD SDS */

type ServiceProviderOffering struct{
    Address         string      `xml:"Address,attr"`
    Port            int         `xml:"Port,attr"`
}

func (spo ServiceProviderOffering) String() string{
    return fmt.Sprintf("%s:%d", spo.Address, spo.Port)
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
    ServiceLocationHTTP []string                `xml:"ServiceLocation>HTTPURL"`
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
    BCGDiscovery        BCGDiscovery
}

/* BCGDiscovery */

type BCGDiscovery struct{
    DomainName          string              `xml:"DomainName,attr"`
    Version             int                 `xml:"Version,attr"`
    BCGList             []*BCG              `xml:"BCG"`
}

type BCG struct{
    Id                  string              `xml:"Id,attr"`
    Name                string              `xml:"Name"`
    TransportMode       TransportMode       `xml:"TransportMode"`
}

type TransportMode struct{
    DVBSTP              []DVBSTPTransport   `xml:"DVBSTP"`
    DVBBINSTP           []DVBBINSTP         `xml:"DVBBINSTP"`
}

type DVBSTPTransport struct{
    Address             string              `xml:"Address,attr"`
    Port                int                 `xml:"Port,attr"`
}

type DVBBINSTP struct{
    DVBSTPTransport
    Source              string              `xml:"Source,attr"`
    PayloadId           PayloadId           `xml:"PayloadId"`
}

type PayloadId struct{
    Id                  string              `xml:"Id,attr"`
    Segments            []PayloadSegment    `xml:"Segment"`
}

type PayloadSegment struct{
    Id                  string              `xml:"ID,attr"`
    Version             int                 `xml:"Version,attr"`
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
        HD:             service.SI.Name[len(service.SI.Name)-2:] == "HD",
        FromPackage:    packagename,
        Url:            url,
        Description:    service.SI.Description,
        Genre:          service.SI.Genre,
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



