package utils

import (
	"cpu/globals"
	"fmt"
	"log/slog"
	"strconv"
)

// Función para traducir una dirección lógica a una dirección física.
func Traducir(direccionLogica uint32, contexto *globals.CPU) (uint32 , error) {
	// Verificar que la dirección lógica esté dentro de los límites del proceso.
    direccionFisica := contexto.Base + direccionLogica

	if direccionFisica > (contexto.Base + contexto.Limite){
        err := EnviarSegmentationFault(globals.TIDyPIDenEjecucion.TID , globals.TIDyPIDenEjecucion.PID) // Se envia el TID al kernel por Segmentation Fault
        
        if err != nil {
            return 0, err // Manejar el error de envío
        }

        slog.Debug("Segmentation Fault enviado con éxito al Kernel.")

        ActualizarContextoEjecucion(globals.TIDyPIDenEjecucion.PID ,globals.TIDyPIDenEjecucion.TID , contexto)

        return 0, nil
    }else{
        return direccionFisica, nil
    }
}

func EnviarSegmentationFault(tid int, pid int) error {
    //Construimos la URL con los parámetros necesarios para la syscall SEGMENTATION_FAULT
    url := fmt.Sprintf("http://" + globals.Config.Ip_kernel + ":" + strconv.Itoa(globals.Config.Port_kernel) + "/SEGMENTATION_FAULT/%d/%d", tid, pid)

    err := EnviarMensajePorParametro(url) // Se llama a la función EnviarMensajePorParametro, que realiza una solicitud HTTP GET a la URL especificada.

    if err != nil { // Si ocurre un error al enviar la solicitud, se registra el error en el log.
        slog.Error("Error enviando mensaje SEGMENTATION_FAULT al Kernel.")
    }

    return err// Retorna el error (si no hubo errores, será nil).    
}