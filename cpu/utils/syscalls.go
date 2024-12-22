package utils

import (
	"cpu/globals"
	"fmt"
	"log/slog"
	"strconv"
	"utils_general"
    "encoding/json"
)

func DUMP_MEMORY() error{ // Tiene que mandarle al kernel la solicitud del DUMP_MEMOPRY, despues kernel se lo solicita a la memoria.
    // Se actualiza el contexto de ejecución.
    pudoActualizar, err := ActualizarContextoEjecucion(globals.TIDyPIDenEjecucion.TID, globals.TIDyPIDenEjecucion.PID, &globals.RegistrosCPU)
    
    if !pudoActualizar || err != nil{
        slog.Error("Error actualizando el contexto de ejecucion")
        return err
    }

    url := fmt.Sprintf("http://" + globals.Config.Ip_kernel + ":" + strconv.Itoa(globals.Config.Port_kernel)  +"/DUMP_MEMORY/%d/%d", globals.TIDyPIDenEjecucion.TID ,globals.TIDyPIDenEjecucion.PID )

    err = EnviarMensajePorParametro(url)
    
    if err!= nil{ // Si ocurre un error al enviar la solicitud, se registra el error en el log.

        slog.Error("Error enviando mensaje al Kernel DUMP_MEMORY desde la CPU.")        
    }

    return err // Retorna el error (si no hubo errores, será nil).
}

func IO(tiempoEnMilisegundos int) error{ // Envía una solicitud al Kernel para que el hilo actual realice una operación de entrada/salida (I/O).
    //Se actualiza el contexto de ejecución.
    pudoActualizar, err := ActualizarContextoEjecucion(globals.TIDyPIDenEjecucion.TID, globals.TIDyPIDenEjecucion.PID, &globals.RegistrosCPU)
    
    if !pudoActualizar || err != nil{
        slog.Error("Error actualizando el contexto de ejecucion")
        return err      
    }

    var ioAEnviar globals.IOparaHacer

    ioAEnviar.PID = globals.TIDyPIDenEjecucion.PID
    ioAEnviar.TID = globals.TIDyPIDenEjecucion.TID
    ioAEnviar.TiempoEnMilisegundos = tiempoEnMilisegundos
    
    slog.Debug("PID a enviar: " +  strconv.Itoa(ioAEnviar.PID))
    slog.Debug("TID a enviar: " +  strconv.Itoa(ioAEnviar.TID))
    slog.Debug("Tiempo a enviar: " +  strconv.Itoa(ioAEnviar.TiempoEnMilisegundos))

    mensaje, err := json.Marshal(ioAEnviar)
    if err != nil {
        slog.Error("Fallo al codificar la estructura ioAEnviar.")
        return fmt.Errorf("error serializando estructura: "+  err.Error())
    }

    slog.Debug("Mensaje a enviar a Kernel: " + string(mensaje))
    respuesta, err := utils_general.EnviarMensaje(globals.Config.Ip_kernel, globals.Config.Port_kernel, mensaje , "CPU", "/IO/") 
    if err != nil {
        slog.Error("Fallo al enviar el mensaje de IO a kernel. " + err.Error())
        return  fmt.Errorf("error serializando estructura: "+  err.Error())
    }
    var respDecodificada globals.MensajeResultado
  
	err = json.Unmarshal(respuesta, &respDecodificada) // Chequeamos unicamente que no haya ocurrido un error.
    if err != nil {
        slog.Error("Fallo al decodificar json de IO. " + err.Error())
        return  fmt.Errorf("Fallo al decodificar json de IO. "+  err.Error())
    }
    slog.Debug("Respuesta decodificada: " + respDecodificada.Status)

    slog.Debug("Se envio la solicitud de IO")
    return nil
}

func PROCESS_CREATE(rutaArchivo string, tamanio int, prioridadHiloMain int) error{ //Envía una solicitud al Kernel para crear un nuevo proceso.
    // Se actualiza el contexto de ejecución.
    pudoActualizar, err := ActualizarContextoEjecucion(globals.TIDyPIDenEjecucion.TID, globals.TIDyPIDenEjecucion.PID, &globals.RegistrosCPU)
    
    if !pudoActualizar || err != nil{
        slog.Error("Error actualizando el contexto de ejecucion")
        return err       
    }

    //rutaArchivo = rutaArchivo + ".pseudo"
    
    // Construcción de la URL que será utilizada para enviar la solicitud al Kernel.Incluye la ruta del archivo, el tamaño del proceso, y la prioridad del hilo principal.
    url := fmt.Sprintf("http://" + globals.Config.Ip_kernel + ":" + strconv.Itoa(globals.Config.Port_kernel) +"/PROCESS_CREATE/%s/%d/%d", rutaArchivo, tamanio, prioridadHiloMain)
    
    slog.Debug("Ruta Archivo: " + rutaArchivo)
    slog.Debug("Tamanio Proceso: " + strconv.Itoa(tamanio))
    slog.Debug("Prioridad del hilo 0: " + strconv.Itoa(prioridadHiloMain))
    // Llamar a la función EnviarMensajePorParametro, la cual realiza una solicitud HTTP GET a la URL construida.
    err = EnviarMensajePorParametro(url) 

    if err!= nil{ // Si ocurre un error al enviar la solicitud, registrarlo en el log y devolver el error.MensajePorParametro(url)

        slog.Error("Error enviando mensaje al Kernel para PROCESS_CREATE: " + err.Error() )        
    }
    
    return err// Retornar el error (será nil si no hubo problemas)
}

func THREAD_CREATE(rutaArchivo string, prioridadHilo int, pidPadre int) error{ // Envía una solicitud al Kernel para crear un nuevo hilo dentro del proceso padre especificado.
    //Se actualiza el contexto de ejecución 
    pudoActualizar, err := ActualizarContextoEjecucion(globals.TIDyPIDenEjecucion.TID, globals.TIDyPIDenEjecucion.PID, &globals.RegistrosCPU)
    
    if !pudoActualizar || err != nil{
        slog.Error("Error actualizando el contexto de ejecucion")
        return err        
    }

    //rutaArchivo = rutaArchivo + ".pseudo" // Agregamos la extension al archivo.
    
    // Construcción de la URL para enviar la solicitud al Kernel. Los parámetros incluyen la ruta del archivo de instrucciones,la prioridad del hilo, y el PID del proceso padre.
    url := fmt.Sprintf("http://" + globals.Config.Ip_kernel + ":" + strconv.Itoa(globals.Config.Port_kernel) + "/THREAD_CREATE/%s/%d/%d", rutaArchivo, prioridadHilo, pidPadre)

    // Llamar a la función EnviarMensajePorParametro, que realiza una solicitud HTTP GET a la URL especificada.
    slog.Debug("Request a enviar: " + url)
    err = EnviarMensajePorParametro(url)

    if err != nil { // Si ocurre un error al enviar la solicitud, se registra el error en el log.
        slog.Error("Error enviando mensaje al Kernel THREAD_CREATE", "error", err.Error())
    }

    return err// Retorna el error (si no hubo errores, será nil).
}

func THREAD_JOIN(tidAEsperar int, tidAbloquear int, pidPadreHiloABloquear int) error{
    //Se actualiza el contexto de ejecución 
    pudoActualizar, err := ActualizarContextoEjecucion(globals.TIDyPIDenEjecucion.TID, globals.TIDyPIDenEjecucion.PID, &globals.RegistrosCPU)
    
    if !pudoActualizar || err != nil{
        slog.Error("Error actualizando el contexto de ejecucion")
        return err       
    }
    // Construicción de la URL para enviar la solicitud al Kernel. Los parámetros incluyen el TID a Esperar, el TID a Bloquear y el PID del Hilo Padre a Bloquear
    url := fmt.Sprintf("http://" + globals.Config.Ip_kernel + ":" + strconv.Itoa(globals.Config.Port_kernel) + "/THREAD_JOIN/%d/%d/%d", tidAEsperar, tidAbloquear, pidPadreHiloABloquear)

    err = EnviarMensajePorParametro(url) // Se llama a la función EnviarMensajePorParametro, que realiza una solicitud HTTP GET a la URL especificada.

    if err != nil { // Si ocurre un error al enviar la solicitud, se registra el error en el log.
        slog.Error("Error enviando mensaje al Kernel THREAD_JOIN")
    }

    return err// Retorna el error (si no hubo errores, será nil).
}

func THREAD_CANCEL(tidSolicitante int, tidACancelar int, pidPadreHiloACancelar int) error {
    //Se actualiza el contexto de ejecución 
    pudoActualizar, err := ActualizarContextoEjecucion(globals.TIDyPIDenEjecucion.TID, globals.TIDyPIDenEjecucion.PID, &globals.RegistrosCPU)
    
    if !pudoActualizar || err != nil{
        slog.Error("Error actualizando el contexto de ejecucion")
        return err       
    }
    // Construicción de la URL para enviar la solicitud al Kernel. Los parámetros incluyen el TID solicitante, el TID a cancelar y PID Padre del Hilp a Ca
    url := fmt.Sprintf("http://" + globals.Config.Ip_kernel + ":" + strconv.Itoa(globals.Config.Port_kernel) + "/THREAD_CANCEL/%d/%d/%d", tidSolicitante, tidACancelar, pidPadreHiloACancelar)

    err = EnviarMensajePorParametro(url) // Se llama a la función EnviarMensajePorParametro, que realiza una solicitud HTTP GET a la URL especificada.

    if err != nil { // Si ocurre un error al enviar la solicitud, se registra el error en el log.
        slog.Error("Error enviando mensaje al Kernel THREAD_CREATE")
    }

    return err// Retorna el error (si no hubo errores, será nil).    
}

func MUTEX_CREATE(nombre string) error{
    //Se actualiza el contexto de ejecución 
    pudoActualizar, err := ActualizarContextoEjecucion(globals.TIDyPIDenEjecucion.TID, globals.TIDyPIDenEjecucion.PID, &globals.RegistrosCPU)

    if !pudoActualizar || err != nil{
        slog.Error("Error actualizando el contexto de ejecucion")
        return err       
    }
    
    url := fmt.Sprintf("http://" + globals.Config.Ip_kernel + ":" + strconv.Itoa(globals.Config.Port_kernel)  +"/MUTEX_CREATE/%d/%s",globals.TIDyPIDenEjecucion.PID, nombre )

    // Llamar a la función EnviarMensajePorParametro, que realiza una solicitud HTTP GET a la URL especificada.
    err = EnviarMensajePorParametro(url)

    if err!= nil{// Si ocurre un error al enviar la solicitud, se registra el error en el log.

        slog.Error("Error enviando mensaje al Kernel MUTEX_CREATE")        
    }
    return err
}

func MUTEX_LOCK(mutex string) error{
    //Se actualiza el contexto de ejecución 
    pudoActualizar, err := ActualizarContextoEjecucion(globals.TIDyPIDenEjecucion.TID, globals.TIDyPIDenEjecucion.PID, &globals.RegistrosCPU)

    if !pudoActualizar || err != nil{
        slog.Error("Error actualizando el contexto de ejecucion")
        return err     
    }
    
    url := fmt.Sprintf("http://" + globals.Config.Ip_kernel + ":" + strconv.Itoa(globals.Config.Port_kernel)  +"/MUTEX_LOCK/%d/%d/%s",globals.TIDyPIDenEjecucion.TID,globals.TIDyPIDenEjecucion.PID, mutex )

    // Llamar a la función EnviarMensajePorParametro, que realiza una solicitud HTTP GET a la URL especificada.
    err = EnviarMensajePorParametro(url)

    if err!= nil{// Si ocurre un error al enviar la solicitud, se registra el error en el log.

        slog.Error("Error enviando mensaje al Kernel MUTEX_LOCK")        
    }

    return err
}

func MUTEX_UNLOCK(mutex string) error{
    //Se actualiza el contexto de ejecución 
    pudoActualizar, err := ActualizarContextoEjecucion(globals.TIDyPIDenEjecucion.TID, globals.TIDyPIDenEjecucion.PID, &globals.RegistrosCPU)

    if !pudoActualizar || err != nil{
        slog.Error("Error actualizando el contexto de ejecucion")
        return err       
    }
    
    url := fmt.Sprintf("http://" + globals.Config.Ip_kernel + ":" + strconv.Itoa(globals.Config.Port_kernel)  +"/MUTEX_UNLOCK/%d/%d/%s",globals.TIDyPIDenEjecucion.TID,globals.TIDyPIDenEjecucion.PID, mutex )

    // Llamar a la función EnviarMensajePorParametro, que realiza una solicitud HTTP GET a la URL especificada.
    err = EnviarMensajePorParametro(url)

    if err!= nil{// Si ocurre un error al enviar la solicitud, se registra el error en el log.

        slog.Error("Error enviando mensaje al Kernel MUTEX_UNLOCK")        
    }
    return err
}

func THREAD_EXIT(tid int, pidPadre int) error{
    //Se actualiza el contexto de ejecución 
    pudoActualizar, err := ActualizarContextoEjecucion(globals.TIDyPIDenEjecucion.TID, globals.TIDyPIDenEjecucion.PID, &globals.RegistrosCPU)
    
    if !pudoActualizar || err != nil{
        slog.Error("Error actualizando el contexto de ejecucion")
        return err       
    }
    
    //Construimos la URL con los parámetros necesarios para la syscall THREAD_EXIT
    url := fmt.Sprintf("http://" + globals.Config.Ip_kernel + ":" + strconv.Itoa(globals.Config.Port_kernel) + "/THREAD_EXIT/%d/%d", tid, pidPadre)

    err = EnviarMensajePorParametro(url) // Se llama a la función EnviarMensajePorParametro, que realiza una solicitud HTTP GET a la URL especificada.

    if err != nil { // Si ocurre un error al enviar la solicitud, se registra el error en el log.
        slog.Error("Error enviando mensaje al Kernel THREAD_EXIT")
    }
    return err // Retorna el error (si no hubo errores, será nil).        
}

func PROCESS_EXIT(tid int, pidPadre int) error{
    
    // Se actualiza el contexto de ejecución.
    pudoActualizar, err := ActualizarContextoEjecucion(globals.TIDyPIDenEjecucion.TID, globals.TIDyPIDenEjecucion.PID, &globals.RegistrosCPU)
    
    if !pudoActualizar || err != nil{
        slog.Error("Error actualizando el contexto de ejecucion")
    return err       
    }

    // Construimos la URL con los parámetros necesarios para la syscall PROCESS_EXIT.
    url := fmt.Sprintf("http://" + globals.Config.Ip_kernel + ":" + strconv.Itoa(globals.Config.Port_kernel) + "/PROCESS_EXIT/%d/%d", tid, pidPadre)

    err = EnviarMensajePorParametro(url) // Enviamos la solicitud al Kernel usando la función EnviarMensajePorParametro.

    if err != nil { // Si ocurre un error al enviar la solicitud, se registra en el log.
        slog.Error("Error enviando mensaje al Kernel PROCESS_EXIT")
    }
    
    return err // Retorna el error (si no hubo errores, sera nil).
}