package utils

import (
	"encoding/json"
	"log/slog"
	"memoria/globals"
	"net/http"
	"strconv"
	"utils_general"
	//"time"
)

func CrearMemoria(){
	switch globals.Config.Scheme{
	case "FIJAS":
		CrearParticionesFijas()
	case "DINAMICAS":
		CrearParticionesDinamicas()
	default:
		slog.Error("El metodo de partición definido en el archivo 'config.json' es inválido")
	}
}

func CrearParticionesFijas(){
	var tamanioAcumulable int = 0

	for i, tamanioParticion := range globals.Config.Partitions{
		particion := &globals.Particion{} // Declaramos una partición.

		particion.Tamanio = tamanioParticion
		particion.Base = uint32(tamanioAcumulable)
		particion.Libre = true
		particion.Limite = tamanioAcumulable + tamanioParticion - 1 // Instanciamos sus variables.

		globals.Particiones = append(globals.Particiones, particion) // Añadimos nuestra nueva partición a la lista de particiones.

		tamanioAcumulable += tamanioParticion // Hacemos que la base de nuestra proxima particion sea el ultimo+1 de la actual.

		slog.Debug("Se creó la partición " + strconv.Itoa(i) + " de base " + strconv.Itoa(int(particion.Base)) + " y tamaño " + strconv.Itoa(tamanioParticion))
	}
}

// Creamos una sola partición del tamaño total de la memoria.
func CrearParticionesDinamicas(){
	particionInicial := &globals.Particion{} // Declaramos la partición.

    particionInicial.Tamanio = globals.Config.Memory_size
    particionInicial.Base = 0
    particionInicial.Libre = true
    particionInicial.Limite = globals.Config.Memory_size - 1

    globals.Particiones = append(globals.Particiones, particionInicial)

    slog.Debug("Se creó una partición dinámica inicial de base 0 y tamaño " + strconv.Itoa(globals.Config.Memory_size) + "." )
}

// Devuelve un puntero a la partición asignada al proceso.
func HayEspacioEnMemoriaUsuario(proceso globals.ProcesoMemoria) (*globals.Particion){
	switch globals.Config.Search_algorithm{
    case "FIRST":
        return FirstFit(proceso)
	case "BEST":
		return BestFit(proceso)
	case "WORST":
		return WorstFit(proceso)
	default:
		slog.Error("El algoritmo de ubicacion definido en el archivo 'config.json' es inválido")
		return nil
	}
}

func FirstFit(proceso globals.ProcesoMemoria) (*globals.Particion){
	for _, particion := range globals.Particiones {
		if particion.Libre && proceso.Tamanio <= particion.Tamanio { // Si la particion está libre y el tamaño del proceso es menor al de la particion, asignamos dicha particion al proceso.
			return particion
		}
	}
	
	slog.Warn("No se encontró una partición con suficiente espacio para asignar al proceso.")
	return nil
}


func BestFit(proceso globals.ProcesoMemoria) *globals.Particion {
	var mejorParticion *globals.Particion
	mejorTamanio := -1
	
	for _, particion := range globals.Particiones { // Buscamos la partición más pequeña en la que pueda entrar un proceso.
		if particion.Libre && particion.Tamanio >= proceso.Tamanio {
			// Si aún no hay una mejorParticion o encontramos una mejor opción
			if mejorTamanio == -1 || particion.Tamanio < mejorTamanio {
				mejorParticion = particion
				mejorTamanio = particion.Tamanio
			}
		}
	}
	
	// Si no encuentro una partición, devuelvo un puntero nulo.
	if mejorParticion == nil {
		return nil
	}

	// Si no encuentro una partición, devuelvo un puntero nulo.
	return mejorParticion
}

func WorstFit(proceso globals.ProcesoMemoria) (*globals.Particion){
	var peorParticion *globals.Particion
	peorTamanio := -1 
	
	for _, particion := range globals.Particiones{ // Buscamos la particion mas grande en la que pueda entrar un proceso.
		if particion.Libre && particion.Tamanio >= proceso.Tamanio{
			if peorTamanio == -1 || particion.Tamanio > peorTamanio{
				peorParticion = particion
				peorTamanio = particion.Tamanio
			}
		} 
	}

	// Si no encuentro una partición, devuelvo un puntero nulo.
	if peorParticion == nil {
		return nil
	}
	
	return peorParticion
}

// Funcion para evaluar si sirve o no realizar la compactacion
func SumaDeHuecosLibres(proceso *globals.ProcesoMemoria) bool {
	suma := 0
	
	for _ , particion := range globals.Particiones {
		if suma < proceso.Tamanio { // Si el contador de Espacio es Menor al Tamaño del proceso y quedan Particiones a revisar, avanza.
			if particion.Libre { // Verifica si la particion analizada esta Libre
				suma += particion.Tamanio // Suma al contador el tamaño de la particion Libre
			}
		} 
		if suma >= proceso.Tamanio {
			return true // Si la sumatoria de huecos es mayoR o igual al tamaño del proceso devolvemos true y dejamos de recorrer.
		}
	}
	
	return false // Si la sumatoria de huecos es menor al tamaño del proceso devolvemos true.
}

func ChequearEspacioMemoria(writer http.ResponseWriter, requestCliente *http.Request) {
	var request globals.ProcesoMemoria

	// Decodificamos la respuesta del cliente dentro de la variable request.
	err := json.NewDecoder(requestCliente.Body).Decode(&request)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest) // Error el decodificar el request del cliente.
		slog.Warn("Error al decodificar el request del cliente en la recepción del mensaje. Chequear Espacio Memoria")
		return
	}

	slog.Debug("Se decodifico el request de chequeo de espacio en memoria del cliente con exito")
	
	// Llamar a la función para verificar espacio en memoria.
	particion := HayEspacioEnMemoriaUsuario(request)
	if particion != nil{
		slog.Debug("Base de Particion " + strconv.Itoa(int(particion.Base)))
	}

	if particion == nil {
		slog.Warn("No se encontró espacio suficiente para el proceso PID: " + strconv.Itoa(request.PID))
		
		// Verificamos si la memoria es compactable
		if globals.Config.Scheme == "DINAMICAS" { // Solo en caso de memoria dinámica.
			if SumaDeHuecosLibres(&request) {
				slog.Info("Memoria fragmentada, pero compactable para el proceso con PID: " + strconv.Itoa(request.PID))
				
				// Enviar una respuesta indicando que es posible compactar
				EnviarRespuestaCompactable(writer)
				return
			}
		}

		// Si no se puede compactar, enviar respuesta FAILED
		utils_general.EnviarStatusFAILED(writer) // Enviar respuesta FAILED si no hay espacio.
		return
	}
	
	// Agregamos el proceso a la memoria y partición si hay espacio.
	AgregarProcesoAMemoriaSistema(request, particion)
	AgregarProcesoAparticion(request, particion)
	
	slog.Info("## Proceso Creado - PID: <" + strconv.Itoa(request.PID) + "> - Tamaño: <" + strconv.Itoa(request.Tamanio) + ">.")
	utils_general.EnviarStatusOK(writer)
}

func AgregarProcesoAparticion(proceso globals.ProcesoMemoria, particion *globals.Particion){
    if globals.Config.Scheme == "DINAMICAS" && proceso.Tamanio < particion.Tamanio {
        // Si la partición es más grande que el proceso, partimos la memoria.

		// Crear una partición restante que representará el espacio libre después de asignar el proceso.
		particionRestante := &globals.Particion{}

		// El tamaño de la particionRestante sera el de la particion antes de asignarle el proceso
		// (osea, antes de "crear" la particion en si) menos el tamaño del proceso que iré a meter en mi partición
		// que llega como parametro.actual
		particionRestante.Tamanio = particion.Tamanio - proceso.Tamanio
        particionRestante.Base = particion.Base + uint32(proceso.Tamanio)
        particionRestante.Libre = true
        particionRestante.Limite = globals.Config.Memory_size - 1 // Se mantendra constante durante toda la ejecución.

		// Ajustar la partición asignada al tamaño del proceso.
		particion.Limite = int(particion.Base) + proceso.Tamanio - 1
		particion.Tamanio = proceso.Tamanio
		particion.Libre = false

        // Insertar la nueva partición restante en la lista de particiones
		globals.Particiones = append(globals.Particiones, particionRestante)
        slog.Info("Se creó una nueva partición dinámica restante de base: " + strconv.Itoa(int(particionRestante.Base)) + ", limite: " + strconv.Itoa(int(particionRestante.Limite)) + " y tamaño" + strconv.Itoa(particionRestante.Tamanio) + ".")
    } else {
        // Si la partición no se subdivide, simplemente se marca como ocupada.
        particion.Libre = false
    }

	for _, procesoLista := range globals.ListaContextosEjecucionProcesos {
		if procesoLista.PID == proceso.PID {
			procesoLista.ParticionAsignada = particion

			slog.Debug("Proceso " + strconv.Itoa(procesoLista.PID) + " asignado a partición de base " + strconv.Itoa(int(procesoLista.ParticionAsignada.Base)) + " y tamaño " + strconv.Itoa(procesoLista.ParticionAsignada.Tamanio))
		}
	}
}

func FinalizacionDeProceso(writer http.ResponseWriter, requestCliente *http.Request) {
	slog.Debug("Entre a FinalizacionDeProceso")
	var request globals.ProcesoMemoria

	err := json.NewDecoder(requestCliente.Body).Decode(&request) // Decodificamos la respuesta del cliente dentro de la variable request.

	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest) // Error el decodificar el request del cliente.
		slog.Warn("Error al decodificar el request del cliente en la recepción del mensaje. (Origen: FinalizacionDeProceso)")
		return
	}
	
	for _, proceso := range globals.ListaContextosEjecucionProcesos {
		if proceso.PID == request.PID { // Si encuentro el proceso guardo el contexto en una variable.
			switch globals.Config.Scheme {
			case "FIJAS":
				slog.Debug("Particion:" + strconv.Itoa(int(proceso.ParticionAsignada.Base)))
				LiberarEspacioFija(proceso)
			case "DINAMICAS":
				LiberarEspacioDinamica(proceso)
			}

			EliminarProcesoDeMemoria(request.PID) // Eliminamos el contexto del proceso de la memoria.
			slog.Info("## Proceso Destruido - PID: <" + strconv.Itoa(proceso.PID) + "> - Tamaño: <" + strconv.Itoa(proceso.Tamanio) + ">.")
			
			slog.Debug("Voy a mandar status OK. Finalize el proceso")
			utils_general.EnviarStatusOK(writer)
			slog.Debug("Mande Status OK desde Finalizacion de Proceso")
			return
		}
	}
	//slog.Debug("Voy a mandar un status failed desde Finalizacion de Proceso")
	//http.Error(writer, "No se encontro el proceso a finalizar", http.StatusBadRequest) // Error el decodificar el request del cliente.
	utils_general.EnviarStatusFAILED(writer)
}

func FinalizacionDeHilo(writer http.ResponseWriter, requestCliente *http.Request){
	var request globals.TIDyPID

 	err := json.NewDecoder(requestCliente.Body).Decode(&request) // Decodificamos la respuesta del cliente dentro de la variable request.
 	if err != nil {
 		http.Error(writer, err.Error(), http.StatusBadRequest) // Error el decodificar el request del cliente.
 		slog.Warn("Error al decodificar el request del cliente en la recepción del mensaje. FinalizacionDeHilo")
 		return
 	}

	for i, hilo := range globals.ListaContextosEjecucionHilos {
		if hilo.TID == request.TID && hilo.PID == request.PID {
			// Si la lista tiene solo un elemento, vaciarla directamente.
			if len(globals.ListaContextosEjecucionHilos) == 1 {
				globals.ListaContextosEjecucionHilos = []*globals.Hilo{}
			} else {
				// Eliminar el hilo usando slicing.
				globals.ListaContextosEjecucionHilos = append(globals.ListaContextosEjecucionHilos[:i], globals.ListaContextosEjecucionHilos[i+1:]...)
			}

			slog.Info("## Hilo Destruido - (PID:TID) - (<" + strconv.Itoa(hilo.PID) + ">:<" + strconv.Itoa(hilo.TID) + ">).")

			utils_general.EnviarStatusOK(writer)
			return
    	}
	}

	utils_general.EnviarStatusFAILED(writer)
}

func LiberarEspacioFija(proceso *globals.Proceso) {
    if proceso == nil {
        slog.Error("LiberarEspacioFija recibió un proceso nulo.")
    }else{
		slog.Debug("Se libera la particion " + strconv.Itoa(int(proceso.ParticionAsignada.Base)))

		if !proceso.ParticionAsignada.Libre {// Función para Borrar Datos de la partición.
			proceso.ParticionAsignada.Libre = true // Marca la partición como libre.
		} else {
			slog.Error("El proceso con PID " + strconv.Itoa(proceso.PID) + " no tiene una partición asiganada. Error el finalizar el proceso en memoria.")
			return
		}	
	}
}

func LiberarEspacioDinamica(proceso *globals.Proceso) {
	if proceso.ParticionAsignada != nil {
		// Función para Borrar Datos de la partición.
		proceso.ParticionAsignada.Libre = true // Marca la partición como libre.
		slog.Debug("Memoria Antes de consolidar")
		for i, particion := range globals.Particiones{
							
			slog.Debug("Particion Numero " + strconv.Itoa(i) + 
			", Base : " + strconv.Itoa(int(particion.Base)) + 
			", Tamanio: " + strconv.Itoa(int(particion.Tamanio)) + 
			" y Limite: " + strconv.Itoa(int(particion.Limite)) + 
			" Estado: " + strconv.FormatBool(particion.Libre))

		}
		particionesConsolidadas := &globals.Particion{} // Definimos la variable de las particiones consolidadas
		for i := 0;  i < len(globals.Particiones); i++ { // Iteramos sobre la lista de particiones 
			particion := globals.Particiones[i]
			slog.Debug(strconv.Itoa(int(particion.Base)))
			slog.Debug(strconv.Itoa(int(proceso.ParticionAsignada.Base)))
			if particion.Base == proceso.ParticionAsignada.Base { // Si las bases coinciden entonces encontramos la particion
				particion.Libre = true
				slog.Debug("Memoria despues de marcar la particion como true")
				for i, particion := range globals.Particiones{
									
					slog.Debug("Particion Numero " + strconv.Itoa(i) + 
					", Base : " + strconv.Itoa(int(particion.Base)) + 
					", Tamanio: " + strconv.Itoa(int(particion.Tamanio)) + "## (<0>:<0>) - Desbloqueado " + 
					" y Limite: " + strconv.Itoa(int(particion.Limite)) + 
					" Estado: " + strconv.FormatBool(particion.Libre))
		
				}

				if i - 1 >= 0 && i + 1 < len(globals.Particiones) { // Si la particion no es la primera de la lista y tampoco es la ultima entonces es una particion que tiene particion atras y adelante
					if globals.Particiones[i-1].Libre && globals.Particiones[i+1].Libre{ // Si tanto la particion de atras como la de adelante estan libres, hay que consolidar las 3 particiones
						particionesConsolidadas.Base = globals.Particiones[i-1].Base // Asigno la base de la particion de atras a la nueva variable de particion
						particionesConsolidadas.Limite = globals.Particiones[i+1].Limite // Asigno el limite de la particion de adelante a la nueva variable de particion
						particionesConsolidadas.Tamanio = globals.Particiones[i-1].Tamanio + globals.Particiones[i].Tamanio + globals.Particiones[i+1].Tamanio // El tamaño de la particion va a ser el tamaño de las 3
						particionesConsolidadas.Libre = true
						
						// Eliminamos las particiones consolidadas de la lista.
						globals.Particiones = append(globals.Particiones[:i-1], append([]*globals.Particion{particionesConsolidadas}, globals.Particiones[i+2:]...)...) // Reemplazamos en la lista por la nueva partcion
						slog.Info("Particiones consolidadas en base" + strconv.Itoa(int(particionesConsolidadas.Base)) + "con tamaño " + strconv.Itoa(particionesConsolidadas.Tamanio))
						slog.Debug("Memoria despues de consolidar")
						for i, particion := range globals.Particiones{
							
							slog.Debug("Particion Numero " + strconv.Itoa(i) + 
							", Base : " + strconv.Itoa(int(particion.Base)) + 
							", Tamanio: " + strconv.Itoa(int(particion.Tamanio)) + 
							" y Limite: " + strconv.Itoa(int(particion.Limite)) + 
							" Estado: " + strconv.FormatBool(particion.Libre))

						}
						return
					}else if globals.Particiones[i-1].Libre { // Si solo la anterior esta libre, hay que consolidar solo con esa
						particionesConsolidadas.Base = globals.Particiones[i-1].Base // Asigno la base de la particion de atras a la nueva variable de particion
						particionesConsolidadas.Limite = globals.Particiones[i].Limite// Asigno el limite de la particion actual a la nueva variable de particion
						particionesConsolidadas.Tamanio = globals.Particiones[i-1].Tamanio + globals.Particiones[i].Tamanio // El tamaño de las particiones es el tamaño de las 2
						particionesConsolidadas.Libre = true
					
						globals.Particiones = append(globals.Particiones[:i-1], append([]*globals.Particion{particionesConsolidadas}, globals.Particiones[i+1:]...)...)
						
						slog.Debug("Particion de atras es la numero: " + strconv.Itoa((i-1)) + " su estado es: " + strconv.FormatBool(globals.Particiones[i-1].Libre) + " y su tamaño es: " + strconv.Itoa(globals.Particiones[i-1].Tamanio))
						slog.Info("Partición consolidada con la anterior, base " + strconv.Itoa(int(particionesConsolidadas.Base)) + " y tamaño " + strconv.Itoa(particionesConsolidadas.Tamanio))
						slog.Debug("Memoria despues de consolidar")
						for i, particion := range globals.Particiones{
							
							slog.Debug("Particion Numero " + strconv.Itoa(i) + 
							", Base : " + strconv.Itoa(int(particion.Base)) + 
							", Tamanio: " + strconv.Itoa(int(particion.Tamanio)) + 
							" y Limite: " + strconv.Itoa(int(particion.Limite)) + 
							" Estado: " + strconv.FormatBool(particion.Libre))

						}
						return
					}else if globals.Particiones[i+1].Libre { // Si solo la siguiente esta libre hay que consolidar solo con esa
						particionesConsolidadas.Base = globals.Particiones[i].Base // La base es el mismo que el de la particion actual
						particionesConsolidadas.Limite = globals.Particiones[i+1].Limite // El lim es el mismo que el de la siguiente 
						particionesConsolidadas.Tamanio = globals.Particiones[i].Tamanio + globals.Particiones[i+1].Tamanio // El tamaño es la suma de las 2
						particionesConsolidadas.Libre = true
						slog.Debug("Particion de adelante es la numero: " + strconv.Itoa((i+1)) + " su estado es: " + strconv.FormatBool(globals.Particiones[i+1].Libre) + " y su tamaño es: " + strconv.Itoa(globals.Particiones[i+1].Tamanio))

						slog.Info("Partición consolidada con la siguiente, base " + strconv.Itoa(int(particionesConsolidadas.Base)) + " y tamaño " + strconv.Itoa(particionesConsolidadas.Tamanio))
						globals.Particiones = append(globals.Particiones[:i], append([]*globals.Particion{particionesConsolidadas}, globals.Particiones[i+2:]...)...)
						slog.Debug("Memoria despues de consolidar")
						for i, particion := range globals.Particiones{
							
							slog.Debug("Particion Numero " + strconv.Itoa(i) + 
							", Base : " + strconv.Itoa(int(particion.Base)) + 
							", Tamanio: " + strconv.Itoa(int(particion.Tamanio)) + 
							" y Limite: " + strconv.Itoa(int(particion.Limite)) + 
							" Estado: " + strconv.FormatBool(particion.Libre))

						}
						return
					}
					
				} else if i + 1 < len(globals.Particiones) { // Si no cumple la condicion de arriba y la siguiente posicion es menor al total de las particiones, estamos frente a la primer particion.
					if globals.Particiones[i+1].Libre {	// Si la particion proxima esta libre hay que consolidar
						particionesConsolidadas.Base = globals.Particiones[i].Base // Mantenemos la base de la particion actual
						particionesConsolidadas.Limite = globals.Particiones[i+1].Limite // Asigno el limite de la particion proxima a la nueva variable de particion
						particionesConsolidadas.Tamanio = globals.Particiones[i].Tamanio + globals.Particiones[i+1].Tamanio //El tamaño de las particiones es el tamaño de las 2
						particionesConsolidadas.Libre = true
	
						globals.Particiones = append(globals.Particiones[:i], append([]*globals.Particion{particionesConsolidadas}, globals.Particiones[i+2:]...)...)
						slog.Info("Partición consolidada con la anterior, base " + strconv.Itoa(int(particionesConsolidadas.Base)) + " y tamaño " + strconv.Itoa(particionesConsolidadas.Tamanio)) 
						slog.Debug("Memoria despues de consolidar")
						
							
							for i, particion := range globals.Particiones{
							
								slog.Debug("Particion Numero " + strconv.Itoa(i) + 
								", Base : " + strconv.Itoa(int(particion.Base)) + 
								", Tamanio: " + strconv.Itoa(int(particion.Tamanio)) + 
								" y Limite: " + strconv.Itoa(int(particion.Limite)) + 
								" Estado: " + strconv.FormatBool(particion.Libre))
	
							}
						
						return
					}
				} else if i - 1 >= 0{ // Si no cumple con ninguna de las otras condiciones, entonces es la ultima particion de la lista y solo tiene particion anterior
					if globals.Particiones[i-1].Libre { //Si la particion anterior esta libre hay que consolidar
						particionesConsolidadas.Base = globals.Particiones[i-1].Base //Asigno la base de la particion de atras a la nueva variable de particion
						particionesConsolidadas.Limite = globals.Particiones[i].Limite // Asigno el limite de la particion actual a la nueva variable de particion
						particionesConsolidadas.Tamanio = globals.Particiones[i-1].Tamanio + globals.Particiones[i].Tamanio //El tamaño de las particiones es el tamaño de las 2
						particionesConsolidadas.Libre = true
	
						globals.Particiones = append(globals.Particiones[:i-1], []*globals.Particion{particionesConsolidadas}...)
						slog.Info("Partición consolidada con la siguiente, base " + strconv.Itoa(int(particionesConsolidadas.Base)) + " y tamaño " + strconv.Itoa(particionesConsolidadas.Tamanio))
						slog.Debug("Memoria despues de consolidar")
						for i, particion := range globals.Particiones{
							
							slog.Debug("Particion Numero " + strconv.Itoa(i) + 
							", Base : " + strconv.Itoa(int(particion.Base)) + 
							", Tamanio: " + strconv.Itoa(int(particion.Tamanio)) + 
							" y Limite: " + strconv.Itoa(int(particion.Limite)) + 
							" Estado: " + strconv.FormatBool(particion.Libre))

						}
						return
					}	
				}
			}
		}

	}}