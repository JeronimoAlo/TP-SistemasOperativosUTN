package main

import (
	"memoria/globals"
	"memoria/utils"
	"net/http"
	"utils_general"
)

func main() {
	// Inicializamos el archivo de configuración .json para el seteo de variables globales.
	globals.Config = utils.IniciarConfiguracion("config.json")

	// Configuramos el logger
	utils_general.ConfigurarLogger("memoria.log", globals.Config.Log_level)

	utils.InicializarMemoria() // Iniciamos la memoria de usuario.

	utils.CrearMemoria() // Creamos las particiones de la memoria segun el esquema del config.

	// Seteamos multiplexor para el servidor.
	mux := http.NewServeMux()
	mux.HandleFunc("/mensaje", utils.RecibirMensajeMemoria)
	mux.HandleFunc("/memoriaDisponible", utils.ChequearEspacioMemoria)
	mux.HandleFunc("/compactar", utils.Compactar_HTTP)
	mux.HandleFunc("/obtenerContexto", utils.ObtenerContextoDeEjecucion)
	mux.HandleFunc("/actualizarContexto", utils.ActualizarContextoDeEjecucion)
	mux.HandleFunc("/creacionDeHilo", utils.CreacionDeHilo)
	mux.HandleFunc("/finalizacionDeProceso", utils.FinalizacionDeProceso)
	mux.HandleFunc("/finalizacionDeHilo", utils.FinalizacionDeHilo)
	mux.HandleFunc("/obtenerInstruccion", utils.ObtenerInstruccion)
	mux.HandleFunc("/leerMemoria", utils.READ_MEM)
	mux.HandleFunc("/escribirMemoria", utils.WRITE_MEM)
	mux.HandleFunc("/memoryDump", utils.DumpMemory)

	go func() {
		//Iniciamos el servidor
		utils_general.IniciarServer(globals.Config.Port, mux)
	}()
	select {} // Bloqueamos el programa para que no termine
}
