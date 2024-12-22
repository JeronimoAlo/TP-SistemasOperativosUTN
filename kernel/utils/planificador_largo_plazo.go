package utils

import (
	"encoding/json"
	"kernel/globals"
	"log/slog"
	"strconv"
	"utils_general"
	"globals_general"
	"net/http")

// Creamos el proceso.
func CrearProceso(pidNuevo int, tamanio int) *globals.PCB {
	// Creamos el PCB
	pcb := &globals.PCB{
		PID:    pidNuevo,
		TIDs:   []int{0},
		Mutex:  []globals.Mutex{},
		Estado: globals.New,
		Tamanio: tamanio,
	}

	return pcb
}

// Creamos el hilo dentro del proceso asociado.
func CrearHilo(pidPadre int, tidNuevo int, prioridad int, instrucciones []globals.Instruccion) *globals.TCB {
	TCBnuevo := &globals.TCB{ // Devolvemos una estructura de TCB con el nuevo hilo creado.
		PID:       pidPadre,
		TID:       tidNuevo,
		Prioridad: prioridad,
		Estado:    globals.Ready,
		Codigo:    instrucciones, // Pseudocódigo del hilo.
		BloqueadoPor: -1,
	}

	globals.HilosDelSistema.Hilos = append(globals.HilosDelSistema.Hilos, TCBnuevo)

	return TCBnuevo
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////Creación de procesos/////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Chequeo de Estado de Cola NEW + Memoria para pasar a READY (Esto debería ejecutarse cuando creamos un proceso a partir de una syscall).
func VerificarColaNew() {
	if len(globals.ColaNewProcesos.Procesos) > 0  && !globals.PlanificadorPausado{ // Se agrega validacion para chequear si el kernel no pausó la ejecucion para poder compactar.
		//Hay procesos en cola NEW, se prueba el envio a memoria.
		proceso := globals.ColaNewProcesos.Procesos[0]

		var procesoMemoria globals.ProcesoMemoria // Declaramos variable la cual tiene un PID y su tamanio, para mandarla a la memoria en formato Json
		procesoMemoria.PID = proceso.PID
		procesoMemoria.Tamanio = proceso.Tamanio

		// Chequeamos si hay espacio en memoria o si es compactable.
		espacioDisponible, compactable := ChequearEspacioMemoria(procesoMemoria)

		switch {
		case espacioDisponible:
			// Hay espacio disponible, pasamos el proceso a READY.
			EncolarProceso(proceso, &globals.ColaReadyProcesos, globals.Ready)
			tcb := EncontrarTCBPorTID(0, procesoMemoria.PID)

			CreacionHiloAMemoria(tcb)

			// Elimino de la cola new el proceso en cuestión.
			globals.ColaNewProcesos.Procesos = globals.ColaNewProcesos.Procesos[1:]
			slog.Info("El proceso con PID " + strconv.Itoa(proceso.PID) + " fue removido de la cola NEW.")

			ProcesarColaReadyHilos()
		case compactable:
			// No hay espacio continuo, pero la memoria es compactable.
			slog.Info("Memoria compactable detectada para proceso PID: " + strconv.Itoa(proceso.PID))

			PausarPlanificacionCortoPlazo() //Setea una variable global de planificador en pausado, luego se utiliza para validar la necesidad de compactacion al finalizar un hilo
			// No se hace la compactacion aca. Se hace al finalizar el hilo en ejecucion.
			if len(globals.ColaExecHilos.Hilos) == 0 {
				if(globals.PlanificadorPausado) { 
					if SolicitarCompactacion() {
						slog.Info("Compactacion finalizada. Reintentando chequear espacio de memoria luego de la compactacion para el proceso en cuestion...")
						globals.PlanificadorPausado = false
						slog.Debug("Planificador a corto plazo despausado")

						VerificarColaNew()
					} else {
						slog.Error("Error en la compactacion al solicitarla desde el Kernel para el proceso")
					}
				}
			}
		default:
			// No hay espacio suficiente y no es compactable.
			slog.Warn("No hay espacio para el proceso PID: " + strconv.Itoa(proceso.PID) + ". Proceso permanece en NEW.")

			ProcesarColaReadyHilos()
		}
	} else { // En el caso de que no haya procesos para procesar.
		slog.Info("No hay procesos nuevos en la cola NEW para enviar a memoria.")

		if globals.Config.Sheduler_algorithm == "CMN" {
			hiloPotencial := SeleccionarHiloColaMultinivel()
			if hiloPotencial != nil {
				ProcesarColaReadyHilos()
			}
		} else if globals.Config.Sheduler_algorithm == "PRIORIDADES" {
			if len(globals.ColaReadyPrioridadHilos.Hilos) > 0 {
				// Variable para almacenar el hilo con la mayor prioridad
				hiloSeleccionado := globals.ColaReadyPrioridadHilos.Hilos[0]
				
				// Buscar el hilo con la mayor prioridad
				for _, hilo := range globals.ColaReadyPrioridadHilos.Hilos {
					if hilo.Prioridad < hiloSeleccionado.Prioridad { // Menor valor de prioridad es mejor
						hiloSeleccionado = hilo
					}
				}

				if hiloSeleccionado != nil {
					ProcesarColaReadyHilos()
				}
			}
		} else {
			if len(globals.ColaReadyFIFO.Hilos) > 0 {
				hilo := globals.ColaReadyFIFO.Hilos[0]

				if hilo != nil {
					ProcesarColaReadyHilos()
				}
			}
		}
	}
}

func PausarPlanificacionCortoPlazo(){
	slog.Debug("Se pausa el Planificador a Corto Plazo")
	globals.PlanificadorPausado = true
}

// Función que verifica si hay espacio en memoria. El primer bool que devuelve es para saber si hay espacio disponible y el segundo para saber si se puede compactar o no.
func ChequearEspacioMemoria(procesoMemoria globals.ProcesoMemoria) (bool, bool) {
	// Codificamos el proceso como JSON
	mensaje, err := json.Marshal(procesoMemoria)
	if err != nil {
		slog.Error("Error codificando mensaje.")
		return false, false
	}

	// Enviamos el proceso a la memoria, para que chequee si existe memoria disponible.
	respMemoriaJson, err := utils_general.EnviarMensaje(globals.Config.Ip_memory, globals.Config.Port_memory, mensaje, "Kernel", "/memoriaDisponible")
	if err != nil {
		slog.Debug("No hay espacio suficiente en memoria para almacenar al proceso: " + string(mensaje) + ".")
		return false, false
	}

	// Decodificamos el estado de la respuesta.
	var respuesta globals_general.CompactacionRespuesta
	err = json.Unmarshal(respMemoriaJson, &respuesta)
	if err != nil {
		slog.Error("Fallo al decodificar la respuesta de memoria para crear el proceso con PID: " + strconv.Itoa(procesoMemoria.PID))
		return false, false
	}

	// Verificamos el estado de la respuesta.
	if respuesta.Status == "OK" {
		slog.Debug("Memoria disponible confirmada para el proceso PID: " + strconv.Itoa(procesoMemoria.PID))
		return true, false
	} else if respuesta.Status == "FAILED" {
		slog.Warn("No hay memoria suficiente para el proceso PID: " + strconv.Itoa(procesoMemoria.PID))
		return false, respuesta.Compactable
	} else {
		slog.Error("Respuesta inesperada del módulo de memoria para crear el proceso PID: " + strconv.Itoa(procesoMemoria.PID))
		return false, false
	}
}

func SolicitarCompactacion() bool  {
	url := "http://" + globals.Config.Ip_memory + ":" + strconv.Itoa(globals.Config.Port_memory) + "/compactar"

	request, err := http.NewRequest("POST", url, nil)
	if err != nil {
		slog.Error("Error creando solicitud")
		return false
	}

	client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		slog.Error("Error enviando solicitud")
		return false
	}

	defer response.Body.Close()

	var result struct {
		Success bool `json:"success"`
	}

	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		slog.Error("error decodificando respuesta")
		return false
	}

	// Imprimimos el código de estado HTTP recibido
	slog.Debug("HTTP Status recibido:" + strconv.Itoa(response.StatusCode))
	if response.StatusCode == 200{
		globals.PlanificadorPausado = false

		slog.Debug("Planificador a corto plazo activado nuevamente.")
		return true
	}else{
		return false
	}
	
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////Finalización de procesos///////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Envia a la memoria el PID a ser eliminado de la memoria, libera el PCB del proceso.
func FinalizarProceso(proceso *globals.PCB) {
	if InformarFinalizacionProcesoAMemoria(proceso.PID) {
		EliminarHilosAsociados(proceso) // Función que elimina/finaliza todos los hilos asociados al proceso que está finalizando.

		DesencolarProceso(proceso)    // Libera el PCB del proceso, eliminandolo de cualquier cola donde esté.
		proceso.Estado = globals.Exit // Pasamos el proceso al estado final.

		for i, pcb := range globals.ProcesosDelSistema.Procesos { // Busca el proceso a eliminar en la lista.
			if pcb.PID == proceso.PID {
				// Si la lista tiene solo un elemento, vaciarla directamente.
				if len(globals.ProcesosDelSistema.Procesos) == 1 {
					globals.ProcesosDelSistema.Procesos = []*globals.PCB{}
				} else {
					// Eliminar el proceso usando slicing.
					globals.ProcesosDelSistema.Procesos = append(globals.ProcesosDelSistema.Procesos[:i], globals.ProcesosDelSistema.Procesos[i+1:]...)
				}
				break
			}
		}
		
		slog.Info("## Finaliza el proceso <" + strconv.Itoa(proceso.PID) + ">.")

		VerificarColaNew() // Intentar inicializar uno de los que estén esperando en estado NEW si los hubiere.
	} else {
		slog.Warn("No se puede finalizar el proceso con PID <" + strconv.Itoa(proceso.PID) + "> en memoria")
	}
}

func EliminarHilosAsociados(proceso *globals.PCB) {
	// Crear una copia de la lista de TIDs antes de la iteración.
	tidsCopia := append([]int(nil), proceso.TIDs...)

	for _, tidPosible := range tidsCopia { 	      // Itero sobre todos los TIDs.
		hilo := HiloAPartirDeID(tidPosible, proceso)  // Busca el hilo asociado al proceso.
		FinalizacionDeHilo(hilo, "FIN PROCESO PADRE") // Luego lo finaliza.
	}
}

func InformarFinalizacionProcesoAMemoria(pid int) bool {
	var procesoAfinalizar globals.ProcesoMemoria // Inicializamos el puntero.
	
	procesoAfinalizar.PID = pid
	procesoAfinalizar.Tamanio = 0 // Ignorar

	mensaje, err := json.Marshal(procesoAfinalizar)


	slog.Debug("El json a enviar a Memoria es: " + string(mensaje))

	if err != nil {
		slog.Error("Error al codificar el json de proceso a finalizar para enviar a memoria")
		return false
	}
	
	respuestaMemoriaJSON, err :=  utils_general.EnviarMensaje(globals.Config.Ip_memory, globals.Config.Port_memory, mensaje, "Kernel", "/finalizacionDeProceso")

	if err != nil {
		slog.Warn("Error al enviar el json de proceso a finalizar a memoria " + err.Error())
		return false
	}

	var respuesta utils_general.StatusRespuesta

	err = json.Unmarshal(respuestaMemoriaJSON, &respuesta)
	
	if err != nil {
		slog.Warn("Error al decodificar finalizar a memoria" + err.Error())
		return false
	}

	return respuesta.Status == "OK"
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////
/////////////////////////////////////////Creación de hilos//////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////////////

/*
Encolo el hilo dependiendo de la prioridad del mismo, al finalizar, cambio el estado del hilo a READY ya que se encuentra
en alguna de las colas de prioridad listo para ser ejecutado.
*/

// Notifica a la memoria la creacion del hilo, envia el TCB del mismo. Encola en ready al hilo
func CreacionHiloAMemoria(hilo *globals.TCB) {
	var hiloMemoria globals.HiloMemoria

	hiloMemoria.PID = hilo.PID
	hiloMemoria.TID = hilo.TID
	hiloMemoria.Codigo = hilo.Codigo

	hiloJson, err := json.Marshal(hiloMemoria)
	if err != nil {
		slog.Error("Error convirtiendo TCB a JSON: " + err.Error())
	}

	// Envia el TCB a la memoria para que lo guarde, no hace falta recibir confirmacion de espacio, los hilos se crean igual.
	respMemoryJson, errorAlEnviarMensaje := utils_general.EnviarMensaje(globals.Config.Ip_memory, globals.Config.Port_memory, hiloJson, "Kernel", "/creacionDeHilo")
	var resp utils_general.StatusRespuesta
	err = json.Unmarshal(respMemoryJson, &resp)

	if err!= nil{
		slog.Error("Error decodificando el json")
		return
	}

	slog.Debug("Respuesta al crear hilo en memoria: " + resp.Status)

	if errorAlEnviarMensaje != nil {
		slog.Error("Fallo al enviar mensaje de creacion de hilo a memoria" + errorAlEnviarMensaje.Error())
		return
	}
	var respuesta utils_general.StatusRespuesta

	err = json.Unmarshal(respMemoryJson, &respuesta)
	if err != nil {
		slog.Error("Error decodificando el json: " + err.Error())
		return
	}
	
	if respuesta.Status == "OK"{
		slog.Info("## (<" + strconv.Itoa(hiloMemoria.PID) + ">:<" + strconv.Itoa(hiloMemoria.TID) + ">) Se crea el Hilo - Estado: READY")

		EncolarHiloEnReady(hilo) // Manda al hilo directamente a la cola de ready y cambia su estado.
		
		return
	}	
}

////////////////////////////////////////////////////////////////////////////////////////////////////////////
///////////////////////////////////////Finalización de hilos////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////////////////////////////

func EliminarTIDDelProceso(proceso *globals.PCB, tid int) {
	tidEncontrado := -1

	// Buscar el TID en la lista de TIDs del proceso.
	for i, t := range proceso.TIDs {
		if t == tid {
			tidEncontrado = i
			break
		}
	}

	// Si se encontró el TID, eliminarlo de la lista.
	if tidEncontrado != -1 {
		if len(proceso.TIDs) == 1 {
			proceso.TIDs = []int{} // Dejar la lista vacía si solo tiene un elemento
		} else {
			proceso.TIDs = append(proceso.TIDs[:tidEncontrado], proceso.TIDs[tidEncontrado+1:]...)
		}

		slog.Debug("TID " + strconv.Itoa(tid) + " eliminado de la lista de IDs proceso con PID " + strconv.Itoa(proceso.PID))
	} else {
		slog.Error("No se encontró el TID " + strconv.Itoa(tid) + " en el proceso con PID " + strconv.Itoa(proceso.PID))
	}

}

func FinalizacionDeHilo(hilo *globals.TCB, motivo string) {
	// El motivo puede ser "FIN HILO" o "FIN PROCESO PADRE"
	hayQueChequearColaNew := false

	if(globals.PlanificadorPausado) { //Comprobamos si es que el planificador esta pausado por necesidad de compactar, es necesario chequearlo al finalizar un hilo
		//Si el planificador esta pausado, entonces hay que solicitar compactacion
		if SolicitarCompactacion() {
			slog.Info("Compactacion finalizada. Reintentando chequear espacio de memoria luego de la compactacion para el proceso en cuestion...")
			hayQueChequearColaNew = true
		} else {
			slog.Error("Error en la compactacion al solicitarla desde el Kernel para el proceso")
		}
	}

	hiloEncontrado := 0 // Contador para saber si se encontró el hilo buscado.
	procesoPadre := EncontrarPCBPorPID(hilo.PID)

	if procesoPadre == nil {
		slog.Error("No se encontró el proceso con PID " + strconv.Itoa(hilo.PID) + " al querer finalizar el hilo con TID " + strconv.Itoa(hilo.TID))
	}

	// Llamada a la función para informar a memoria la finalización del hilo
    if !InformarFinalizacionHiloAMemoria(hilo.PID, hilo.TID) {
        slog.Error("Error al enviar la finalización del hilo a memoria.")
	}
	
	DesencolarHilo(hilo)                  // Desencolamos el hilo dado que ya fue finalizado.
	CambiarEstadoHilo(hilo, globals.Exit) // Cambia el estado del hilo a Exit.
	EliminarTIDDelProceso(procesoPadre, hilo.TID) // Eliminamos el TID de la lista de TIDs del proceso.

	//Elimina el hilo del la lista global de todos los hilos
	for i, tcb := range globals.HilosDelSistema.Hilos { // Busca el hilo a eliminar en la lista.
		if tcb.TID == hilo.TID && tcb.PID == hilo.PID {
			// Verificar si la lista tiene solo un elemento
			if len(globals.HilosDelSistema.Hilos) == 1 {
				globals.HilosDelSistema.Hilos = []*globals.TCB{} // Asignar una lista vacía
			} else {
				// Eliminar el hilo usando slicing
				globals.HilosDelSistema.Hilos = append(globals.HilosDelSistema.Hilos[:i], globals.HilosDelSistema.Hilos[i+1:]...)
			}
			
			hiloEncontrado = 1
		
			slog.Info("## (<" + strconv.Itoa(hilo.PID) + ">:<" + strconv.Itoa(hilo.TID) + ">) Finaliza el hilo")
		
			break
		}
		
	}

	if hiloEncontrado == 0 {
		slog.Debug("No se encontro en ningun hilo en el sistema con el TID: " + strconv.Itoa(hilo.TID))
		return
	}

	if hilo.TID != 0 {
		for _, hiloBlocked := range globals.ColaBlockedHilos.Hilos { // Recorremos la lista de hilos bloqueados.
			if hiloBlocked.BloqueadoPor == hilo.TID && hiloBlocked.PID == hilo.PID {
				DesbloquearHilo(hiloBlocked)

				slog.Debug("Al finalizar el hilo con TID " + strconv.Itoa(hilo.TID) + " el hilo con TID " + strconv.Itoa(hiloBlocked.TID) + " fue removido de la cola blocked.")
			}
		}
	}

	if hilo.TID == 0 && motivo == "FIN HILO"{
		procesoPadre := EncontrarPCBPorPID(hilo.PID)

		FinalizarProceso(procesoPadre) // Si el hilo que estamos finalizando tiene el TID 0, entonces disparamos la finalización del proceso padre.
	}

	if hayQueChequearColaNew {
		VerificarColaNew() 
	} else if motivo == "FIN PROCESO PADRE" {
		return // Si entramos a esta funcion por el fin del proceso padre no chequeamos ninguna cola
	} else {
		ProcesarColaReadyHilos()
	}
}

func InformarFinalizacionHiloAMemoria(pid int, tid int, ) bool {
	hiloAfinalizar := &globals.HiloMemoria{} // Inicializamos el puntero.
	
	hiloAfinalizar.PID = pid
	hiloAfinalizar.TID = tid

	mensaje, err := json.Marshal(hiloAfinalizar)

	if err != nil {
		slog.Error("Error al codificar el json de hilo a finalizar para enviar a memoria")
		return false
	}

	respuestaMemoriaJSON, err :=  utils_general.EnviarMensaje(globals.Config.Ip_memory, globals.Config.Port_memory, mensaje, "Kernel", "/finalizacionDeHilo")

	if err != nil {
		slog.Warn("Error al enviar el json de hilo a finalizar a memoria" + err.Error())
		return false
	}

	var respuesta utils_general.StatusRespuesta

	err = json.Unmarshal(respuestaMemoriaJSON, &respuesta)
	
	if err != nil {
		slog.Warn("Error al decodificar el json de hilo a finalizar a memoria" + err.Error())
		return false
	}

	return respuesta.Status == "OK"
}