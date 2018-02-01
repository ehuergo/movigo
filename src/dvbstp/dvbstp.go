package dvbstp

import (
    "log"
    "dvbstp/sds"
    "io"
)


func ReadSDSFiles(r io.Reader, howmany int) [][]byte{
    log.Printf("Will read %d SD&S files from %s", howmany, r)

    sdsmuxer := sds.NewSDSMuxer(r)
    files := make([][]byte, 0)
    for len(files) < howmany{
        file, err := sdsmuxer.NextFile(); if err != nil{
            log.Printf("Error reading file: %s. Starting over", err)
            files = make([][]byte, 0) //Start over if errors found. We don't want dupe files
            continue
        }
        files = append(files, file)
    }

    log.Printf("All %d files read.", len(files))

    r.(io.Closer).Close()

    return files
}

