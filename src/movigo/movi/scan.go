package movi

import (
    "log"
    "os"
    "io"
    "fmt"
    "path/filepath"
    "encoding/xml"
    "time"
    "movigo/dvbstp"
    "movigo/epg"
)

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

    //os.Chtimes(movi.cachedir, time.Now(), time.Now())
}

func (movi *Movi) LoadCaches() bool{

    if err := loadJSON(movi.cachedir + "/spd.json", &movi.spd); err != nil{
        log.Println(err)
        return false
    }
    if err := loadJSON(movi.cachedir + "/sp.json", &movi.sp); err != nil{
        log.Println(err)
        return false
    }
    if err := loadJSON(movi.cachedir + "/bd.json", &movi.bd); err != nil{
        log.Println(err)
        return false
    }
    if err := loadJSON(movi.cachedir + "/pd.json", &movi.pd); err != nil{
        log.Println(err)
        return false
    }
    if err := loadJSON(movi.cachedir + "/bcg.json", &movi.bcg); err != nil{
        log.Println(err)
        return false
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

    log.Println("Caches loaded")


    return true
    //log.Println(movi)
}

func(movi *Movi) Scan(gr func(string) io.Reader, prefix string, scandays int, force bool) bool{

    getreader = gr

    entrypoint := fmt.Sprintf("%s%s", prefix, EntryPointURI)

    if force || movi.sp == nil || movi.spd == nil{
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

    if force || movi.bd == nil || movi.pd == nil || movi.bcg == nil{
        if ok := movi.FindDiscoveryFiles(getreader(nexturi)); !ok{
            log.Println("Some discovery files are missing", nexturi)
            return false
        }
    }else{
        log.Println("Discovery files already cached")
    }

    // EPG
    if force || len(movi.epgfiles) == 0{
        movi.ReadEPG(scandays, prefix)
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

func (movi *Movi) ReadEPG(scandays int, prefix string){
    movi.epgfiles = make(map[uint16]*epg.EPGFile)
    epguris := movi.bcg.GetEPGAddresses()
    log.Println(epguris)
    movi.epgfiles = make(map[uint16]*epg.EPGFile)
    for i, uri := range epguris{
        if scandays > 0 && i > scandays{
            break
        }
        log.Println("URI", uri)
        movi.ReadEPGFile(getreader(prefix + uri))
    }
}

func (movi *Movi) ReadEPGFile(r io.Reader){
    for k, file := range epg.ReadMulticastEPG(r).Files{
        _, exists := movi.epgfiles[k]; if exists{
            for _, program := range file.Programs{
                movi.epgfiles[k].Programs = append(movi.epgfiles[k].Programs, program)
            }
        }else{
            movi.epgfiles[k] = file
        }
    }
}
