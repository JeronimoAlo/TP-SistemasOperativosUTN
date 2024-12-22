package utils

import (
	"kernel/globals"
	"reflect"
	"testing"
)

func TestCambiarEstadoProceso(t *testing.T) {
	// Crear un proceso simulado con un estado inicial
	proceso := &globals.PCB{
		PID:    1,
		Estado: globals.New, // Estado inicial
	}

	// Cambiar el estado del proceso a Blocked
	CambiarEstadoProceso(proceso, globals.Blocked)

	// Verificar si el estado ha cambiado correctamente
	if proceso.Estado != globals.Blocked {
		t.Errorf("Se esperaba estado %v, pero se obtuvo %v", globals.Blocked, proceso.Estado)
	}
}

func TestCambiarEstadoHilo(t *testing.T) {
	// Crear un hilo simulado con un estado inicial
	hilo := &globals.TCB{
		TID:    1,
		Estado: globals.New,
	}

	// Cambiar el estado del proceso a Blocked
	CambiarEstadoHilo(hilo, globals.Blocked)

	// Verificar si el estado ha cambiado correctamente
	if hilo.Estado != globals.Blocked {
		t.Errorf("Se esperaba estado %v, pero se obtuvo %v", globals.Blocked, hilo.Estado)
	}
}

func TestEncolarProceso(t *testing.T) {
	// Creamos un nuevo proceso.
	proceso := &globals.PCB{
		PID:    1,
		TIDs:   []int{0},
		Estado: globals.New, // Estado inicial.
	}

	// Crear una cola de procesos para el test.
	cola := &globals.ColaProcesos{
		Procesos: []*globals.PCB{},
	}

	// Testear encolado en el estado NEW.
	EncolarProceso(proceso, cola, globals.New)

	if cola.Procesos[0].PID != proceso.PID {
		t.Errorf("Se esperaba que el proceso con PID %d fuera el primero en la cola, pero se encontró %d", proceso.PID, cola.Procesos[0].PID)
	}
	if cola.Procesos[0].Estado != globals.New {
		t.Errorf("Se esperaba que el estado del proceso fuera %d, pero se encontró %d", globals.New, cola.Procesos[0].Estado)
	}

	// Testear encolado en el estado READY.
	EncolarProceso(proceso, cola, globals.Ready)

	if len(cola.Procesos) != 2 {
		t.Errorf("Se esperaba que la cola tuviera 2 procesos, pero tiene %d", len(cola.Procesos))
	}
	if cola.Procesos[1].Estado != globals.Ready {
		t.Errorf("Se esperaba que el estado del segundo proceso fuera %d, pero se encontró %d", globals.Ready, cola.Procesos[1].Estado)
	}

	// Testear encolado en el estado EXEC.
	EncolarProceso(proceso, cola, globals.Exec)

	if len(cola.Procesos) != 3 {
		t.Errorf("Se esperaba que la cola tuviera 3 procesos, pero tiene %d", len(cola.Procesos))
	}
	if cola.Procesos[2].Estado != globals.Exec {
		t.Errorf("Se esperaba que el estado del tercer proceso fuera %d, pero se encontró %d", globals.Exec, cola.Procesos[2].Estado)
	}

	// Testear encolado en el estado BLOCKED.
	EncolarProceso(proceso, cola, globals.Blocked)

	if len(cola.Procesos) != 4 {
		t.Errorf("Se esperaba que la cola tuviera 4 procesos, pero tiene %d", len(cola.Procesos))
	}
	if cola.Procesos[3].Estado != globals.Blocked {
		t.Errorf("Se esperaba que el estado del cuarto proceso fuera %d, pero se encontró %d", globals.Blocked, cola.Procesos[3].Estado)
	}

	// Testear encolado en el estado EXIT.
	EncolarProceso(proceso, cola, globals.Exit)

	if len(cola.Procesos) != 5 {
		t.Errorf("Se esperaba que la cola tuviera 5 procesos, pero tiene %d", len(cola.Procesos))
	}
	if cola.Procesos[4].Estado != globals.Exit {
		t.Errorf("Se esperaba que el estado del quinto proceso fuera %d, pero se encontró %d", globals.Exit, cola.Procesos[4].Estado)
	}

	// Testear encolado con un estado desconocido
	EncolarProceso(proceso, cola, 99) // Estado desconocido.

	// Verificar que el proceso no fue añadido a la cola.
	if len(cola.Procesos) != 5 { // Debería seguir siendo 5.
		t.Errorf("Se esperaba que la cola tuviera 5 procesos, pero tiene %d después de intentar encolar un estado desconocido", len(cola.Procesos))
	}
}

func TestEncolarHilo(t *testing.T) {
	// Creamos un nuevo hilo.
	hilo := &globals.TCB{
		TID:    1,
		Estado: globals.New, // Estado inicial.
	}

	// Crear una cola de hilos para el test.
	cola := &globals.ColaHilos{
		Hilos: []*globals.TCB{},
	}

	// Testear encolado en el estado NEW.
	EncolarHilo(hilo, cola, globals.New)

	if len(cola.Hilos) != 0 {
		t.Errorf("Se esperaba que la cola no tenga procesos, ya que no existe cola New para hilos, pero tiene %d.", len(cola.Hilos))
	}

	// Testear encolado en el estado READY.
	EncolarHilo(hilo, cola, globals.Ready)

	if len(cola.Hilos) != 0 {
		t.Errorf("Se esperaba que la cola no tenga procesos; ya que utilizaremos otro mecanismo para introducir hilos en Ready; pero tiene %d", len(cola.Hilos))
	}

	// Testear encolado en el estado EXEC.
	EncolarHilo(hilo, cola, globals.Exec)

	if len(cola.Hilos) != 1 {
		t.Errorf("Se esperaba que la cola tuviera 1 hilo, pero tiene %d", len(cola.Hilos))
	}
	if cola.Hilos[0].Estado != globals.Exec {
		t.Errorf("Se esperaba que el estado del primer hilo fuera %d, pero se encontró %d", globals.Exec, cola.Hilos[0].Estado)
	}

	// Testear encolado en el estado BLOCKED.
	EncolarHilo(hilo, cola, globals.Blocked)

	if len(cola.Hilos) != 2 {
		t.Errorf("Se esperaba que la cola tuviera 2 hilos, pero tiene %d", len(cola.Hilos))
	}
	if cola.Hilos[1].Estado != globals.Blocked {
		t.Errorf("Se esperaba que el estado del segundo hilo fuera %d, pero se encontró %d", globals.Blocked, cola.Hilos[1].Estado)
	}

	// Testear encolado en el estado EXIT.
	EncolarHilo(hilo, cola, globals.Exit)

	if len(cola.Hilos) != 3 {
		t.Errorf("Se esperaba que la cola tuviera 3 hilos, pero tiene %d", len(cola.Hilos))
	}
	if cola.Hilos[2].Estado != globals.Exit {
		t.Errorf("Se esperaba que el estado del tercer hilo fuera %d, pero se encontró %d", globals.Exit, cola.Hilos[2].Estado)
	}

	// Testear encolado con un estado desconocido
	EncolarHilo(hilo, cola, 99) // Estado desconocido.

	// Verificar que el proceso no fue añadido a la cola.
	if len(cola.Hilos) != 3 { // Debería seguir siendo 3.
		t.Errorf("Se esperaba que la cola tuviera 3 hilos, pero tiene %d después de intentar encolar un estado desconocido", len(cola.Hilos))
	}
}

//func TestDesencolarHilo(t *testing.T) 

/*
func TestDesencolarHilo(t *testing.T) {
	var colaARetirar *globals.ColaHilos

	// Identificar la cola correcta según el estado y prioridad del hilo.
	switch hilo.Estado {
	case globals.Ready:
		// Según la prioridad del hilo, seleccionamos la cola de Ready correspondiente.
		switch hilo.Prioridad {
		case 0:
			colaARetirar = &globals.ColaReadyPrioridad0Hilos
		case 1:
			colaARetirar = &globals.ColaReadyPrioridad1Hilos
		case 2:
			colaARetirar = &globals.ColaReadyPrioridad2Hilos
		default:
			slog.Error("Prioridad del hilo desconocida: " + strconv.Itoa(hilo.Prioridad))
			return
		}
	case globals.Blocked:
		colaARetirar = &globals.ColaBlockedHilos

	case globals.Exec:
		colaARetirar = &globals.ColaExecHilos

	default:
		slog.Error("Estado del hilo desconocido: " + hilo.Estado.String())
		return
	}

	for i, tcb := range colaARetirar.Hilos { // Busca el hilo a eliminar en la lista.
		if tcb.TID == hilo.TID {
			// Eliminar el Hilo usando slicing.
			colaARetirar.Hilos = append(colaARetirar.Hilos[:i], colaARetirar.Hilos[i+1:]...) // Cuando lo encuentra lo borra de la lista con Slice.

			slog.Info("Hilo con TID <", strconv.Itoa(hilo.TID), "> eliminado de la cola: "+colaARetirar.String())
			break
		}
	}
	slog.Debug("No se encontro en ninguna cola el hilo con TID: " + strconv.Itoa(hilo.TID))
}
*/

// Test de DesencolarProceso
func TestDesencolarProceso(t *testing.T) {
	// Inicializamos colas de procesos.
	globals.ColaNewProcesos = globals.ColaProcesos{Procesos: []*globals.PCB{}}
	globals.ColaReadyProcesos = globals.ColaProcesos{Procesos: []*globals.PCB{}}
	globals.ColaExecProcesos = globals.ColaProcesos{Procesos: []*globals.PCB{}}
	globals.ColaBlockedProcesos = globals.ColaProcesos{Procesos: []*globals.PCB{}}

	// Crear un nuevo proceso
	proceso := &globals.PCB{
		PID:    1,
		TIDs:   []int{0},
		Estado: globals.New, // Estado inicial.
	}

	// Encolamos el proceso en la cola NEW.
	globals.ColaNewProcesos.Procesos = append(globals.ColaNewProcesos.Procesos, proceso)

	// Desencolamos el proceso de la cola donde esté.
	DesencolarProceso(proceso)

	// Verificamos que el proceso haya sido eliminado de la cola NEW.
	if len(globals.ColaNewProcesos.Procesos) != 0 {
		t.Errorf("Se esperaba que la cola NEW estuviera vacía, pero tiene %d procesos", len(globals.ColaNewProcesos.Procesos))
	}

	// Volvemos a encolar el proceso para probar otra cola.
	proceso.Estado = globals.Ready
	globals.ColaReadyProcesos.Procesos = append(globals.ColaReadyProcesos.Procesos, proceso)

	// Desencolamos el proceso de la cola donde esté.
	DesencolarProceso(proceso)

	// Verificar que el proceso fue eliminado de la cola READY
	if len(globals.ColaReadyProcesos.Procesos) != 0 {
		t.Errorf("Se esperaba que la cola READY estuviera vacía, pero tiene %d procesos", len(globals.ColaReadyProcesos.Procesos))
	}

	// Volvemos a encolar el proceso para probar otra cola.
	proceso.Estado = globals.Exec // Asignar un estado válido
	globals.ColaExecProcesos.Procesos = append(globals.ColaExecProcesos.Procesos, proceso)

	// Desencolamos el proceso de la cola donde esté.
	DesencolarProceso(proceso)

	// Verificamos que el proceso fue eliminado de la cola EXEC.
	if len(globals.ColaExecProcesos.Procesos) != 0 {
		t.Errorf("Se esperaba que la cola EXEC estuviera vacía, pero tiene %d procesos", len(globals.ColaExecProcesos.Procesos))
	}

	// Probamos con un proceso que no existe en la cola.
	procesoNoExistente := &globals.PCB{PID: 999, Estado: globals.Ready}
	DesencolarProceso(procesoNoExistente)

	// Verificamos que no se produce un error y que las colas no cambian
	if len(globals.ColaReadyProcesos.Procesos) != 0 {
		t.Errorf("Se esperaba que la cola READY continuara vacía, pero tiene %d procesos", len(globals.ColaReadyProcesos.Procesos))
	}
}

// func TestCrearProceso(t *testing.T) {

// 	result := CrearProceso(5) // Almacenamos el resultado.

// 	// Verificar los valores de los campos PID, TIDs asociados, Prioridad y Estado
// 	if result.PID != 5 {
// 		t.Errorf("Se esperaba PID 5, pero se obtuvo %d", result.PID)
// 	}
// 	if !reflect.DeepEqual(result.TIDs, []int{0}) {
// 		t.Errorf("Los TIDS no coinciden. Se esperaba %v, pero se obtuvo %v", []int{0}, result.TIDs)
// 	}
// 	/*if !reflect.DeepEqual(result.Mutex, []sync.Mutex{}) {
// 		t.Errorf("Los Mutex no coinciden. Se esperaba %v, pero se obtuvo %v", []sync.Mutex{}, result.Mutex)
// 	}*/
	
// 	if result.Estado != globals.New {
// 		t.Errorf("Se esperaba Estado 'New', pero se obtuvo %v", result.Estado)
// 	}
// }

func TestCrearHilo(t *testing.T) {
	instrucciones := []globals.Instruccion{
		{Operacion: "ADD", Parametros: []string{"5", "10"}},
		{Operacion: "SUB", Parametros: []string{"20", "4"}},
		{Operacion: "MULT", Parametros: []string{"3", "7"}},
	}

	result := CrearHilo(10, 1, 2, instrucciones) // Almacenamos el resultado.

	// Verificar los valores de los campos PID, TID, Prioridad y Estado
	if result.PID != 10 {
		t.Errorf("Se esperaba PID 10, pero se obtuvo %d", result.PID)
	}
	if result.TID != 1 {
		t.Errorf("Se esperaba TID 1, pero se obtuvo %d", result.TID)
	}
	if result.Prioridad != 2 {
		t.Errorf("Se esperaba Prioridad 2, pero se obtuvo %d", result.Prioridad)
	}
	if result.Estado != globals.Ready {
		t.Errorf("Se esperaba Estado 'Ready', pero se obtuvo %v", result.Estado)
	}

	// Verificar que las instrucciones se cargaron correctamente
	if !reflect.DeepEqual(result.Codigo, instrucciones) {
		t.Errorf("Las instrucciones no coinciden. Se esperaba %v, pero se obtuvo %v", instrucciones, result.Codigo)
	}
}

func TestEncontrarPCBPorTID(t *testing.T) {
	// Inicializamos los procesos en el sistema para la prueba.
	globals.ProcesosDelSistema = globals.ColaProcesos{
		Procesos: []*globals.PCB{
			{PID: 1, TIDs: []int{101, 102}, Estado: globals.New, Tamanio: 1024},
			{PID: 2, TIDs: []int{201, 202}, Estado: globals.New, Tamanio: 2048},
			{PID: 3, TIDs: []int{301, 302}, Estado: globals.New, Tamanio: 3072},
		},
	}

	// Caso 1: TID que existe.
	tidExistente := 102
	pidPadre := 1

	pcb := EncontrarPCBPorTID(tidExistente, pidPadre)

	if pcb == nil {
		t.Errorf("Se esperaba encontrar el PCB con TID %d, pero se devolvió nil", tidExistente)
	} else if pcb.PID != 1 {
		t.Errorf("Se esperaba el PCB con PID 1, pero se obtuvo PCB con PID %d", pcb.PID)
	}

	// Caso 2: TID que no existe
	tidInexistente := 999
	pidPadreInexistente := 2

	pcb = EncontrarPCBPorTID(tidInexistente, pidPadreInexistente)

	if pcb != nil {
		t.Errorf("Se esperaba nil para TID %d, pero se devolvió un PCB con PID %d", tidInexistente, pcb.PID)
	}
}