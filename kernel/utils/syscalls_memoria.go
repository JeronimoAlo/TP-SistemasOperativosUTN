package utils

import (
	"encoding/json"
	"kernel/globals"
	"log/slog"
	"strconv"
	"utils_general"
	//"time"
)

// "dump" de un proceso: capturar el estado de la memoria y otros recursos de un proceso en un momento determinado
func DUMP_MEMORY(tid int, pidPadre int) {
	hilo := EncontrarTCBPorTID(tid, pidPadre)    // Hilo es un puntero al tcb.
	proceso := EncontrarPCBPorTID(tid, pidPadre) // Proceso es un puntero al pcb.

	if proceso == nil {
		slog.Warn("Proceso con PID " + strconv.Itoa(pidPadre) + " no encontrado en el sistema.")
		return
	}

	if hilo == nil {
		slog.Warn("Al querer realizar un DUMP_MEMORY para el hilo con TID " + strconv.Itoa(tid) + " nos encontramos con que el mismo no existe en el sistema para el PID " + strconv.Itoa(pidPadre))
		return
	}

	slog.Info("## (<" + strconv.Itoa(proceso.PID) + ">:<" + strconv.Itoa(hilo.TID) + ">) - Solicitó syscall: DUMP_MEMORY.")

	DesencolarHilo(hilo)
	BloquearHilo(hilo, tid, "DUMP_MEMORY") // Bloqueamos el hilo con tu motivo.

	//Interrupción al CPU por case "DUMP_MEMORY"
	globals.HiloSolicitoDUMP = true
	EnviarInterrupcionACPU(hilo, "DUMP_MEMORY")
}

func SolicitarDumpMemoria(pid int, tid int) bool {
	var hiloProceso globals.HiloMemoria

	hiloProceso.PID = pid
	hiloProceso.TID = tid

	mensaje, err :=  json.Marshal(hiloProceso)

	if err != nil{
		slog.Warn("Error codificando el mensaje al dumpear memoria.")
		return false	
	}
	
	respuestaMemoriaJSON, err2 := utils_general.EnviarMensaje(globals.Config.Ip_memory, globals.Config.Port_memory, mensaje, "Kernel", "/memoryDump")

	if err2 != nil{
		slog.Warn("Error recibiendo la respuesta de Memoria al querer dumpear memoria.")
		return false
	}

	var respuesta utils_general.StatusRespuesta

	err3 := json.Unmarshal(respuestaMemoriaJSON, &respuesta)

	if err3 != nil{
		slog.Warn("Error decodificando la respuesta de Memoria al dumpear memoria.")
		return false
	}

	return respuesta.Status == "OK"
}