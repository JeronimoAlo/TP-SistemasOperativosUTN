package utils

import (
	"cpu/globals"
	"encoding/json"
	"log/slog"
	"net/http"
	"utils_general"
    "strconv"
)

func RecibirInterrupcion(writer http.ResponseWriter, requestCliente *http.Request){
    var request globals.InterrupcionDeHilo

	err := json.NewDecoder(requestCliente.Body).Decode(&request) // Decodificamos la respuesta del cliente dentro de la variable request.

	if err != nil {
		http.Error(writer, err.Error(), http.StatusBadRequest) // Error el decodificar el request del cliente.
		slog.Warn("Error al decodificar el request del cliente en la recepción del mensaje. Recibir Interrupcion")
		return
	}

    globals.ListaDeInterrupcionesDeHilo = append(globals.ListaDeInterrupcionesDeHilo, request)  //Se agrega la interrupcion a la lista de interrupciones dentro de CPU

    slog.Info("## Llega interrupcion al puerto Interrupt")
    
    EnviarStatusOK(writer)
}

func HayInterrupciones(tidEnEjecucion int, pidEnEjecucion int) bool{ // Chequeo si existe interrupciones del hilo ejecutandose.
    if(len(globals.ListaDeInterrupcionesDeHilo) == 0){
        return false
    }else {
        for i, interrupcionHilo := range globals.ListaDeInterrupcionesDeHilo{
            if interrupcionHilo.TID == tidEnEjecucion && interrupcionHilo.PID == pidEnEjecucion{
                slog.Debug("Encontre una interrupcion para el hilo en ejecucion y el tamaño de la lista de interrupciones es: " +  strconv.Itoa(len(globals.ListaDeInterrupcionesDeHilo)))
                return true
            }else{
                if len(globals.ListaDeInterrupcionesDeHilo) == 1 {
                    // Si la lista tiene un solo elemento, simplemente la vaciamos
                    globals.ListaDeInterrupcionesDeHilo = []globals.InterrupcionDeHilo{}
                } else {
                    // Si hay más de un elemento, eliminamos el elemento en el índice i
                    globals.ListaDeInterrupcionesDeHilo = append(globals.ListaDeInterrupcionesDeHilo[:i], globals.ListaDeInterrupcionesDeHilo[i+1:]...)
                }
            }
        }
        return false
    }
}

// Función que avisa a Kernel que ya desalojo el hilo que le habia pedido que desaloje.
func AvisoDeDesalojoDeHilo(hiloAdesalojar globals.InterrupcionDeHilo) error{
    
    request , err := json.Marshal(hiloAdesalojar) 

    if err!= nil{
        slog.Error("Error codificando el json: " + err.Error())
        return err
    }

    respuestaJson, err := utils_general.EnviarMensaje(globals.Config.Ip_kernel, globals.Config.Port_kernel, request, "CPU","/desalojoHilo")

    slog.Debug("El json que me llegó es: " + string(respuestaJson))

    if err != nil{
        slog.Error("Error enviando el json: " + err.Error())
        return err
    }

    var respuesta utils_general.StatusRespuesta

    err = json.Unmarshal(respuestaJson, &respuesta)

    if err != nil{
        slog.Error("Error decodificando el json: " + err.Error())
        return err
    }    

    if respuesta.Status != "OK"{
        slog.Error("Error al enviar desalojo: " + respuesta.Status )
        return err
    }

    return err
}

//CHECK INTERRUPT
func ChequeoInterrupciones(tid int, pid int) error {
    for i, interrupcionHilo := range globals.ListaDeInterrupcionesDeHilo{
         // range recorre la lista y devuelve la posición y el contenido.
        if interrupcionHilo.TID == tid && interrupcionHilo.PID == pid { // Si encuentra una interrupción para el TID actual, modifica el contexto de ejecución.
            pudoActualizar, err := ActualizarContextoEjecucion(pid, tid, &globals.RegistrosCPU)

            if !pudoActualizar || err != nil {
                slog.Error("Error al actualizar contexto de ejecución al atender interrupción.")                
            }


            globals.HayHiloEjecutandoEnCPU = false
            if len(globals.ListaDeInterrupcionesDeHilo) == 1 {
                // Si la lista tiene un solo elemento, simplemente la vaciamos
                globals.ListaDeInterrupcionesDeHilo = []globals.InterrupcionDeHilo{}
            } else if i >= 0 && i < len(globals.ListaDeInterrupcionesDeHilo) {
                globals.ListaDeInterrupcionesDeHilo = append(globals.ListaDeInterrupcionesDeHilo[:i], globals.ListaDeInterrupcionesDeHilo[i+1:]...)
            } else {
                slog.Error("Índice fuera de límites: " + strconv.Itoa(i))
            }
            slog.Debug("Se libero la interrupcion de la lista de interrupciones")

            err = AvisoDeDesalojoDeHilo(interrupcionHilo) // Se le envia al Kernel el motivo de la interrupción.

            if err != nil {
                slog.Error("Error al dar aviso del desalojo de un hilo al atender interrupción")   
                return err             
            }               

            return nil
        }else{

            if len(globals.ListaDeInterrupcionesDeHilo) == 1 {
                // Si la lista tiene un solo elemento, simplemente la vaciamos
                globals.ListaDeInterrupcionesDeHilo = []globals.InterrupcionDeHilo{}
            }else if i >= 0 && i < len(globals.ListaDeInterrupcionesDeHilo) {
                globals.ListaDeInterrupcionesDeHilo = append(globals.ListaDeInterrupcionesDeHilo[:i], globals.ListaDeInterrupcionesDeHilo[i+1:]...)
            } else {
                slog.Error("Índice fuera de límites: " + strconv.Itoa(i))
            }
                       
            slog.Debug("Se descarta interrupción por no corresponder al TID actual")
        }        
    }
    return nil
}