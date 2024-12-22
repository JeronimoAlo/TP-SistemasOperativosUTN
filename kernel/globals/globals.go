package globals

type KernelConfig struct {
	Ip_memory          string `json:"ip_memory"`          // IP a la cual se deberá conectar con la memoria.
	Port_memory        int    `json:"port_memory"`        // Puerto al cual se deberá conectar con la memoria.
	Ip_cpu             string `json:"ip_cpu"`             // IP a la cual se deberá conectar con la CPU.
	Port_cpu           int    `json:"port_cpu"`           // Puerto al cual se deberá conectar con la CPU.
	Sheduler_algorithm string `json:"sheduler_algorithm"` // Define el algoritmo de planificación de corto plazo. (FIFO / PRIORIDADES / CMN).
	Quantum            int    `json:"quantum"`            // Tiempo en milisegundos del quantum para utilizar bajo el algoritmo RR en Colas Multinivel.
	Log_level          string `json:"log_level"`          // Nivel de detalle máximo a mostrar.
	Port               int    `json:"port"`               // Puerto en el cual se escuchará la conexión de módulo.
}

var Config *KernelConfig

type Estado int // Tipo de dato para referenciar al estado actual de un proceso.

// Estructura básica de un PCB para administrar los procesos.
type PCB struct {
	PID     int           // Identificador único del proceso.
	TIDs    []int         // Lista de TIDs asociados al proceso.
	Mutex   []Mutex 	  // Lista de mutex creados por el proceso.
	Estado  Estado        // Estado actual del proceso.
	Tamanio int           // Tamaño en bytes del proceso.
}

type Instruccion struct {
	Operacion  string	`json:"Operacion"`
	Parametros []string `json:"Parametros"`
}

// Estructura básica de un TCB para administrar los hilos de un proceso.
type TCB struct {
	PID          int           // Identificador del proceso padre.
	TID          int           // Identificador del hilo dentro del proceso.
	Prioridad    int           // Prioridad del hilo dentro del sistema.
	Estado       Estado        // Estado actual del hilo.
	Codigo       []Instruccion // Pseudocódigo del hilo.
	BloqueadoPor int           // Referencia al hilo que bloqueó este hilo.
}

// iota nos permite asignar automáticamente un valor incremental desde la primer variable a la última.
const (
	New     Estado = iota // Estado "New" (Valor 0) - Proceso o hilo recién creado
	Ready                 // Estado "Ready" (Valor 1) - Proceso o hilo listo para ejecución
	Exec                  // Estado "Exec" (Valor 2) - Proceso o hilo en ejecución
	Blocked               // Estado "Blocked" (Valor 3) - Proceso o hilo bloqueado, esperando recurso o I/O
	Exit                  // Estado "Exit" (Valor 4) - Proceso o hilo terminado
)

// Método String() para el tipo Estado (Esto es para los logs)
func (estado Estado) String() string {
	switch estado {
	case New:
		return "New"
	case Ready:
		return "Ready"
	case Exec:
		return "Exec"
	case Blocked:
		return "Blocked"
	case Exit:
		return "Exit"
	default:
		return "Desconocido"
	}
}

// Tipos de datos para las colas de procesos e hilos.
type ColaProcesos struct {
	Procesos []*PCB
	Nombre   string
}

type ColaHilos struct {
	Hilos  []*TCB
	Nombre string
}

type ColaHilosMN struct {
	Prioridad int
	Hilos  []*TCB
	Nombre string
}

// Lista de procesos general La vamos a usar para almacenar todos los procesos que hay en curso en el sistema.
var ProcesosDelSistema = ColaProcesos{
	Procesos: []*PCB{},
	Nombre: "Cola-Procesos-Del-Sistema",
}

////////////////////////////////////////////////////
///////////////////////MUTEX////////////////////////
////////////////////////////////////////////////////

//Estructura para manejar los mutex. Contiene información sobre su estado
type Mutex struct {
	ID             int
	HiloOwner      *TCB          // Hilo que posee el mutex.
	Bloqueados     []*TCB        // Cola de hilos bloqueados.
	NombreRecurso  string        // Nombre del recurso a proteger.
 }

 var Lista_mutex_global = make(map[int]*Mutex) // Tabla global de mutex para consultarlos y gestionarlos.
 // [1:0xc0000ae080 2:0xc0000ae0a0] ejemplo de un print de lista_mutex_global.

////////////////////////////////////////////////////
/////////////////COLAS DE PROCESOS//////////////////
////////////////////////////////////////////////////

// Cola de procesos en estado "New".
var ColaNewProcesos = ColaProcesos{
	Procesos: []*PCB{},
	Nombre:   "New-Procesos",
}

// Cola de procesos en estado "Ready".
var ColaReadyProcesos = ColaProcesos{
	Procesos: []*PCB{},
	Nombre:   "Ready-Procesos",
}

// Cola de procesos en estado "Running"
var ColaExecProcesos = ColaProcesos{
	Procesos: []*PCB{},
	Nombre:   "Exec-Procesos",
}

// Cola de procesos en estado "Blocked"
var ColaBlockedProcesos = ColaProcesos{
	Procesos: []*PCB{},
	Nombre:   "Blocked-Procesos",
}

// Método String() para el tipo ColaHilos (Lo usamos para los logs).
func (proceso ColaProcesos) String() string {
	switch proceso.Nombre {
	case "New-Procesos":
		return "Cola New Procesos"
	case "Ready-Procesos":
		return "Cola Ready Procesos"
	case "Exec-Procesos":
		return "Cola Exec Procesos"
	case "Blocked-Procesos":
		return "Cola Blocked Procesos"
	default:
		return "Cola Procesos Desconocida"
	}
}
type TIDyPIDaEjecutar struct{
    TID int
    PID int
}

////////////////////////////////////////////////////
///////////////////COLAS DE HILOS///////////////////
////////////////////////////////////////////////////

// Hilos del sistema. La vamos a usar para almacenar todos los hilos que hay en curso en el sistema.
var HilosDelSistema = ColaHilos{
	Hilos:  []*TCB{},
	Nombre: "HilosDelSistema",
}

// Cola de Hilos en Ready para ejecutar en FIFO.
var ColaReadyFIFO = ColaHilos{
	Hilos:  []*TCB{},
	Nombre: "ReadyFIFO-Hilos",
}

// Se define una lista de niveles de prioridad.
var ColaReadyPrioridadHilos = ColaHilos{
	Hilos:  []*TCB{},
	Nombre: "ReadyPrioridad-Hilos",
}

var ColasMultinivel []ColaHilosMN

var ColaBlockedHilos = ColaHilos{
	Hilos:  []*TCB{},
	Nombre: "Blocked-Hilos",
}

var ColaExecHilos = ColaHilos{
	Hilos:  []*TCB{},
	Nombre: "Exec-Hilos",
}

// Método String() para el tipo ColaHilos (Lo usamos para los logs).
func (cola ColaHilos) String() string {
	switch cola.Nombre {
	case "ReadyFIFO-Hilos":
		return "Cola Hilos Ready FIFO"
	case "ReadyPrioridad-Hilos":
		return "Cola Hilos Ready Prioridad"
	case "Blocked-Hilos":
		return "Cola Hilos Blockeados"
	case "Exec-Hilos":
		return "Cola Hilos en Ejecución"
	default:
		return "Cola Hilos Desconocida"
	}
}

/////////////////////////////////////
////////// ENTRADA / SALIDA /////////
/////////////////////////////////////

type DispositivoIO struct {
    ColaES []*TCB  // Cola FIFO de hilos esperando por E/S.
	TiempoES map[*TCB]int // Mapa que asocia cada hilo con su tiempo de E/S.
    EnUso  bool    // Indica si el dispositivo está en uso o no.
}

var DispositivoEntradaSalida = DispositivoIO{
    ColaES: []*TCB{},
	TiempoES: make(map[*TCB]int), // Inicializa el mapa de tiempos.
    EnUso:  false,
}

//Estructura para enviar proceso a memoria
type ProcesoMemoria struct{
	PID int
	Tamanio int
}

//Estructura para enviar hilo a memoria
type HiloMemoria struct{
	TID int	`json:"TID"`
	PID int	`json:"PID"`
	Codigo       []Instruccion `json:"Codigo"`// Pseudocódigo del hilo.
}


type InterrupcionDeHilo struct{
	TID int
	PID int
	Motivo string
}


var DesalojoRealizado bool = false

var PlanificadorPausado bool = false

type IOparaHacer struct{
    TID int
    PID int
    TiempoEnMilisegundos int
}

var HiloSolicitoDUMP bool = false