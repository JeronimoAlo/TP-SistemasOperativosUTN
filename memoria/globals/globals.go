package globals

type MemoryConfig struct {
	Port             int    `json:"port"`             // Puerto en el cual se escuchará la conexión de módulo.
	Memory_size      int    `json:"memory_size"`      // Tamaño expresado en bytes del espacio de usuario de la memoria.
	Instruction_path string `json:"instruction_path"` // Carpeta donde se encuentran los archivos de pseudocódigo.
	Response_delay   int    `json:"response_delay"`   // Tiempo en milisegundos que se deberá esperar antes de responder a las solicitudes de CPU.
	Ip_kernel        string `json:"ip_kernel"`        // IP a la cual se deberá conectar con la Kernel.
	Port_kernel      int    `json:"port_kernel"`      // Puerto al cual se deberá conectar con la Kernel.
	Ip_cpu           string `json:"ip_cpu"`           // IP a la cual se deberá conectar con la CPU.
	Port_cpu         int    `json:"port_cpu"`         // Puerto al cual se deberá conectar con la CPU.
	Ip_filesystem    string `json:"ip_filesystem"`    // IP a la cual se deberá conectar con la Filesystem.
	Port_filesystem  int    `json:"port_filesystem"`  // Puerto al cual se deberá conectar con la Filesystem.
	Scheme           string `json:"scheme"`           // Esquema de particiones de memoria a utilizar FIJAS / DINAMICAS.
	Search_algorithm string `json:"search_algorithm"` // Algoritmo de busqueda de huecos de memoria: FIRST / BEST / WORST
	Partitions       []int  `json:"partitions"`       // Lista ordenada con las particiones a generar en el algoritmo de Particiones Fijas.
	Log_level        string `json:"log_level"`        // Nivel de detalle máximo a mostrar.
}

var Config *MemoryConfig

type MensajeObtenerContexto struct {
	PID string `json:"PID"`
	TID string `json:"TID"`
}

type TIDyPID struct {
	PID int
	TID int
}

type Proceso struct {
	PID int
	ContextoEjecucionProceso *ContextoEjecucionProceso
	Tamanio int
	ParticionAsignada *Particion
}

type ContextoEjecucionProceso struct {
	Base   uint32 // Dirección base del proceso
	Limite uint32 // Límite del proceso
}

type ContextoEjecucionHilo struct {
	PID    int	
	TID    int
	PC     uint32 // Program Counter
	AX     uint32 // Registro de propósito general
	BX     uint32 // Registro de propósito general
	CX     uint32 // Registro de propósito general
	DX     uint32 // Registro de propósito general
	EX     uint32 // Registro de propósito general
	FX     uint32 // Registro de propósito general
	GX     uint32 // Registro de propósito general
	HX     uint32 // Registro de propósito general
	Base   uint32 // Dirección base del proceso
	Limite uint32 // Límite del proceso
}

type RespuestaContextoProcesoHilo struct {
	PC     uint32 // Program Counter
	AX     uint32 // Registro de propósito general
	BX     uint32 // Registro de propósito general
	CX     uint32 // Registro de propósito general
	DX     uint32 // Registro de propósito general
	EX     uint32 // Registro de propósito general
	FX     uint32 // Registro de propósito general
	GX     uint32 // Registro de propósito general
	HX     uint32 // Registro de propósito general
	Base   uint32 // Dirección base del proceso
	Limite uint32 // Límite del proceso
}

type SolicitudDump struct{
	ContenidoAgrabar ProcesoDump
	Tamanio int
	Nombre string
	HiloSolicitante int
	ProcesoSolicitante int
}

type ProcesoDump struct{
	ContenidoBytes []byte
}

type Hilo struct {
	PID                   int
	TID                   int
	ContextoEjecucionHilo ContextoEjecucionHilo
	Codigo       []Instruccion // Pseudocódigo del hilo.
}

type SolicitudInstruccion struct{
    TID int
    PID int
    PC uint32
}

// Estructura de un proceso que es enviado por el kernel, se usa para decodificar el json enviado por el kernel
type ProcesoMemoria struct{
	PID int	//`json:"PID"`
	Tamanio int //`json:"Tamanio"`
}

type Instruccion struct {
	Operacion  string	`json:"Operacion"`
	Parametros []string `json:"Parametros"`
}

// Estructura de un hilo que es enviado por el kernel, se usa para decodificar el json enviado por el kernel
type HiloMemoria struct{
	TID int	`json:"TID"`
	PID int	`json:"PID"`
	Codigo       []Instruccion `json:"Codigo"`// Pseudocódigo del hilo.
}

type InstruccionDesdeMemoria struct{
    Operacion  string
    Parametros []string
    NoHayMasInstrucciones bool
	PC uint32 // Program Counter
	AX uint32 // Registro de propósito general
	BX uint32 // Registro de propósito general
	CX uint32 // Registro de propósito general
	DX uint32 // Registro de propósito general
	EX uint32 // Registro de propósito general
	FX uint32 // Registro de propósito general
	GX uint32 // Registro de propósito general
	HX uint32 // Registro de propósito general
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

var ListaContextosEjecucionHilos []*Hilo

var ListaContextosEjecucionProcesos []*Proceso

type LecturaEscrituraMemoria struct { // Utilizada para las funciones READ_MEM y WRITE_MEM
	DireccionFisica uint32
	ValorAenviar uint32
	PID int
	TID int
}

type Particion struct { // Cada particion tendra su estado (libre o no), su tamanio total y su inicio.
	Tamanio int
	Libre bool
	Base uint32
	Limite int
}

var Particiones []*Particion // Declaramos que particiones será un slice de structs Partición.

// Declarar Memoria a nivel de paquete (global).
var Memoria []byte

type ValorAresponder struct {
	Valor uint32
}