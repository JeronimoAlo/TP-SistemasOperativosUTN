package utils

import (
	"log/slog"
	"memoria/globals"
	"net/http"
	"strconv"
)

//Funcion que realiza la compactación de Memoria cuando la solicita el kernel.
func Compactar_HTTP(writer http.ResponseWriter, requestCliente *http.Request) {
	// Verificamos si la memoria está configurada como DINÁMICA.
	if globals.Config.Scheme != "DINAMICAS" {
		http.Error(writer, `{"status": "failed", "error": "Memoria no configurada como dinámica"}`, http.StatusBadRequest)
		slog.Error("Intento de compactación en memoria no dinámica.")
		return
	}

	slog.Debug("Estado de la particiones antes de compactar:")
	for i, particion := range globals.Particiones{
							
		slog.Debug("Particion Numero " + strconv.Itoa(i) + 
		", Base : " + strconv.Itoa(int(particion.Base)) + 
		", Tamanio: " + strconv.Itoa(int(particion.Tamanio)) + 
		" y Limite: " + strconv.Itoa(int(particion.Limite)) + 
		" Estado: " + strconv.FormatBool(particion.Libre))

	}
	// Iniciamos el proceso de compactación.
	if CompactarMemoria() {
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte(`{"status": "success"}`))

		slog.Info("Compactación completada con éxito.")
		slog.Debug("Estado de la particiones luego de compactar:")
		for i, particion := range globals.Particiones{
							
			slog.Debug("Particion Numero " + strconv.Itoa(i) + 
			", Base : " + strconv.Itoa(int(particion.Base)) + 
			", Tamanio: " + strconv.Itoa(int(particion.Tamanio)) + 
			" y Limite: " + strconv.Itoa(int(particion.Limite)) + 
			" Estado: " + strconv.FormatBool(particion.Libre))

		}
	} else {
		http.Error(writer, `{"status": "failed"}`, http.StatusInternalServerError)
		slog.Error("Error durante la compactación de memoria.")
	}
}

func CompactarMemoria() bool {
	var memoriaCompactada []*globals.Particion
	espacioLibreTotal := 0
	var baseLibre uint32 = 0

	// Crear un mapeo de las particiones antiguas a las nuevas.
	particionMap := make(map[*globals.Particion]*globals.Particion)

	// Recorremos las particiones y compactamos las libres.
	for _, particion := range globals.Particiones {
		if particion.Libre {
			espacioLibreTotal += particion.Tamanio // Sumamos el espacio libre.
		} else {
			// Reubicamos la partición ocupada.
			nuevaParticion := &globals.Particion{
				Tamanio: particion.Tamanio,
				Libre:   false,
				Base:    baseLibre,
				Limite:  int(baseLibre) + particion.Tamanio - 1,
			}

			// Guardar la relación entre la partición antigua y la nueva.
			particionMap[particion] = nuevaParticion

			memoriaCompactada = append(memoriaCompactada, nuevaParticion)
			baseLibre += uint32(particion.Tamanio)
		}
	}

	// Agrega una única partición libre con el espacio total al final.
	if espacioLibreTotal > 0 {
		particionLibre := &globals.Particion{
			Tamanio: espacioLibreTotal,
			Libre:   true,
			Base:    baseLibre,
			Limite:  int(baseLibre) + espacioLibreTotal - 1,
		}
		memoriaCompactada = append(memoriaCompactada, particionLibre)
	}

	// Actualizar la lista global de particiones.
	globals.Particiones = memoriaCompactada

	// Actualizar las referencias de las particiones en los procesos.
	for _, proceso := range globals.ListaContextosEjecucionProcesos {
		if proceso.ParticionAsignada != nil {
			if nuevaParticion, ok := particionMap[proceso.ParticionAsignada]; ok {
				proceso.ParticionAsignada = nuevaParticion
			}
		}
	}

	return true
}
