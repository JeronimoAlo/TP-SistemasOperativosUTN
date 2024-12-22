package utils

import (
	"cpu/globals"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"utils_general"
    "io"
	"strconv"
)

// Función que recibe cómo parámetro la ubicación del .json con las configuraciones
func IniciarConfiguracion(filePath string) *globals.CpuConfig {
	var config *globals.CpuConfig
	configFile, err := os.Open(filePath) // Abrimos el .json y lo guardamos en la variable "configFile".

	if err != nil {
		fmt.Println(err.Error())
	}
	defer configFile.Close() // Cerramos el archivo y liberamos los recursos.

	jsonParser := json.NewDecoder(configFile) // Leemos y decodificamos datos .json.
	jsonParser.Decode(&config)                // Decodifica y convierte los datos leídos del .json en la estructura definida para el manejo de configs.

	return config
}

func RecibirMensajeCPU(writer http.ResponseWriter, requestCliente *http.Request) {
	var request utils_general.BodyRequest // Variable local para almacenar el request decodificado.

	err := json.NewDecoder(requestCliente.Body).Decode(&request) // Decodificamos la respuesta del cliente dentro de la variable request.

	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest) // Error el decodificar el request del cliente.
		slog.Warn("Error al decodificar el request del cliente en la recepción del mensaje.")
		return
	}
	slog.Debug("Se decodifico el request del cliente con exito")

	respuesta, err := json.Marshal(fmt.Sprintf("Hola %s! Soy la CPU. El mensaje que enviaste fue: %s", request.Emisor, request.Mensaje)) // Convierte el mensaje de respuesta en formato JSON

	if err != nil {
		http.Error(writer, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		slog.Warn("Error al codificar los datos como JSON")
		return
	}
	writer.WriteHeader(http.StatusOK)
	writer.Write(respuesta)
	slog.Debug("Se envio la respuesta al cliente con exito.")
}

func EnviarRespuesta(writer http.ResponseWriter, respuesta []byte) {
	writer.WriteHeader(http.StatusOK)
	writer.Write(respuesta)
	slog.Debug("Se envio la respuesta al cliente con exito.")
}

func RecibirTIDyPIDaEjecutar(writer http.ResponseWriter, requestCliente *http.Request){ // Recibe el TID y PID a ejecutar desde el kernel.
    var request globals.TIDyPIDaEjecutar // Variable local para almacenar el request decodificado.

	err := json.NewDecoder(requestCliente.Body).Decode(&request) // Decodificamos la respuesta del cliente dentro de la variable request.

	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest) // Error el decodificar el request del cliente.
		slog.Warn("Error al decodificar el request del cliente en la recepción del mensaje. RecibirMensajeMemoria")
		return
	}

	slog.Debug("Se decodifico el request del cliente con exito. RecibirTIDyPIDaEjecutar")

    globals.TIDyPIDenEjecucion.PID = request.PID // Setea en la variable global el PID y el TID en ejecucion en la CPU.
    globals.TIDyPIDenEjecucion.TID = request.TID
	
	slog.Debug("Meti el PID " + strconv.Itoa(globals.TIDyPIDenEjecucion.PID) + " y el TID " + strconv.Itoa(globals.TIDyPIDenEjecucion.TID) + " desde la funcion RecibirTIDyPIDaEjecutar()")

	utils_general.EnviarStatusOK(writer)

	go EjecutarHilo() // Comienza el ciclo de ejecucion de las instrucciones del hilo.
}

func EnviarMensajePorParametro(url string) error {
    // Creamos una solicitud HTTP GET.
    client := &http.Client{}
    req, err := http.NewRequest("GET", url, nil)

	slog.Debug("url a hacer el get: " + url)

    if err != nil {
        return err
    }

    // Enviamos la solicitud.
    resp, err := client.Do(req)
	slog.Debug("Termine de mandar el request")


    if err != nil {
        return err
    }

    defer resp.Body.Close()

    // Verificamos el código de estado.
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("error en la solicitud: %s", resp.Status)
    }

    // Leemos el cuerpo de la respuesta.
    body, err := io.ReadAll(resp.Body)

	slog.Debug("El json de la respuesta de enviar por parametro es: "+ string(body))

    if err != nil {
        return fmt.Errorf("error al leer el cuerpo de la respuesta: %v", err)
    }

     // Loggeamos el contenido de la respuesta.
     slog.Debug("Contenido de la respuesta", "response", string(body))

    // Todo fue exitoso.
      return nil
}

func EnviarStatusOK(writer http.ResponseWriter) {
	var mensajeExito globals.MensajeResultado

	mensajeExito.Status = "OK"

	respuesta, err := json.Marshal(mensajeExito)

	if err != nil {
		http.Error(writer, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		slog.Warn("Error al codificar los datos como JSON")
		return
	}
	EnviarRespuesta(writer, respuesta)
}