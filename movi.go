package main

import (
    "encoding/xml"
    "log"
    "fmt"
    "sort"
)

const (
    EntryPointURI = "239.0.2.129:3937"
)

type Movi struct{
    area        Area
    DomainName  string
    spd         *ServiceProviderDiscovery
    sp          *ServiceProvider
    bd          *BroadcastDiscovery
    pd          *PackageDiscovery
    bcg         *BCGDiscovery
}

func NewMovi(area Area) *Movi{
    movi := &Movi{}

    movi.area = area
    movi.DomainName = fmt.Sprintf("DEM_%d.imagenio.es", area)

    return movi
}

func(movi *Movi) Scan(prefix string) bool{

    entrypoint := fmt.Sprintf("%s%s", prefix, EntryPointURI)

    movi.FindAreaServiceProvider(entrypoint); if movi.sp == nil{
        log.Fatal("No service provider found for ", movi.DomainName)
    }

    log.Printf("Found %d SP offerings for area %s: %s\n", len(movi.sp.Offering), movi.area.String(), movi.sp.Offering)

    offering := movi.sp.Offering[0]
    nexturi := fmt.Sprintf("%s%s", prefix, offering)

    r := NewDVBSTPReader(nexturi)
    files := r.ReadFiles(3)

    for _, file := range files{
        //log.Println(string(file))
        disco := &ServiceDiscovery{}
        err := xml.Unmarshal(file, disco); if err != nil{
            log.Println(err)
        }
        //log.Printf("%+v", disco)
        if disco.BroadcastDiscovery.Version != 0{
            movi.bd = &disco.BroadcastDiscovery
            log.Println("Found BroadcastDiscovery with", len(disco.BroadcastDiscovery.ServiceList), "services")
        }else if disco.PackageDiscovery.Version != 0{
            movi.pd = &disco.PackageDiscovery
            log.Println("Found PackageDiscovery with", len(disco.PackageDiscovery.PackageList), "packages")
        }else if disco.BCGDiscovery.Version != 0{
            movi.bcg = &disco.BCGDiscovery
            log.Println("Found BCGDiscovery with", len(disco.BCGDiscovery.BCGList), "providers")
            log.Printf("%+v\n", movi.bcg.BCGList[0])
            log.Printf("%+v\n", movi.bcg.BCGList[1])
        }
    }

    if movi.bd == nil && movi.pd == nil{
        return false
    }

    log.Printf("%+v\n",movi)
    return true
}

func (movi *Movi) FindAreaServiceProvider(path string){
    r := NewDVBSTPReader(path)
    files := r.ReadFiles(1)

    spd_raw := files[0]

    //log.Println(string(spd_raw))

    sd := &ServiceDiscovery{}
    xml.Unmarshal(spd_raw, sd)
    movi.spd = &sd.ServiceProviderDiscovery

    //log.Printf("%+v\n",movi.spd)

    for _, provider := range movi.spd.ServiceProviders{
        //log.Printf("%+v\n", provider)
        if provider.DomainName == movi.DomainName{
            movi.sp = provider
        }
    }
}

func (movi *Movi) ListPackages(){
    for _, x := range movi.pd.PackageList{
        log.Println("\n->", x.PackageName, len(x.Services))
        for _, ser := range x.Services{
            si := movi.bd.GetServiceByTextualID(ser.TextualID.ServiceName); if si == nil{
                log.Println("Service spec not found for TextualID", ser.TextualID.ServiceName)
            }else{
                log.Println(ser.LogicalChannelNumber, ser.TextualID.ServiceName, si.SI.Name)
            }
        }
    }
}

func (movi *Movi) GetChannelGroups(packages map[string]string) map[int]*ChannelGroup{

    groups := make(map[int]*ChannelGroup)

    channels := movi.GetChannelList(packages)

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

func (movi *Movi) GetChannelList(packages map[string]string) []*LogicalChannel{

    channels := make([]*LogicalChannel, 0)

    for _, x := range movi.pd.PackageList{

        friendlyname := x.PackageName
        var ok bool

        if packages != nil{
            friendlyname, ok = packages[x.PackageName]; if !ok{
                continue
            }
        }

        for _, service := range x.Services{

            si := movi.bd.GetServiceByTextualID(service.TextualID.ServiceName); if si == nil{
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
