package utils

import (
	"log/slog"
	"net/http"
	"strconv"
	"strings"
)

func MUTEX_CREATE_HTTP(w http.ResponseWriter, r *http.Request){
	// Obtenemos los parámetros de la URL.
	// Ejemplo: /MUTEX_CREATE/pidACrearMutex/nombreRecursoAProteger.
	
	params := strings.Split(r.URL.Path, "/")
	if len(params) < 4 {
		http.Error(w, "Parametros insuficientes.", http.StatusBadRequest) // Devolvemos cómo respuesta que los parpametros son insuficientes.
		slog.Debug("Parámetros insuficientes al llamar MUTEX_CREATE_HTTP")
		return
	}

	pidACrearMutex, err := strconv.Atoi(params[2])
	if err != nil {
		http.Error(w, "PID inválido.", http.StatusBadRequest)
		return
	}
	
	nombreDeMutex := params[3]

	MUTEX_CREATE(pidACrearMutex, nombreDeMutex)

	w.WriteHeader(http.StatusOK)
    w.Write([]byte("Mutex creado correctamente."))
}

func MUTEX_LOCK_HTTP(w http.ResponseWriter, r *http.Request){
	// Obtenemos los parámetros de la URL.
	// Ejemplo: /MUTEX_LOCK/tidALockearMutex/pidALockearMutex/NombreRecurso.

	params := strings.Split(r.URL.Path, "/")
	if len(params) < 5 {
		http.Error(w, "Parametros insuficientes.", http.StatusBadRequest) // Devolvemos cómo respuesta que los parpametros son insuficientes.
		slog.Debug("Parámetros insuficientes al llamar MUTEX_LOCK_HTTP")
		return
	}

	tidALockearMutex, err1 := strconv.Atoi(params[2])
	if err1 != nil {
		http.Error(w, "TID inválido.", http.StatusBadRequest)
		return
	}

	pidALockearMutex, err2 := strconv.Atoi(params[3])
	if err2 != nil {
		http.Error(w, "PID inválido.", http.StatusBadRequest)
		return
	}

	nombreRecurso := params[4]

	MUTEX_LOCK(tidALockearMutex, pidALockearMutex, nombreRecurso)

	// Devolver una respuesta HTTP.
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("MUTEX LOCK realizado correctamente."))
}

func MUTEX_UNLOCK_HTTP(w http.ResponseWriter, r *http.Request){
	// Obtenemos los parámetros de la URL.
	// Ejemplo: /MUTEX_UNLOCK/tidADesockearMutex/pidADesockearMutex/mutex.

	params := strings.Split(r.URL.Path, "/")
	if len(params) < 5 {
		http.Error(w, "Parametros insuficientes.", http.StatusBadRequest) // Devolvemos cómo respuesta que los parpametros son insuficientes.
		slog.Debug("Parámetros insuficientes al llamar MUTEX_UNLOCK_HTTP")
		return
	}

	tidADesockearMutex, err1 := strconv.Atoi(params[2])
	if err1 != nil {
		http.Error(w, "TID inválido.", http.StatusBadRequest)
		return
	}

	pidADesockearMutex, err2 := strconv.Atoi(params[3])
	if err2 != nil {
		http.Error(w, "PID inválido.", http.StatusBadRequest)
		return
	}

	mutex := params[4]

	MUTEX_UNLOCK(tidADesockearMutex,pidADesockearMutex,mutex)

	// Devolver una respuesta HTTP.
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("MUTEX UNLOCK realizado correctamente."))
}