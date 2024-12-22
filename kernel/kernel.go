package main

import (
	"kernel/globals"
	"kernel/utils"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"utils_general"
)

func main() {
	// Inicializamos el archivo de configuración .json para el seteo de variables globales.
	globals.Config = utils.IniciarConfiguracion("config.json")

	// Configuramos el logger.
	utils_general.ConfigurarLogger("kernel.log", globals.Config.Log_level)

	// Verificamos si se pasaron los argumentos necesarios
	if len(os.Args) < 3 {
		slog.Error("Faltan parámetros para iniciar el kernel, chequear que se haya pasado el nombre del archivo en pseudocodigo y el tamanio del proceso inicial.")
	}

	var rutaArchivoInstrucciones = "../instrucciones/"
	//rutaArchivoInstrucciones =  rutaArchivoInstrucciones + os.Args[1] + ".pseudo" 
	rutaArchivoInstrucciones =  rutaArchivoInstrucciones + os.Args[1] // Primer argumento: Ruta al archivo de instrucciones
	tamanioProcesoInicial, err := strconv.Atoi(os.Args[2]) // Segundo argumento: Tamaño del proceso

	if err != nil {
		slog.Error("El tamaño del proceso inicial no es un número válido.")
		return
	}

	utils.InicioDelSistema(rutaArchivoInstrucciones, tamanioProcesoInicial)

	// Seteamos multiplexor para el servidor.
	mux := http.NewServeMux()
	mux.HandleFunc("/mensaje", utils.RecibirMensajeKernel)
	mux.HandleFunc("/PROCESS_CREATE/", utils.PROCESS_CREATE_HTTP)
	mux.HandleFunc("/PROCESS_EXIT/", utils.PROCESS_EXIT_HTTP)
	mux.HandleFunc("/THREAD_CREATE/", utils.THREAD_CREATE_HTTP)
	mux.HandleFunc("/THREAD_JOIN/", utils.THREAD_JOIN_HTTP)
	mux.HandleFunc("/THREAD_CANCEL/", utils.THREAD_CANCEL_HTTP)
	mux.HandleFunc("/THREAD_EXIT/", utils.THREAD_EXIT_HTTP)
	mux.HandleFunc("/MUTEX_CREATE/", utils.MUTEX_CREATE_HTTP)
	mux.HandleFunc("/MUTEX_LOCK/", utils.MUTEX_LOCK_HTTP)
	mux.HandleFunc("/MUTEX_UNLOCK/", utils.MUTEX_UNLOCK_HTTP)
	mux.HandleFunc("/DUMP_MEMORY/", utils.DUMP_MEMORY_HTTP)
	mux.HandleFunc("/IO/", utils.IO_HTTP)
	mux.HandleFunc("/desalojoHilo", utils.RecibirFinEjecucion)
	mux.HandleFunc("/SEGMENTATION_FAULT", utils.SEGMENTATION_FAULT_HTTP)

	go func() {
		//Iniciamos el servidor
		utils_general.IniciarServer(globals.Config.Port, mux)
	}()
	select {} // Bloqueamos el programa para que no termine
}