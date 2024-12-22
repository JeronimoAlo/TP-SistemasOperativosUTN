package utils

import (
	"cpu/globals"
	"encoding/json"
	"log/slog"
	"strconv"
	"utils_general"
	"errors"
)

func ActualizarContextoEjecucion(pid int , tid int , cpu *globals.CPU) (bool, error){
    var request globals.MensajeActualizarContexto

    slog.Debug("El PC que voy a mandar a la memoria es: " + strconv.Itoa(int(cpu.PC)))
    
    request.PID = globals.TIDyPIDenEjecucion.PID
    request.TID = globals.TIDyPIDenEjecucion.TID
    request.AX = cpu.AX
    request.BX = cpu.BX
    request.CX = cpu.CX
    request.DX = cpu.DX
    request.EX = cpu.EX
    request.FX = cpu.FX
    request.GX = cpu.GX
    request.HX = cpu.HX
    request.PC = cpu.PC
    
    requestAenviar, err := json.Marshal(request)

    if err != nil{
        slog.Error("Error al codificar el mensaje del contexto")
    }

    respMemoria, err := utils_general.EnviarMensaje(globals.Config.Ip_memory, globals.Config.Port_memory, requestAenviar ,"CPU" ,"/actualizarContexto")

    if err != nil{
        slog.Error("Error al solicitar actualizar el contexto")
    }
    
    var respuesta utils_general.StatusRespuesta
    
    err2 := json.Unmarshal(respMemoria, &respuesta)

    if err2 != nil{
        slog.Error("Error al decodificar actualizar el contexto")
    }

    slog.Info("## TID: <"+strconv.Itoa(globals.TIDyPIDenEjecucion.TID)+"> - Actualizo Contexto Ejecución")

    return respuesta.Status == "OK", err2
}

func ObtenerContextoEjecucion(pid int , tid int , cpu *globals.CPU) error {
    var solicitudContexto globals.SolicitudContexto 

	pidString := strconv.Itoa(pid)
	tidString := strconv.Itoa(tid)

	solicitudContexto.PID = pidString
	solicitudContexto.TID = tidString
    
    slog.Debug("El contexto que voy a pedir es para el PID: " + pidString + " y TID: " +tidString )

	mensaje, errSolContexto := json.Marshal(solicitudContexto)

    if errSolContexto != nil {
        slog.Error("Fallo al codificar la solicitud del contexto antes de enviar a memoria la peticion.")
    }

	respMemoria, err := utils_general.EnviarMensaje(globals.Config.Ip_memory, globals.Config.Port_memory, mensaje ,"CPU" ,"/obtenerContexto")

	if err != nil{
		slog.Error("Fallo al enviar el mensaje a la memoria para el fetch desde la CPU: " + string(mensaje))
	}

	var respuesta globals.CPU
    	
	err2 := json.Unmarshal(respMemoria, &respuesta) // Chequeamos unicamente que no haya ocurrido un error y almacenamos la respuesta de la memoria.

	if err2 != nil{
		slog.Error("Fallo al decodificar.")
	}

	cpu.PC = respuesta.PC
	cpu.AX = respuesta.AX
	cpu.BX = respuesta.BX
	cpu.CX = respuesta.CX
	cpu.DX = respuesta.DX
	cpu.EX = respuesta.EX
	cpu.FX = respuesta.FX
	cpu.GX = respuesta.GX
	cpu.HX = respuesta.HX
	cpu.Base = respuesta.Base
	cpu.Limite = respuesta.Limite

    slog.Info("## TID: <" + tidString +  "> - Solicito Contexto Ejecución")
	
    return nil
}

func ObtenerDireccionLogica(cpu *globals.CPU, registro string) (uint32, error){
    switch registro{
    case "PC":
        return cpu.PC, nil
    case "AX":
        return cpu.AX, nil
    case "BX":
        return cpu.BX, nil
    case "CX":
        return cpu.CX, nil
    case "DX":
        return cpu.DX, nil
    case "EX":
        return cpu.EX, nil
    case "FX":
        return cpu.FX, nil
    case "GX":
        return cpu.GX, nil
    case "HX":
        return cpu.HX, nil
    case "BASE":
        return cpu.Base, nil
    case "LIMITE":
        return cpu.Limite, nil
    default:
        return 0, errors.New("Registro '"+registro+"' no válido")
    }


}