package main

import (
    "encoding/binary"
    "net"
    "io"
    "os"
    "log"
//    "strings"
    "hash/crc32"
)

const (
    DVBSTP_HEADER_LEN = 12
    DVBSTP_CRC_LEN  = 4

    DVBSTP_DGRAM_LEN = 1400
)

type DVBSTP struct{
    Ver                 uint8
    Resrv               uint8
    Enc                 uint8
    CRCPresent          bool

    SegmentSize         uint32
    PayloadID           uint8
    SegmentID           uint16
    SegmentVersion      uint8
    SectionNumber       uint16
    LastSectionNumber   uint16

    Compr               uint8
    P                   bool
    PrivateHdrlen       uint8

    Payload             []byte

    CRC                 uint32

    ServiceProviderID   net.IP

}

func NewDVBSTPMessage(payload []byte) *DVBSTP{
    m := &DVBSTP{}
    m.Ver = payload[0] & 0xc0
    m.Resrv = payload[0] & 0x38
    m.Enc = payload[0] & 0x6
    m.CRCPresent = payload[0]& 0x1 == 0x1

    m.SegmentSize = binary.BigEndian.Uint32(append([]byte{0}, payload[1:4]...))

    m.PayloadID = payload[4]
    m.SegmentID = binary.BigEndian.Uint16(payload[5:7])
    m.SegmentVersion = payload[7]

    m.SectionNumber = binary.BigEndian.Uint16(payload[8:10]) >> 4
    m.LastSectionNumber = binary.BigEndian.Uint16(payload[9:11]) & 0x3ff

    m.Compr = payload[11] & 0xe0
    m.P = payload[11] & 0x10 == 0x1
    m.PrivateHdrlen = payload[11] & 0x0f

    //m.ServiceProviderID = net.IP(payload[12:16])
    //m.Payload = payload[m.Hdrlen: len(payload) - int(m.Hdrlen) - 4]

    if m.CRCPresent{
        m.CRC = binary.BigEndian.Uint32(payload[len(payload) - DVBSTP_CRC_LEN:])
        m.Payload = payload[DVBSTP_HEADER_LEN:len(payload) - DVBSTP_CRC_LEN]
    }else{
        m.Payload = payload[DVBSTP_HEADER_LEN:]
    }

    return m
}


type DVBSTPStreamReader struct{
    path            string
}

func NewDVBSTPStreamReader(path string) *DVBSTPStreamReader{
    r := &DVBSTPStreamReader{}
    r.path = path

    return r
}

func (r *DVBSTPStreamReader) ReadFiles(howmany int) [][]byte{

    log.Printf("Requested to read %d files from %s", howmany, r.path)

    files := make([][]byte, 0)

    f, err := os.Open(r.path); if err != nil{
        log.Fatal(err)
    }
    defer f.Close()

    reset := false

    for{
        msgs := make([]*DVBSTP, 0)
        nextlen := uint32(DVBSTP_DGRAM_LEN)
        nextcrc := uint32(0)
        nextsection := uint16(0)

        for{
            b := make([]byte, nextlen)
            n, err := io.ReadFull(f, b); if err != nil{
                log.Fatal(n, err)
            }
//            log.Println("Requested", nextlen, "read", n, "OK?", uint32(n) == nextlen)

            msg := NewDVBSTPMessage(b)
//            log.Printf("%+v\n", msg)
            //log.Printf("%+v\n%s\n", msg, string(msg.Payload))

            if msg.SectionNumber == msg.LastSectionNumber - 1{
                alreadyread := uint32(DVBSTP_DGRAM_LEN - DVBSTP_HEADER_LEN) * uint32(msg.SectionNumber + 1)
                nextlen = uint32(msg.SegmentSize - alreadyread + DVBSTP_HEADER_LEN + DVBSTP_CRC_LEN) //Last Segment includes CRC
            }else{
                nextlen = DVBSTP_DGRAM_LEN
            }

            if len(msgs) == 0 && msg.SectionNumber != 0{ //Looking for first section if msg slice is still empty
//                log.Printf("Skip this section %d/%d\n", msg.SectionNumber, msg.LastSectionNumber)
                continue
            }

            if len(msgs) == 0{
                log.Printf("File starts. Size %d bytes. Sections %d.", msg.SegmentSize, msg.LastSectionNumber + 1)
            }

            if msg.SectionNumber != nextsection{
                log.Printf("Wrong SECTION Wrong SECTION %d should be %d\n", msg.SectionNumber, nextsection)
                reset = true
                nextsection = 0
                break
            }

            nextsection++

//            log.Println("SEGMENT PAYLOAD", strings.Replace(string(msg.Payload), "\n", "", -1))

            msgs = append(msgs, msg)

            if msg.SectionNumber == msg.LastSectionNumber{
                log.Printf("This was the last section (%d). CRC should be 0x%x", msg.SectionNumber, msg.CRC)
                nextcrc = msg.CRC
                nextsection = 0
                break
            }
        }

        if reset{
            log.Println("RESET detected. Start over.")
            files = make([][]byte, 0)
            reset = false
            continue
        }

        file := make([]byte, 0)
        for _, msg := range msgs{
            //log.Println("FILE", string(file))
            file = append(file, msg.Payload...)
        }

        filecrc := crc32.ChecksumIEEE(file)

        log.Printf("Finished read file. Read %d bytes. CRC 0x%x is OK: %t", len(file), filecrc, filecrc == nextcrc)

        files = append(files, file)
        if len(files) == howmany{
            break
        }
    }

    return files
}











