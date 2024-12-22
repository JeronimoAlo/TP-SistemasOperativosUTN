package utils

import (
	"encoding/json"
	"log/slog"
	"memoria/globals"
	"net/http"
	"strconv"
	"utils_general"
	"time"
	"fmt"
	"bytes"
)

func NombreArchivoMetadata(request globals.HiloMemoria) string {
	pid := strconv.Itoa(request.PID)
	tid := strconv.Itoa(request.TID)
	tiempo := time.Now().Format("15:04:05") // Seteamos el formato que utiliza Go por defecto para las constantes de tiempo.

	nombre := fmt.Sprintf("%s-%s-%s", pid, tid, tiempo) + ".dmp"

	return nombre
}

func DumpMemory(writer http.ResponseWriter, requestCliente *http.Request) {
	var request globals.HiloMemoria

	err := json.NewDecoder(requestCliente.Body).Decode(&request) // Decodificamos la respuesta del cliente dentro de la variable request.

	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest) // Error el decodificar el request del cliente.
		slog.Warn("Error al decodificar el request del cliente en la recepción del mensaje DumpMemory")
		return
	}

	slog.Info("## Memory Dump solicitado - (PID:TID) - (<" + strconv.Itoa(request.PID) + ">:<" + strconv.Itoa(request.TID) + ">)")

	err = CreacionDeArchivoMemoryDump(request)

	if err != nil {
		utils_general.EnviarStatusFAILED(writer)
	}

	utils_general.EnviarStatusOK(writer) // Por el momento, unicamente devuelve un estado "OK".
}

func BytesAEnviar(proceso globals.HiloMemoria) ([]byte){ // Si bien el struct se llama HiloMemoria, usamos la estructura para poder usar esta funcion en CrearDump
	bytesAenviar := []byte{}

	for _, procesoEnLista := range globals.ListaContextosEjecucionProcesos{
		if proceso.PID == procesoEnLista.PID{
			for i := procesoEnLista.ContextoEjecucionProceso.Base; i < procesoEnLista.ContextoEjecucionProceso.Limite; i++{
				bytesAenviar = append(bytesAenviar, globals.Memoria[i])
			}
			return bytesAenviar
		}
	}
	return nil
}

func CrearDump(request globals.HiloMemoria, dump globals.SolicitudDump) (globals.SolicitudDump, error) {
	dump.Nombre = NombreArchivoMetadata(request)
	dump.HiloSolicitante = request.TID
	dump.ProcesoSolicitante = request.PID
	dump.ContenidoAgrabar.ContenidoBytes = BytesAEnviar(request)
	//dump.ContenidoAgrabar.Hilos = CrearListaDeHilos(request)

	tamanio, err := EncontrarTamanioDeProceso(request)
	if err != nil {
		slog.Error("Error obteniendo tamaño del proceso.")
		return globals.SolicitudDump{}, err
	}

	dump.Tamanio = tamanio
	return dump, nil
}

// Obsoleta luego de modificar que es lo que envia memoria al FS

// func CrearListaDeHilos(request globals.HiloMemoria) ([]globals.Hilo){ // Creamos una lista con todos los 
// 	var listaHilosDelProceso []globals.Hilo
// 	for _, hilo := range(globals.ListaContextosEjecucionHilos){
// 		if hilo.PID == request.PID{
// 			listaHilosDelProceso = append(listaHilosDelProceso, *hilo)
// 		}
// 	}
	
// 	return listaHilosDelProceso
// }

func EncontrarTamanioDeProceso(request globals.HiloMemoria) (int, error){
	for _, proceso := range(globals.ListaContextosEjecucionProcesos){
		if proceso.PID == request.PID{
			slog.Debug("El tamaño del proceso a hacer dump es: " + strconv.Itoa(proceso.Tamanio))

			return proceso.Tamanio, nil
		}
	}
	return -1, fmt.Errorf("no se halló el tamaño del proceso al querer dumpear")
}

func CreacionDeArchivoMemoryDump(request globals.HiloMemoria) error {
	var dump globals.SolicitudDump

	dump, err := CrearDump(request, dump)
	if err != nil {
		slog.Error("Error obteniendo tamaño del proceso.")
		return err
	}

	dumpJson, err0 := json.Marshal(dump)
	if err0 != nil {
		slog.Error("Error convirtiendo dump a JSON: " + err0.Error())
	}

	url := "http://" + globals.Config.Ip_filesystem + ":" + strconv.Itoa(globals.Config.Port_filesystem) + "/memoryDump"

	// Crear la solicitud HTTP POST
	req, err := http.NewRequest("POST", url, bytes.NewReader(dumpJson))
	if err != nil {
		slog.Error("Error creando solicitud")
		return fmt.Errorf("error creando solicitud: %w", err)
	}

	client := &http.Client{}
	response, err := client.Do(req)

	if err != nil {
		slog.Error("Error enviando solicitud")
		return fmt.Errorf("error enviando solicitud %w", err)
	}

	defer response.Body.Close()

	// Verificar el código de estado HTTP
	if response.StatusCode == http.StatusBadRequest {
		slog.Error("Respuesta con código 400 - Bad Request")
		return fmt.Errorf("error: bad request")
	} 
	// Decodificar la respuesta si el código es 200
	var result struct {
		Success string `json:"success"`
	}

	if err := json.NewDecoder(response.Body).Decode(&result); err != nil {
		slog.Error("Error decodificando respuesta.")
		return fmt.Errorf("error decodificando respuesta: %w", err)
	}

	// Si todo es exitoso, devolver nil indicando que no hubo errores
	return nil
}