package globals

type FileSystemConfig struct {
	Port 				int 	`json:"port"` // Puerto en el cual se escuchará la conexión de módulo.
	Ip_memory 			string 	`json:"ip_memory"` // IP a la cual se deberá conectar con la Memoria.
	Port_memory 		int 	`json:"port_memory"` // Puerto al cual se deberá conectar con la Memoria.
	Mount_dir 			string 	`json:"mount_dir"` // Path a partir del cual van a encontrarse los archivos del FS.
	Block_size 			uint32 	`json:"block_size"` // Tamaño de los bloques del FS.
	Block_count 		int 	`json:"block_count"` // Cantidad de bloques del FS.
	Block_access_delay 	int 	`json:"block_access_delay"` // Tiempo en milisegundos que se deberá esperar luego de cada acceso a bloques (de datos o punteros).
	Log_level 			string 	`json:"log_level"` // Nivel de detalle máximo a mostrar.
}

var Config *FileSystemConfig

// Estructura Archivo MetaData
type Metadata struct {
	IndexBlock uint32 `json:"index_block"`
	Size       uint32 `json:"size"`
}

type ConfigDelay struct{
	Delay int `json:"block_access_delay"`
}

type Instruccion struct {
	Operacion  string	`json:"Operacion"`
	Parametros []string `json:"Parametros"`
}

type HiloAEscribir struct{
	TID int	`json:"TID"`
	PID int	`json:"PID"`
	Codigo []Instruccion `json:"Codigo"`// Pseudocódigo del hilo.
}

var (
	BitmapPath  string
	BloquesPath string
)

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

type IndiceDeBloques struct {
	PunterosABloques []uint32
}