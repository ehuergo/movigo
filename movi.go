package main

import (
    "encoding/xml"
    "log"
    "fmt"
    "sort"
    "dvbstp"
    "io"
    "io/ioutil"
    "epg"
    "encoding/json"
    "path/filepath"
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
    epgfiles    map[uint16]*epg.EPGFile
    save        bool
}

func NewMovi(area Area) *Movi{
    movi := &Movi{}

    movi.area = area
    movi.DomainName = fmt.Sprintf("DEM_%d.imagenio.es", area)

    movi.save = true

    movi.LoadCaches()

    return movi
}

func (movi *Movi) SaveCaches(){

    if !movi.save{ return }

    jspd, err := json.Marshal(movi.spd); if err == nil{
        ioutil.WriteFile("cache/spd.json", jspd, 0644)
    }

    jsp, err := json.Marshal(movi.sp); if err == nil{
        ioutil.WriteFile("cache/sp.json", jsp, 0644)
    }

    jbd, err := json.Marshal(movi.bd); if err == nil{
        ioutil.WriteFile("cache/bd.json", jbd, 0644)
    }

    jpd, err := json.Marshal(movi.pd); if err == nil{
        ioutil.WriteFile("cache/pd.json", jpd, 0644)
    }

    jbcg, err := json.Marshal(movi.bcg); if err == nil{
        ioutil.WriteFile("cache/bcg.json", jbcg, 0644)
    }

    for _, file := range movi.epgfiles{
        j, err := json.Marshal(file); if err == nil{
            ioutil.WriteFile("cache/epg/" + file.File.ServiceURL + ".json", j, 0644)
        }
    }
}

func (movi *Movi) LoadCaches() bool{
    jspd, err := ioutil.ReadFile("cache/spd.json"); if err == nil{
        err = json.Unmarshal(jspd, &movi.spd); if err != nil{
            log.Println(err)
        }
    }
    jsp, err := ioutil.ReadFile("cache/sp.json"); if err == nil{
        json.Unmarshal(jsp, &movi.sp)
    }

    jbd, err := ioutil.ReadFile("cache/bd.json"); if err == nil{
        json.Unmarshal(jbd, &movi.bd)
    }

    jpd, err := ioutil.ReadFile("cache/pd.json"); if err == nil{
        json.Unmarshal(jpd, &movi.pd)
    }

    jbcg, err := ioutil.ReadFile("cache/bcg.json"); if err == nil{
        json.Unmarshal(jbcg, &movi.bcg)
    }

    movi.epgfiles = make(map[uint16]*epg.EPGFile)
    files, err := filepath.Glob("cache/epg/*.json")
    for _, name := range files{
        j, err := ioutil.ReadFile(name); if err == nil{
            epgf := &epg.EPGFile{}
            //log.Println(string(j))
            json.Unmarshal(j, epgf)
            //log.Println(epgf.File.ServiceId, epgf.Programs[0])
            movi.epgfiles[epgf.File.ServiceId] = epgf
        }
    }

    if movi.spd != nil && movi.sp != nil && movi.sp != nil && movi.pd != nil && movi.bcg != nil && len(movi.epgfiles) > 0{
        movi.save = false
    }

    //log.Println(movi)

    return true
}

func(movi *Movi) Scan(getreader func(string) io.Reader, prefix string) bool{

    entrypoint := fmt.Sprintf("%s%s", prefix, EntryPointURI)

    if movi.sp == nil || movi.spd == nil{
        movi.FindAreaServiceProvider(getreader(entrypoint)); if movi.sp == nil{
            log.Fatal("No service provider found for ", movi.DomainName)
        }
    }else{
        log.Println("Service provider files already cached")
    }

    log.Printf("Found %d SP offerings for area %s: %s\n", len(movi.sp.Offering), movi.area.String(), movi.sp.Offering)

    offering := movi.sp.Offering[0]
    nexturi := fmt.Sprintf("%s%s", prefix, offering)

    if movi.bd == nil || movi.pd == nil || movi.bcg == nil{
        if ok := movi.FindDiscoveryFiles(getreader(nexturi)); !ok{
            log.Fatal("Some discovery files are missing", nexturi)
        }
    }else{
        log.Println("Discovery files already cached")
    }

    // TV-Anytime
    // Now-Next
    //nownexturi := movi.bcg.GetNowNextAddress()
    //files := dvbstp.ReadBIMFiles(prefix + nownexturi, 10)
    //files := dvbstp.ReadSDSFiles(prefix + 
    //log.Println(files)
    // EPG
    if len(movi.epgfiles) == 0{
        epguris := movi.bcg.GetEPGAddresses()
        log.Println(epguris)
        movi.epgfiles = make(map[uint16]*epg.EPGFile)
        for i, uri := range epguris{
            if i > 2{
                break
            }
            log.Println("URI", uri)
            for k, file := range epg.ReadEPG(getreader(prefix + uri)).Files{
                movi.epgfiles[k] = file
            }
            //files := dvbstp.ReadSDSFiles(getreader(prefix + uri), 1)
            //log.Println(files)
        }
    }else{
        log.Println("EPG files already cached")
    }

    //log.Println(movi.epgfiles)

    movi.SaveCaches()

    //log.Printf("%+v\n",movi)
    return true
}

func (movi *Movi) FindDiscoveryFiles(r io.Reader) bool{
    files := dvbstp.ReadSDSFiles(r, 3)

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
            //log.Printf("%+v\n", movi.bcg.BCGList[0])
            //log.Printf("%+v\n", movi.bcg.BCGList[1])
        }
    }

    if movi.bd == nil || movi.pd == nil || movi.bcg == nil{
        return false
    }

    return true
}

func (movi *Movi) FindAreaServiceProvider(r io.Reader){
    files := dvbstp.ReadSDSFiles(r, 1)

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

func (movi *Movi) GetUniqueChannels() []*LogicalChannel{

    channels := make([]*LogicalChannel, 0)

    groups := movi.GetChannelGroups(nil)
    for _, group := range groups{
        if len(group.HD) > 0{
            if len(group.SD) > 0{
                group.SD[0].Number += 1000
                channels = append(channels, group.SD[0])
                if group.HD[0].EPG == nil && group.SD[0].EPG != nil{
                    group.HD[0].EPG = group.SD[0].EPG
                }
            }
            channels = append(channels, group.HD[0])
        }else if len(group.SD) > 0{
                channels = append(channels, group.SD[0])
        }
    }

    return channels
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
            channel := NewLogicalChannel(friendlyname, service, si, movi.epgfiles[uint16(service.TextualID.ServiceName)])
            //log.Println(channel)
            channels = append(channels, channel)
        }
    }

    sort.Slice(channels, func(i, j int) bool { return channels[i].Number < channels[j].Number })

    return channels
}
