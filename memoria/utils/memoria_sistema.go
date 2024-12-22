package utils

import (
	"encoding/json"
	"log/slog"
	"memoria/globals"
	"net/http"
	"strconv"
	"utils_general"
	"time"
)

func AgregarProcesoAMemoriaSistema(proceso globals.ProcesoMemoria, particion *globals.Particion){
	var pcb globals.Proceso
	contextoPCB := &globals.ContextoEjecucionProceso{}

	pcb.PID = proceso.PID

	contextoPCB.Base = particion.Base // Almacenamos el valor de la base del proceso creado en memoria.
	contextoPCB.Limite = particion.Base + uint32(particion.Tamanio) // Almacenamos el valor del límite del proceso creado en memoria.
	pcb.Tamanio =  proceso.Tamanio
	
	pcb.ContextoEjecucionProceso = contextoPCB

	globals.ListaContextosEjecucionProcesos = append(globals.ListaContextosEjecucionProcesos, &pcb)

	//slog.Debug("PID " + strconv.Itoa(pcb.PID) + " agregado a la memoria del sistema.")
	//slog.Debug("Base de la particion del PID " + strconv.Itoa(int(contextoPCB.Base)) + " mapeada OK.")
	//slog.Debug("Limite de la particion del PID " + strconv.Itoa(int(contextoPCB.Limite)) + " mapeada OK.")
}

func AgregarHiloAMemoria(hilo globals.HiloMemoria) { // Va a recibir el struct ProcesoMemoria para
	var tcb globals.Hilo // Creamos el TCB que almacenaremos en memoria

	tcb.PID = hilo.PID // Asignamos el PID que recibimos a nuestro TCB
	tcb.TID = hilo.TID
	tcb.Codigo = hilo.Codigo

	var contextoTCB globals.ContextoEjecucionHilo

	contextoTCB.AX = 0
	contextoTCB.BX = 0
	contextoTCB.CX = 0
	contextoTCB.DX = 0
	contextoTCB.EX = 0
	contextoTCB.FX = 0
	contextoTCB.GX = 0
	contextoTCB.HX = 0
	contextoTCB.PC = 0

	tcb.ContextoEjecucionHilo = contextoTCB // Inicializamos los registros a almacenar en 0

	globals.ListaContextosEjecucionHilos = append(globals.ListaContextosEjecucionHilos, &tcb) // Encolamos nuestro TCB en la lista de contextos de ejecucion de memoria.
	
	slog.Info("## Hilo Creado - (PID:TID) - (<" + strconv.Itoa(hilo.PID) + ">:<" + strconv.Itoa(hilo.TID) + ">).")
}

// Elimina un proceso de memoria por su PID.
func EliminarProcesoDeMemoria(PID int) {
    // Buscar el proceso en la lista de contextos de ejecución de procesos.

    for j, proceso := range globals.ListaContextosEjecucionProcesos {
        if proceso.PID == PID {
			if len(globals.ListaContextosEjecucionProcesos) == 1 {
				globals.ListaContextosEjecucionProcesos = []*globals.Proceso{}
			}else{
				globals.ListaContextosEjecucionProcesos = append(globals.ListaContextosEjecucionProcesos[:j], globals.ListaContextosEjecucionProcesos[j+1:]...)
			}
            
			slog.Debug("Proceso con PID: " + strconv.Itoa(PID) + " eliminado de memoria.")
			return
        }
    }

    slog.Error("Proceso con PID: " + strconv.Itoa(PID) + " no encontrado en memoria.")
}

func ObtenerContextoDeEjecucion(writer http.ResponseWriter, requestCliente *http.Request) {
	var contador int
	var request globals.MensajeObtenerContexto

	err := json.NewDecoder(requestCliente.Body).Decode(&request) // Decodificamos la respuesta del cliente dentro de la variable request.

	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest) // Error el decodificar el request del cliente.
		slog.Warn("Error al decodificar el request del cliente en la recepción del mensaje.(Origen: ObtenerContextoDeEjecucion)")
		return
	}

	// Deserializar el JSON del campo request.Mensaje en la variable mensaje
	pid, _ := strconv.Atoi(request.PID)
	tid, _ := strconv.Atoi(request.TID)

	var contextoHilo globals.ContextoEjecucionHilo
	var contextoProceso globals.ContextoEjecucionProceso

	contador = 0

	// Buscar el contexto del hilo correspondiente al TID
	for _, hilo := range globals.ListaContextosEjecucionHilos {
		if hilo.TID == tid && hilo.PID == pid{ //Si encuentro el Hilo guardo el contexto en una variable
			contextoHilo = hilo.ContextoEjecucionHilo

			contador += 1

			break
		}
	}

	// Buscar el contexto del proceso correspondiente al PID
	for _, proceso := range globals.ListaContextosEjecucionProcesos {
		if proceso.PID == pid { // Si encuentro el proceso gurado el contexto en una variable
			contextoProceso = *proceso.ContextoEjecucionProceso

			contador += 1

			break
		}
	}

	if contador == 2 { // Chequeo que exista el par proceso/hilo del hilo del que quiero obtener el contexto
		// Crear la estructura combinada
		respuestaCombinada := globals.RespuestaContextoProcesoHilo{
			PC:     contextoHilo.PC,
			AX:     contextoHilo.AX,
			BX:     contextoHilo.BX,
			CX:     contextoHilo.CX,
			DX:     contextoHilo.DX,
			EX:     contextoHilo.EX,
			FX:     contextoHilo.FX,
			GX:     contextoHilo.GX,
			HX:     contextoHilo.HX,
			Base:   contextoProceso.Base,
			Limite: contextoProceso.Limite,
		}

		// Codificar la estructura combinada en JSON
		respuesta, err := json.Marshal(respuestaCombinada)
		if err != nil {
			slog.Debug("Error codificando la respuesta final a JSON: " + err.Error())
		}

		// Tiempo de retardo en la petición.
		time.Sleep(time.Duration(globals.Config.Response_delay) * time.Millisecond)

		utils_general.EnviarRespuesta(writer, respuesta)
		
		slog.Info("## Contexto <Solicitado> - (PID:TID) - (<" + request.PID +">:<" + request.TID +">)")
	}else{
		slog.Error("No se encontro el par proceso / hilo en memoria solicitado.")
		utils_general.EnviarStatusFAILED(writer)
	}	
}

func ActualizarContextoDeEjecucion(writer http.ResponseWriter, requestCliente *http.Request) {
	var request globals.ContextoEjecucionHilo

	err := json.NewDecoder(requestCliente.Body).Decode(&request) // Decodificamos la respuesta del cliente dentro de la variable request.

	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest) // Error el decodificar el request del cliente.
		slog.Warn("Error al decodificar el request del cliente en la recepción del mensaje. (Origen: ActualizarContextoDeEjecucion)")
		return
	}

	tid := request.TID
	pid := request.PID

	hiloEncontrado := 0 // Si vale 0 es que el hilo no se encontró, si vale 1 es que sí.

	for _, hilo := range globals.ListaContextosEjecucionHilos {
		if hilo.TID == tid && hilo.PID == pid {
			hilo.ContextoEjecucionHilo.PC = request.PC
			hilo.ContextoEjecucionHilo.AX = request.AX
			hilo.ContextoEjecucionHilo.BX = request.BX
			hilo.ContextoEjecucionHilo.CX = request.CX
			hilo.ContextoEjecucionHilo.DX = request.DX
			hilo.ContextoEjecucionHilo.EX = request.EX
			hilo.ContextoEjecucionHilo.FX = request.FX
			hilo.ContextoEjecucionHilo.GX = request.GX
			hilo.ContextoEjecucionHilo.HX = request.HX

			hiloEncontrado = 1
			
			
		}
	}

	// Tiempo de retardo en la respuesta.
	time.Sleep(time.Duration(globals.Config.Response_delay) * time.Millisecond)

	if hiloEncontrado == 1 {
		utils_general.EnviarStatusOK(writer)
		slog.Info("## Contexto <Actualizado> - (PID:TID) - (<"+strconv.Itoa(request.PID)+">:<"+strconv.Itoa(request.TID)+">)")
	} else { 
		slog.Warn("No se puede actualizar el contexto del hilo <" + strconv.Itoa(request.PID) + ">:<" + strconv.Itoa(request.TID) + ">, debido a que no se encuentra en el sistema.")
		utils_general.EnviarStatusFAILED(writer)
	}
}