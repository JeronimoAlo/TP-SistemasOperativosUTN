package main

import (
	"filesystem/utils"
	"filesystem/globals"
	"utils_general"
	"net/http"
)

func main() {
	// Inicializamos el archivo de configuración .json para el seteo de variables globales.
	globals.Config = utils.IniciarConfiguracion("config.json")

	// Configuramos el logger
	utils_general.ConfigurarLogger("filesystem.log", globals.Config.Log_level)

	// Inicializamos los archivos necesarios para el FS.
	utils.InicializarArchivos()

	// Seteamos multiplexor para el servidor.
	mux := http.NewServeMux()
	
	//mux.HandleFunc("/mensaje", utils.RecibirMensajeFileSystem)
	mux.HandleFunc("/memoryDump", utils.RecibirMemoryDump)

	//Iniciamos el servidor
	utils_general.IniciarServer(globals.Config.Port, mux)
}
