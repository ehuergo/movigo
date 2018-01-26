package dvbstp

import (
    "log"
    "dvbstp/sds"
)


func ReadFiles(path string, howmany int) [][]byte{
    log.Printf("Will read %d files from %s", howmany, path)

    sdsmuxer := sds.NewSDSMuxer(path)
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

    return files
}
