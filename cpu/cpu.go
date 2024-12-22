package main

import(
	"cpu/utils"
	"cpu/globals"
	"utils_general"
	"net/http"
)

func main() {
	// Inicializamos el archivo de configuración .json para el seteo de variables globales.
	globals.Config = utils.IniciarConfiguracion("config.json")

	// Configuramos el logger
	utils_general.ConfigurarLogger("cpu.log", globals.Config.Log_level)

	// Seteamos multiplexor para el servidor.
	mux := http.NewServeMux()
	mux.HandleFunc("/mensaje", utils.RecibirMensajeCPU)
	mux.HandleFunc("/ejecucionTIDyPID", utils.RecibirTIDyPIDaEjecutar)
	mux.HandleFunc("/interrupciones", utils.RecibirInterrupcion)

	//Iniciamos el servidor
	utils_general.IniciarServer(globals.Config.Port, mux)
	
	// go func() {
	// 	//Iniciamos el servidor
	// 	utils_general.IniciarServer(globals.Config.Port, mux)
	// }()
	// select {} // Bloqueamos el programa para que no termine



	
}