package main

import (
    "fmt"
)

func DumpIPTVSimple(channels []*LogicalChannel, xyaddr string, xyport int) []byte{

    fmt.Println("#EXTM3U")
    for _, c := range channels{

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
    }

    return nil
}
