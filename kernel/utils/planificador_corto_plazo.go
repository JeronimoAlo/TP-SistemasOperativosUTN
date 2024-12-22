package utils

import (
	"encoding/json"
	"kernel/globals"
	"log/slog"
	"net/http"
	"strconv"
	"utils_general"
	"time"
)

func BloquearHilo(hilo *globals.TCB, hiloQueBloqueaTID int, motivoBloqueo string) {
	hilo.BloqueadoPor = hiloQueBloqueaTID // Asocia el hilo que lo bloquea.

	ptrProceso := EncontrarPCBPorPID(hilo.PID) // Buscamos el proceso del hilo a bloquear.
	
	slog.Info("## (<" + strconv.Itoa(ptrProceso.PID) + ">:<" + strconv.Itoa(hilo.TID) + ">) - Bloqueado por: <" + motivoBloqueo + ">.")

	EncolarHilo(hilo, &globals.ColaBlockedHilos, globals.Blocked)             // Agrega el hilo a la cola de hilos bloqueados y cambia su estado.
}

func DesbloquearHilo(hilo *globals.TCB) {
	DesencolarHilo(hilo) // Remover el hilo de la cola de bloqueados.

	ptrProceso := EncontrarPCBPorPID(hilo.PID) // Buscar el proceso correspondiente al hilo.

	// Cambiar el estado del hilo a READY.
	hilo.BloqueadoPor = -1    // Reseteamos el TID del hilo que lo bloqueaba.

	EncolarHiloEnReady(hilo) // Encolamos el hilo en ready.
	
	if ProcesoTieneTodosHilosReady(ptrProceso) && ptrProceso.Estado == globals.Blocked {
		DesencolarProceso(ptrProceso) // Removemos el proceso de la cola de bloqueados.
		EncolarProceso(ptrProceso, &globals.ColaReadyProcesos, globals.Ready)
	}
	if len(globals.ColaExecHilos.Hilos) == 0 {
		Replanificar()
	}
	
	slog.Info("## (<" + strconv.Itoa(ptrProceso.PID) + ">:<" + strconv.Itoa(hilo.TID) + ">) - Desbloqueado.")
}

// Función para desalojo de hilos (Correr cuando llega un nuevo hilo)
func DesalojarHiloSiNecesario(tcbNuevo *globals.TCB) {
	hiloExec := globals.ColaExecHilos.Hilos[0]

	if tcbNuevo.Prioridad < hiloExec.Prioridad { // Si el nuevo hilo tiene mayor prioridad que el que está ejecutando.
		EnviarInterrupcionACPU(hiloExec, "Desalojado") // Avisamos a la CPU que debe desalojar el hilo en EXEC.
	}
}

func EnviarInterrupcionACPU(hilo *globals.TCB, motivo string) bool {
	slog.Info("Enviando interrupción a la CPU para desalojar el hilo " + strconv.Itoa(hilo.TID) + "...")

	var interrupcion globals.InterrupcionDeHilo

	interrupcion.TID = hilo.TID
	interrupcion.PID = hilo.PID
	interrupcion.Motivo = motivo

	request, err1:= json.Marshal(interrupcion)
	if err1 != nil{
		slog.Error("Error al codificar el request de interrupcion: " + err1.Error())
	}
	respuestaJson, err := utils_general.EnviarMensaje(globals.Config.Ip_cpu, globals.Config.Port_cpu, request, "Kernel", "/interrupciones")

	if err!= nil{
		slog.Error("Error al enviar el request de interrupcion: " + err.Error())
	}

	var respuesta utils_general.StatusRespuesta
	err2 := json.Unmarshal(respuestaJson, &respuesta)
	if err2!= nil{
		slog.Error("Error al decodificar la respuesta de interrupcion: " + err.Error())
	}

	//Si la CPU recibio correctamtne la solicitud de interrupcion devuelve True. NO significa que ya se haya hecho la interrupcion, solo que la CPU lo recibio y la guardo
	return respuesta.Status == "OK"
}

func DesalojarHiloExec() {
	hiloExec := globals.ColaExecHilos.Hilos[0]

	EnviarInterrupcionACPU(hiloExec, "Quantum Expirado")
}

// Algoritmos del planificador de corto plazo para Hilos. Esta función se encarga de mandar hilos a ejecutar.
func ProcesarColaReadyHilos() {
	if !globals.PlanificadorPausado{

	switch globals.Config.Sheduler_algorithm {
		case "FIFO": // Ejecucion de Hilos en Ready por Metodo FIFO.
			{
				if len(globals.ColaReadyFIFO.Hilos) > 0 { // Si hay hilos pendientes en ejecución en la cola FIFO.
					hilo := globals.ColaReadyFIFO.Hilos[0]                  // Tomamos el primero.
					DesencolarHilo(hilo)                                    // Sacamos el hilo de la cola donde esté.
					EncolarHilo(hilo, &globals.ColaExecHilos, globals.Exec) // Lo pasamos a exec.
					Ejecucion(hilo)
				} else {
					slog.Info("No hay procesos en la cola de ready FIFO actualmente")
				}
			}
		case "PRIORIDADES": // Ejecucion de Hilos en Ready por Prioridad.
			{
				hiloAEjecutar := SeleccionarHiloPorPrioridad()
	
				if hiloAEjecutar != nil {
					EncolarHilo(hiloAEjecutar, &globals.ColaExecHilos, globals.Exec)
					Ejecucion(hiloAEjecutar)
				} else {
					slog.Info("No hay procesos en la cola de ready PRIORIDADES actualmente")
				}
			}
		case "CMN": // Ejecucion de Hilos en Ready por Colas Multinivel
			{
				ProcesarColasMultinivel()
			}
		}
	}else{
		slog.Debug("El planificador de corto plazo esta Pausado")
	}

}

// func ChequeoHayHiloEnCPU(hilo *globals.TCB){
// 	if globals_general.HayHiloEjecutandoEnCPU {
// 		return
// 	}

// 	Ejecucion(hilo)
// }

func Ejecucion(hilo *globals.TCB){ // Función que manda a CPU el nuevo Hilo y Proceso a ejecutar
	var tIDyPIDaEjecutar globals.TIDyPIDaEjecutar
 
	tIDyPIDaEjecutar.PID = hilo.PID
	tIDyPIDaEjecutar.TID = hilo.TID
	
	request, err1:= json.Marshal(tIDyPIDaEjecutar)

	if err1!=nil{
		slog.Error("Error codificando a json.")
	}

	respuestaJson, err := utils_general.EnviarMensaje(globals.Config.Ip_cpu, globals.Config.Port_cpu, request, "Kernel", "/ejecucionTIDyPID")

	if err!= nil{
		slog.Error("Error enviando el mensaje a la CPU para ejecutar un hilo.")
		return
	}

	var respuesta utils_general.StatusRespuesta
	
	if err4 := json.Unmarshal(respuestaJson, &respuesta); err4 != nil {
		slog.Warn("Error al decodificar la respuesta JSON desde CPU en Ejecucion:", "error", err4.Error(), "respuesta", string(respuestaJson))
		return
	}
}

func ProcesarColasMultinivel() {
	quantum := globals.Config.Quantum 					// Quantum del sistema.
	numNivelesPrioridad := len(globals.ColasMultinivel) // Número de colas (niveles de prioridad)

	if len(globals.ColaExecHilos.Hilos) != 0 {
		return
	} else {
		for prioridad := 0; prioridad < numNivelesPrioridad; prioridad++ {
			cola := globals.ColasMultinivel[prioridad]
	
			// Si hay hilos en la cola actual.
			if len(cola.Hilos) > 0 {
				hilo := cola.Hilos[0] // Tomamos el primer hilo (Round Robin)
				DesencolarHilo(hilo)  // Sacarlo de la cola actual
	
				// Ejecutar el hilo con el quantum correspondiente
				//go func(hilo *globals.TCB, quantum int) {
				EncolarHilo(hilo, &globals.ColaExecHilos, globals.Exec)
				go EjecutarConQuantum(hilo, quantum)
				//}(hilo, quantum)
	
				return // Procesamos un hilo por nivel de prioridad en cada ciclo
			}
		}
	}
}

func EjecutarConQuantum(hilo *globals.TCB, quantum int) {
	Ejecucion(hilo) // Mandamos a ejecutar el hilo a la CPU.

	time.Sleep(time.Duration(quantum) * time.Millisecond)

	if(len(globals.ColaExecHilos.Hilos) != 0){
		if (hilo.TID == globals.ColaExecHilos.Hilos[0].TID && hilo.PID == globals.ColaExecHilos.Hilos[0].PID) { // Si el hilo no cambió durante el tiempo de quantum definido, 
																											// se debe desalojar con motivo "Quantum Finalizado"
			// Desalojamos el hilo después del quantum
			DesalojarHiloExec()
			
			slog.Info("Hilo <" + strconv.Itoa(hilo.PID) + ":" + strconv.Itoa(hilo.TID) + "> desalojado por fin de quantum")
		} 
	}
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////////////////////////////EJECUCIÓN DE HILOS////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////////////

func RecibirFinEjecucion(writer http.ResponseWriter, requestCliente *http.Request) {
	var request globals.InterrupcionDeHilo

	err := json.NewDecoder(requestCliente.Body).Decode(&request) // Decodificamos la respuesta del cliente dentro de la variable request.

	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest) // Error el decodificar el request del cliente.
		slog.Warn("Error al decodificar el request del cliente en la recepción del mensaje.")
		return
	}

	hilo := EncontrarTCBPorTID(request.TID, request.PID)
	
	slog.Info("Hilo " + strconv.Itoa(hilo.TID) + " finalizó su ejecución. Motivo: " + request.Motivo + ".")

	switch request.Motivo {
	case "Finalizado":
		FinalizacionDeHilo(hilo, "FIN HILO") // Fin de la ejecución, finalizamos el hilo.
	case "Desalojado":
		DesencolarHiloExec(hilo)

		EncolarHiloEnReady(hilo)

		if(globals.PlanificadorPausado) { //Comprobamos si es que el planificador esta pausado por necesidad de compactar, es necesario chequearlo al finalizar un hilo
			//Si el planificador esta pausado, entonces hay que solicitar compactacion
			if SolicitarCompactacion() {
				slog.Info("Compactacion finalizada. Reintentando chequear espacio de memoria luego de la compactacion para el proceso en cuestion...")
				VerificarColaNew()
				return
			} else {
				slog.Error("Error en la compactacion al solicitarla desde el Kernel para el proceso")
				return
			}
		}

	case "Quantum Expirado":
		if !globals.HiloSolicitoDUMP {
			DesencolarHiloExec(hilo)
			// Sacamos de ejecución el hilo que estaba (Debería ser el que nos está pasando por parámetro en la función).
			EncolarHiloEnReady(hilo) // Vuelve a Ready si su quantum se agotó.

			if(globals.PlanificadorPausado) { //Comprobamos si es que el planificador esta pausado por necesidad de compactar, es necesario chequearlo al finalizar un hilo
				//Si el planificador esta pausado, entonces hay que solicitar compactacion
				if SolicitarCompactacion() {
					slog.Info("Compactacion finalizada. Reintentando chequear espacio de memoria luego de la compactacion para el proceso en cuestion...")
					VerificarColaNew()
					return
				} else {
					slog.Error("Error en la compactacion al solicitarla desde el Kernel para el proceso")
					return
				}
			}
		} else {
			globals.HiloSolicitoDUMP = false
			globals.DesalojoRealizado = true
			
			// Replanificar si es necesario.
			if NecesitaReplanificar(request.Motivo) {
				Replanificar()
			}

			utils_general.EnviarStatusOK(writer)
			
			dumpExitoso := SolicitarDumpMemoria(hilo.PID, hilo.TID) // Devuelve un booleano si sale todo OK.

			if dumpExitoso {
				DesbloquearHilo(hilo)
				slog.Info("## (<" + strconv.Itoa(hilo.PID) + ">:<" + strconv.Itoa(hilo.TID) + ">) - DUMP_MEMORY finalizado exitosamente.")
			} else {
				proceso := EncontrarPCBPorPID(hilo.PID)
				FinalizarProceso(proceso)
				slog.Error("## (<" + strconv.Itoa(proceso.PID) + ">:<" + strconv.Itoa(hilo.TID) + ">) - Error en DUMP_MEMORY. Proceso enviado a EXIT.")
			}

			return
		}		
	case "IO":
		break // Hace un break porque la funcion de IO ya lo enconla en ready cuando termina la entrada y salida
	case "DUMP_MEMORY":
		if globals.HiloSolicitoDUMP {
			globals.DesalojoRealizado = true

			// Replanificar si es necesario.
			if NecesitaReplanificar(request.Motivo) {
				Replanificar()
			}

			utils_general.EnviarStatusOK(writer)

			dumpExitoso := SolicitarDumpMemoria(hilo.PID, hilo.TID) // Devuelve un booleano si sale todo OK.

			if dumpExitoso {
				DesbloquearHilo(hilo)
				slog.Info("## (<" + strconv.Itoa(hilo.PID) + ">:<" + strconv.Itoa(hilo.TID) + ">) - DUMP_MEMORY finalizado exitosamente.")
			} else {
				slog.Debug("El dump fallo")
				proceso := EncontrarPCBPorPID(hilo.PID)
				FinalizarProceso(proceso)
				DesbloquearHilo(hilo)
				slog.Error("## (<" + strconv.Itoa(proceso.PID) + ">:<" + strconv.Itoa(hilo.TID) + ">) - Error en DUMP_MEMORY. Proceso se envia a EXIT.")
			}

			return
		}
	}	

	globals.DesalojoRealizado = true // Variable que se utiliza en el kernel para confirmar si ya se desalojo el proceso, se usa en la solicitud de compactacion
	
	// Replanificar si es necesario.
	if NecesitaReplanificar(request.Motivo) {
		Replanificar()
	}

	utils_general.EnviarStatusOK(writer)
}

func NecesitaReplanificar(motivo string) bool { // Dependiendo el motivo, analizamos si es necesario replanificar o no.
	return motivo == "Desalojado" || motivo == "Quantum Expirado" || motivo == "IO" || motivo == "DUMP_MEMORY"
}

func Replanificar() {
	slog.Info("Replanificando...")

	var hilo *globals.TCB

	if len(globals.ColaExecHilos.Hilos) > 0 {
		hilo = globals.ColaExecHilos.Hilos[0]
	}else{
		hilo = SeleccionarSiguienteHilo() // Función que devuelve el siguiente hilo a ejecutar en el sistema.
	}
	
	if hilo != nil {
		EncolarHilo(hilo, &globals.ColaExecHilos, globals.Exec)
		if globals.Config.Sheduler_algorithm == "CMN" {
			EjecutarConQuantum(hilo, globals.Config.Quantum)
		} else {
			Ejecucion(hilo)
		}
	} else {
		slog.Info("No hay hilos en Ready, CPU en espera.")
	}
}

func SeleccionarSiguienteHilo() *globals.TCB {
	switch globals.Config.Sheduler_algorithm {
	case "FIFO":
		if len(globals.ColaReadyFIFO.Hilos) > 0 {
			hilo := globals.ColaReadyFIFO.Hilos[0] // Tomamos el primero en la cola.
			DesencolarHilo(hilo)                   // Lo sacamos de la cola READY.
			return hilo
		} else {
			slog.Info("No hay hilos en la cola de READY FIFO actualmente.")
			return nil
		}

	case "PRIORIDADES":
		hilo := SeleccionarHiloPorPrioridad()
		if(hilo != nil){
			return hilo
		}else{
			return nil
		}

	case "CMN":
		hilo := SeleccionarHiloColaMultinivel()
		if(hilo != nil){
			DesencolarHilo(hilo)
			return hilo
		}
		return hilo

	default:
		slog.Error("Algoritmo de planificación no soportado: " + globals.Config.Sheduler_algorithm)
		return nil
	}
}

func SeleccionarHiloPorPrioridad() *globals.TCB {
	if len(globals.ColaReadyPrioridadHilos.Hilos) > 0 {
		
		// Variable para almacenar el hilo con la mayor prioridad
		hiloSeleccionado := globals.ColaReadyPrioridadHilos.Hilos[0]
		
		// Buscar el hilo con la mayor prioridad
		for _, hilo := range globals.ColaReadyPrioridadHilos.Hilos {
			if hilo.Prioridad < hiloSeleccionado.Prioridad { // Menor valor de prioridad es mejor
				hiloSeleccionado = hilo
			}
		}
		
		DesencolarHilo(hiloSeleccionado)
		return hiloSeleccionado
	}else{
		slog.Info("No hay hilos en la cola de Prioridades actualmente.")
		return nil
	} 
}

func SeleccionarHiloColaMultinivel() *globals.TCB {
	for _, colaMultinivel := range globals.ColasMultinivel { // Si ya existe una cola cuya prioridad corresponda a la del hilo, simplemente se encola el hilo en dicha cola
		if len(colaMultinivel.Hilos) != 0 {			
			return colaMultinivel.Hilos[0]
		}
	}

	slog.Info("No hay hilos en ninguna cola de multinivel")
	return nil
}

func DesencolarHiloExec(hilo *globals.TCB) {
	hiloEncontrado := 0
	colaARetirar := &globals.ColaExecHilos

	for i, tcb := range colaARetirar.Hilos { // Busca el hilo a eliminar en la lista.
		if tcb.TID == hilo.TID {
			// Eliminar el Hilo usando slicing.
			colaARetirar.Hilos = append(colaARetirar.Hilos[:i], colaARetirar.Hilos[i+1:]...) // Cuando lo encuentra lo borra de la lista con Slice.
			hiloEncontrado = 1 // Encontramos y eliminamos el hilo.
	
			slog.Info("Hilo con PID <" + strconv.Itoa(hilo.PID) + "> y TID <" + strconv.Itoa(hilo.TID) + "> eliminado de la cola: " + colaARetirar.String())
			break
		}
	}

	if hiloEncontrado == 0 {
		slog.Debug("No se encontro en ninguna cola el hilo con TID: " + strconv.Itoa(hilo.TID))
	}
}