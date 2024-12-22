package utils

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"memoria/globals"
	"net/http"
	"os"
	"strconv"
	"utils_general"
	"time"
	"globals_general"
)

// Función que recibe cómo parámetro la ubicación del .json con las configuraciones
func IniciarConfiguracion(filePath string) *globals.MemoryConfig {
	var config *globals.MemoryConfig
	configFile, err := os.Open(filePath) // Abrimos el .json y lo guardamos en la variable "configFile".

	if err != nil {
		fmt.Println(err.Error())
	}
	defer configFile.Close() // Cerramos el archivo y liberamos los recursos.

	jsonParser := json.NewDecoder(configFile) // Leemos y decodificamos datos .json.
	jsonParser.Decode(&config)                // Decodifica y convierte los datos leídos del .json en la estructura definida para el manejo de configs.

	return config
}

func EnviarRespuestaCompactable(writer http.ResponseWriter) {
	respuesta := globals_general.CompactacionRespuesta{
		Status:      "FAILED",
		Compactable: true,
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
	json.NewEncoder(writer).Encode(respuesta)
}

func RecibirMensajeMemoria(writer http.ResponseWriter, requestCliente *http.Request) {
	var request utils_general.BodyRequest // Variable local para almacenar el request decodificado.

	err := json.NewDecoder(requestCliente.Body).Decode(&request) // Decodificamos la respuesta del cliente dentro de la variable request.

	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest) // Error el decodificar el request del cliente.
		slog.Warn("Error al decodificar el request del cliente en la recepción del mensaje. RecibirMensajeMemoria")
		return
	}
	slog.Debug("Se decodifico el request del cliente con exito")

	respuesta, err := json.Marshal(fmt.Sprintf("Hola %s! Soy la memoria. El mensaje que enviaste fue: %s", request.Emisor, request.Mensaje)) // Convierte el mensaje de respuesta en formato JSON

	if err != nil {
		http.Error(writer, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		slog.Warn("Error al codificar los datos como JSON")
		return
	}

	writer.WriteHeader(http.StatusOK)
	writer.Write(respuesta)
	slog.Debug("Se envio la respuesta al cliente con exito desde RecibirMensajeMemoria")
}

func CreacionDeHilo(writer http.ResponseWriter, requestCliente *http.Request) {
	var request globals.HiloMemoria

	 err := json.NewDecoder(requestCliente.Body).Decode(&request) // Decodificamos la respuesta del cliente dentro de la variable request.

	 	if err != nil {
	  		http.Error(writer, err.Error(), http.StatusBadRequest) // Error el de codificar el request del cliente.
	   		slog.Warn("Error al decodificar el request del cliente en la recepción del mensaje. CreacionDeHilo")
	  	 	return
	   	}

	AgregarHiloAMemoria(request) //Agrega el hilo a la lista de hilos en mmemoria y le setea los valores en 0 a sus registros
	utils_general.EnviarStatusOK(writer) // Devuelve un "OK", no necesito ninguna validacion
} 

func ObtenerInstruccion(writer http.ResponseWriter, requestCliente *http.Request){
	var request globals.SolicitudInstruccion //	Como tendremos que recibir el hilo (TID) y el PC (registro), recibiremos una estructura donde ambos esten contenidos.

	err := json.NewDecoder(requestCliente.Body).Decode(&request) // Decodificamos la respuesta del cliente dentro de la variable request.

	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest) // Error el decodificar el request del cliente.
		slog.Warn("Error al decodificar el request del cliente en la recepción del mensaje. (Origen: ObtenerInstruccion)")
		return
	}

	slog.Debug("Mensaje listo para ser enviado como OK")

	hilo := HiloAPartirDeTID(request.TID, request.PID)

	// Tiempo de retardo en la respuesta.
	time.Sleep(time.Duration(globals.Config.Response_delay) * time.Millisecond)

	var pcInt int = int(request.PC)

	if pcInt < 0 || pcInt>= len(hilo.Codigo) { //Si el PC no esta en rango, devuelve error
		utils_general.EnviarStatusFAILED(writer)
		return
    }

	instruccion := hilo.Codigo[pcInt] // Obtenemos la instrucción actual.

	// Registrar la instrucción que se está obteniendo.
	slog.Info("## Obtener instrucción - (PID:TID) - (<" + strconv.Itoa(request.PID) + ">:<" + strconv.Itoa(request.TID) + ">) - Instrucción: <" + instruccion.Operacion + "> " + fmt.Sprint(instruccion.Parametros))
	
	var respuesta globals.InstruccionDesdeMemoria
	// Preparamos la respuesta a enviar al CPU
	respuesta.NoHayMasInstrucciones = pcInt == len(hilo.Codigo) + 1 //Confirmamos si es la ultima instruccion por ejecutar o no
	respuesta.Operacion = hilo.Codigo[pcInt].Operacion 
	respuesta.Parametros = hilo.Codigo[pcInt].Parametros
	
	//Si esta en rango,  Codificamos la respuesta con la instruccion como JSON y la enviamos.
	writer.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(writer).Encode(respuesta); err != nil {
		utils_general.EnviarStatusFAILED(writer)
		return
	}
}

func HiloAPartirDeTID(tid int, pid int) (*globals.Hilo) {
	for  _ , hilo := range globals.ListaContextosEjecucionHilos{
		
		if (hilo.TID == tid  && hilo.PID == pid){
			return hilo
		}
	}
	return nil 
}