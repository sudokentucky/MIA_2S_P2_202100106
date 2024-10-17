package structs

import (
	"encoding/binary"
	"fmt"
	"os"
)

// PointerBlock : Estructura para guardar los bloques de apuntadores
type PointerBlock struct {
	B_pointers [16]int64 // Apuntadores a bloques de carpetas o datos
}

// FindFreePointer busca el primer apuntador libre en un bloque de apuntadores y devuelve su índice
func (pb *PointerBlock) FindFreePointer() (int, error) {
	for i, pointer := range pb.B_pointers {
		if pointer == -1 { // -1 es un apuntador libre
			return i, nil
		}
	}
	return -1, fmt.Errorf("no hay apuntadores libres en el bloque de apuntadores")
}

// SetPointer establece un valor específico en un índice dado
func (pb *PointerBlock) SetPointer(index int, value int64) error {
	if index < 0 || index >= len(pb.B_pointers) {
		return fmt.Errorf("índice fuera de rango")
	}
	pb.B_pointers[index] = value
	return nil
}

// GetPointer obtiene el valor de un apuntador en un índice dado
func (pb *PointerBlock) GetPointer(index int) (int64, error) {
	if index < 0 || index >= len(pb.B_pointers) {
		return -1, fmt.Errorf("índice fuera de rango")
	}
	return pb.B_pointers[index], nil
}

// IsFull verifica si todos los apuntadores están ocupados
func (pb *PointerBlock) IsFull() bool {
	for _, pointer := range pb.B_pointers {
		if pointer == -1 {
			return false
		}
	}
	return true
}

// CountFreePointers cuenta cuántos apuntadores libres hay en el bloque
func (pb *PointerBlock) CountFreePointers() int {
	count := 0
	for _, pointer := range pb.B_pointers {
		if pointer == -1 {
			count++
		}
	}
	return count
}

// Encode serializa el PointerBlock en el archivo en la posición dada
func (pb *PointerBlock) Encode(file *os.File, offset int64) error {
	// Mover el cursor del archivo a la posición deseada
	_, err := file.Seek(offset, 0)
	if err != nil {
		return fmt.Errorf("error buscando la posición en el archivo: %w", err)
	}

	// Escribir la estructura PointerBlock en el archivo
	err = binary.Write(file, binary.BigEndian, pb)
	if err != nil {
		return fmt.Errorf("error escribiendo el PointerBlock: %w", err)
	}
	return nil
}

// Decode deserializa el PointerBlock desde el archivo en la posición dada
func (pb *PointerBlock) Decode(file *os.File, offset int64) error {
	// Mover el cursor del archivo a la posición deseada
	_, err := file.Seek(offset, 0)
	if err != nil {
		return fmt.Errorf("error buscando la posición en el archivo: %w", err)
	}

	// Leer la estructura PointerBlock desde el archivo
	err = binary.Read(file, binary.BigEndian, pb)
	if err != nil {
		return fmt.Errorf("error leyendo el PointerBlock: %w", err)
	}
	return nil
}
