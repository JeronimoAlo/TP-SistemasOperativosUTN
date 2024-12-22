package utils

import (
	"cpu/globals"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"utils_general"
)

// FETCH.
func FetchAMemoria() (globals.InstruccionDesdeMemoria, bool, error) {
	var solicitudInstruccion globals.SolicitudInstruccion // Seteamos la solicitud de instrucción para solicitar a memoria.
	solicitudInstruccion.PC = globals.RegistrosCPU.PC
	solicitudInstruccion.PID = globals.TIDyPIDenEjecucion.PID
	solicitudInstruccion.TID = globals.TIDyPIDenEjecucion.TID

	request, err := json.Marshal(solicitudInstruccion)
	if err != nil {
		slog.Error("Error al codificar el TID y PID en ejecucion")
	}

	respuestaJson, err := utils_general.EnviarMensaje(globals.Config.Ip_memory, globals.Config.Port_memory, request, "CPU", "/obtenerInstruccion")
	if err != nil {
		slog.Error("Error al obtener la proxima Instruccion a ejecutar por la cpu")
	}

	var instruccionDesdeMemoria globals.InstruccionDesdeMemoria

	err = json.Unmarshal(respuestaJson, &instruccionDesdeMemoria)
	if err != nil {
		slog.Error("Error al decodificar el json de la instruccion que envio la memoria a la CPU.")
	}

	slog.Info("## TID: <" + strconv.Itoa(globals.TIDyPIDenEjecucion.TID) + "> - FETCH - Program Counter: <" + strconv.FormatUint(uint64(globals.RegistrosCPU.PC), 10) + ">")
	return instruccionDesdeMemoria, instruccionDesdeMemoria.NoHayMasInstrucciones, err

}

// DECODE
// Decode interpreta la instrucción desde memoria y la prepara para la ejecución.
func Decode(instruccion globals.InstruccionDesdeMemoria) (globals.InstruccionDecodificada, error) {
	// Crear la estructura de la instrucción decodificada
	decodificada := globals.InstruccionDecodificada{
		Operacion:          instruccion.Operacion,
		Parametros:         instruccion.Parametros,
		RequiereTraduccion: false,
	}

	// Validar que la instrucción tiene los parámetros correctos (opcional)
	switch instruccion.Operacion {
	case "PROCESS_CREATE":
		if len(instruccion.Parametros) != 3 { // Estas operaciones necesitan 3 parametros si o si, si no debería fallar
			return globals.InstruccionDecodificada{}, fmt.Errorf("la instrucción %s requiere 3 parámetros", instruccion.Operacion)
		}
	case "SET", "SUM", "SUB", "JNZ", "THREAD_CREATE":
		if len(instruccion.Parametros) != 2 { // Estas operaciones necesitan 2 parametros si o si, si no debería fallar
			return globals.InstruccionDecodificada{}, fmt.Errorf("la instrucción %s requiere 2 parámetros", instruccion.Operacion)
		}
	case "READ_MEM", "WRITE_MEM":
		if len(instruccion.Parametros) != 2 { // Estas operaciones necesitan 2 parametros si o si, si no debería fallar
			return globals.InstruccionDecodificada{}, fmt.Errorf("la instrucción %s requiere 2 parámetros (Registro Dirección y Registro Datos)", instruccion.Operacion)
		}
		decodificada.RequiereTraduccion = true //Como READ_MEM y WRITE_MEM requieren una traduccion entonce seteamos el decodificada.RequiereTraduccion en true, luego la etapa de execute debería de traducirlas (usando la mmu)

	case "LOG", "IO", "THREAD_JOIN", "THREAD_CANCEL", "MUTEX_CREATE", "MUTEX_LOCK", "MUTEX_UNLOCK":
		if len(instruccion.Parametros) != 1 { // Estas operaciones necesitan 1 parametros si o si, si no debería fallar
			return globals.InstruccionDecodificada{}, fmt.Errorf("la instrucción %s requiere 1 parámetros", instruccion.Operacion)
		}
	case "DUMP_MEMORY", "THREAD_EXIT", "PROCESS_EXIT":
		if len(instruccion.Parametros) != 0 { // Estas operacionesno tienen que tener paramteros, si no debería fallar
			return globals.InstruccionDecodificada{}, fmt.Errorf("la instrucción %s requiere 2 parámetros", instruccion.Operacion)
		}
	default:
		// Instrucción no reconocida
		return globals.InstruccionDecodificada{}, fmt.Errorf("instrucción no reconocida: %s", instruccion.Operacion)
	}
	// Devolver la instrucción decodificada

	slog.Debug("OPERACION DECODIFICADA " + decodificada.Operacion)
	for i, parametro := range decodificada.Parametros {
		slog.Debug("Contenido del numero de parametro " + strconv.Itoa(i) + " y su valor: " + parametro)
	}

	return decodificada, nil
}

// EXECUTE
func EjecutarInstruccion(instruccion globals.InstruccionDecodificada) bool {
	var termino = false
	tid := globals.TIDyPIDenEjecucion.TID
	pid := globals.TIDyPIDenEjecucion.PID

	if instruccion.RequiereTraduccion {
		switch instruccion.Operacion {
		case "READ_MEM":
			READ_MEM(&globals.RegistrosCPU, instruccion.Parametros[0], instruccion.Parametros[1])

		case "WRITE_MEM":
			WRITE_MEM(&globals.RegistrosCPU, instruccion.Parametros[0], instruccion.Parametros[1])
		}
	} else {
		switch instruccion.Operacion {
		case "SET":
			instruccionInt, _ := strconv.Atoi(instruccion.Parametros[1])
			SET(&globals.RegistrosCPU, instruccion.Parametros[0], instruccionInt)
		case "SUM":
			SUM(&globals.RegistrosCPU, instruccion.Parametros[0], instruccion.Parametros[1])
		case "SUB":
			SUB(&globals.RegistrosCPU, instruccion.Parametros[0], instruccion.Parametros[1])
		case "JNZ":
			instruccionInt, _ := strconv.Atoi(instruccion.Parametros[1])
			JNZ(&globals.RegistrosCPU, instruccion.Parametros[0], instruccionInt)
		case "LOG":
			LOG(&globals.RegistrosCPU, instruccion.Parametros[0])
		case "THREAD_CREATE":
			param1, _ := strconv.Atoi(instruccion.Parametros[1])
			param2 := pid
			THREAD_CREATE(instruccion.Parametros[0], param1, param2)
		case "THREAD_EXIT":
			THREAD_EXIT(tid, pid)
			termino = true
		case "THREAD_JOIN":
			param1, _ := strconv.Atoi(instruccion.Parametros[0])
			param2 := tid
			param3 := pid

			THREAD_JOIN(param1, param2, param3)
		case "THREAD_CANCEL":
			param0, _ := strconv.Atoi(instruccion.Parametros[0])

			THREAD_CANCEL(tid, param0, pid)
		case "PROCESS_CREATE":
			param1, _ := strconv.Atoi(instruccion.Parametros[1])
			param2, _ := strconv.Atoi(instruccion.Parametros[2])

			PROCESS_CREATE(instruccion.Parametros[0], param1, param2)
		case "PROCESS_EXIT":
			termino = true
			PROCESS_EXIT(tid, pid)
		case "MUTEX_CREATE":
			MUTEX_CREATE(instruccion.Parametros[0])
		case "MUTEX_LOCK":
			MUTEX_LOCK(instruccion.Parametros[0])
		case "MUTEX_UNLOCK":
			MUTEX_UNLOCK(instruccion.Parametros[0])
		case "IO":
			param1, _ := strconv.Atoi(instruccion.Parametros[0])
			IO(param1)
		case "DUMP_MEMORY":
			err := DUMP_MEMORY()
			if err != nil{
				slog.Error("Error al hacer el DUMP_MEMORY" + err.Error())
			}
		default:
			slog.Error("Instruccion no reconocida en funcion EjecutarInstruccion")
		}
		parametros := strings.Join(instruccion.Parametros, ", ") // Convierto instruccion.Parametros (lista) en un string con los elementos de dicha lista separados por comas.

		slog.Info("## TID: <" + strconv.Itoa(tid) + "> - Ejecutando: <" + instruccion.Operacion + "> - <" + parametros + ">.")
	}
	return termino
}

func EjecutarHilo() {
	var termino = false

	globals.HayHiloEjecutandoEnCPU = true

	ObtenerContextoEjecucion(globals.TIDyPIDenEjecucion.PID, globals.TIDyPIDenEjecucion.TID, &globals.RegistrosCPU) // Se obtiene el contexto del hilo y el proceso a ejecutar.

	instruccionDesdeMemoria, _, err := FetchAMemoria()
	if err != nil {
		slog.Error("Error al hacer el fetch a memoria: ", "error", err.Error())
		return
	}

	PCantesDeEjecucion := globals.RegistrosCPU.PC

	instruccionDecodificada, _ := Decode(instruccionDesdeMemoria)
	termino = EjecutarInstruccion(instruccionDecodificada)
	//slog.Debug("Termine de ejecutar la instruccion antes del for: " +  instruccionDecodificada.Operacion)

	if PCantesDeEjecucion == globals.RegistrosCPU.PC {
		//slog.Debug("Entre a aumentar el PC antes del for")
		globals.RegistrosCPU.PC++ // Al finalizar el ciclo, el PC deberá ser actualizado sumándole 1
		// en caso de que éste no haya sido modificado por la instrucción.
	}

	if instruccionDecodificada.Operacion == "THREAD_EXIT" || instruccionDecodificada.Operacion == "PROCESS_EXIT" {
		termino = true
		globals.HayHiloEjecutandoEnCPU = false
		return
	}

	// slog.Debug("Estado antes del FOR",
	// "termino", termino,
	// "hayInterrupciones", HayInterrupciones(globals.TIDyPIDenEjecucion.TID, globals.TIDyPIDenEjecucion.PID))

	if HayInterrupciones(globals.TIDyPIDenEjecucion.TID, globals.TIDyPIDenEjecucion.PID) {
		slog.Debug("El tid que voy a mandar en ejecucion es: " + strconv.Itoa(globals.TIDyPIDenEjecucion.TID) + " el pid es: " + strconv.Itoa(globals.TIDyPIDenEjecucion.PID))
		err2 := ChequeoInterrupciones(globals.TIDyPIDenEjecucion.TID, globals.TIDyPIDenEjecucion.PID)

		if err2 != nil {
			slog.Error("Error al chequear interrupciones")
		}
		return
	}

	// slog.Debug("Antes de entrar al for: " )
	// slog.Debug("BIS: Estado antes del FOR",
	// "termino", termino,
	// "hayInterrupciones", HayInterrupciones(globals.TIDyPIDenEjecucion.TID, globals.TIDyPIDenEjecucion.PID))

	for !termino && !HayInterrupciones(globals.TIDyPIDenEjecucion.TID, globals.TIDyPIDenEjecucion.PID) { // Mientras que hay mas instrucciones para ejecutar del hilo o no haya interrupciones, continua el ciclo de ejecución.
		PCantesDeEjecucion := globals.RegistrosCPU.PC // Guardamos el valor del PC antes de hacer la ejecucion, para controlar si este cambio por alguna instruccion la ejecución.

		instruccionDesdeMemoria, _, err = FetchAMemoria()
		instruccionDecodificada, errDecodificacion := Decode(instruccionDesdeMemoria)

		if errDecodificacion != nil {
			slog.Error("Error al decodificar la instruccion a ejecutar", "error", errDecodificacion.Error())
		}

		if !termino {
			termino = EjecutarInstruccion(instruccionDecodificada)
			//slog.Debug("Termine de ejecutar la instruccion DESPUES del for: " +  instruccionDecodificada.Operacion)
			if PCantesDeEjecucion == globals.RegistrosCPU.PC {
				//slog.Debug("Entre a aumentar el PC dentro del for")
				globals.RegistrosCPU.PC++ // Al finalizar el ciclo, el PC deberá ser actualizado sumándole 1
				// en caso de que éste no haya sido modificado por la instrucción.
			}
			if instruccionDecodificada.Operacion == "THREAD_EXIT" || instruccionDecodificada.Operacion == "PROCESS_EXIT" {
				termino = true
				globals.HayHiloEjecutandoEnCPU = false
				return
			}

			// Si hay interrupciones entonces el hilo finalizo por una interrupcion, ya sea de prioridad o de quantum.
			if HayInterrupciones(globals.TIDyPIDenEjecucion.TID, globals.TIDyPIDenEjecucion.PID) {
				slog.Debug("El tid que voy a mandar en ejecucion es: " + strconv.Itoa(globals.TIDyPIDenEjecucion.TID) + " el pid es: " + strconv.Itoa(globals.TIDyPIDenEjecucion.PID))
				err := ChequeoInterrupciones(globals.TIDyPIDenEjecucion.TID, globals.TIDyPIDenEjecucion.PID) // Chequea interrupciones, manda a actualizar el contexto , se desaloja el hilo en ejecucion y se avisa al Kernel.

				if err != nil {
					slog.Error("Error al chequear las interrupciones.")
				}
				return
			}
		}

		if termino { // Si no hay mas instrucciones entonces el hilo finalizó la ejecucion.
			globals.HayHiloEjecutandoEnCPU = false
			return
		}

		if err != nil {
			slog.Error("Error al hacer el fetch a memoria.")
		}
	}
}
