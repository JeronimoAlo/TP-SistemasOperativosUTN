package utils

import (
	"log/slog"
	"strconv"
)

//////THREADS//////
func THREAD_CREATE(rutaArchivo string, prioridadHilo int, pidPadre int){
	pcbPadre := EncontrarPCBPorPID(pidPadre)
	proximoTID := len(pcbPadre.TIDs) // len nos trae el proximo TID ya que es incremental.

	slog.Info("## (<" + strconv.Itoa(pidPadre) + ">) - Solicitó syscall: THREAD_CREATE.")

	instrucciones, err := cargarCodigoDesdeArchivo(rutaArchivo) // Carga las instrucciones del archivo de pseudocodigo en una lista de instrucciones

	if err != nil {
		slog.Error("Error cargando el archivo de pseudocódigo al crear el hilo: <" + strconv.Itoa(pidPadre) + ">:<" + strconv.Itoa(proximoTID) + ">." + err.Error()) // Notifica si hubo un error al leer el archivo
		return
	}

	nuevoHilo := CrearHilo(pidPadre, proximoTID, prioridadHilo, instrucciones)
	pcbPadre.TIDs = append(pcbPadre.TIDs, nuevoHilo.TID) // Agregamos el nuevo TID a la lista de hilos del proceso.

	CreacionHiloAMemoria(nuevoHilo) // Mandamos a la memoria información del hilo y lo encola a la cola de ready correspondiente.
}

/*
THREAD_JOIN, esta syscall recibe como parámetro un TID, mueve el hilo que la invocó al estado BLOCK hasta que el TID pasado por parámetro finalice.
En caso de que el TID pasado por parámetro no exista o ya haya finalizado, esta syscall no hace nada y el hilo que la invocó continuará su ejecución.
*/
func THREAD_JOIN(tidAEsperar int, tidAbloquear int, pidPadreHiloABloquear int) {
	hiloABloquear := EncontrarTCBPorTID(tidAbloquear, pidPadreHiloABloquear) //Busca al TCB del hilo invocador, para poder bloquearlo.

	slog.Info("## (<" + strconv.Itoa(pidPadreHiloABloquear) + ">:<" + strconv.Itoa(tidAbloquear) + ">) - Solicitó syscall: THREAD_JOIN.")

	if hiloABloquear != nil { // Si no es nil, continua, si no no hace nada.
		BloquearHilo(hiloABloquear, tidAEsperar, "THREAD_JOIN")
		EnviarInterrupcionACPU(hiloABloquear, "Desalojado") 
	}
}

//THREAD_CANCEL, esta syscall recibe como parámetro un TID con el objetivo de finalizarlo pasando al mismo al estado EXIT.
//Se deberá indicar a la Memoria la finalización de dicho hilo. En caso de que el TID pasado por parámetro no exista o ya haya finalizado,
//esta syscall no hace nada. Finalmente, el hilo que la invocó continuará su ejecución.

func THREAD_CANCEL(tidSolicitante int, tidACancelar int, pidPadreHiloACancelar int) {
	slog.Info("## (<" + strconv.Itoa(pidPadreHiloACancelar) + ">:<" + strconv.Itoa(tidSolicitante) + ">) - Solicitó syscall: THREAD_CANCEL.")

	// Obtener el PCB del proceso padre
	procesoPadre := EncontrarPCBPorPID(pidPadreHiloACancelar)

	if procesoPadre == nil {
		slog.Warn("No se encontró el proceso con PID <" + strconv.Itoa(pidPadreHiloACancelar) + "> (THREAD_CANCEL)")
		return
	}

	// Validar si ambos TIDs están en la lista de TIDs del proceso
	if !existeTIDEnProceso(procesoPadre.TIDs, tidSolicitante) && !existeTIDEnProceso(procesoPadre.TIDs, tidACancelar) {
		slog.Warn("El TID solicitante y TID a cancelar no pertenecen al proceso con PID <" + strconv.Itoa(pidPadreHiloACancelar) + "> (THREAD_CANCEL)")
		return
	}

	hiloACancelar := EncontrarTCBPorTID(tidACancelar, pidPadreHiloACancelar)

	if hiloACancelar != nil { // Si encontro al hilo lo finaliza, si no no hace nada.
		FinalizacionDeHilo(hiloACancelar, "FIN HILO") // Se finaliza el hilo y se avisa a memoria de la finalización.
	} else {
		slog.Warn("El TID <" + strconv.Itoa(tidACancelar) + "> no existe o ya ha finalizado. (THREAD_CANCEL)")
	}
}

//THREAD_EXIT, esta syscall finaliza al hilo que lo invocó,
//pasando el mismo al estado EXIT. Se deberá indicar a la Memoria la finalización de dicho hilo.
func THREAD_EXIT(tid int, pidPadre int) {
	slog.Info("## (<" + strconv.Itoa(pidPadre) + ">:<" + strconv.Itoa(tid) + ">) - Solicitó syscall: THREAD_EXIT.")

	hiloAFinalizar := EncontrarTCBPorTID(tid, pidPadre)

	if hiloAFinalizar != nil { // Si encontró al hilo lo finaliza, si no no hace nada.
		FinalizacionDeHilo(hiloAFinalizar, "FIN HILO") // Se finaliza el hilo y se avisa a memoria de la finalización.
	} else {
		slog.Warn("No se puede finalizar el hilo con TID <" + strconv.Itoa(tid) + "> ya que no fue encontrado. (THREAD_EXIT)")
	}
}