package sds

import (
    "io"
    "hash/crc32"
    "github.com/cheggaaa/pb"
    "log"
    "fmt"
    "errors"
    "readers"
)

const (
    DGRAM_LEN = 1400
)

type SDSMuxer struct{
    sr            *SegmentReader
}

func NewSDSMuxer(uri string) SDSMuxer{

    var bytereader io.Reader

    if len(uri) > 7 && uri[:7] == "http://"{
        bytereader = readers.GetHttpReader(uri)
    }else if len(uri) > 8 && uri[:8] == "https://"{
        bytereader = readers.GetHttpReader(uri)
    }else{
        bytereader = readers.GetFilesystemReader(uri)
    }

    sr := &SegmentReader{bytereader, uri}

    return SDSMuxer{sr}
}

func (mux SDSMuxer) NextFile() ([]byte, error){

    msgs := make([]*SDSSegment, 0)
    nextlen := DGRAM_LEN
    var prevmsg *SDSSegment

    var bar *pb.ProgressBar

    newbar := func(size int){
        if bar != nil{
            bar.Finish()
        }
        bar = pb.New(size)
        bar.ShowSpeed = true
        bar.SetWidth(120)
        //log.Println("BAR STARTS")
        bar.Start()
    }

    endbar := func(){
        if bar != nil{
            bar.Finish()
        }
        //log.Println("BAR FINISH")
    }
    writebar := func(b []byte){
        if bar != nil{
            bar.Write(b)
        }
    }

    for{
        msg := mux.sr.NextSegment(nextlen)
        if bar == nil && msg.SectionNumber != 0{
            discard := msg.SegmentSize - (msg.SectionNumber * (DGRAM_LEN - SEGMENT_HEADER_LEN))
            log.Printf("We're in the middle of a file. Discarding first %d bytes", discard)
            newbar(discard)
        }

        // Next length. Should be DGRAM_LEN except for the last section
        if msg.SectionNumber == msg.LastSectionNumber - 1{
            read := (DGRAM_LEN - SEGMENT_HEADER_LEN) * (msg.SectionNumber + 1)
            nextlen = msg.SegmentSize - read + SEGMENT_HEADER_LEN + SEGMENT_CRC_LEN //last segment includes CRC
        }else{
            nextlen = DGRAM_LEN
        }

        if len(msgs) == 0 && msg.SectionNumber != 0{ //Looking for the first section
            writebar(msg.Payload)
            continue
        }

        if len(msgs) > 0{
            if msg.SectionNumber != prevmsg.SectionNumber + 1{
                endbar()
                log.Printf("WRONG SECTION SHOULD BE %d FOUND %d\n", prevmsg.SectionNumber + 1, msg.SectionNumber)
                return nil, errors.New("Error in sequence")
            }
        }else{
            endbar()
            log.Printf("New file starts. Size is %d bytes. Has %d sections.\n", msg.SegmentSize, msg.LastSectionNumber + 1)
            newbar(msg.SegmentSize)
        }

        writebar(msg.Payload)

        msgs = append(msgs, msg)

        prevmsg = msg

        if msg.SectionNumber == msg.LastSectionNumber{
            break
        }
    }

    endbar()

    data := make([]byte, 0)
    for _, msg := range msgs{
        data = append(data, msg.Payload...)
    }

    datacrc := crc32.ChecksumIEEE(data)

    log.Printf("File complete (%d/%d) -> (%d/%d). CRC 0x%x matches? %t", 
        len(msgs), prevmsg.LastSectionNumber + 1, 
        len(data), prevmsg.SegmentSize,
        prevmsg.CRC, prevmsg.CRC == datacrc)

    if prevmsg.CRC != datacrc{
        return nil, errors.New(fmt.Sprintf("CRC mismatch 0x%x != 0x%x", datacrc, prevmsg.CRC))
    }

    return data, nil
}
