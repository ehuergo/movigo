package movi

import (
    "io"
    "encoding/json"
    "log"
    "fmt"
    "sort"
    "time"
    "io/ioutil"
    "../dvbstp"
    "../epg"
    //"golang.org/x/text/search"
    //"golang.org/x/text/language"
)

const (
    EntryPointURI = "239.0.2.129:3937"
)

var (
    getreader   func(string) io.Reader
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
    channels    []*LogicalChannel
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
                //if channel.HD{
                //    packages[x.PackageName].HD++
                //}else{
                //    packages[x.PackageName].SD++
                //}else{
                //}
            }
        }
    }
    return packages
}

func (movi *Movi) GetChannelGroups(packages map[string]string, ignore map[string]bool) map[int]*ChannelGroup{

    groups := make(map[int]*ChannelGroup)

    channels := movi.GetChannelList(packages, ignore, false, 1000)

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
func (movi *Movi) GetChannelList2(SDoffset int) []*LogicalChannel{
    servicenums := make(map[int]map[int]int)
    channels := make([]*LogicalChannel, 0)

        for _, pkg := range movi.pd.PackageList{
            for _, service := range pkg.Services{
                servicenums[service.TextualID.ServiceName][service.LogicalChannelNumber]++
            }
        }
    //sort.Slice(servicenums, func(i, j int) bool { return channels[i].Number < channels[j].Number })

    return channels
}

// Get all available channels in the specified packages, or from any package if nil
// If unique == true then try to skip duplicate channels
// If SDoffset > 0 then allow duplicate HD and SD channels, but move the SDs to Number+SDoffset 
func (movi *Movi) GetChannelList(packages map[string]string, ignore map[string]bool, unique bool, SDoffset int) []*LogicalChannel{

    channelmap := make(map[int]*LogicalChannel)

    for _, pkg := range movi.pd.PackageList{

        name := pkg.PackageName
        if packages != nil{
            ok := false
            name, ok = packages[pkg.PackageName]; if !ok{
                continue
            }
        }

        if ignore != nil{
            _, found := ignore[pkg.PackageName]; if found{
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
                copyEPG(c, channel)

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

/*
// EPG
func (movi *Movi) programMatches(searcher *search.Matcher, program *epg.Program, name string, season string, episode string, title string, exact bool) bool{
    if name != ""{
        s := program.Title
        if program.IsSerie && program.ParsedSerie != nil{
            if program.ParsedSerie.ParsedName  != ""{
                s = program.ParsedSerie.ParsedName
            }
        }
        //log.Println(program, program.ParsedSerie)
        start, end := searcher.IndexString(s, name)
        //log.Println("S IS", s, start, end)
        if start == -1{
            return false
        }else if exact{
            if start > 4 || end < len(s) - 2{
                return false
            }
        }
    }
    if program.ParsedSerie == nil{
        return true
    }

    if season != "" && season != program.ParsedSeason{
        return false
    }
    if episode != "" && episode != program.ParsedEpisode{
        return false
    }

    if title != ""{
        start, _ := searcher.IndexString(program.Title, name)
        if start == -1{
            return false
        }
    }

    return true
}

func (movi *Movi) FindCurrent() map[string]*epg.Program{

    programs := make(map[string]*epg.Program, 0)

    channels := movi.GetChannelList(nil, nil, true, 1000)
    for _, channel := range channels{
        programs[channel.Name] = channel.EPGNow()
    }

    return programs
}

func (movi *Movi) FindProgram(name string, season string, episode string, title string, exact bool) map[string][]*epg.Program{

    matches := make(map[string][]*epg.Program, 0)

    if name + season + episode + title == ""{
        return matches
    }

    searcher := search.New(language.Spanish, search.IgnoreCase, search.IgnoreDiacritics) //, search.WholeWord)

    channels := movi.GetChannelList(nil, nil, true, 1000)
    for _, channel := range channels{
        if channel.EPG == nil{
            continue
        }
        //log.Println(channel.Name)

        for _, program := range channel.EPG.Programs{
            if movi.programMatches(searcher, program, name, season, episode, title, exact){
                _, ok := matches[channel.Name]; if !ok{
                    matches[channel.Name] = make([]*epg.Program, 0)
                }
                matches[channel.Name] = append(matches[channel.Name], program)
            }
        }
    }

    return matches
}
*/

func (movi *Movi) GetChannelByUri(addrinfo string) *LogicalChannel{
    channels := movi.GetChannelList(nil, nil, true, 1000)
    for _, channel := range channels{
        if addrinfo == channel.Url.Raw(){
            return channel
        }
    }

    return nil
}

func (movi *Movi) GetChannelByName(name string) *LogicalChannel{
    channels := movi.GetChannelList(nil, nil, true, 1000)
    for _, channel := range channels{
        if name == channel.Name{
            return channel
        }
    }

    return nil
}

func copyEPG(c1, c2 *LogicalChannel){
    if c1.EPG == nil && c2.EPG != nil{
        c1.EPG = c2.EPG
    }else if c1.EPG != nil && c2.EPG == nil{
        c2.EPG = c1.EPG
    }
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

