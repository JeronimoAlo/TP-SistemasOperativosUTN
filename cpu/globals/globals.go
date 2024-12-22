package globals

var HayHiloEjecutandoEnCPU bool = false

type CpuConfig struct {
	Ip_memory 	string 	`json:"ip_memory"` // IP a la cual se deberá conectar con la Memoria.
	Port_memory int 	`json:"port_memory"` // Puerto al cual se deberá conectar con la Memoria.
	Ip_kernel 	string 	`json:"ip_kernel"` // IP a la cual se deberá conectar con la Kernel.
	Port_kernel int 	`json:"port_kernel"` // Puerto al cual se deberá conectar con la Kernel.
	Port 		int 	`json:"port"` // Puerto en el cual se escuchará las conexiones
	Log_level 	string 	`json:"log_level"` // Nivel de detalle máximo a mostrar.
}

var Config *CpuConfig

type SolicitudContexto struct {
    PID         string  `json:"PID"`
    TID         string  `json:"TID"`
}

// Definimos los registros que utilizara la CPU para Funcionar
// unit32: conjunto de todos los enteros de 32 bits sin signo. Desde el 0 al 4294967295.
type CPU struct {
    PC      uint32  // Program Counter
    AX      uint32  // Registro de propósito general
    BX      uint32  // Registro de propósito general
    CX      uint32  // Registro de propósito general
    DX      uint32  // Registro de propósito general
    EX      uint32  // Registro de propósito general
    FX      uint32  // Registro de propósito general
    GX      uint32  // Registro de propósito general
    HX      uint32  // Registro de propósito general
    Base    uint32  // Dirección base del proceso
    Limite  uint32  // Límite del proceso
}

type EstadoCPU int

// Definicion de Estados en Global (Fetch, Decode, Execute y Check Interrupt.)
const (
    Fetch EstadoCPU = iota       // Estado "Fetch" (Valor 0) 
    Decode                   // Estado "Decode" (Valor 1)
    Execute                  // Estado "Execute" (Valor 2)
    CheckInterrupt               // Estado "Check Interrupt" (Valor 3)
)


// Estructura que representa una instrucción a ejecutar por la CPU usada en la función "Ejecutar Instruccion".
type Instruccion struct {
    Operacion         string // La operación a realizar (SET, READ_MEM, etc.)
    Registro          string // El registro involucrado (AX, BX, etc.)
    Valor             int    // Un valor numérico para operaciones como SET
    RegistroDatos     string // Para operaciones de lectura y escritura de memoria
    RegistroDireccion string // Para operaciones de lectura y escritura que usan registros como direcciones
    Instruccion       int    // Usado para JNZ, representa la instrucción a saltar si corresponde
    RegistroDestino   string // Usado para operaciones SUM y SUB
    RegistroOrigen    string // Usado para operaciones SUM y SUB
}

// Estructura de la MMU para manejar la traducción de direcciones
type MMU struct{}

type TIDyPIDaEjecutar struct{
    TID int
    PID int
}

type SolicitudInstruccion struct{
    TID int
    PID int
    PC uint32
}

type InstruccionDesdeMemoria struct{
    Operacion  string
    Parametros []string
    NoHayMasInstrucciones bool
}

type InstruccionDecodificada struct{
    Operacion  string
    Parametros []string
    RequiereTraduccion bool
}

var TIDyPIDenEjecucion TIDyPIDaEjecutar

var RegistrosCPU CPU
var FinDeEjecucion bool = false

var HayInterrupciones bool = false

type LecturaEscrituraMemoria struct { // Utilizada para las funciones READ_MEM y WRITE_MEM
	DireccionFisica uint32
	ValorAenviar uint32
    PID int
    TID int
}

type MensajeActualizarContexto struct {
	PID int    `json:"PID"` // PID
	TID int    `json:"TID"` // TID
	PC  uint32 `json:"PC"`  // Program Counter
	AX  uint32 `json:"AX"`  // Registro de propósito general
	BX  uint32 `json:"BX"`  // Registro de propósito general
	CX  uint32 `json:"CX"`  // Registro de propósito general
	DX  uint32 `json:"DX"`  // Registro de propósito general
	EX  uint32 `json:"EX"`  // Registro de propósito general
	FX  uint32 `json:"FX"`  // Registro de propósito general
	GX  uint32 `json:"GX"`  // Registro de propósito general
	HX  uint32 `json:"HX"`  // Registro de propósito general
}

type InterrupcionDeHilo struct{
	TID int
	PID int
    Motivo string
}

var ListaDeInterrupcionesDeHilo []InterrupcionDeHilo

type MensajeResultado struct {
	Status string
}

type ValorAresponder struct {
    Valor uint32
}

type IOparaHacer struct{

    TID int
    PID int
    TiempoEnMilisegundos int

}