package utils

import (
	"kernel/globals"
	"log/slog"
	"strconv"
	"utils_general"
)

func PROCESS_CREATE(rutaArchivo string, tamanio int, prioridadHiloMain int) {
	slog.Info("## CPU - Solicitó syscall: PROCESS_CREATE.")

	listaPIDs := MapPID(globals.ProcesosDelSistema.Procesos) // Mapea todos los PIDs del sistema en una lista, esto sirve para chequear que no estemos creando un PID que ya exista
	nuevoPID, err := utils_general.Max(listaPIDs)            // Asigna el valor maximo que haya de PID a una variable, para luego sumarle 1

	if err != nil {
		slog.Debug(err.Error())
	} else {
		instrucciones, err := cargarCodigoDesdeArchivo(rutaArchivo) // Carga las instrucciones del archivo de pseudocodigo en una lista de instrucciones

		if err != nil {
			slog.Error("Error cargando el archivo: " + err.Error())
			return
		}

		nuevoPID = nuevoPID + 1                                                                // Suma 1 al numero de PID, para que no se repita
		pcb := CrearProceso(nuevoPID, tamanio)                                                 // Crea el PCB del nuevo proceso.
		globals.ProcesosDelSistema.Procesos = append(globals.ProcesosDelSistema.Procesos, pcb) // Agregamos el nuevo proceso a la lista de procesos del sistema.

		CrearHilo(nuevoPID, 0, prioridadHiloMain, instrucciones) // Crea el TCB del  hilo main con el pid del proceso padre, la prioridad del hilo y su codigo

		EncolarProceso(pcb, &globals.ColaNewProcesos, globals.New) // Encola el proceso en new.
		
		VerificarColaNew() // Verifica si se puede mover algún proceso de NEW a READY.
	}
}

func PROCESS_EXIT(tid int, pidPadre int) {
	// Validamos que el TID sea 0, de lo contrario, se rechaza la syscall.
	if tid != 0 {
		slog.Warn("PROCESS_EXIT solo puede ser llamado por el TID 0, el TID recibido es el: " + strconv.Itoa(tid))
		return
	}

	tcbAsociado := EncontrarTCBPorTID(tid, pidPadre)

	if tcbAsociado == nil {
		slog.Warn("No se encontró un TCB asociado para el TID: " + strconv.Itoa(tid) + ". Al querer procesar PROCESS_EXIT")
		return
	}

	procesoAFinalizar := EncontrarPCBPorTID(tid, tcbAsociado.PID) // Buscamos el proceso a finalizar (Formato PCB).

	if procesoAFinalizar == nil {
		slog.Warn("No se encontró ningún proceso relacionado con el TID que solicitó la syscall PROCESS_EXIT.")
		return
	}

	slog.Info("## (<" + strconv.Itoa(procesoAFinalizar.PID) + ">:<" + strconv.Itoa(tid) + ">) - Solicitó syscall: PROCESS_EXIT.")

	FinalizarProceso(procesoAFinalizar) // Finalizamos el proceso y todos sus hilos asociados.
}