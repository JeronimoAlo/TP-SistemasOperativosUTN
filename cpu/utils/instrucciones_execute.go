package utils

import (
	"cpu/globals"
	"encoding/json"
	"log/slog"
	"strconv"
	"utils_general"
    "fmt"
)

func READ_MEM(cpu *globals.CPU, registroDatos string, registroDireccion string) (bool, error) {
    // Obtener la dirección lógica de los registros
    // slog.Debug("El valor de AX es " + strconv.Itoa(int(cpu.AX)) + " y el de DX es " + strconv.Itoa(int(cpu.DX)) + " antes del READ_MEM")

    direccionLogicaDireccion, err := ObtenerDireccionLogica(cpu, registroDireccion) // Guardo en direccionLogicaDireccion la direccion logica de mi registroDireccion
    if err != nil{
        slog.Error("El registro '" + registroDireccion + "' ingresado es incorrecto.")
        return false, fmt.Errorf("error obteniendo dirección lógica de %s: %w", registroDireccion, err)
    }

    direccionFisicaDireccion, err := Traducir(direccionLogicaDireccion, cpu) // Traducimos la direccionLogicaDatos a la direccionFisica
    if err != nil{
        slog.Error("Fallo al querer traducir la direccion logica '" + strconv.FormatUint(uint64(direccionLogicaDireccion), 10) + "' a direccion fisica.")
        return false, fmt.Errorf("error traduciendo dirección lógica %d: %w", direccionLogicaDireccion, err)
    }

    //Creamos una estructura que guardara la direccionFisica y el valor de registro de registroDatos a enviar a memoria.
    EstructuraEnviarAmemoria := globals.LecturaEscrituraMemoria{
        DireccionFisica: direccionFisicaDireccion,
        PID: globals.TIDyPIDenEjecucion.PID,
        TID: globals.TIDyPIDenEjecucion.TID,
    }

    slog.Debug("Direccion Fisica a escribir: " + strconv.Itoa(int(EstructuraEnviarAmemoria.DireccionFisica)))
    slog.Debug("Valor a escribir en Direccion Fisica: " + strconv.Itoa(int(EstructuraEnviarAmemoria.ValorAenviar)))
    slog.Debug("PID: " + strconv.Itoa(int(EstructuraEnviarAmemoria.PID)))
    slog.Debug("TID: " + strconv.Itoa(int(EstructuraEnviarAmemoria.TID)))

    mensaje, err := json.Marshal(EstructuraEnviarAmemoria) // Generamos un mensaje en JSON que contiene direccionFisica y el valor de memoria de registroDireccion en un "mensaje".
    if err != nil {
        slog.Error("Fallo al codificar la estructura Registro/Direccion Fisica antes de enviar a memoria la peticion.")
        return false, fmt.Errorf("error serializando estructura LecturaEscrituraMemoria: %w", err)
    }

    // Enviamos a memoria el "mensaje" generado y almacenamos su respuesta en respMemoria.
	respMemoria, err := utils_general.EnviarMensaje(globals.Config.Ip_memory, globals.Config.Port_memory, mensaje , "CPU", "/leerMemoria")
	if err != nil{
		slog.Error("Fallo al enviar el mensaje a la memoria: " + string(mensaje))
        return false, fmt.Errorf("error enviando mensaje a memoria: %w", err)         
	}

	var valor globals.ValorAresponder 
	err = json.Unmarshal(respMemoria, &valor) // Chequeamos unicamente que no haya ocurrido un error.

    slog.Debug("El supuesto valor de AX en memoria " + strconv.Itoa(int(valor.Valor)))
    SET(cpu,registroDatos, int(valor.Valor))   

	if err != nil{ // Verificamos que la respuesta de la memoria haya sido "OK".
        slog.Error("Error al escribir el valor del memoria '" +strconv.Itoa(int(direccionFisicaDireccion) ) + "' en la direccion de memoria del '" + strconv.Itoa(int(direccionLogicaDireccion ))+ "' .")
        return false, fmt.Errorf("error deserializando respuesta de memoria: %w", err)
    }else {
        slog.Info("## TID: <" + strconv.Itoa(globals.TIDyPIDenEjecucion.TID) + "> - Acción: <LEER> - Dirección Física: <" + strconv.FormatUint(uint64(direccionFisicaDireccion), 10) + ">")
	
        return true, nil
    }
}

func WRITE_MEM(cpu *globals.CPU, registroDireccion string, registroDatos string) (bool, error){
    var valor uint32
    
    // Guardo el valor de registroDatos en una variable valor dependiendo el tipo de registro.
    switch registroDatos{
    case "AX":
        valor = cpu.AX
    case "BX":
        valor = cpu.BX
    case "CX":
        valor = cpu.CX
    case "DX":
        valor = cpu.DX
    case "EX":
        valor = cpu.EX
    case "FX":
        valor = cpu.FX
    case "GX":
        valor = cpu.GX
    case "HX":
        valor = cpu.HX
    default:
        slog.Error("El registro '" + registroDatos + "' ingresado no existe.")
        return false, fmt.Errorf("registro '%s' es inválido", registroDatos)
    }
    
    slog.Debug("Valor en cada Operacion: " + strconv.Itoa(int(valor)))
    slog.Debug("Registro Direccion :" + registroDireccion)

    // Obtenemos la direccion logica del registro direccion, donde queremos guardar el valor de RegistroDatos
    direccionLogica, err := ObtenerDireccionLogica(cpu, registroDireccion)
    if err != nil{
        slog.Error("El registro '" + registroDireccion + "' ingresado es invalido.")
        return false, fmt.Errorf("error obteniendo dirección lógica de %s: %w", registroDireccion, err)
    }

    slog.Debug("Direccion Logica: " + strconv.Itoa(int(direccionLogica)))

    // Traducimos la direccionLogica a la direccionFisica
    direccionFisica, err := Traducir(direccionLogica, cpu)
    if err != nil{
        slog.Error("Fallo al querer traducir la direccion logica '" + strconv.FormatUint(uint64(direccionLogica), 10) + "' a direccion fisica.")
        return false, fmt.Errorf("error traduciendo dirección lógica %d: %w", direccionLogica, err)
    }

    //Creamos una estructura que guardara la direccionFisica y el valor de registro de registroDatos a enviar a memoria.
    EstructuraEnviarAmemoria := globals.LecturaEscrituraMemoria{
        DireccionFisica: direccionFisica,
        ValorAenviar:    valor,
        PID: globals.TIDyPIDenEjecucion.PID,
        TID: globals.TIDyPIDenEjecucion.TID,
    }

    slog.Debug("Direccion Fisica a escribir: " + strconv.Itoa(int(EstructuraEnviarAmemoria.DireccionFisica)))
    slog.Debug("Valor a escribir en Direccion Fisica: " + strconv.Itoa(int(EstructuraEnviarAmemoria.ValorAenviar)))
    slog.Debug("PID: " + strconv.Itoa(int(EstructuraEnviarAmemoria.PID)))
    slog.Debug("TID: " + strconv.Itoa(int(EstructuraEnviarAmemoria.TID)))

    // Generamos un mensaje en JSON que contiene direccionFisica y el valor de registro de registroDatos en un "mensaje".
    mensaje, err := json.Marshal(EstructuraEnviarAmemoria)
    if err != nil {
        slog.Error("Fallo al codificar la estructura Registro/Direccion Fisica antes de enviar a memoria la peticion.")
        return false, fmt.Errorf("error serializando estructura: %w", err)
    }

    // Enviamos a memoria el "mensaje" generado y almacenamos su respuesta en respMemoria.
	respMemoria, err := utils_general.EnviarMensaje(globals.Config.Ip_memory, globals.Config.Port_memory, mensaje , "CPU", "/escribirMemoria") 
	if err != nil{
		slog.Error("Fallo al enviar el mensaje a la memoria: " + string(mensaje))
        return false, fmt.Errorf("error enviando mensaje a memoria: %w", err)
	}

	var respuesta utils_general.StatusRespuesta
	err = json.Unmarshal(respMemoria, &respuesta) // Chequeamos unicamente que no haya ocurrido un error.

	if err != nil{
		slog.Error("Fallo al decodificar.")
        return false, fmt.Errorf("error decodificando respuesta de memoria: %w", err)
	}
    
    if respuesta.Status == "OK"{ // Verificamos que la respuesta de la memoria haya sido "OK".
        slog.Info("## TID: <"+ strconv.Itoa(globals.TIDyPIDenEjecucion.TID) + "> - Acción: <ESCRIBIR> - Dirección Física: <" + strconv.FormatUint(uint64(direccionFisica), 10) + ">")
        return true, nil
    }else{
        slog.Error("Error al escribir el valor del registro '"+registroDatos+"' en la direccion de memoria del '"+registroDireccion+"' .")
        return false, fmt.Errorf("memoria devolvió error: %s", respuesta.Status)
    } 
}

func SET(cpu *globals.CPU, registro string, valor int) { // SET Asigna un valor a uno de los registros de la CPU.
    switch registro {
    case "AX":
        cpu.AX = uint32(valor) //Asigna el valor al registro AX
	case "BX":
        cpu.BX = uint32(valor) //Asigna el valor al registro BX    
    case "CX":
        cpu.CX = uint32(valor) //Asigna el valor al registro CX
    case "DX":
        cpu.DX = uint32(valor) //Asigna el valor al registro DX
	case "EX":
        cpu.EX = uint32(valor) //Asigna el valor al registro EX
	case "FX":
        cpu.FX = uint32(valor) //Asigna el valor al registro FX        
    case "GX":
        cpu.GX = uint32(valor) //Asigna el valor al registro GX        
	case "HX":
        cpu.HX = uint32(valor) //Asigna el valor al registro HX        
    }
    slog.Debug("Valor en cada Operacion: " + strconv.Itoa(valor))
}

func SUM(cpu *globals.CPU, registroDestino string, registroOrigen string) { // Suma el valor de un registro origen al valor de un registro destino en la CPU
    var valorOrigen int //Guada el valor del resgistro de origen convertido a int para hacer la suma
    
    // Obtiene el valor del registro de origen y lo convierte a int.
	switch registroOrigen {
    case "AX":
        valorOrigen = int(cpu.AX)
    case "BX":
        valorOrigen = int(cpu.BX)
	case "CX":
        valorOrigen = int(cpu.CX)
    case "DX":
        valorOrigen = int(cpu.DX)
	case "EX":
        valorOrigen = int(cpu.EX)
	case "FX":
        valorOrigen = int(cpu.FX)
    case "GX":
        valorOrigen = int(cpu.GX)
	case "HX":
        valorOrigen = int(cpu.HX)
	case "Base":
        valorOrigen = int(cpu.Base)
	case "Limite":
		valorOrigen = int(cpu.Limite)
    }

    // Suma el valor de origen al registro destino
    switch registroDestino {
    case "AX":
        cpu.AX += uint32(valorOrigen)
    case "BX":
        cpu.BX += uint32(valorOrigen)
    case "CX":
        cpu.CX += uint32(valorOrigen)
    case "DX":
        cpu.DX += uint32(valorOrigen)
	case "EX":
        cpu.EX += uint32(valorOrigen)
	case "FX":
        cpu.FX += uint32(valorOrigen)
    case "GX":
        cpu.GX += uint32(valorOrigen)
	case "HX":
        cpu.HX += uint32(valorOrigen)
	case "Base":
        cpu.Base += uint32(valorOrigen)
	case "Limite":
		cpu.Limite += uint32(valorOrigen)
    }   
}

func SUB(cpu *globals.CPU, registroDestino string, registroOrigen string) {//Resta el valor de un registro a otro y almacena el resultado en el registro de destino.
	var valorOrigen int//Guada el valor del resgistro de origen convertido a int para hacer la resta
    
	switch registroOrigen {// Se obtiene el valor del registro de origen dependiendo de su nombre.
    case "AX":
        valorOrigen = int(cpu.AX)
    case "BX":
        valorOrigen = int(cpu.BX)
	case "CX":
        valorOrigen = int(cpu.CX)
    case "DX":
        valorOrigen = int(cpu.DX)
	case "EX":
        valorOrigen = int(cpu.EX)
	case "FX":
        valorOrigen = int(cpu.FX)
    case "GX":
        valorOrigen = int(cpu.GX)
	case "HX":
        valorOrigen = int(cpu.HX)
	case "Base":
        valorOrigen = int(cpu.Base)
	case "Limite":
		valorOrigen = int(cpu.Limite)
    }

    switch registroDestino { // Se resta el valor del registro de origen al valor del registro de destino y el resultado se almacena en el registro de destino.
    case "AX":
        cpu.AX -= uint32(valorOrigen)
    case "BX":
        cpu.BX -= uint32(valorOrigen)
    case "CX":
        cpu.CX -= uint32(valorOrigen)
    case "DX":
        cpu.DX -= uint32(valorOrigen)
	case "EX":
        cpu.EX -= uint32(valorOrigen)
	case "FX":
        cpu.FX -= uint32(valorOrigen)
    case "GX":
        cpu.GX -= uint32(valorOrigen)
	case "HX":
        cpu.HX -= uint32(valorOrigen)
	case "Base":
        cpu.Base -= uint32(valorOrigen)
	case "Limite":
		cpu.Limite -= uint32(valorOrigen)
    }
}

func JNZ(cpu *globals.CPU, registro string, instruccion int) {// Realiza un salto condicional si el valor del registro indicado no es cero
    var valor int // Guarda en forma temporal el valor del registro indicado	

    switch registro {//Obtiene el valor del registro y lo convierte a int
    case "AX":
        valor = int(cpu.AX)
    case "BX":
        valor = int(cpu.BX)
    case "CX":
        valor = int(cpu.CX)
    case "DX":
        valor = int(cpu.DX)
	case "EX":
        valor = int(cpu.EX)
	case "FX":
        valor = int(cpu.FX)
    case "GX":
        valor = int(cpu.GX)
	case "HX":
        valor = int(cpu.HX)
	case "Base":
        valor = int(cpu.Base)
	case "Limite":
		valor = int(cpu.Limite)
    }
    
    // Si el valor no es igual a cero, se actualiza el Program Counter (PC) con la instrucción dada
    if valor != 0 {
        cpu.PC = uint32(instruccion) // Actualiza el Program Counter
    }
}

func LOG(cpu *globals.CPU, registro string) { //Imprime el valor de un registro especifico en un archivo de logs.
    var valor int // Guarda en forma temporal el valor del registro indicado.
	
    switch registro { // Dependiendo del nombre del registro, se obtiene su valor y se convierte a int.
    case "AX":
        valor = int(cpu.AX)
    case "BX":
        valor = int(cpu.BX)
	case "CX":
        valor = int(cpu.CX)
    case "DX":
        valor = int(cpu.DX)
	case "EX":
        valor = int(cpu.EX)
	case "FX":
        valor = int(cpu.FX)
    case "GX":
        valor = int(cpu.GX)
	case "HX":
        valor = int(cpu.HX)
	case "Base":
        valor = int(cpu.Base)
	case "Limite":
		valor = int(cpu.Limite)
    }

    slog.Debug("Proceso en ejecucion: " + strconv.Itoa(globals.TIDyPIDenEjecucion.PID))
    slog.Debug("Hilo en ejecucion: " + strconv.Itoa(globals.TIDyPIDenEjecucion.TID))
    slog.Info("LOG de " + registro + ": " + strconv.Itoa(valor)) // Escribir en el archivo de logs.
}