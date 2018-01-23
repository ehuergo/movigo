package main

import (
    "encoding/binary"
    "net"
    "log"
//    "strings"
)

const (
    DVBSTP_HEADER_LEN = 12
    DVBSTP_CRC_LEN  = 4
)

type DVBSTP struct{
    Ver                 uint8
    Resrv               uint8
    Enc                 uint8
    CRCPresent          bool

    SegmentSize         int
    PayloadID           uint8
    SegmentID           uint16
    SegmentVersion      uint8
    SectionNumber       int
    LastSectionNumber   int

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

    m.SegmentSize = int(binary.BigEndian.Uint32(append([]byte{0}, payload[1:4]...)))

    m.PayloadID = payload[4]
    m.SegmentID = binary.BigEndian.Uint16(payload[5:7])
    m.SegmentVersion = payload[7]

    m.SectionNumber = int(binary.BigEndian.Uint16(payload[8:10]) >> 4)
    m.LastSectionNumber = int(binary.BigEndian.Uint16(payload[9:11]) & 0x3ff)

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


type DVBSTPReader struct{
    path            string
}

func NewDVBSTPReader(path string) *DVBSTPReader{
    r := &DVBSTPReader{}
    r.path = path

    return r
}

func (r *DVBSTPReader) ReadFiles(howmany int) [][]byte{
    log.Printf("Will read %d files from %s", howmany, r.path)

    filereader := NewDVBSTPFileReader(r.path)
    files := make([][]byte, 0)
    for len(files) < howmany{
        file, err := filereader.NextFile(); if err != nil{
            log.Printf("Error reading file: %s. Starting over", err)
            files = make([][]byte, 0) //Start over if errors found. We don't want dupe files
            continue
        }
        files = append(files, file)
    }

    log.Printf("All %d files read.", len(files))

    return files
}
