package main

type Area uint8

const (
    CATALUNYA               Area = 1
    CASTILLA_Y_LEON         Area = 4
    COMUNIDAD_VALENCIANA    Area = 6
    BALEARES                Area = 10
    MURCIA                  Area = 12
    ASTURIAS                Area = 13
    ANDALUCIA               Area = 15
    MADRID                  Area = 19
    GALICIA                 Area = 24
    CANTABRIA               Area = 29
    LA_RIOJA                Area = 31
    EXTREMADURA             Area = 32
    ARAGON                  Area = 34
    NAVARRA                 Area = 35
    PAIS_VASCO              Area = 36
    CANARIAS                Area = 37
    CASTILLA_LA_MANCHA      Area = 38
)

func (l Area) String() string{
    switch l{
        case CATALUNYA:
            return "Catalunya"
        case CASTILLA_Y_LEON:
            return "Castilla y Leon"
        case COMUNIDAD_VALENCIANA:
            return "Comunidad Valenciana"
        case BALEARES:
            return "Baleares"
        case MURCIA:
            return "Murcia"
        case ASTURIAS:
            return "Asturias"
        case ANDALUCIA:
            return "Andalucia"
        case MADRID:
            return "Madrid"
        case GALICIA:
            return "Galicia"
        case CANTABRIA:
            return "Cantabria"
        case LA_RIOJA:
            return "La Rioja"
        case EXTREMADURA:
            return "Extremadura"
        case ARAGON:
            return "Aragon"
        case NAVARRA:
            return "Navarra"
        case PAIS_VASCO:
            return "Pais Vasco"
        case CANARIAS:
            return "Canarias"
        case CASTILLA_LA_MANCHA:
            return "Castilla la Mancha"
    }

    return "Unknown"
}

