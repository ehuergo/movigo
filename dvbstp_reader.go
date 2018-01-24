package main

import (
    "io"
    "os"
    "log"
    "fmt"
    "errors"
    "net/http"
    "hash/crc32"
)

const (
    DVBSTP_DGRAM_LEN = 1400
)


type DVBSTPMessageReader struct{
    r               io.Reader
    uri             string
}

func (reader DVBSTPMessageReader) NextDatagram(length int) []byte{
    b := make([]byte, length)
    n, err := io.ReadFull(reader.r, b); if err != nil{
        log.Fatal(err)
    }

    if n != length{
        log.Fatal(n, "!=", length)
    }

    //log.Println("Requested", length, "read", n, "OK?", n == length)

    return b
}

func (reader DVBSTPMessageReader) NextMessage(length int) *DVBSTP{
    dgram := reader.NextDatagram(length)
    msg := NewDVBSTPMessage(dgram)

    return msg
}


type DVBSTPFileReader struct{
    msgreader       *DVBSTPMessageReader
}

func NewDVBSTPFileReader(uri string) DVBSTPFileReader{

    var bytereader io.Reader

    if len(uri) > 7 && uri[:7] == "http://"{
        bytereader = getHttpReader(uri)
    }else if len(uri) > 8 && uri[:8] == "https://"{
        bytereader = getHttpReader(uri)
    }else{
        bytereader = getFilesystemReader(uri)
    }

    msgreader := &DVBSTPMessageReader{bytereader, uri}

    return DVBSTPFileReader{msgreader}
}

func (reader DVBSTPFileReader) NextFile() ([]byte, error){

    msgs := make([]*DVBSTP, 0)
    nextlen := DVBSTP_DGRAM_LEN
    var prevmsg *DVBSTP

    for{
        msg := reader.msgreader.NextMessage(nextlen)

        // Next length. Should be DVBSTP_DGRAM_LEN except for the last section
        if msg.SectionNumber == msg.LastSectionNumber - 1{
            read := (DVBSTP_DGRAM_LEN - DVBSTP_HEADER_LEN) * (msg.SectionNumber + 1)
            nextlen = msg.SegmentSize - read + DVBSTP_HEADER_LEN + DVBSTP_CRC_LEN //last segment includes CRC
        }else{
            nextlen = DVBSTP_DGRAM_LEN
        }

        if len(msgs) == 0 && msg.SectionNumber != 0{ //Looking for the first section
            continue
        }

        if len(msgs) > 0{
            if msg.SectionNumber != prevmsg.SectionNumber + 1{
                log.Printf("WRONG SECTION SHOULD BE %d FOUND %d\n", prevmsg.SectionNumber + 1, msg.SectionNumber)
                return nil, errors.New("Error in sequence")
            }
        }else{
            log.Printf("New file starts. Size is %d bytes. Has %d sections.\n", msg.SegmentSize, msg.LastSectionNumber + 1)
        }

        msgs = append(msgs, msg)

        prevmsg = msg

        if msg.SectionNumber == msg.LastSectionNumber{
            break
        }
    }

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

func getHttpReader(uri string) io.Reader{
    get, err := http.Get(uri); if err != nil{
        log.Fatal(err)
    }

    return get.Body
}

func getFilesystemReader(uri string) io.Reader{
    log.Println(uri)
    f, err := os.Open(uri); if err != nil{
        log.Fatal(err)
    }

    return f
}
