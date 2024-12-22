package utils_general

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"net/http"
	"strconv"
	"bytes"
	"globals_general"
	"encoding/json"
)	

type StatusRespuesta struct{
	Status string
}

// Defino el struct del body request que vamos a usar para enviar y recibir mensajes.
type BodyRequest struct {
	Mensaje string `json:"mensaje"`
	Emisor string `json:"emisor"`
	TipoSolicitud string `json:"tipoSolicitud"`
}

func ConfigurarLogger(ruta_log string, log_level string) {
	//Creamos el archivo "modulo".log en modo escritura, si ocurre algún error finalizamos con panic.
	logFile, err := os.OpenFile(ruta_log, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)

	if err != nil {
		panic(err)
	}

    // Usa io.MultiWriter para escribir a múltiples destinos: consola y archivo.
    multiWriter := io.MultiWriter(os.Stdout, logFile)

	// Convertir el Log_Level del config al tipo slog.Level.
	level, err := convertirLogLevelDeConfig(log_level)

    // Crear un nuevo manejador de registros (handler) para colocar.
    handler := slog.NewTextHandler(multiWriter, &slog.HandlerOptions{
        Level:level, // Toma el valor del level_log que se define arriba con convertirLogLevelDeConfig.
    })

	// Configura slog para que use el manejador que creamos anteriormente.
	slog.SetDefault(slog.New(handler))

	// Escribimos en el log el warning que obtenemos por no setear el log_level
	if err != nil {
		slog.Warn(err.Error())
	}

	slog.Debug("Se ha configurado correctamente el logger y el archivo de configuración")
}

// El objetivo de esta función es seterar dinámicamente el nivel de log que deseamos tener en el sistema.
func convertirLogLevelDeConfig(levelStr string) (slog.Level, error) {
	switch levelStr {
	case "DEBUG":
		return slog.LevelDebug, nil
	case "INFO":
		return slog.LevelInfo, nil
	case "WARN":
		return slog.LevelWarn, nil
	case "ERROR":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("no hay un Log_level seteado, se coloca INFO por default")
	}
}

func IniciarServer(port int, mux *http.ServeMux) {
	err := http.ListenAndServe(":" + strconv.Itoa(port), mux) // Levantamos el servidor.
	if err != nil {
		slog.Error(err.Error())
		return
	}
}

//Devuelve un error, en caso que lo haya.
func EnviarMensaje(IP_receptor string, port int, mensaje []byte, emisor string, endpoint string) ([]byte, error){
	cliente := &http.Client{} // Creamos la variable de tipo *http.Client para hacer el envío del mensaje.
	url := "http://" + IP_receptor + ":" + strconv.Itoa(port) + endpoint // Seteamos la URL donde vamos a enviar.
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(mensaje))  // Generamos la request de tipo POST y generamos un array de bytes con el mensaje a enviar.

	if err != nil {
		slog.Error(err.Error()) // Devolvemos error si hay un problema al generar la http request.
		return []byte{0}, err
	}
	
	slog.Debug("HTTP request POST generada correctamente.")

	req.Header.Set("Content-Type", "application/json") // Seteamos que el body a enviar en el header es un JSON.
	respuesta, err := cliente.Do(req)                  // Ejecuta la request y almacena la respuesta.

	if err != nil {
		//slog.Error(err.Error()) // Error en la request en caso de haberlo.
		return []byte{0}, err
	}
	
	slog.Debug("Request enviada correctamente al servidor.")

	// Verificar el código de estado de la respuesta.
	if respuesta.StatusCode != http.StatusOK {
		//errMsg := fmt.Sprintf("Error: %d", respuesta.StatusCode)
		//slog.Error(errMsg) //TODO: Borre estos logs para que no jodan en la entrega.
		err = fmt.Errorf("status Error")
		return []byte{0}, err
	}

	bodyBytes, err := io.ReadAll(respuesta.Body) // Leemos la respuesta que nos dió el servidor.

	if err != nil {
		slog.Error(err.Error()) // Log del error en caso de haberlo.
		return bodyBytes, err
	}
	
	return bodyBytes, nil
}

// Función para encontrar el máximo de una lista de enteros.
func Max(nums []int) (int, error) {
	if len(nums) == 0 {
		return 0, nil
	}

	max := nums[0]
	for _, num := range nums[1:] {
		if num > max {
			max = num
		}
	}
	return max, nil
}

func EnviarRespuesta(writer http.ResponseWriter, respuesta []byte) {
	writer.WriteHeader(http.StatusOK)
	writer.Write(respuesta)
	slog.Debug("Se envio la respuesta al cliente con exito desde Enviar Respuesta")
}

func EnviarStatusOK(writer http.ResponseWriter) {
	var mensajeExito globals_general.MensajeResultado

	mensajeExito.Status = "OK"

	respuesta, err := json.Marshal(mensajeExito)

	if err != nil {
		http.Error(writer, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		slog.Warn("Error al codificar los datos como JSON")
		return
	}
	
	EnviarRespuesta(writer, respuesta)
}

func EnviarStatusFAILED(writer http.ResponseWriter) {
	var mensajeFAILED globals_general.MensajeResultado

	mensajeFAILED.Status = "FAILED"

    respuesta, err := json.Marshal(mensajeFAILED)
    if err != nil {
        http.Error(writer, "Error al codificar los datos como JSON", http.StatusInternalServerError)
        slog.Warn("Error al codificar los datos como JSON")
        return
    }

	writer.WriteHeader(http.StatusBadRequest)
	writer.Write(respuesta)
	slog.Debug("No se envió la respuesta correctamente desde EnviarStatusFAILED")
}