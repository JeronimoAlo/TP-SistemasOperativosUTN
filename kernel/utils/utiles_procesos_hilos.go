package utils

import (
	"kernel/globals"
	"log/slog"
	"strconv"
	"fmt"
	"sort"
)

// Función para cambiar el estado de un proceso.
func CambiarEstadoProceso(proceso *globals.PCB, nuevoEstado globals.Estado) {
	proceso.Estado = nuevoEstado // Cambiamos el estado del hilo.
	slog.Info("Estado del proceso " + strconv.Itoa(proceso.PID) + " cambio a " + nuevoEstado.String())
}

// Función para cambiar el estado de un hilo.
func CambiarEstadoHilo(hilo *globals.TCB, nuevoEstado globals.Estado) {
	hilo.Estado = nuevoEstado // Cambiamos el estado del hilo.
	slog.Info("Estado del hilo con PID <" + strconv.Itoa(hilo.PID) + "> y TID <" + strconv.Itoa(hilo.TID) + "> cambio a " + nuevoEstado.String())
}

// Función para añadir procesos en una las colas (NEW/READY/EXEC/BLOQUEADO).
func EncolarProceso(proceso *globals.PCB, colaEstado *globals.ColaProcesos, nuevoEstado globals.Estado) { // Recibe el proceso a encolar, la cola a la que se debe agregar y el nuevo estado al que queremos cambiar.
	switch nuevoEstado { // Cambiamos el estado del proceso en cuestión dependiento el valor de nuevoEstado.
	case globals.New:
		proceso.Estado = globals.New
		slog.Info("## (<" + strconv.Itoa(proceso.PID) + ">:0) Se crea el proceso - Estado: NEW")
	case globals.Ready:
		proceso.Estado = globals.Ready
		slog.Debug("Se encoló el proceso con PID " + strconv.Itoa(proceso.PID) + " en estado: Ready")
	case globals.Exec:
		proceso.Estado = globals.Exec
		slog.Debug("Se encoló el proceso con PID " + strconv.Itoa(proceso.PID) + " en estado: Exec")
	case globals.Blocked:
		proceso.Estado = globals.Blocked
		slog.Debug("Se encoló el proceso con PID " + strconv.Itoa(proceso.PID) + " en estado: Blocked")
	case globals.Exit:
		proceso.Estado = globals.Exit
		slog.Debug("Se encoló el proceso con PID " + strconv.Itoa(proceso.PID) + " en estado: Exit")
	default:
		slog.Warn("Estado desconocido para el proceso con PID " + strconv.Itoa(proceso.PID))
		return
	}

	colaEstado.Procesos = append(colaEstado.Procesos, proceso) // Encolamos el proceso a la cola indicada.
}

// Función para añadir hilos en una de las colas (Prioridad0, Prioridad1, Prioridad2, Blocked).
func EncolarHilo(hilo *globals.TCB, colaEstado *globals.ColaHilos, nuevoEstado globals.Estado) { // Recibe el hilo a encolar, la cola a la que se debe agregar y el nuevo estado al que queremos cambiar.
	switch nuevoEstado { // Cambiamos el estado del proceso en cuestión dependiento el valor de nuevoEstado.
	case globals.New:
		slog.Debug("Los hilos no se encolan en cola de New")
		return
	case globals.Ready:
		slog.Debug("Los hilos se deben pasar al estado Ready mediante la función 'EncolarHiloEnReady'")
		return
	case globals.Exec:
		hilo.Estado = globals.Exec
		slog.Info("Se encoló el hilo con PID " + strconv.Itoa(hilo.PID) + " TID " + strconv.Itoa(hilo.TID) + " en estado: Exec")
	case globals.Blocked:
		hilo.Estado = globals.Blocked
		slog.Info("Se encoló el hilo con PID " + strconv.Itoa(hilo.PID) + " TID " + strconv.Itoa(hilo.TID) + " en estado: Blocked")
	case globals.Exit:
		hilo.Estado = globals.Exit
		slog.Info("Se encoló el hilo con PID " + strconv.Itoa(hilo.PID) + " TID " + strconv.Itoa(hilo.TID) + " en estado: Exit")
	default:
		slog.Warn("Estado desconocido para el hilo con TID " + strconv.Itoa(hilo.TID))
		return
	}

	colaEstado.Hilos = append(colaEstado.Hilos, hilo) // Encolamos el hilo a la cola indicada.
}

// Función para desencolar un hilo en base a una cola dada. No cambia el estado del hilo, solo lo elimina de la cola en la que se encuentra.
func DesencolarHilo(hilo *globals.TCB) {
	hiloEncontrado := 0
	var colaARetirar *globals.ColaHilos
	var colaARetirarCMN *globals.ColaHilosMN
	var indiceColaARetirarCMN int
	
	// Identificar la cola correcta según el estado y prioridad del hilo.
	switch hilo.Estado {
	case globals.Ready:
		if globals.Config.Sheduler_algorithm == "PRIORIDADES" {
			colaARetirar = &globals.ColaReadyPrioridadHilos
		} else if (globals.Config.Sheduler_algorithm == "CMN") {
			for i, colaMultinivel := range globals.ColasMultinivel { // Si ya existe una cola cuya prioridad corresponda a la del hilo, simplemente se encola el hilo en dicha cola
				if (colaMultinivel.Prioridad == hilo.Prioridad){
					colaARetirarCMN = &colaMultinivel
					indiceColaARetirarCMN = i
				}
				
			}
		} else { // Sino, vamos por FIFO.
			colaARetirar = &globals.ColaReadyFIFO
		}
	case globals.Blocked:
		colaARetirar = &globals.ColaBlockedHilos
	case globals.Exec:
		colaARetirar = &globals.ColaExecHilos
	default:
		slog.Error("Estado del hilo desconocido: " + hilo.Estado.String())
		return
	}

	if globals.Config.Sheduler_algorithm != "CMN" {
		for i, tcb := range colaARetirar.Hilos { // Busca el hilo a eliminar en la lista.
			if tcb.TID == hilo.TID && tcb.PID == hilo.PID {
				
				if len(colaARetirar.Hilos) == 1{
					colaARetirar.Hilos = []*globals.TCB{}
				}else{
					// Eliminar el Hilo usando slicing.
					colaARetirar.Hilos = append(colaARetirar.Hilos[:i], colaARetirar.Hilos[i+1:]...) // Cuando lo encuentra lo borra de la lista con Slice.
				}
				hiloEncontrado = 1 // Encontramos y eliminamos el hilo.
	
				slog.Info("Hilo con PID <" + strconv.Itoa(hilo.PID) + "> y TID <" + strconv.Itoa(hilo.TID) + "> eliminado de la cola: " + colaARetirar.String())
				break
			}
		}
	} else {
		if colaARetirarCMN != nil {
			for i, tcb := range colaARetirarCMN.Hilos {				
				if tcb.TID == hilo.TID && tcb.PID == hilo.PID {
					if len(colaARetirarCMN.Hilos) == 1 {
						colaARetirarCMN.Hilos = []*globals.TCB{}
					}else{
							
						// Eliminar el Hilo usando slicing.
						colaARetirarCMN.Hilos = append(colaARetirarCMN.Hilos[:i], colaARetirarCMN.Hilos[i+1:]...) // Cuando lo encuentra lo borra de la lista con Slice.

					}
					hiloEncontrado = 1 // Encontramos y eliminamos el hilo.

					globals.ColasMultinivel[indiceColaARetirarCMN] = *colaARetirarCMN
					slog.Debug("La cola multinivel " + globals.ColasMultinivel[indiceColaARetirarCMN].Nombre + " tiene tamaño " + strconv.Itoa(len(globals.ColasMultinivel[indiceColaARetirarCMN].Hilos)))
			
					slog.Info("Hilo con PID <" + strconv.Itoa(hilo.PID) + "> y TID <" + strconv.Itoa(hilo.TID) + "> eliminado de la cola: " + colaARetirarCMN.Nombre)
					break
				}
			}
		} else {
			for i, tcb := range colaARetirar.Hilos { // Busca el hilo a eliminar en la lista.
				if tcb.TID == hilo.TID && tcb.PID == hilo.PID {
					// Eliminar el Hilo usando slicing.
					if len(colaARetirar.Hilos) == 1{
						colaARetirar.Hilos = []*globals.TCB{}
					} else{
						colaARetirar.Hilos = append(colaARetirar.Hilos[:i], colaARetirar.Hilos[i+1:]...) // Cuando lo encuentra lo borra de la lista con Slice.
					}
					
					hiloEncontrado = 1 // Encontramos y eliminamos el hilo.
					
					slog.Info("Hilo con PID <" + strconv.Itoa(hilo.PID) + "> y TID <" + strconv.Itoa(hilo.TID) + "> eliminado de la cola: " + colaARetirar.String())
					break
				}
			}
		}
	}
	
	if hiloEncontrado == 0 {
		slog.Debug("No se encontro en ninguna cola el hilo con TID: " + strconv.Itoa(hilo.TID))
	}
}

// Función para desencolar un proceso en base a una cola dada.
func DesencolarProceso(proceso *globals.PCB) {
	procesoEncontrado := 0 // Contador para avisar que encontramos el proceso que estabamos buscando.
	var colaARetirar *globals.ColaProcesos

	// Determina de qué cola se va a retirar el proceso según su estado actual.
	switch proceso.Estado {
	case globals.New:
		colaARetirar = &globals.ColaNewProcesos
	case globals.Ready:
		colaARetirar = &globals.ColaReadyProcesos
	case globals.Exec:
		colaARetirar = &globals.ColaExecProcesos
	case globals.Blocked:
		colaARetirar = &globals.ColaBlockedProcesos
	default:
		slog.Error("Estado no válido para el proceso PID: %d" + strconv.Itoa(proceso.PID))
		return
	}

	for i, pcb := range colaARetirar.Procesos { // Busca el proceso a eliminar en la cola.
		if pcb.PID == proceso.PID {
			// Eliminar el proceso usando slicing.
			colaARetirar.Procesos = append(colaARetirar.Procesos[:i], colaARetirar.Procesos[i+1:]...) // Cuando lo encuentra lo borra de la lista con Slice.
			procesoEncontrado = 1

			slog.Info("Proceso con PID: " + strconv.Itoa(proceso.PID) + " eliminado de la cola " + colaARetirar.String())
			break
		}
	}

	if procesoEncontrado == 0{
		slog.Debug("No se encontro en ninguna cola el proceso con PID: " + strconv.Itoa(proceso.PID))
	}
}

// Función que devuelve el tcb de un hilo a partir de su tid y un proceso.
func HiloAPartirDeID(tid int, proceso *globals.PCB) *globals.TCB {
	for _, tcb := range globals.HilosDelSistema.Hilos { // Busca el hilo en la lista de todos los del sistema.
		if (tcb.TID == tid) && (tcb.PID == proceso.PID) {
			return tcb // tcb es un puntero al hilo, lo devuelve.
		}
	}

	slog.Debug("No se encontró ningún hilo con el TID: " + strconv.Itoa(tid))
    return nil
}

// Función auxiliar para verificar si todos los hilos del proceso están en READY.
func ProcesoTieneTodosHilosReady(proceso *globals.PCB) bool {
	for _, tid := range proceso.TIDs {
		tcb := EncontrarTCBPorTID(tid, proceso.PID)

		if tcb.Estado != globals.Ready {
			return false // Si encontramos un hilo que no está listo, retornamos false.
		}
	}
	return true // Todos los hilos están listos.
}

func EncolarHiloEnReadyPorPrioridad(hilo *globals.TCB) {
	// Agrega el hilo a la cola general.
	globals.ColaReadyPrioridadHilos.Hilos = append(globals.ColaReadyPrioridadHilos.Hilos, hilo)

	slog.Debug("Hilo " + strconv.Itoa(hilo.TID) + " del PID " + strconv.Itoa(hilo.PID) + " se encoló con prioridad " + strconv.Itoa(hilo.Prioridad))

	CambiarEstadoHilo(hilo, globals.Ready) // Cambiamos el estado del hilo a "Ready".
}

func EncolarHiloEnReadyPorCMN(hilo *globals.TCB) {
	var nuevaCola *globals.ColaHilosMN
	
	for i, colaMultinivel := range globals.ColasMultinivel { // Si ya existe una cola cuya prioridad corresponda a la del hilo, simplemente se encola el hilo en dicha cola
		if (colaMultinivel.Prioridad == hilo.Prioridad){
			colaMultinivel.Hilos = append(colaMultinivel.Hilos, hilo)

			globals.ColasMultinivel[i] = colaMultinivel
			
			slog.Info("Hilo con PID " + strconv.Itoa(hilo.PID) + ", TID " + strconv.Itoa(hilo.TID) + " y prioridad " + strconv.Itoa(hilo.Prioridad) + " agregado a cola existente de nombre " + colaMultinivel.Nombre)
			return
		}
	}
	
	nuevaCola = &globals.ColaHilosMN{ // Si no se encuentra una cola de la prioridad indicada por el hilo, se crea dicha cola dentro de las colas multinivel del sistema
		Prioridad: hilo.Prioridad,	  // y se encola el hilo en la nueva cola
		Hilos: []*globals.TCB{},
		Nombre: fmt.Sprintf("Cola Prioridad %d", hilo.Prioridad),
	}

	nuevaCola.Hilos = append(nuevaCola.Hilos, hilo)
    globals.ColasMultinivel= append(globals.ColasMultinivel, *nuevaCola)

	slog.Info("Hilo con PID " + strconv.Itoa(hilo.PID) + ", TID " + strconv.Itoa(hilo.TID) + " y prioridad " + strconv.Itoa(hilo.Prioridad) + " agregado a la nueva cola creada de nombre " + nuevaCola.Nombre)

	sort.Slice(globals.ColasMultinivel, func(i, j int) bool { // Ordenamos las colas por prioridad
		return globals.ColasMultinivel[i].Prioridad < globals.ColasMultinivel[j].Prioridad
	})

	slog.Debug("Colas multinivel ordenadas por prioridad.")
}

// Función para encolar hilos en ready dependiendo del algoritmo utilizado en el config.
func EncolarHiloEnReady(hilo *globals.TCB) {
	if hilo.BloqueadoPor != -1 { // Esto significa que el hilo sigue bloqueado.
		return
	}
	
	switch globals.Config.Sheduler_algorithm {
	case "FIFO":
		globals.ColaReadyFIFO.Hilos = append(globals.ColaReadyFIFO.Hilos, hilo)
		hilo.Estado = globals.Ready
		slog.Debug("Hilo " + strconv.Itoa(hilo.TID) + " del PID " + strconv.Itoa(hilo.PID) + " se encolo en Ready FIFO")
	case "PRIORIDADES":
		EncolarHiloEnReadyPorPrioridad(hilo)
		hilo.Estado = globals.Ready

		if len(globals.ColaExecHilos.Hilos) > 0 {
			DesalojarHiloSiNecesario(hilo)
		}
	case "CMN":
		EncolarHiloEnReadyPorCMN(hilo)
		
		hilo.Estado = globals.Ready

		if len(globals.ColaExecHilos.Hilos) > 0 {
			DesalojarHiloSiNecesario(hilo)
		}

		slog.Debug("Hilo " + strconv.Itoa(hilo.TID) + " del PID " + strconv.Itoa(hilo.PID) + " se encolo en Ready Cola Multinivel")
	default:
		slog.Error("No existe el algoritmo: " + globals.Config.Sheduler_algorithm + " para el planificador de corto plazo. Por favor revise el archivo de configuracion.")
	}
}

// Función para mapear los PID de una lista de PCB. Esto lo necesitamos para no repetir PIDs al crear un proceso nuevo.
func MapPID(pcbs []*globals.PCB) []int {
	var pids []int
	for _, pcb := range pcbs {
		pids = append(pids, pcb.PID)
	}
	return pids
}