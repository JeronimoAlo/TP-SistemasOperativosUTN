package utils_general

import (
	"testing"
)

func TestMax(t *testing.T) {
	// Caso normal.

	lista_numeros := []int{1,4,6,9,15}
	
	result, err := Max(lista_numeros) // Almacenamos el resultado.

	if err != nil || result == 0 { // Si hay un error
		t.Errorf("Se produjo un error: %v", err)
	} else if result != 15 { // Si no hay error, verificamos el resultado
		t.Errorf("Se esperaba que el mayor sea 15, pero se obtuvo %d", result)
	}

	// Caso de lista vacía
	emptyList := []int{}
	_, err = Max(emptyList)

	if err == nil {
		t.Errorf("Se esperaba un error debido a que la lista está vacía, pero no se produjo.")
	} else if err.Error() != "la lista está vacía" {
		t.Errorf("Se esperaba el error 'la lista está vacía', pero se obtuvo: %v", err)
	}
}