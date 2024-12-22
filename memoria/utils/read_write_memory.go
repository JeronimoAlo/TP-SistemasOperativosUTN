package utils

import (
	"encoding/json"
	"log/slog"
	"memoria/globals"
	"net/http"
	"strconv"
	"utils_general"
	"time"
)

// InicializarMemoria inicializa la variable Memoria con el tamaño especificado en la configuración global.
func InicializarMemoria() {
	globals.Memoria = make([]byte, globals.Config.Memory_size) // Memoria de usuario.
}

func READ_MEM(writer http.ResponseWriter, requestCliente *http.Request) {
	var request globals.LecturaEscrituraMemoria

	// Decodificamos el request en la variable request
	err := json.NewDecoder(requestCliente.Body).Decode(&request)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		slog.Warn("Error al decodificar el request del cliente en la recepción del mensaje. READ_MEM")
		return
	}

	// Obtener la dirección física como índice
	index := int(request.DireccionFisica)

	// Verificamos que haya al menos 4 bytes disponibles desde esa posición
	if index < 0 || index+4 > len(globals.Memoria) {
		utils_general.EnviarStatusFAILED(writer)
		return
	}

	// Leer los 4 bytes en orden como están en la memoria y combinarlos como uint32
	valor := uint32(globals.Memoria[index]) | //Primer byte: Lo dejamos en su posición sin desplazamiento, ya que es el byte menos significativo.
		uint32(globals.Memoria[index+1])<<8 | //Segundo byte: Desplazamos 8 bits hacia la izquierda (<< 8) para que ocupe la posición de "segundo byte" en el uint32.
		uint32(globals.Memoria[index+2])<<16 | //Tercer byte: Desplazamos 16 bits hacia la izquierda (<< 16) para que ocupe la posición de "tercer byte" en el uint32.
		uint32(globals.Memoria[index+3])<<24 //Cuarto byte: Desplazamos 24 bits hacia la izquierda (<< 24) para que ocupe la posición de "cuarto byte" en el uint32.
	// Combinamos todos estos bytes utilizando el operador |, que realiza una operación "OR bit a bit". Esto nos permite ensamblar los 4 bytes en un solo valor uint32

	var valorAresponder globals.ValorAresponder
	// Devolvemos el valor como JSON
	valorAresponder.Valor = valor
	slog.Debug("Valor pedido de read_mem es: " + strconv.Itoa( int(valorAresponder.Valor)))
	mensaje, err := json.Marshal(valorAresponder)
    if err != nil {
        slog.Error("Fallo al codificar la estructura valorAresponder.")
    }
	
	// Retraso de respuesta
	time.Sleep(time.Duration(globals.Config.Response_delay) * time.Millisecond)

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	writer.Write(mensaje)

	slog.Info("## <Lectura> - (PID:TID) - (<" + strconv.Itoa(request.PID) + ">:<" + strconv.Itoa(request.TID) + ">) - Dir. Física: <" + strconv.Itoa(int(request.DireccionFisica)) + "> - Tamaño: <4BYTES>") //FIXME: si no es esto, a que se refiere con tamaño?? Creo que esta bien.
}

func WRITE_MEM(writer http.ResponseWriter, requestCliente *http.Request) {
	var request globals.LecturaEscrituraMemoria // Contiene el valor que queremos escribir y la dirección física

	// Decodificar el request
	err := json.NewDecoder(requestCliente.Body).Decode(&request)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest)
		slog.Warn("Error al decodificar el request del cliente en la recepción del mensaje. WRITE_MEM")
		return
	}

	// Convertir la dirección física a un índice en memoria.
	index := int(request.DireccionFisica)

	if request.ValorAenviar == 0 {
		if index < 0 || index >= len(globals.Memoria) {
		http.Error(writer, "Dirección fuera de los límites de la memoria", http.StatusBadRequest)
		slog.Error("Error: Dirección fuera del rango de memoria al escribir un byte")
		utils_general.EnviarStatusFAILED(writer)
		return
		}
	} else {
		if index < 0 || index+3 >= len(globals.Memoria) {
		http.Error(writer, "Dirección fuera de los límites de la memoria", http.StatusBadRequest)
		slog.Error("Error: Dirección fuera del rango de memoria al escribir un uint32")
		utils_general.EnviarStatusFAILED(writer)
	return
		}
	}
	slog.Debug("Direccion a escribir en memoria: " + strconv.Itoa(int(request.DireccionFisica)))
	slog.Debug("Valor recibido en memoria: " + strconv.Itoa(int(request.ValorAenviar)))
	// Descomponer el valor `uint32` en 4 bytes y escribirlos en Memoria.
	globals.Memoria[index] = byte(request.ValorAenviar)         // Primer byte
	globals.Memoria[index+1] = byte(request.ValorAenviar >> 8)  // Segundo byte
	globals.Memoria[index+2] = byte(request.ValorAenviar >> 16) // Tercer byte
	globals.Memoria[index+3] = byte(request.ValorAenviar >> 24) // Cuarto byte 

	// Log detallado de los valores escritos
	slog.Debug("## Valores escritos en memoria:")
	slog.Debug("Memoria[" + strconv.Itoa(index) + "] = " + strconv.Itoa(int(globals.Memoria[index])))
	slog.Debug("Memoria[" + strconv.Itoa(index+1) + "] = " + strconv.Itoa(int(globals.Memoria[index+1])))
	slog.Debug("Memoria[" + strconv.Itoa(index+2) + "] = " + strconv.Itoa(int(globals.Memoria[index+2])))
	slog.Debug("Memoria[" + strconv.Itoa(index+3) + "] = " + strconv.Itoa(int(globals.Memoria[index+3])))
	
	// Retraso en la respuesta.
	time.Sleep(time.Duration(globals.Config.Response_delay) * time.Millisecond)

	// Confirmación de escritura exitosa.
	utils_general.EnviarStatusOK(writer)

	slog.Info("## <Escritura> - (PID:TID) - (<" + strconv.Itoa(request.PID) + ">:<" + strconv.Itoa(request.TID) + ">) - Dir. Física: <" + strconv.Itoa(int(request.DireccionFisica)) + "> - Tamaño: <4BYTES>")
}