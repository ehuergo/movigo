package epg

// All credits to: https://github.com/MovistarTV/tv_grab_es_movistartv/blob/master/tv_grab_es_movistartv.py#L610

import (
    "fmt"
    "time"
    "regexp"
    "encoding/binary"
    "encoding/json"
)

type Chunk struct{
    End bool
    Size uint32
    Type uint8
    FileId uint16
    ChunkNumber uint16
    ChunkCount uint8
    Data []byte
}

func ParseChunk(data []byte) *Chunk{
    chunk := &Chunk{}
    chunk.End = data[0] == 0x01
    chunk.Size = Uint32(append([]byte{0}, data[1:4]...))
    chunk.Type = data[4]
    chunk.FileId = Uint16(data[5:7])
    chunk.ChunkNumber = Uint16(data[8:10]) / 0x10
    chunk.ChunkCount = data[10]

    //fmt.Printf("CHUNK %+v\n", chunk)
    chunk.Data = data[12:]

    return chunk
}

type File struct{
    Size        uint16
    ServiceId   uint16
    ServiceVer  uint8
    ServiceURL  string
    DataStart   uint16
    Data        []byte
    Checksum    uint32
}

func ParseFile(data []byte) *File{
    file := &File{}
    file.Size = Uint16(data[1:3])
    if file.Size == 0{ return nil }
    file.ServiceId = Uint16(data[3:5])
    file.ServiceVer = data[5]
    file.DataStart = uint16(data[6] + 7)
    file.ServiceURL = string(data[7:file.DataStart])
    file.Checksum = Uint32(data[len(data)-4:])
    file.Data = data[file.DataStart:len(data)-4]

    return file
}

type Program struct{
    Start       time.Time
    Duration    time.Duration
    TitleEnd    uint8
    Pid         uint32
    Genre       uint8
    Age         uint8
    Title       string
    IsSerie     bool

    Serie
}

func (p *Program) MarshalJSON() ([]byte, error){
    return json.Marshal(&struct{
        Start       int64
        Duration    float64
        TitleEnd    uint8
        Pid         uint32
        Genre       uint8
        Age         uint8
        Title       string
        IsSerie     bool
        Serie
    }{
        Start:      p.Start.Unix(),
        Duration:   p.Duration.Seconds(),
        TitleEnd:   p.TitleEnd,
        Pid:        p.Pid,
        Genre:      p.Genre,
        Age:        p.Age,
        Title:      p.Title,
        IsSerie:    p.IsSerie,
        Serie:      p.Serie,
    })
}

func (p *Program) UnmarshalJSON(data []byte) error{
    aux := &struct{
        Start       int64
        Duration    float64
        TitleEnd    uint8
        Pid         uint32
        Genre       uint8
        Age         uint8
        Title       string
        IsSerie     bool
        Serie
    }{}

    err := json.Unmarshal(data, aux); if err != nil{
        return err
    }

    p.Start =     time.Unix(aux.Start, 0)
    p.Duration =  time.Duration(aux.Duration) * 1000000000
    p.TitleEnd =  aux.TitleEnd
    p.Pid =       aux.Pid
    p.Genre =     aux.Genre
    p.Age =       aux.Age
    p.Title =     aux.Title
    p.IsSerie =   aux.IsSerie
    p.Serie =     aux.Serie

    //fmt.Println(p)

    return nil
}

func ParsePrograms(data []byte) []*Program{
    progs := make([]*Program, 0)
    off := 0
    //log.Println(data)
    for off + 108 < len(data){
    //for off > 31 && off + int(data[31]) + 33 < len(data){
        //log.Println(data[off:off+24])
        prog := &Program{}
        prog.Pid = Uint32(data[off:off+4])
        prog.Start = time.Unix(int64(Uint32(data[off+4:off+8])), 0)
        prog.Duration = time.Duration(Uint16(data[off+8:off+10])) * 1000000000
        prog.Genre = data[off+20]
        prog.Age = data[off+24]
        prog.TitleEnd = data[off+31] + 32
        prog.Title = decodetitle(data[off+32:off+int(prog.TitleEnd)]) //string(data[32:prog.TitleEnd]) 

        off += int(prog.TitleEnd)

        prog.IsSerie = data[off] == 0xf1

        if prog.IsSerie{
            soff := 0
            prog.Serie, soff = ParseSerie(prog.Title, data[off+1:], off+1)

            off += soff
        }
        off += int(data[off+3]) + 4
        //fmt.Printf("PROG % 5d/% 05d %+v\n", off, len(data), prog)
        progs = append(progs, prog)
    }

    return progs
}

type Serie struct{
    SerieId         uint16
    Episode         uint8
    Year            uint16
    Season          uint8
    Title           string
    TitleEnd        uint8

    *ParsedSerie
}

type ParsedSerie struct{
    ParsedName      string
    ParsedTitle     string
    ParsedSeason    string
    ParsedEpisode   string
}

func decodeSerieTitle(title string) *ParsedSerie{
    r := regexp.MustCompile(`^(.*) T([^ ]).*Ep. ([^ ]) ?-? ?(.*)$`)
    x := r.FindStringSubmatch(title)

    if len(x) < 4{ return nil }

    s := &ParsedSerie{}

    s.ParsedName = x[1]
    s.ParsedSeason = x[2]
    s.ParsedEpisode = x[3]
    s.ParsedTitle = x[4]

    return s
}

func ParseSerie(progtitle string, data []byte, off int) (Serie, int){
    serie := Serie{}
    serie.ParsedSerie = decodeSerieTitle(progtitle)
    //fmt.Println("SERIEDATA", data[:16])
    title_end := int(data[11] + 13)
    serie.SerieId = Uint16(data[4:6])
    serie.Episode = data[7]
    serie.Year = Uint16(data[8:10])
    serie.Season = data[10]
    //serie.Title = decodetitle(data[12:title_end - off])
    fmt.Printf("%s %+v %+v\n", progtitle, serie.ParsedSerie, serie)
    return serie, title_end
}

func decodetitle(b[]byte) string{
    for i, x := range b{
        b[i] = x ^0x15
    }

    return string(b)
}

func Uint16(b []byte) uint16{
    return binary.BigEndian.Uint16(b)
}

func Uint32(b []byte) uint32{
    return binary.BigEndian.Uint32(b)
}

