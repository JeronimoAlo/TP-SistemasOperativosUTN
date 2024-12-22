package utils

import (
	"cpu/globals"
	//"cpu/utils"
	"encoding/json"
	"log/slog"
	"net/http"
	//"strconv"
	//"strings"
	"utils_general"
	//"io"
	//"bytes"
	
)

func IO_HTTP(w http.ResponseWriter, r *http.Request){
	// Obtenemos los parámetros de la URL.
	// Ejemplo: /IO/tidAhacerIO/pidPadreHiloAHacerIO/tiempoEnMilisegundos
	/*
	params := strings.Split(r.URL.Path, "/")
	if len(params) < 5 {
		http.Error(w, "Parametros insuficientes.", http.StatusBadRequest) // Devolvemos cómo respuesta que los parpametros son insuficientes.
		slog.Debug("Parámetros insuficientes al llamar IO_HTTP")
		return
	}

	tidAhacerIO, err := strconv.Atoi(params[2])
	if err != nil {
		http.Error(w, "TID a hacer io invalido.", http.StatusBadRequest)
	}

	pidPadreHiloAHacerIO, err := strconv.Atoi(params[3])
	if err != nil {
		http.Error(w, "PID del hilo a hacer io inválido.", http.StatusBadRequest)
		return
	}

	tiempoEnMilisegundos,err2 := strconv.Atoi(params[4])
	if err2!= nil{
		http.Error(w, "Tiempo a hacer io invalido.", http.StatusBadRequest)
	}
	*/


	var ioAHacer globals.IOparaHacer
	
	err := json.NewDecoder(r.Body).Decode(&ioAHacer) // Decodificamos la respuesta del cliente dentro de la variable request.

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest) // Error el decodificar el request del cliente.
		
		slog.Warn("Error al decodificar el request del cliente en la recepción del mensaje. Recibir Interrupcion")
		return
	}

	utils_general.EnviarStatusOK(w)
	
	slog.Debug("Envie estatus OK y esta pendiente la IO")
	go IO(ioAHacer.TID, ioAHacer.PID, ioAHacer.TiempoEnMilisegundos)


}