package utils

import (
	"encoding/json"
	"filesystem/globals"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"
	"utils_general"
	"bytes"
	"encoding/gob"
	"path/filepath"
)

// Función que recibe cómo parámetro la ubicación del .json con las configuraciones
func IniciarConfiguracion(filePath string) *globals.FileSystemConfig {
	var config *globals.FileSystemConfig
	configFile, err := os.Open(filePath) // Abrimos el .json y lo guardamos en la variable "configFile".

	if err != nil {
		slog.Error("Error al abrir archivo de configuración con el path: " + filePath)
		panic(err) // Finalizar si no se puede cargar la configuración
	}

	defer configFile.Close() // Cerramos el archivo y liberamos los recursos.

	jsonParser := json.NewDecoder(configFile) // Leemos y decodificamos datos .json.
	jsonParser.Decode(&config) // Decodifica y convierte los datos leídos del .json en la estructura definida para el manejo de configs.

	return config
}

// Inicia el FileSystem, creando los archivos bitmap.dat y bloques.dat en caso de no existir
func InicializarArchivos() {
	// Ruta a los archivos
	globals.BitmapPath = fmt.Sprintf("%s/bitmap.dat", globals.Config.Mount_dir)
	globals.BloquesPath = fmt.Sprintf("%s/bloques.dat", globals.Config.Mount_dir)

	// Crear el directorio mount_dir si no existe
	if err := os.MkdirAll(globals.Config.Mount_dir, os.ModePerm); err != nil {
		slog.Error("Error al crear el directorio mount_dir", "error", err)
		return
	}

	// Validar y crear bitmap.dat si no existe
	if _, err := os.Stat(globals.BitmapPath); os.IsNotExist(err) {
		file, err := os.Create(globals.BitmapPath)
		if err != nil {
			slog.Error("Error al crear bitmap.dat", "error", err)
			return
		}

		// Inicializar el archivo con ceros
		tamañoBitmap := uint32((globals.Config.Block_count + 7) / 8) // Redondeo hacia arriba.
		file.Write(make([]byte, tamañoBitmap))
		file.Close()
	}

	// Validar y crear bloques.dat si no existe
	if _, err := os.Stat(globals.BloquesPath); os.IsNotExist(err) {
		file, err := os.Create(globals.BloquesPath)
		if err != nil {
			slog.Error("Error al crear bloques.dat", "error", err)
			return
		}

		// Inicializar el archivo con ceros
		tamañoBloques := uint32(globals.Config.Block_count) * globals.Config.Block_size
		file.Write(make([]byte, tamañoBloques))
		file.Close()

	}

	slog.Debug("Inicialice los archivos bitmap.dat y bloques.dat")
}

// Leer BitMap.
func LeerBitmap() ([]byte, error) {
	data, err := os.ReadFile(globals.BitmapPath) // Lee el contenido del archivo y lo devuelve como un slice de bytes.

	if err != nil {
		return nil, err
	}
	return data, nil
}

// Escribir Archivo BitMap.
func ActualizarBitmap(bitmap []byte) error {
	return os.WriteFile(globals.BitmapPath, bitmap, 0644)
}

// Función para encontrar el primer bloque libre.
func EncontrarBloqueLibre(bitmap *[]byte) (int, error) {
	for i, byteValue := range *bitmap {
		for bit := 0; bit < 8; bit++ {
			if byteValue&(1<<bit) == 0 {
				return i*8 + bit, nil
			}
		}
	}

	return -1, fmt.Errorf("no hay bloques libres disponibles")
}

func AsignarBloque(bitmap *[]byte, nombreArchivo string) int {
	bloque, err := EncontrarBloqueLibre(bitmap)

	if err != nil {
		slog.Error("## Error: No hay bloques libres - Archivo: " + nombreArchivo)
		return -1
	}
	
	MarcarBloque(bitmap, bloque)
	slog.Info("## Bloque asignado: " + strconv.Itoa(bloque) + " - Archivo: " + nombreArchivo)

	return bloque
}

// Marcar un bloque como ocupado.
func MarcarBloque(bitmap *[]byte, bloque int) {
	byteIndex := bloque / 8
	bitIndex := bloque % 8
	(*bitmap)[byteIndex] |= (1 << bitIndex) // Desreferencia el puntero y establece el bit correspondiente en 1
}

func CrearArchivoMetadata(indexBlock uint32, size uint32, nombreArchivo string) error {
    // Verificamos si existe la carpeta /files
    filesDir := fmt.Sprintf("%s/files", globals.Config.Mount_dir)
    if _, err := os.Stat(filesDir); os.IsNotExist(err) {
        if err := os.MkdirAll(filesDir, os.ModePerm); err != nil {
            slog.Error("Error al crear directorio /files", "path", filesDir, "error", err)
            return err
        }
    }
	
    metadata := globals.Metadata{
        IndexBlock: indexBlock,
        Size:       size,
    }
	
	slog.Debug("Voy a meter en el .dmp IndexBLock: " + strconv.Itoa(int(metadata.IndexBlock)) + " y el size: " + strconv.Itoa(int(metadata.Size)))

    // Obtener la ruta completa del archivo
    filePath := fmt.Sprintf("%s/%s", filesDir, nombreArchivo)

    // Crear los directorios que componen el path del archivo
    dirPath := filepath.Dir(filePath) // Extraer la parte del directorio del path
    if _, err := os.Stat(dirPath); os.IsNotExist(err) {
        if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
            slog.Error("Error al crear directorio para el archivo", "path", dirPath, "error", err)
            return err
        }
    }

    // Verificamos si existe el archivo
    if _, err := os.Stat(filePath); os.IsNotExist(err) {
        // Creamos el archivo
        file, err := os.Create(filePath)
        if err != nil {
            slog.Error("Error al crear archivo de metadata", "filePath", filePath, "error", err)
            return err
        }
        defer file.Close()

        // Escribir metadata en formato JSON
        encoder := json.NewEncoder(file)
        if err := encoder.Encode(metadata); err != nil {
            slog.Error("Error al escribir metadata en archivo", "filePath", filePath, "error", err)
            return err
        }

        slog.Info("## Archivo Creado: " + nombreArchivo + " - Tamaño: <" + strconv.Itoa(int(size)) + ">")
    } else {
        slog.Warn("El archivo ya existe, no se sobrescribirá", "filePath", filePath)
    }

    return nil
}


// Funcion que recibe la solicitud de memory dump desde la memoria, llama a creacionArchivoDump para crear el archivo.
func RecibirMemoryDump(writer http.ResponseWriter, requestCliente *http.Request) {
	slog.Debug("Recibi una solicitud de DUMP_MEMORY")

	var request globals.SolicitudDump

	err := json.NewDecoder(requestCliente.Body).Decode(&request) // Decodificamos la respuesta del cliente dentro de la variable request.

	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest) // Error el decodificar el request del cliente.
		slog.Warn("Error al decodificar el request del cliente en la recepción del mensaje. DumpMemory")
		return
	}

	err = CreacionArchivoDump(request.ContenidoAgrabar, uint32(request.Tamanio), request.Nombre)

	if err!= nil {
		slog.Error("Error al crear el archivo de dump " + err.Error())
		http.Error(writer, "Error al crear el archivo de DUMP: " + err.Error(), http.StatusBadRequest)
		return
	}

	utils_general.EnviarStatusOK(writer) // Por el momento, unicamente devuelve un estado "OK".
	slog.Info("## Fin de solicitud - Archivo:"+request.Nombre)
}

// Escribe el contenido Bloque por Bloque
func EscribirEnBloque(rutaBloques string, bloque int, contenido []byte) error {
	file, err := os.OpenFile(rutaBloques, os.O_WRONLY, 0644)
	
	if err != nil {
		return err
	}
	defer file.Close() //Asegura que se ejecute file.close() al final de función

	offset := bloque * len(contenido)
	file.Seek(int64(offset), 0)
	_, err = file.Write(contenido)
	if err != nil {
		return err
	}

	slog.Info("## Acceso Bloque - Bloque File System: " + strconv.Itoa(bloque))
	return nil
}

//Funcion que recorre el bitmap enviado por parametro y cuenta la cantidad de bloques que esten libres, osea, bits en 0
func CantidadDeBloquesLibres(bitmap *[]byte) int{
	cantidadDeBloquesLibres := 0

	// Recorremos cada byte del bitmap
	for _, byteValue := range *bitmap {
		// Recorremos los 8 bits de cada byte
		for i := 0; i < 8; i++ {
			// Si el bit en la posición i es 0, incrementamos el contador
			if (byteValue & (1 << i)) == 0 {
				cantidadDeBloquesLibres++
			}
		}
	}
	return cantidadDeBloquesLibres
}

//Crea el archivo del dump memory solicitado por la memoria.
func CreacionArchivoDump(datosParaGuardarEnDump globals.ProcesoDump, tamanioEnBytes uint32, nombreArchivo string) error{
	bitmap, err  := LeerBitmap()

	if(err!= nil){
		slog.Error("Error al leer el archivo de bitmap")
		return err
	}

	cantidadDeBloquesLibres := CantidadDeBloquesLibres(&bitmap)

	slog.Debug("Cantidad de bloques libres antes de hacer el dump: " + strconv.Itoa(cantidadDeBloquesLibres))

	cantBloquesAOcuparPorArchivo := int(tamanioEnBytes / globals.Config.Block_size) // Se determina por el tamanio del archivo y el tamanio del bloque
	if tamanioEnBytes%globals.Config.Block_size != 0{
		cantBloquesAOcuparPorArchivo++ //Esto es por si la division da con resto, entonces ocupa un bloque mas
	}

	slog.Debug("Cantidad de bloques a ocupar por archivo antes de hacer el dump: " + strconv.Itoa(cantBloquesAOcuparPorArchivo))
	
	tamanioDePuntero := 4 // Tamanio del puntero a bloques en bytes (Indicado por el enunciado)
	tamanioTablaPunteros := tamanioDePuntero * cantBloquesAOcuparPorArchivo // La necesitamos para saber cuantos bloques de punteros se necesitan.
	cantBloquesDePunterosNecesarios := tamanioTablaPunteros / int(globals.Config.Block_size)
	if tamanioTablaPunteros%int(globals.Config.Block_size) != 0{
		cantBloquesDePunterosNecesarios++ //Esto es por si la division da con resto, entonces ocupa un bloque mas
	}

	slog.Debug("Cantidad de bloques de punteros a ocupar por archivo antes de hacer el dump: " + strconv.Itoa(cantBloquesDePunterosNecesarios))

	var err2 error
	if cantidadDeBloquesLibres < cantBloquesAOcuparPorArchivo + cantBloquesDePunterosNecesarios { // Verificamos que haya la cantidad de bloques necesarios para almacenar los bloques de datos y de punteros
		slog.Error("No espacio suficiente para almacenar el archivo DUMP.")
		err2 = fmt.Errorf("no espacio suficiente para almacenar el archivo DUMP")
		slog.Debug(fmt.Sprintf("Bloques libres: %d, Bloques requeridos: %d", cantidadDeBloquesLibres, cantBloquesAOcuparPorArchivo+cantBloquesDePunterosNecesarios))

		return err2
	}

	//Asignamos un bloque a la tabla de indice
	bloqueAsignadoIndice:=  AsignarBloque(&bitmap, nombreArchivo)

	var indiceDeIndice globals.IndiceDeBloques
	indiceDeIndice.PunterosABloques = append(indiceDeIndice.PunterosABloques, uint32(bloqueAsignadoIndice)) //Esto es solo para poder usar la funcion de escribir en el bloque

	//Asignacion de bloques al archivo y al archivo de indice
	var indice globals.IndiceDeBloques

	for j := 0; j < cantBloquesAOcuparPorArchivo; j++ {
		bloqueAsignadoArchivo := AsignarBloque(&bitmap, nombreArchivo)
		if bloqueAsignadoArchivo == -1 {
			slog.Error("error al asignar bloque para el archivo")
			return fmt.Errorf("error al asignar bloque para el archivo")
		}
		indice.PunterosABloques = append(indice.PunterosABloques, uint32(bloqueAsignadoArchivo))
	}
	dataIndice , err1 := SerializarIndice(indice)
	if err1 != nil {
        slog.Error("Error al serializar datos para el el indice.")
        return err1
	}

	//Escribimos en el bloque del indice asignado, los datos del indice
	err4 := EscribirArchivoEnBloques(indiceDeIndice, dataIndice, int(globals.Config.Block_size), globals.BloquesPath, "INDICE", nombreArchivo, 0)

	if err4 != nil {
		slog.Error("Error al escribir el bloque de indice.")
        return err4
	}

	//Se crea el archivo de metadata con los datos del bloqeu de indice y el tamanio del proceso
	err6 := CrearArchivoMetadata(uint32(bloqueAsignadoIndice), tamanioEnBytes, nombreArchivo)
	if err6 != nil {
		slog.Error("Error al crear el archivo Metadata.")
        return err6
	}

	//Escribimos los bloques asignados a ese archivo
	err3 := EscribirArchivoEnBloques(indice, datosParaGuardarEnDump.ContenidoBytes, int(globals.Config.Block_size),globals.BloquesPath,"DATOS",nombreArchivo,uint32(bloqueAsignadoIndice))
	if err3 != nil {
		slog.Error("Error al escribir el archivo en los bloques")
		return err3
	}

	// Guardar bitmap actualizado
    err = EscribirBitmap(bitmap)
    if err != nil {
        slog.Error("Error al guardar el bitmap actualizado: " + err.Error())
        return err
    }

    slog.Info("Archivo dump creado exitosamente")
	return nil
}

func EscribirBitmap(bitmap []byte) error {
    // Ruta del archivo donde se guarda el bitmap
    archivoBitmap := globals.BitmapPath

    // Abrir el archivo en modo escritura
    file, err := os.OpenFile(archivoBitmap, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
    if err != nil {
		slog.Error("No se pudo abrir el archivo del bitmap.")
        return fmt.Errorf("no se pudo abrir el archivo del bitmap: %w", err)
    }
    defer file.Close()

    // Escribir los datos del bitmap en el archivo
    _, err = file.Write(bitmap)
    if err != nil {
		slog.Error("Error al escribir en el archivo del bitmap")
        return fmt.Errorf("error al escribir en el archivo del bitmap: %w", err)
    }

    return nil
}


func SerializarProcesoDump(proceso globals.ProcesoDump) ([]byte, error) {
    buffer := new(bytes.Buffer)
    encoder := gob.NewEncoder(buffer)
    err := encoder.Encode(proceso)

    if err != nil {
        return nil, err
    }

    return buffer.Bytes(), nil
}


func SerializarIndice(indice globals.IndiceDeBloques)([]byte, error) {
    buffer := new(bytes.Buffer)
    encoder := gob.NewEncoder(buffer)
    err := encoder.Encode(indice.PunterosABloques)

    if err != nil {
        return nil, err
    }

    return buffer.Bytes(), nil
}

//Esta funcion escribe los datos que le envies de array de bytes en los bloques que este asignados en el indice que le envies
func EscribirArchivoEnBloques(indice globals.IndiceDeBloques, data []byte, blockSize int, filePath string, tipoDeBloqueAEscribir string , nombreArchivo string, bloqueIndice uint32) error {
    file, err := os.OpenFile(filePath, os.O_WRONLY, 0644)

    if err != nil {
        return fmt.Errorf("error al abrir el archivo: %w", err)
	}
    defer file.Close()

    // Índice de datos que se están escribiendo
    dataIndex := 0
/*
	if tipoDeBloqueAEscribir == "DATOS" {
		slog.Info("## Acceso Bloque - Archivo:" +  nombreArchivo + "- Tipo Bloque: INDICE - Bloque File System "+ strconv.Itoa(int(bloqueIndice)))
	}
*/	
	//Tiempo de acceso a 1 bloque de datos a escribir
	// esta una vez mas el time sleep porque en el enunciado dice:deberán tener N+1 accesos a bloques, donde N es el número de bloques de datos del archivo y
	// los retardos deberán realizarse luego de acceder a cada bloque.

    // Iterar sobre los bloques asignados
    for _, bloqueID := range indice.PunterosABloques {
        if dataIndex >= len(data) {
			dataIndex = 0
            //break // No hay más datos para escribir
        }

        // Calcular la posición del bloque en el archivo
        offset := int64(bloqueID) * int64(blockSize)

        // Moverse al offset correspondiente
        _, err := file.Seek(offset, 0)

        if err != nil {
            return fmt.Errorf("error al mover el puntero de archivo: %w", err)
        }

        // Calcular cuántos bytes escribir en este bloque
        bytesRestantes := len(data) - dataIndex
        bytesPorEscribir := blockSize

        if bytesRestantes < blockSize {
            bytesPorEscribir = bytesRestantes
        }
		if tipoDeBloqueAEscribir == "INDICE" {
			slog.Info("## Acceso Bloque - Archivo:" +  nombreArchivo + "- Tipo Bloque: INDICE - Bloque File System "+ strconv.Itoa(int(indice.PunterosABloques[0])) )
		}else{
			slog.Info("## Acceso Bloque - Archivo:" +  nombreArchivo + "- Tipo Bloque: DATOS - Bloque File System "+ strconv.Itoa(int(bloqueID)) )
		}
		
        // Escribir los datos al bloque
        _, err = file.Write(data[dataIndex : dataIndex+bytesPorEscribir])
        if err != nil {
            return fmt.Errorf("error al escribir en el bloque %d: %w", bloqueID, err)
        }

        // Actualizar el índice de datos
        dataIndex += bytesPorEscribir
		time.Sleep( time.Duration(globals.Config.Block_access_delay) * time.Millisecond) //Tiempo de acceso a 1 bloque de datos a escribir
    }
	
    return nil
}