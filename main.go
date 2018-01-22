package main

import (
    "log"
//    "encoding/json"
    "os"
    "sort"
)

func main(){

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

    groups := scanner.GetChannelGroups(nil)

    var keys []int
    for k := range groups{
        keys = append(keys, k)
    }
    sort.Ints(keys)

    //for _, group := range groups{
    for _, k := range keys{
        group := groups[k]
        log.Println(group.Number, len(group.SD), len(group.HD))
    }
    //log.Println("GROUPS", groups)
    //channels := scanner.GetChannelList(nil) //packages)
    //DumpIPTVSimple(channels, "172.16.10.9", 9998)
}
