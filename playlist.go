package main

import (
    "fmt"
)

//type PlaylistGenerator interface{
//    Dump() []byte
//}
//
//type IPTVSimplePlaylist struct{
//}
//
//func (p *IPTVSimplePlaylist) Dump() []byte{
//
//    return nil
//}

func DumpIPTVSimple(channels []*LogicalChannel) []byte{
    //channels := sd.GetPackageChannels(packages)

    fmt.Println("#EXTM3U")
    for _, c := range channels{

        fmt.Printf("#EXTINF:-1 tvg-logo=\"%s\" tvg-chno=\"%d\" group-title=\"%s\", %s\n",
            c.GetLogoPath(),
            c.Number,
            c.FromPackage,
            c.Name)

        //fmt.Println(c.GetUDPXYString(address, port))
        fmt.Println(c.Url.AsRTP())
    }

    return nil
}
