package utils

import (
	"kernel/globals"
	"log/slog"
	"strconv"
	"time"
)

func IO(tid int, pidPadre int, tiempoES int) {
	hilo := EncontrarTCBPorTID(tid, pidPadre) // Hilo es un puntero al tcb.
	proceso := EncontrarPCBPorTID(tid, pidPadre) // Proceso es un puntero al pcb.

	if proceso == nil {
		slog.Warn("Proceso con PID " + strconv.Itoa(pidPadre) + " no encontrado en el sistema.")
		return
	}

	if hilo == nil {
		slog.Warn("Al solicitar IO para el hilo con TID " + strconv.Itoa(tid) + " nos encontramos con que el mismo no existe en el sistema para el PID " + strconv.Itoa(pidPadre))
		return
	}

	// Logueamos la solicitud de E/S.
	slog.Info("## (<" + strconv.Itoa(hilo.PID) + ">:<" + strconv.Itoa(hilo.TID) + ">) - Solicitó syscall: IO por " + strconv.Itoa(tiempoES) + " ms.")

	// Bloqueamos el hilo mientras se espera la operación de E/S.
	DesencolarHilo(hilo)
	BloquearHilo(hilo, hilo.TID, "IO")

	// Encolamos el hilo en la cola de E/S.
	SolicitarIO(hilo, tiempoES)

	//Interrupción al CPU por IO
	EnviarInterrupcionACPU(hilo, "IO")

	// Si el dispositivo no está en uso, empezamos a procesar la cola.
	if !globals.DispositivoEntradaSalida.EnUso {
		ProcesarColaES()
	}else {
		slog.Info("El dispositivo de IO está en uso, se bloqueó el hilo que lo solicitó hasta que finalice.")
	}
}

func SolicitarIO(hilo *globals.TCB, tiempoES int) {
	globals.DispositivoEntradaSalida.ColaES = append(globals.DispositivoEntradaSalida.ColaES, hilo) // Agregamos el hilo a la lista de espera para E/S.
	globals.DispositivoEntradaSalida.TiempoES[hilo] = tiempoES                                      // Guardamos el tiempo de E/S para el hilo en cuestión.
}

func ProcesarColaES() {
	if len(globals.DispositivoEntradaSalida.ColaES) == 0 {
		return // Si la lista a procesar está vacía, no devuelvo nada.
	}

	globals.DispositivoEntradaSalida.EnUso = true

	// Tomar el primer hilo en la cola (FIFO).
	hilo := globals.DispositivoEntradaSalida.ColaES[0]

	// Remover el hilo de la cola.
	globals.DispositivoEntradaSalida.ColaES = globals.DispositivoEntradaSalida.ColaES[1:]

	// Recuperamos el tiempo de E/S para este hilo.
	tiempoES, existe := globals.DispositivoEntradaSalida.TiempoES[hilo]
	if !existe {
		slog.Warn("No se encontró tiempo de E/S para el hilo " + strconv.Itoa(hilo.TID) + " al querer procesar la cola de entrada/salida")
		globals.DispositivoEntradaSalida.EnUso = false // Liberamos el dispositivo si hay un error.
		return
	}

	// Simular la espera de la operación de E/S.
	time.Sleep(time.Duration(tiempoES) * time.Millisecond)

	slog.Info("El dispositivo de IO finalizó para el hilo con TID " + strconv.Itoa(hilo.TID) + " y PID " + strconv.Itoa(hilo.PID))

	// Desbloquear el hilo cuando la operación de E/S finalice y lo encola en Ready.
	DesbloquearHilo(hilo)

	slog.Info("## (<" + strconv.Itoa(hilo.PID) + ">:<" + strconv.Itoa(hilo.TID) + ">) finalizó IO y pasa a READY")

	// Procesar el siguiente hilo en la cola si lo hay
	if len(globals.DispositivoEntradaSalida.ColaES) > 0 {
		ProcesarColaES() // Llamamos recursivamente para procesar el siguiente hilo.
	} else {
		globals.DispositivoEntradaSalida.EnUso = false // Marcar el dispositivo como no en uso si no hay más hilos.
	}
}