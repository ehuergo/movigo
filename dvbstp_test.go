package main

import (
    "testing"
    "os"
    "log"
    "io"
)

func TestMessage(t *testing.T){
    f, err := os.Open("p3.raw"); if err != nil{
        t.Fatal(err)
    }

    defer f.Close()

    //FIXME: Assumed UDP messages of max 1400bytes. Safe when reading UDP directly

    defaultmsglen := 1400
    nextmsglen := defaultmsglen
    msgs := make([]*DVBSTP, 0)

    for{
        log.Println("Will read", nextmsglen)
        b := make([]byte, nextmsglen)
        n, err := io.ReadFull(f, b); if err != nil{
            t.Fatal(err)
        }

        msg := NewDVBSTPMessage(b)

        log.Printf("%+v\n%s\n", msg, string(msg.Payload))

        if msg.SectionNumber == msg.LastSectionNumber - 1{
            alreadyread := uint32(defaultmsglen - 12 ) * uint32(msg.SectionNumber + 1)
            log.Println("segmentsize", msg.SegmentSize, "sectionnumber", msg.SectionNumber, "alreadyread", alreadyread)
            //nextmsglen = int(msg.SegmentSize / uint32((defaultmsglen - 12) * int(msg.SectionNumber)))
            nextmsglen = int(msg.SegmentSize - alreadyread + 12 + 4) //+CRC
        }else{
            nextmsglen = defaultmsglen
        }
        continue

        //if len(msgs) == 0 && msg.SectionNumber != 0{
        //    log.Println("Looking for Section 0. This is", msg.SectionNumber)

        //    rem := msg.SegmentSize - uint32(((defaultmsglen - 12) * int(msg.SectionNumber)))

        //    if int(rem + 12) > defaultmsglen{
        //        nextmsglen = 1400
        //    }else{
        //        nextmsglen = int(rem) + 12
        //    }

        //    log.Println(rem)
        //    continue
        //}
            

//            remaining := msg.SegmentSize - uint32((uint16(defaultmsglen - 12) * (msg.SectionNumber + 1)))
//            log.Println("REM", remaining, "N", n)
//            if remaining < uint32(defaultmsglen - 12){
//                nextmsglen = int(remaining + 12)
//            }
//                continue
//            }





        //if msg.SectionNumber == msg.LastSectionNumber{
        //    //FIX OFFSET
        //    msgs = append(msgs, msg)
        //    break
        //    t.Fatal("Need to fix offset")
        //}

        remaining := msg.SegmentSize - uint32((uint16(defaultmsglen) * (msg.SectionNumber + 1)))
        log.Println("REM", remaining, "N", n)
        if remaining < uint32(defaultmsglen){
            nextmsglen = int(remaining)
        }

        msgs = append(msgs, msg)
    }

    log.Println(msgs)

    var xmlpayload []byte

    for _, msg := range msgs{
        xmlpayload = append(xmlpayload, msg.Payload...)
    }

    log.Println(string(xmlpayload))



    //log.Printf("%d %+v %s", n, msg, msg.Payload)

    /*
    b := make([]byte, 10)
    n, err := io.ReadFull(f, b); if err != nil{
        t.Fatal(err)
    }
    //log.Println(n, b, b[1:4])

    msg := NewDVBSTPMessage(b)

    log.Printf("%+v", msg)

    pktb := make([]byte, msg.SegmentSize)
    n, err = io.ReadFull(f, pktb); if err != nil{
        t.Fatal(err)
    }

    log.Println(n, string(pktb))
    */
}
