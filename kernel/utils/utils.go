package utils

import (
	"bufio"
	"encoding/json"
	"fmt"
	"kernel/globals"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"unicode"
	"utils_general"
)

/////////////////////////////////////////////////////////////////////////////////////////////////////////////
//////////////////INICIO DEL SISTEMA A PARTIR DE UN ARCHIVO DE INSTRUCCIONES ESPECIFICADO////////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////////////

func InicioDelSistema(rutaArchivoInstrucciones string, tamanioProcesoInicial int) {
	pcbNuevo := CrearProceso(0, tamanioProcesoInicial)
	globals.ProcesosDelSistema.Procesos = append(globals.ProcesosDelSistema.Procesos, pcbNuevo) // Agregamos el nuevo proceso a la lista de procesos del sistema.

	instrucciones, _ := cargarCodigoDesdeArchivo(rutaArchivoInstrucciones) // Carga las instrucciones del archivo de pseudocodigo en una lista de instrucciones

	CrearHilo(pcbNuevo.PID, 0, 0, instrucciones)

	EncolarProceso(pcbNuevo, &globals.ColaNewProcesos, globals.New) // Encola el proceso en new.
	VerificarColaNew()
}

/////////////////////////////////////////////////////////////////////////////////////////////////////////////
///////////////LECTURA DE ARCHIVO DE CONFIGURACIÓN Y DEFINICIÓN DE API PARA ENVÍO DE MENSAJE/////////////////
/////////////////////////////////////////////////////////////////////////////////////////////////////////////

// Función que recibe cómo parámetro la ubicación del .json con las configuraciones
func IniciarConfiguracion(filePath string) *globals.KernelConfig {
	var config *globals.KernelConfig
	configFile, err := os.Open(filePath) // Abrimos el .json y lo guardamos en la variable "configFile".

	if err != nil {
		fmt.Println(err.Error())
	}
	defer configFile.Close() // Cerramos el archivo y liberamos los recursos.

	jsonParser := json.NewDecoder(configFile) // Leemos y decodificamos datos .json.
	jsonParser.Decode(&config)                // Decodifica y convierte los datos leídos del .json en la estructura definida para el manejo de configs.

	return config
}

func RecibirMensajeKernel(writer http.ResponseWriter, requestCliente *http.Request) {
	var request utils_general.BodyRequest // Variable local para almacenar el request decodificado.

	err := json.NewDecoder(requestCliente.Body).Decode(&request) // Decodificamos la respuesta del cliente dentro de la variable request.

	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest) // Error el decodificar el request del cliente.
		slog.Warn("Error al decodificar el request del cliente en la recepción del mensaje.")
		return
	}
	slog.Debug("Se decodifico el request del cliente con exito")

	respuesta, err := json.Marshal(fmt.Sprintf("Hola %s! Soy el kernel. El mensaje que enviaste fue: %s", request.Emisor, request.Mensaje)) // Convierte el mensaje de respuesta en formato JSON

	if err != nil {
		http.Error(writer, "Error al codificar los datos como JSON", http.StatusInternalServerError)
		slog.Warn("Error al codificar los datos como JSON")
		return
	}
	writer.WriteHeader(http.StatusOK)
	writer.Write(respuesta)
	slog.Debug("Se envio la respuesta al cliente con exito")
}

/*
El formato de cada instrucción dentro del archivo deberá ser el siguiente: "OPERACION PARAM1 PARAM2"
*/
// Cargamos el código de un proceso o hilo desde un archivo.
func cargarCodigoDesdeArchivo(filepath string) ([]globals.Instruccion, error) {
	rutaCompleta := "../instrucciones/" + filepath

	file, err := os.Open(rutaCompleta) // Tratamos de abrir el archivo con la ruta que nos enviaron desde la CPU.

	if err != nil {
		slog.Error("Error abriendo archivo: " + err.Error())
		return nil, err // Si hay algún error abriendo el archivo, lo devolvemos.
	}

	defer file.Close() // Programamos cerrar el archivo al final de la función.

	var instrucciones []globals.Instruccion // Variable para almacenar las instrucciones.

	scanner := bufio.NewScanner(file) // Levantamos un scanner para leer línea por línea del archivo.
	for scanner.Scan() {              // Recorre todo el archivo.
		linea := scanner.Text() // Levantamos la línea actual que leyó el scanner.

		// Ignorar líneas vacías.
		if len(linea) == 0 {
			continue
		}

		// Partir la línea en Operacion y Parámetros manualmente.
		var operacion string    // Para almacenar la operación.
		var parametros []string // Para almacenar los parámetros (Son 2).
		word := ""              // Cadena temporal para acumular caracteres.
		parsingParams := false  // Para diferenciar parámetros de operaciones.

		for i := 0; i < len(linea); i++ { // Iteramos sobre cada caracter de la línea.
			c := rune(linea[i]) // Almacenamos el caracter en ASCII/UNICODE.

			if unicode.IsSpace(c) { // Si leemos un espacio es que terminamos de leer una palabra.
				if word != "" {
					if parsingParams { // Si estamos en la fase de parámetros...
						parametros = append(parametros, word) // Agregamos la palabra a la lista.
					} else { // Sino, es una operacion y la próxima palabra que vamos a leer es una operación.
						operacion = word
						parsingParams = true
					}
					word = "" // Reiniciamos la palabra a leer
				}
			} else { // Si no leímos un espacio, sumamos ese caracter al conjunto word (Palabra que estamos leyendo).
				word += string(c)
			}
		}

		// Agregar el último parámetro, si es que lo hay (Esto es por si no pusimos un espacio al final del último parámetro.)
		if word != "" {
			// Misma lógica que antes.
			if parsingParams {
				parametros = append(parametros, word)
			} else {
				operacion = word
			}
		}

		instruccion := globals.Instruccion{
			Operacion:  operacion,
			Parametros: parametros,
		}
		instrucciones = append(instrucciones, instruccion) // Copiamos el struct instrucciones a instruccion.
	}

	if err := scanner.Err(); err != nil {

		slog.Error("Error abriendo archivo: " + err.Error())
		return nil, err
	}

	return instrucciones, nil // Si sale todo bien no devolvemos ningún error y las instrucciones leídas.
}

// Función que devuelve el puntero al PCB del proceso padre dado un TID.
func EncontrarPCBPorTID(tid int, pidPadre int) *globals.PCB { 
	// Iteramos sobre todos los procesos en el sistema.
	for _, proceso := range globals.ProcesosDelSistema.Procesos {
		// Verificamos si el TID está en la lista de TIDs del proceso (Por cada proceso dentro del sistema, itero sobre sus TIDs).
		for _, tcbTid := range proceso.TIDs {
			if tcbTid == tid && pidPadre == proceso.PID {
				// Si encontramos el TID, retornamos el puntero al PCB del proceso padre.
				return proceso
			}
		}
	}

	// Si no se encuentra el TID, retornamos nil y un log de advertencia
	slog.Warn("No se encontró un proceso padre para el TID:" + strconv.Itoa(tid))
	return nil
}

// Funcion que devuelve al PCB del proceso padre dado un PID.
func EncontrarPCBPorPID(pid int) *globals.PCB { 
	for _, proceso := range globals.ProcesosDelSistema.Procesos {
		if proceso.PID == pid {
			return proceso
		}
	}

	// Si no se encuentra el PID, retornamos nil y un log de advertencia.
	slog.Warn("No se encontró un proceso padre para el PID:" + strconv.Itoa(pid))
	return nil
}

// Funcion que devuelve al TCB del hilo dado un TID.
func EncontrarTCBPorTID(tid int, pidPadre int) *globals.TCB { 
	for _, hilo := range globals.HilosDelSistema.Hilos {
		if hilo.TID == tid && hilo.PID == pidPadre {
			return hilo
		}
	}

	// Si no se encuentra el TID, retornamos nil y un log de advertencia.
	slog.Warn("No se encontró un hilo para el TID: " + strconv.Itoa(tid) + " dentro del PID: " + strconv.Itoa(pidPadre))
	return nil
}

// Función auxiliar para verificar si un TID está en la lista de TIDs del proceso.
func existeTIDEnProceso(tids []int, tid int) bool {
	for _, t := range tids {
		if t == tid {
			return true
		}
	}
	return false
}