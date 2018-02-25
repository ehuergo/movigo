package movistarxml

import (
    "encoding/xml"
    "time"
    "io/ioutil"
    "sort"
)

type Export struct{
    Pases       []Pase  `xml:"pase"`
}

type Pase struct{
    Cadena      string  `xml:"cadena,attr"`
    Fecha       string  `xml:"fecha,attr"`
    Hora        string  `xml:"hora"`
    Time        time.Time
    DescCorta   string  `xml:"descripcion_corta"`
    Titulo      string  `xml:"titulo"`
    TipoFicha   string  `xml:"tipo_ficha"`
    SinopCorta  string  `xml:"sinopsis_corta"`
    SinopLarga  string  `xml:"sinopsis_larga"`
    Web         string  `xml:"web"`
}

func Unmarshal(data []byte, v interface{}) error{
    return xml.Unmarshal(data, v)
}

func ReadFile(path string) (*Export, error){
    data, err := ioutil.ReadFile(path); if err != nil{
        return nil, err
    }

    export := &Export{}
    export.Pases = make([]Pase, 0)

    Unmarshal(data, export)

    if export != nil{
        Sort(export)
    }

    return export, nil
}

func Sort(export *Export){
    sort.Slice(export.Pases, func(i, j int) bool{
            return export.Pases[i].Cadena + export.Pases[i].Fecha + export.Pases[i].Hora < export.Pases[j].Cadena + export.Pases[j].Fecha + export.Pases[j].Hora
    })
}
