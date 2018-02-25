package dvbstp

import (
    "fmt"
    "time"
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

/* TV-Anytime MiView TV Present / Following  */

type TVAMain struct{
    ProgramDescription      ProgramDescription  `xml:"ProgramDescription"`
}

type ProgramDescription struct{
    ProgramLocationTable    ProgramLocationTable    `xml:"ProgramLocationTable"`
}

type ProgramLocationTable struct{
    Schedule        Schedule        `xml:"Schedule"`
}

type Schedule struct{
    ServiceIDRef    string          `xml:"serviceIDRef,attr"`
    Version         int             `xml:"Version"`
    ScheduleEventList   []ScheduleEvent `xml:"ScheduleEvent"`
}

type ScheduleEvent struct{
    Program         Program         `xml:"Program"`
}

type Program struct{
    Crid            string `xml:"crid,attr"`
    ProgramInfo     int    `xml:"ProgramInfo,attr"`
    Title           string `xml:"InstanceDescription>Title"`
    Genre           string `xml:"InstanceDescription>Genre>Name"`
    ParentalRating  string `xml:"InstanceDescription>ParentalGuidance>ParentalRating>Name"` 
    PublishedStartTime time.Time `xml:"PublishedStartTime"`
}

/* BCGDiscovery */

type BCGDiscovery struct{
    DomainName          string              `xml:"DomainName,attr"`
    Version             int                 `xml:"Version,attr"`
    BCGList             []BCG               `xml:"BCG"`
}

func (d *BCGDiscovery) GetNowNext() *BCG{
    for _, bcg := range d.BCGList{
        if bcg.Id == "p_f"{
            return &bcg
        }
    }

    return nil
}

func (d *BCGDiscovery) getEPG() *BCG{
    for _, bcg := range d.BCGList{
        if bcg.Id == "EPG"{
            return &bcg
        }
    }
    return nil
}

func (d *BCGDiscovery) GetNowNextAddress() string{
    nn := d.GetNowNext()
    if len(nn.TransportMode.DVBSTP) > 0{
        return fmt.Sprintf("%s:%d", nn.TransportMode.DVBSTP[0].Address, nn.TransportMode.DVBSTP[0].Port)
    }

    return ""
}

func (d *BCGDiscovery) GetEPGAddresses() []string{
    addresses := make([]string, 0)
    epg := d.getEPG()
    if len(epg.TransportMode.DVBBINSTP) > 0{
        for _, dvbbinstp := range epg.TransportMode.DVBBINSTP{
            addresses = append(addresses, fmt.Sprintf("%s:%d", dvbbinstp.Address, dvbbinstp.Port))
        }
    }

    return addresses
}

func (d *BCGDiscovery) GetList(){
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
    HD                  int
    SD                  int
}

