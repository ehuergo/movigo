package main

import (
    "fmt"
    "log"
    "sort"
)

func DumpGroupsAsIPTVSimple(groups map[int]*ChannelGroup, xyaddr string, xyport int) []byte{
    var keys []int
    for k := range groups{
        keys = append(keys, k)
    }
    sort.Ints(keys)

    for _, k := range keys{
        group := groups[k]
        if len(group.HD) > 0{
            dumpIPTVSimpleChannel(group.HD[0], xyaddr, xyport)
        }else if len(group.SD) > 0{
            dumpIPTVSimpleChannel(group.SD[0], xyaddr, xyport)
        }else{
            log.Println("WARNING: No SD or HD channels in group", group)
        }
    }

    return nil
}

func dumpIPTVSimpleChannel(c *LogicalChannel, xyaddr string, xyport int) []byte{
    fmt.Printf("#EXTINF:-1 tvg-logo=\"%s\" tvg-chno=\"%d\" group-title=\"%s\", %s\n",
        c.GetLogoPath(),
        c.Number,
        c.FromPackage,
        c.Name)

    if xyaddr != ""{
        fmt.Println(c.Url.AsUDPXY(xyaddr, xyport))
    }else{
        fmt.Println(c.Url.AsRTP())
    }

    return nil
}

func DumpIPTVSimple(channels []*LogicalChannel, xyaddr string, xyport int) []byte{

    fmt.Println("#EXTM3U")
    for _, c := range channels{
        dumpIPTVSimpleChannel(c, xyaddr, xyport)
    }

    return nil
}
