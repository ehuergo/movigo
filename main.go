package main

import (
    "log"
//    "encoding/json"
    "os"
)

func main(){
    //r := NewDVBSTPStreamReader("p1.raw")
    //files := r.ReadFiles(1)

    //spd_raw := files[0]

    //log.Println(string(spd_raw))

    //sd := &ServiceDiscovery{}
    //xml.Unmarshal(spd_raw, sd)
    //spd := sd.ServiceProviderDiscovery

    //log.Printf("%+v\n",spd)

    scanner := NewMovistarScanner(MADRID)
    ok := scanner.Scan(os.Args[1]); if !ok{
        log.Fatal("Something went wrong scanning %s", MADRID)
    }

    packages := map[string]string{
        "UTX32": "TDT",
        "UTX64": "Extra",
    }

    log.Println(packages)

    //scanner.ListPackages()

    channels := scanner.GetChannelList(nil) //packages)
    DumpIPTVSimple(channels)
}

func main1(){

        //j, err := json.MarshalIndent(pd, "", "  "); if err != nil{

    //sd.ListPackages()

    //sd.GenerateIPTVSimpleList(packages, "172.16.10.9", 9998)

}
