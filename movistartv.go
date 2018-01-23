package main

import (
    "encoding/xml"
    "log"
    "fmt"
    "sort"
)

type Area uint8

const (
    CATALUNYA               Area = 1
    CASTILLA_Y_LEON         Area = 4
    COMUNIDAD_VALENCIANA    Area = 6
    BALEARES                Area = 10
    MURCIA                  Area = 12
    ASTURIAS                Area = 13
    ANDALUCIA               Area = 15
    MADRID                  Area = 19
    GALICIA                 Area = 24
    CANTABRIA               Area = 29
    LA_RIOJA                Area = 31
    EXTREMADURA             Area = 32
    ARAGON                  Area = 34
    NAVARRA                 Area = 35
    PAIS_VASCO              Area = 36
    CANARIAS                Area = 37
    CASTILLA_LA_MANCHA      Area = 38
)

func (l *Area) String() string{
    switch *l{
        case CATALUNYA:
            return "Catalunya"
        case CASTILLA_Y_LEON:
            return "Castilla y Leon"
        case COMUNIDAD_VALENCIANA:
            return "Comunidad Valenciana"
        case BALEARES:
            return "Baleares"
        case MURCIA:
            return "Murcia"
        case ASTURIAS:
            return "Asturias"
        case ANDALUCIA:
            return "Andalucia"
        case MADRID:
            return "Madrid"
        case GALICIA:
            return "Galicia"
        case CANTABRIA:
            return "Cantabria"
        case LA_RIOJA:
            return "La Rioja"
        case EXTREMADURA:
            return "Extremadura"
        case ARAGON:
            return "Aragon"
        case NAVARRA:
            return "Navarra"
        case PAIS_VASCO:
            return "Pais Vasco"
        case CANARIAS:
            return "Canarias"
        case CASTILLA_LA_MANCHA:
            return "Castilla la Mancha"
    }

    return "Unknown"
}

type MovistarScanner struct{
    area        Area
    DomainName  string
    spd         *ServiceProviderDiscovery
    sp          *ServiceProvider
    bd          *BroadcastDiscovery
    pd          *PackageDiscovery
}

func NewMovistarScanner(area Area) *MovistarScanner{
    s := &MovistarScanner{}

    s.area = area
    s.DomainName = fmt.Sprintf("DEM_%d.imagenio.es", area)

    return s
}

func (s *MovistarScanner) Scan(path string) bool{
    s.ScanServiceProvider(path); if s.sp == nil{
        log.Fatal("No service provider found for ", s.DomainName)
    }

    r := NewDVBSTPReader("samples/all2.raw")
    files := r.ReadFiles(3)

    for _, file := range files{
        //log.Println(string(file))
        disco := &ServiceDiscovery{}
        xml.Unmarshal(file, disco)
        //log.Printf("%+v", disco)
        if disco.BroadcastDiscovery.Version != 0{
            s.bd = &disco.BroadcastDiscovery
            log.Println("Found BroadcastDiscovery with", len(disco.BroadcastDiscovery.ServiceList), "services")
        }else if disco.PackageDiscovery.Version != 0{
            s.pd = &disco.PackageDiscovery
            log.Println("Found PackageDiscovery with", len(disco.PackageDiscovery.PackageList), "packages")
        }
    }

    if s.bd == nil && s.pd == nil{
        return false
    }

    log.Printf("%+v\n",s)
    return true
}

func (s *MovistarScanner) ScanServiceProvider(path string){
    r := NewDVBSTPReader(path)
    files := r.ReadFiles(1)

    spd_raw := files[0]

    //log.Println(string(spd_raw))

    sd := &ServiceDiscovery{}
    xml.Unmarshal(spd_raw, sd)
    s.spd = &sd.ServiceProviderDiscovery

    //log.Printf("%+v\n",s.spd)

    for _, provider := range s.spd.ServiceProviders{
        //log.Printf("%+v\n", provider)
        if provider.DomainName == s.DomainName{
            s.sp = provider
        }
    }
}

func (s *MovistarScanner) ListPackages(){
    for _, x := range s.pd.PackageList{
        log.Println("\n->", x.PackageName, len(x.Services))
        for _, ser := range x.Services{
            si := s.bd.GetServiceByTextualID(ser.TextualID.ServiceName); if si == nil{
                log.Println("Service spec not found for TextualID", ser.TextualID.ServiceName)
            }else{
                log.Println(ser.LogicalChannelNumber, ser.TextualID.ServiceName, si.SI.Name)
            }
        }
    }
}

func (scanner *MovistarScanner) GetChannelGroups(packages map[string]string) map[int]*ChannelGroup{

    groups := make(map[int]*ChannelGroup)

    channels := scanner.GetChannelList(packages)

    for _, channel := range channels{
        group, ok := groups[channel.Number]; if !ok{
            group = &ChannelGroup{
                Number:     channel.Number,
                SD:         make([]*LogicalChannel, 0),
                HD:         make([]*LogicalChannel, 0),
            }
            groups[channel.Number] = group
        }
        if channel.HD{
            group.HD = append(group.HD, channel)
        }else{
            group.SD = append(group.SD, channel)
        }
    }

    return groups

}

func (scanner *MovistarScanner) GetChannelList(packages map[string]string) []*LogicalChannel{

    channels := make([]*LogicalChannel, 0)

    for _, x := range scanner.pd.PackageList{

        friendlyname := x.PackageName
        var ok bool

        if packages != nil{
            friendlyname, ok = packages[x.PackageName]; if !ok{
                continue
            }
        }

        for _, service := range x.Services{

            si := scanner.bd.GetServiceByTextualID(service.TextualID.ServiceName); if si == nil{
                log.Println("No channel found for service", service)
                continue
            }

            //log.Println(service, si)
            channel := NewLogicalChannel(friendlyname, service, si)
            //log.Println(channel)
            channels = append(channels, channel)
        }
    }

    sort.Slice(channels, func(i, j int) bool { return channels[i].Number < channels[j].Number })

    return channels
}
