package utils

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func PROCESS_CREATE_HTTP(w http.ResponseWriter, r *http.Request) {
	// Obtenemos los parámetros de la URL.
	// Ejemplo: /PROCESS_CREATE/rutaArchivo/Tamaño/PrioridadHilo.

	// Separamos la ruta del archivo, tamaño y prioridad del hilo de los parámetros en la URL.
	params := strings.Split(r.URL.Path, "/")
	if len(params) < 5 {
		slog.Debug("Entre al bad request 1")
		http.Error(w, "Parametros insuficientes.", http.StatusBadRequest) // Devolvemos cómo respuesta que los parpametros son insuficientes.
		slog.Debug("Parámetros insuficientes al llamar PROCESS_CREATE_HTTP")
		return
	}

	rutaArchivo := params[2]
	rutaCompleta := "../instrucciones/" + rutaArchivo

	_, err := os.Open(rutaCompleta) // Tratamos de abrir el archivo con la ruta que nos enviaron desde la CPU.

	if err != nil {
		slog.Debug("Entre al bad request 2")
		http.Error(w, "Ruta del archivo inválida", http.StatusBadRequest)
		return // Si hay algún error abriendo el archivo, lo devolvemos.
	}

	tamanio, err := strconv.Atoi(params[3])
	if err != nil {
		slog.Debug("Entre al bad request 3")
		http.Error(w, "Tamaño inválido", http.StatusBadRequest)
		return
	}

	prioridadHilo, err := strconv.Atoi(params[4])
	if err != nil {
		slog.Debug("Entre al bad request 4")
		http.Error(w, "Prioridad inválida", http.StatusBadRequest)
		return
	}

	// Llamamos a la función original.
	PROCESS_CREATE(rutaArchivo, tamanio, prioridadHilo)

	// Devolver una respuesta HTTP.
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Proceso creado con éxito"))
}

func PROCESS_EXIT_HTTP(w http.ResponseWriter, r *http.Request) {
	// Obtenemos los parámetros de la URL.
	// Ejemplo: /PROCESS_EXIT/tid/pidPadre.

	// Separamos la ruta del archivo, tamaño y prioridad del hilo de los parámetros en la URL.
	params := strings.Split(r.URL.Path, "/")
	if len(params) < 4 {
		http.Error(w, "Parametros insuficientes.", http.StatusBadRequest) // Devolvemos cómo respuesta que los parpametros son insuficientes.
		slog.Debug("Parámetros insuficientes al llamar PROCESS_EXIT_HTTP")
		return
	}

	tid, err := strconv.Atoi(params[2])
	if err != nil {
		http.Error(w, "TID Inválido", http.StatusBadRequest)
		return
	}

	// Validamos que sea el hilo con TID 0 el que está solicitando finalizar el proceso.
	if tid != 0 {
		http.Error(w, "ERROR: Únicamente el TID 0 puede finalizar el proceso padre. LE PASASTE "+strconv.Itoa(tid)+".", http.StatusBadRequest)
		return
	}

	pidPadre, err := strconv.Atoi(params[3])
	if err != nil {
		http.Error(w, "PID Inválido", http.StatusBadRequest)
		return
	}

	// Llamamos a la función original.
	PROCESS_EXIT(tid, pidPadre)

	// Devolver una respuesta HTTP.
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Proceso finalizado con éxito"))
}