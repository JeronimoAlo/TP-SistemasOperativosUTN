package utils

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

func DUMP_MEMORY_HTTP(w http.ResponseWriter, r *http.Request){
	// Obtenemos los parámetros de la URL.
	// Ejemplo: /DUMP_MEMORY/tidAHacerDump/pidPadreHiloAHacerDump

	params := strings.Split(r.URL.Path, "/")
	if len(params) < 4 {
		http.Error(w, "Parametros insuficientes.", http.StatusBadRequest) // Devolvemos cómo respuesta que los parpametros son insuficientes.
		slog.Debug("Parámetros insuficientes al llamar DUMP_MEMORY_HTTP")
		return
	}

	tidAHacerDump, err := strconv.Atoi(params[2])
	if err != nil {
		http.Error(w, "TID a hacer dump invalido.", http.StatusBadRequest)
	}

	pidPadreHiloAHacerDump, err := strconv.Atoi(params[3])
	if err != nil {
		http.Error(w, "PID del hilo a hacer dump inválido.", http.StatusBadRequest)
		return
	}

	DUMP_MEMORY(tidAHacerDump,pidPadreHiloAHacerDump)

	w.WriteHeader(http.StatusOK)
    w.Write([]byte("DUMP MEMORY realizado correctamente."))
}

func SEGMENTATION_FAULT_HTTP(w http.ResponseWriter, r *http.Request){
	// Obtenemos los parámetros de la URL.
	// Ejemplo: /SEGMENTATION_FAULT/tid/pidPadre.

	params := strings.Split(r.URL.Path, "/")
	if len(params) < 4 {
		http.Error(w, "Parametros insuficientes.", http.StatusBadRequest) // Devolvemos cómo respuesta que los parpametros son insuficientes.
		slog.Debug("Parámetros insuficientes al llamar SEGMENTATION_FAULT_HTTP")
		return
	}

	tidAFinalizar, err := strconv.Atoi(params[2])
	if err != nil {
		http.Error(w, "TID a finalizar invalido.", http.StatusBadRequest)
	}

	pidPadreAFinalizar, err := strconv.Atoi(params[3])
	if err != nil {
		http.Error(w, "PID del hilo a finalizar inválido.", http.StatusBadRequest)
		return
	}

	PROCESS_EXIT(tidAFinalizar, pidPadreAFinalizar)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("PROCCESS_EXIT realizado correctamente debido a SEGMENTATION_FAULT."))
}