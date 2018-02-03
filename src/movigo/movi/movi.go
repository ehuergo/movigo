package movi

import (
    "os"
    "io"
    "log"
    "fmt"
    "sort"
    "time"
    "io/ioutil"
    "path/filepath"
    "encoding/json"
    "encoding/xml"
    "movigo/dvbstp"
    "movigo/epg"
)

const (
    EntryPointURI = "239.0.2.129:3937"
)

type Movi struct{
    area        Area
    DomainName  string
    spd         *dvbstp.ServiceProviderDiscovery
    sp          *dvbstp.ServiceProvider
    bd          *dvbstp.BroadcastDiscovery
    pd          *dvbstp.PackageDiscovery
    bcg         *dvbstp.BCGDiscovery
    epgfiles    map[uint16]*epg.EPGFile
    cachedir    string
    epgcachedir string
}

func NewMovi(area Area, cachedays int) *Movi{
    movi := &Movi{}

    movi.area = area
    movi.DomainName = fmt.Sprintf("DEM_%d.imagenio.es", area)

    movi.cachedir = "/tmp/movicache"
    movi.epgcachedir = "/tmp/movicache/epg"

    cachehours, _ := time.ParseDuration(fmt.Sprintf("%dh", cachedays * 24))
    expired := movi.getCacheAge().Add(cachehours).Before(time.Now())
    if !expired{
        movi.LoadCaches()
    }

    return movi
}

func (movi *Movi) getCacheAge() time.Time{
    d, err := os.Open(movi.cachedir); if err != nil{
        if os.IsNotExist(err){
            os.MkdirAll(movi.epgcachedir, 0755)
            log.Println("Created cache directories", movi.epgcachedir)
            return time.Now().Add(-1 * 30  * 24 * time.Hour) //Old enough to force cache save
        }
        log.Println(err)
    }

    di, err := d.Stat(); if err != nil{
        log.Println(err)
    }

    return di.ModTime()
}

func (movi *Movi) SaveCaches(){

    saveJSON(movi.cachedir + "/spd.json", movi.spd)
    saveJSON(movi.cachedir + "/sp.json", movi.sp)
    saveJSON(movi.cachedir + "/bd.json", movi.bd)
    saveJSON(movi.cachedir + "/pd.json", movi.pd)
    saveJSON(movi.cachedir + "/bcg.json", movi.bcg)
    saveJSON(movi.cachedir + "/bd.json", movi.bd)

    for _, file := range movi.epgfiles{
        saveJSON(movi.epgcachedir + "/" + file.File.ServiceURL + ".json", file)
    }
}

func (movi *Movi) LoadCaches(){

    if err := loadJSON(movi.cachedir + "/spd.json", &movi.spd); err != nil{
        log.Println(err)
    }
    if err := loadJSON(movi.cachedir + "/sp.json", &movi.sp); err != nil{
        log.Println(err)
    }
    if err := loadJSON(movi.cachedir + "/bd.json", &movi.bd); err != nil{
        log.Println(err)
    }
    if err := loadJSON(movi.cachedir + "/pd.json", &movi.pd); err != nil{
        log.Println(err)
    }
    if err := loadJSON(movi.cachedir + "/bcg.json", &movi.bcg); err != nil{
        log.Println(err)
    }

    movi.epgfiles = make(map[uint16]*epg.EPGFile)

    files, _ := filepath.Glob(movi.epgcachedir + "/*.json")
    for _, name := range files{
        epgf := &epg.EPGFile{}
        if err := loadJSON(name, epgf); err != nil{
            log.Println(err)
            continue
        }
        movi.epgfiles[epgf.File.ServiceId] = epgf
    }

    //log.Println(movi)
}

func(movi *Movi) Scan(getreader func(string) io.Reader, prefix string, scandays int) bool{

    entrypoint := fmt.Sprintf("%s%s", prefix, EntryPointURI)

    if movi.sp == nil || movi.spd == nil{
        movi.FindAreaServiceProvider(getreader(entrypoint)); if movi.sp == nil{
            log.Println("No service provider found for ", movi.DomainName)
            return false
        }
    }else{
        log.Println("Service provider files already cached")
    }

    log.Printf("Found %d SP offerings for area %s: %s\n", len(movi.sp.Offering), movi.area.String(), movi.sp.Offering)

    offering := movi.sp.Offering[0]
    nexturi := fmt.Sprintf("%s%s", prefix, offering)

    if movi.bd == nil || movi.pd == nil || movi.bcg == nil{
        if ok := movi.FindDiscoveryFiles(getreader(nexturi)); !ok{
            log.Println("Some discovery files are missing", nexturi)
            return false
        }
    }else{
        log.Println("Discovery files already cached")
    }

    // EPG
    if len(movi.epgfiles) == 0{
        epguris := movi.bcg.GetEPGAddresses()
        log.Println(epguris)
        movi.epgfiles = make(map[uint16]*epg.EPGFile)
        for i, uri := range epguris{
            if i > scandays{
                break
            }
            log.Println("URI", uri)
            for k, file := range epg.ReadMulticastEPG(getreader(prefix + uri)).Files{
                movi.epgfiles[k] = file
            }
            //files := dvbstp.ReadSDSFiles(getreader(prefix + uri), 1)
            //log.Println(files)
        }
    }else{
        log.Println("EPG files already cached")
    }

    movi.SaveCaches()
    return true
}

func (movi *Movi) FindAreaServiceProvider(r io.Reader){
    files := dvbstp.ReadSDSFiles(r, 1)

    spd_raw := files[0]

    //log.Println(string(spd_raw))

    sd := &dvbstp.ServiceDiscovery{}
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

func (movi *Movi) FindDiscoveryFiles(r io.Reader) bool{
    files := dvbstp.ReadSDSFiles(r, 3)

    for _, file := range files{
        //log.Println(string(file))
        disco := &dvbstp.ServiceDiscovery{}
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
        }
    }

    if movi.bd == nil || movi.pd == nil || movi.bcg == nil{
        return false
    }

    return true
}


func (movi *Movi) GetPackages() map[string][]*LogicalChannel{
    packages := make(map[string][]*LogicalChannel)
    for _, x := range movi.pd.PackageList{
        log.Println("\n->", x.PackageName, len(x.Services))
        packages[x.PackageName] = make([]*LogicalChannel, 0)
        for _, service := range x.Services{
            si := movi.bd.GetServiceByTextualID(service.TextualID.ServiceName); if si == nil{
                log.Println("Service spec not found for TextualID", service.TextualID.ServiceName)
            }else{
                channel := NewLogicalChannel(x.PackageName, service, si, movi.epgfiles[uint16(service.TextualID.ServiceName)])
                packages[x.PackageName] = append(packages[x.PackageName], channel)
            }
        }
    }
    return packages
}

func (movi *Movi) GetChannelGroups(packages map[string]string) map[int]*ChannelGroup{

    groups := make(map[int]*ChannelGroup)

    channels := movi.GetChannelList(packages, false, 1000)

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

// Get all available channels in the specified packages, or from any package if nil
// If unique == true then try to skip duplicate channels
// If SDoffset > 0 then allow duplicate HD and SD channels, but move the SDs to Number+SDoffset 
func (movi *Movi) GetChannelList(packages map[string]string, unique bool, SDoffset int) []*LogicalChannel{

    channelmap := make(map[int]*LogicalChannel)

    for _, pkg := range movi.pd.PackageList{

        name := pkg.PackageName
        if packages != nil{
            ok := false
            name, ok = packages[pkg.PackageName]; if !ok{
                continue
            }
        }

        for _, service := range pkg.Services{

            si := movi.bd.GetServiceByTextualID(service.TextualID.ServiceName); if si == nil{
                log.Println("No channel found for service", service)
                continue
            }

            channel := NewLogicalChannel(name, service, si, movi.epgfiles[uint16(service.TextualID.ServiceName)])

            if unique{
                c, found := channelmap[channel.Number]; if found{

                    if c.HD{
                        continue
                    }else{
                        if channel.HD{
                            if SDoffset > 0{
                                channelmap[channel.Number + SDoffset] = channelmap[channel.Number]
                                channelmap[channel.Number + SDoffset].Number += SDoffset
                            }
                            channelmap[channel.Number] = channel
                        }
                    }
                }
            }

            channelmap[channel.Number] = channel
        }
    }
    channels := make([]*LogicalChannel, 0)
    for num, _ := range channelmap{
        channels = append(channels, channelmap[num])
    }

    sort.Slice(channels, func(i, j int) bool { return channels[i].Number < channels[j].Number })

    return channels
}


func saveJSON(path string, i interface{}) error{
    j, err := json.Marshal(i); if err != nil{
        return err
    }
    ioutil.WriteFile(path, j, 0644)
    return nil
}

func loadJSON(path string, i interface{}) error{
    j, err := ioutil.ReadFile(path); if err != nil{
        return err
    }
    return json.Unmarshal(j, i)
}

