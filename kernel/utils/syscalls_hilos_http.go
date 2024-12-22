package utils

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func THREAD_CREATE_HTTP(w http.ResponseWriter, r *http.Request) {
	// Obtenemos los parámetros de la URL.
	// Ejemplo: /THREAD_CREATE/rutaArchivo/prioridadHilo/pidPadre.

	// Separamos la ruta del archivo, tamaño y prioridad del hilo de los parámetros en la URL.
	params := strings.Split(r.URL.Path, "/")
	if len(params) < 5 {
		http.Error(w, "Parametros insuficientes.", http.StatusBadRequest) // Devolvemos cómo respuesta que los parpametros son insuficientes.
		slog.Debug("Parámetros insuficientes al llamar THREAD_CREATE_HTTP")
		return
	}

	rutaArchivo := params[2]
	rutaCompleta := "../instrucciones/" + rutaArchivo

	_, err := os.Open(rutaCompleta) // Tratamos de abrir el archivo con la ruta que nos enviaron desde la CPU.

	if err != nil {
		http.Error(w, "Ruta del archivo inválida", http.StatusBadRequest)
		return // Si hay algún error abriendo el archivo, lo devolvemos.
	}

	prioridadHilo, err := strconv.Atoi(params[3])
	if err != nil {
		http.Error(w, "Prioridad de hilo inválida.", http.StatusBadRequest)
		return
	}

	pidPadre, err := strconv.Atoi(params[4])
	if err != nil {
		http.Error(w, "PID padre inválido.", http.StatusBadRequest)
		return
	}

	// Llamamos a la función original.
	THREAD_CREATE(rutaArchivo, prioridadHilo, pidPadre)

	// Devolver una respuesta HTTP.
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Hilo creado con éxito."))
}

func THREAD_JOIN_HTTP(w http.ResponseWriter, r *http.Request) {
	// Obtenemos los parámetros de la URL.
	// Ejemplo: /THREAD_JOIN/tidAEsperar/tidAbloquear/pidPadreHiloABloquear.

	// Separamos la ruta del archivo, tamaño y prioridad del hilo de los parámetros en la URL.
	params := strings.Split(r.URL.Path, "/")
	if len(params) < 5 {
		http.Error(w, "Parametros insuficientes.", http.StatusBadRequest) // Devolvemos cómo respuesta que los parpametros son insuficientes.
		slog.Debug("Parámetros insuficientes al llamar THREAD_JOIN_HTTP")
		return
	}

	tidAEsperar, err := strconv.Atoi(params[2])
	if err != nil {
		http.Error(w, "TID a esperar inválido.", http.StatusBadRequest)
		return
	}

	tidAbloquear, err := strconv.Atoi(params[3])
	if err != nil {
		http.Error(w, "TID a bloquear inválido.", http.StatusBadRequest)
		return
	}

	pidPadreHiloABloquear, err := strconv.Atoi(params[4])
	if err != nil {
		http.Error(w, "PID padre del hilo a bloquear inválido.", http.StatusBadRequest)
		return
	}

	// Llamamos a la función original.
	THREAD_JOIN(tidAEsperar, tidAbloquear, pidPadreHiloABloquear)

	// Devolver una respuesta HTTP.
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Thread join realizado correctamente."))
}

func THREAD_CANCEL_HTTP(w http.ResponseWriter, r *http.Request) {
	// Obtenemos los parámetros de la URL.
	// Ejemplo: /THREAD_CANCEL/tidSolicitante/tidACancelar/pidPadreHiloACancelar.

	// Separamos la ruta del archivo, tamaño y prioridad del hilo de los parámetros en la URL.
	params := strings.Split(r.URL.Path, "/")
	if len(params) < 5 {
		http.Error(w, "Parametros insuficientes.", http.StatusBadRequest) // Devolvemos cómo respuesta que los parpametros son insuficientes.
		slog.Debug("Parámetros insuficientes al llamar THREAD_CANCEL_HTTP")
		return
	}

	tidSolicitante, err := strconv.Atoi(params[2])
	if err != nil {
		http.Error(w, "TID solicitante inválido.", http.StatusBadRequest)
	}

	tidACancelar, err := strconv.Atoi(params[3])
	if err != nil {
		http.Error(w, "TID a cancelar inválido.", http.StatusBadRequest)
		return
	}

	pidPadreHiloACancelar, err := strconv.Atoi(params[4])
	if err != nil {
		http.Error(w, "PID padre del hilo a cancelar inválido.", http.StatusBadRequest)
		return
	}

	// Llamamos a la función original.
	THREAD_CANCEL(tidSolicitante, tidACancelar, pidPadreHiloACancelar)

	// Devolver una respuesta HTTP.
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Thread cancel realizado correctamente."))
}

func THREAD_EXIT_HTTP(w http.ResponseWriter, r *http.Request) {
	// Obtenemos los parámetros de la URL.
	// Ejemplo: /THREAD_EXIT/tidAFinalizar/pidPadreHiloAFinalizar.

	// Separamos la ruta del archivo, tamaño y prioridad del hilo de los parámetros en la URL.
	params := strings.Split(r.URL.Path, "/")
	if len(params) < 4 {
		http.Error(w, "Parametros insuficientes.", http.StatusBadRequest) // Devolvemos cómo respuesta que los parpametros son insuficientes.
		slog.Debug("Parámetros insuficientes al llamar THREAD_EXIT_HTTP")
		return
	}

	tidAFinalizar, err := strconv.Atoi(params[2])
	if err != nil {
		http.Error(w, "TID a finalizar inválido.", http.StatusBadRequest)
	}

	pidPadreHiloAFinalizar, err := strconv.Atoi(params[3])
	if err != nil {
		http.Error(w, "PID del hilo a finalizar inválido.", http.StatusBadRequest)
		return
	}

	// Llamamos a la función original.
	THREAD_EXIT(tidAFinalizar, pidPadreHiloAFinalizar)

	// Devolver una respuesta HTTP.
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Thread exit realizado correctamente."))
}