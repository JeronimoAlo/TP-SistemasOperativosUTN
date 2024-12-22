package utils

import (
	"kernel/globals"
	"log/slog"
	"strconv"
)

func MUTEX_CREATE(pid int, nombreRecurso string) {
	slog.Info("## (<" + strconv.Itoa(pid) + ">) - Solicitó syscall: MUTEX_CREATE.")

	proceso := EncontrarPCBPorPID(pid) // Proceso será un puntero al PCB.

	if proceso == nil {
		slog.Warn("Al querer crear un MUTEX para el proceso con PID " + strconv.Itoa(pid) + " nos encontramos con que el mismo no existe en el sistema.")
		return
	}

	nuevo_mutex := &globals.Mutex{
		ID:         len(globals.Lista_mutex_global) + 1, // Generamos ID único del mutex.
		HiloOwner:  nil,                                 // Inicializamos en 0.
		Bloqueados: []*globals.TCB{},                    // Inicializamos una lista vacía.
		NombreRecurso: nombreRecurso,                    // Seteamos el nombre del recurso
	} // Creamos la estructura del mutex.

	proceso.Mutex = append(proceso.Mutex, *nuevo_mutex) // Lo agregamos a la lista de los mutex del proceso que lo creó.

	globals.Lista_mutex_global[nuevo_mutex.ID] = nuevo_mutex // Agregamos el mutex a la lista global edl sistema.

	slog.Info("Mutex creado correctamente, el ID del mismo es: " + strconv.Itoa(nuevo_mutex.ID) + " y el recurso a proteger es " + nombreRecurso)
}

func MUTEX_LOCK(tid int, pidPadre int, nombreRecurso string) {
	mutexID := -1

	hilo := EncontrarTCBPorTID(tid, pidPadre) // Hilo es un puntero al tcb.
	proceso := EncontrarPCBPorPID(pidPadre) // Obtenemos el proceso asociado al hilo.

	if proceso == nil {
		slog.Warn("Proceso con PID " + strconv.Itoa(pidPadre) + " no encontrado en el sistema.")
		return
	}

	if hilo == nil {
		slog.Warn("Al querer bloquear el MUTEX para el hilo con TID " + strconv.Itoa(tid) + " nos encontramos con que el mismo no existe en el sistema para el PID " + strconv.Itoa(pidPadre))
		return
	}

	slog.Info("## (<" + strconv.Itoa(hilo.PID) + ">:<" + strconv.Itoa(hilo.TID) + ">) - Solicitó syscall: MUTEX_LOCK.")

	// Verificación: Si el nombre del recurso a proteger no está en la lista de mutex del proceso, logueamos un error.
	encontrado := false
	for _, mutex := range proceso.Mutex {
		if mutex.NombreRecurso == nombreRecurso {
			encontrado = true
			mutexID = mutex.ID

			break
		}
	}

	if !encontrado {
		slog.Warn("El mutex que protege el recurso " + nombreRecurso + " no pertenece al proceso con PID " + strconv.Itoa(pidPadre))
		return
	}

	// Verificamos si el mutexID existe en la lista global.
	mutex, exists := globals.Lista_mutex_global[mutexID] // Chequeamos que el mutexID sea válido.
	if !exists {
		slog.Warn("El mutex con ID " + strconv.Itoa(mutexID) + " no existe en el sistema. Finalizando el hilo que solicitó MUTEX_LOCK...")
		//FinalizacionDeHilo(hilo, "FIN HILO") // Si el mutex no existe, finalizamos el hilo.
		EnviarInterrupcionACPU(hilo, "Finalizar")
		return
	} else if mutex.HiloOwner == nil {
		mutex.HiloOwner = hilo // Si el mutex no está tomado, asigna el mutex a este hilo.
		slog.Info("El mutex con ID " + strconv.Itoa(mutexID) + " fue asignado correctamente al hilo con TID " + strconv.Itoa(hilo.TID))
	} else if mutex.HiloOwner != nil {
		mutex.Bloqueados = append(mutex.Bloqueados, hilo) // Agrega el hilo a la cola de bloqueados del mutex en cuestión.
		slog.Warn("El mutex con ID " + strconv.Itoa(mutexID) + " está en uso por el hilo con TID " + strconv.Itoa(mutex.HiloOwner.TID) + " se bloquea el hilo con TID "  + strconv.Itoa(hilo.TID))

		BloquearHilo(hilo, hilo.TID, "MUTEX")

		EnviarInterrupcionACPU(hilo, "Desalojado")
	} else {
		slog.Debug("Revisar Syscall Mutex Lock")
	}
}

func MUTEX_UNLOCK(tid int, pidPadre int, nombreRecurso string) {
	hilo := EncontrarTCBPorTID(tid, pidPadre) // Hilo es un puntero al tcb.
	proceso := EncontrarPCBPorPID(pidPadre) // Obtenemos el proceso asociado al hilo.

	if proceso == nil {
		slog.Warn("Proceso con PID " + strconv.Itoa(pidPadre) + " no encontrado en el sistema.")
		return
	}

	if hilo == nil {
		slog.Warn("Al querer desbloquear el MUTEX para el hilo con TID " + strconv.Itoa(tid) + " nos encontramos con que el mismo no existe en el sistema para el PID " + strconv.Itoa(pidPadre))
		return
	}

	encontrado := false
	var mutexID int
	for _, mutex := range proceso.Mutex {
		if mutex.NombreRecurso == nombreRecurso {
			encontrado = true
			mutexID = mutex.ID

			break
		}
	}

	if(!encontrado){
		slog.Error("No encontre el MUTEX para hacer unlock de nombre: " + nombreRecurso + " asociado al hilo con TID " + strconv.Itoa(tid))
		//FinalizacionDeHilo(hilo, "FIN HILO") // Si no existe, el hilo termina.
		EnviarInterrupcionACPU(hilo, "Finalizar")
	}
	
	mutex := globals.Lista_mutex_global[mutexID]

	slog.Info("## (<" + strconv.Itoa(hilo.PID) + ">:<" + strconv.Itoa(hilo.TID) + ">) - Solicitó syscall: MUTEX_UNLOCK.")

	if mutex.HiloOwner.TID == tid {
		if len(mutex.Bloqueados) > 0 {
			// Desbloquear al primer hilo en la cola de bloqueados
			siguienteHilo := mutex.Bloqueados[0]
			mutex.Bloqueados = mutex.Bloqueados[1:] // Eliminamos por slice el primer hilo.

			DesbloquearHilo(siguienteHilo) // Desbloqueamos el hilo y lo pasamos a ready.

			mutex.HiloOwner = siguienteHilo // Asignamos el primer hilo de la cola de bloqueados como el owner del mutex.
		} else {
			mutex.HiloOwner = nil // Liberar el mutex si no hay hilos bloqueados
		}
	} else { // Si el que pidió el mutex no es el hilo owner, no hacemos nada.
		slog.Info("(<" + strconv.Itoa(hilo.PID) + ">:<" + strconv.Itoa(hilo.TID) + ">) - no tiene asignado el mutex en este momento.")
		return
	}

	slog.Info("Mutex unlock realizado correctamente.")
}