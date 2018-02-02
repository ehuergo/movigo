package sds

import (
    "log"
    "io"
)

type SegmentReader struct{
    r               io.Reader
}

func (sr SegmentReader) NextDatagram(length int) []byte{
    b := make([]byte, length)
    //n, err := sr.r.Read(b); if err != nil{
    n, err := io.ReadFull(sr.r, b); if err != nil{
        log.Fatal(err)
    }

    //log.Println("Requested", length, "read", n, "OK?", n == length)

    if n != length{
        log.Fatal(n, "!=", length)
    }

    return b[:n]
}

func (sr SegmentReader) NextSegment(length int) *SDSSegment{
    dgram := sr.NextDatagram(length)
    msg := NewSDSSegment(dgram)

    //log.Printf("%d/%d (%d)", msg.SectionNumber, msg.LastSectionNumber, msg.SegmentSize)

    return msg
}

