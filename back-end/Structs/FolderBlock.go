package structs

import (
	"backend/utils" // Asegúrate de ajustar el path del package "utils"
	"fmt"
	"os"
	"strings"
)

// FolderBlock representa un bloque de carpeta con 4 contenidos
type FolderBlock struct {
	B_content [4]FolderContent // 4 * 16 = 64 bytes
	// Total: 64 bytes
}

// FolderContent representa el contenido dentro de un bloque de carpeta
type FolderContent struct {
	B_name  [12]byte // Nombre del archivo o carpeta
	B_inodo int32    // Número de inodo al que apunta
	// Total: 16 bytes
}

// Encode serializa la estructura FolderBlock en un archivo binario en la posición especificada
func (fb *FolderBlock) Encode(file *os.File, offset int64) error {
	// Utilizamos la función WriteToFile del paquete utils
	err := utils.WriteToFile(file, offset, fb)
	if err != nil {
		return fmt.Errorf("error writing FolderBlock to file: %w", err)
	}
	return nil
}

// Decode deserializa la estructura FolderBlock desde un archivo binario en la posición especificada
func (fb *FolderBlock) Decode(file *os.File, offset int64) error {
	// Utilizamos la función ReadFromFile del paquete utils
	err := utils.ReadFromFile(file, offset, fb)
	if err != nil {
		return fmt.Errorf("error reading FolderBlock from file: %w", err)
	}
	return nil
}
func NewFolderBlock(selfInodo, parentInodo int32, additionalContents map[string]int32) *FolderBlock {
	fb := &FolderBlock{}

	// Asignar los primeros dos contenidos: . y ..
	copy(fb.B_content[0].B_name[:], ".") //A sí mismo
	fb.B_content[0].B_inodo = selfInodo

	copy(fb.B_content[1].B_name[:], "..") //Padre
	fb.B_content[1].B_inodo = parentInodo

	// Asignar contenidos adicionales a partir del índice 2.
	i := 2
	for name, inodo := range additionalContents {
		if i >= len(fb.B_content) {
			break // Evitamos exceder la capacidad del bloque.
		}

		// Copiar el nombre al array de 12 bytes.
		copy(fb.B_content[i].B_name[:], name)
		fb.B_content[i].B_inodo = inodo
		i++
	}

	// Rellenar los espacios restantes con nombres vacíos.
	for ; i < len(fb.B_content); i++ {
		copy(fb.B_content[i].B_name[:], "-")
		fb.B_content[i].B_inodo = -1
	}

	return fb
}

// tienen un B_inodo diferente de -1.
func (fb *FolderBlock) IsFull() bool {
	for _, content := range fb.B_content {
		if content.B_inodo == -1 {
			return false // Si encontramos un B_inodo con -1, el bloque no está lleno.
		}
	}
	return true // Si no hay ningún B_inodo con -1, el bloque está lleno.
}

// Print imprime los atributos del bloque de carpeta
func (fb *FolderBlock) Print() {
	for i, content := range fb.B_content {
		name := string(content.B_name[:])
		fmt.Printf("Content %d:\n", i+1)
		fmt.Printf("  B_name: %s\n", name)
		fmt.Printf("  B_inodo: %d\n", content.B_inodo)
	}
}

// RenameInFolderBlock busca un archivo o carpeta dentro del FolderBlock
func (fb *FolderBlock) RenameInFolderBlock(oldName string, newName string) error {
	// Iterar sobre los contenidos del bloque, empezando desde el índice 2 (evitando . y ..)
	for i := 2; i < len(fb.B_content); i++ {
		content := &fb.B_content[i]                                     // Apuntamos al contenido actual
		currentName := strings.Trim(string(content.B_name[:]), "\x00 ") // Nombre actual del contenido, eliminando caracteres nulos

		// Si encontramos el nombre que queremos cambiar
		if strings.EqualFold(currentName, oldName) && content.B_inodo != -1 {
			// Validamos que el nuevo nombre no exceda los 12 bytes
			if len(newName) > 12 {
				return fmt.Errorf("el nuevo nombre '%s' es demasiado largo, máximo 12 caracteres", newName)
			}

			// Copiar el nuevo nombre al campo B_name
			copy(content.B_name[:], newName)

			// Rellenar con ceros los espacios sobrantes del nombre nuevo (si el nuevo nombre es más corto que el anterior)
			for j := len(newName); j < 12; j++ {
				content.B_name[j] = 0
			}

			fmt.Printf("Nombre cambiado de '%s' a '%s' en la posición %d\n", oldName, newName, i+1)
			return nil // Nombre cambiado exitosamente
		}
	}

	return fmt.Errorf("el nombre '%s' no fue encontrado en los inodos 3 o 4", oldName)
}
