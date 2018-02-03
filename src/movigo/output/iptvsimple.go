package output

import (
    "fmt"
    "log"
    "sort"
    "movigo/movi"
)

func DumpGroupsAsIPTVSimple(groups map[int]*movi.ChannelGroup, prefix string) []byte{
    var keys []int
    data := []byte("#EXTM3U\n")

    for k := range groups{
        keys = append(keys, k)
    }
    sort.Ints(keys)

    for _, k := range keys{
        group := groups[k]
        if len(group.HD) > 0{
            data = append(data, dumpIPTVSimpleChannel(group.HD[0], prefix)...)
        }else if len(group.SD) > 0{
            data = append(data, dumpIPTVSimpleChannel(group.SD[0], prefix)...)
        }else{
            log.Println("WARNING: No SD or HD channels in group", group)
        }
    }

    return data
}

func dumpIPTVSimpleChannel(c *movi.LogicalChannel, prefix string) []byte{

    extinf := fmt.Sprintf("#EXTINF:-1 tvg-chid=\"%d\" tvg-logo=\"%s\" tvg-chno=\"%d\" group-title=\"%s\", %s\n",
        c.Id,
        c.GetLogoPath(),
        c.Number,
        c.FromPackage,
        c.Name)

    url := fmt.Sprintf("%s%s\n", prefix, c.Url.Raw())

    return append([]byte(extinf), []byte(url)...)
}

func DumpIPTVSimple(channels []*movi.LogicalChannel, prefix string) []byte{

    data := []byte("#EXTM3U\n")

    for _, c := range channels{
        data = append(data, dumpIPTVSimpleChannel(c, prefix)...)
    }

    return data
}
