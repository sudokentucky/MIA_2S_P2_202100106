package commands

import (
	structures "backend/Structs"
	globals "backend/globals"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
)

// Unmount estructura para representar el comando unmount
type Unmount struct {
	id string // ID de la partición a desmontar
}

// ParserUnmount parsea el comando unmount y devuelve una instancia de Unmount junto con un buffer de salida
func ParserUnmount(tokens []string) (string, error) {
	var outputBuffer bytes.Buffer
	cmd := &Unmount{}

	// Parsear el argumento -id
	for _, token := range tokens {
		if strings.HasPrefix(token, "-id=") {
			cmd.id = strings.TrimPrefix(token, "-id=")
		}
	}

	// Validar que el ID no esté vacío
	if cmd.id == "" {
		return "", errors.New("faltan parámetros requeridos: -id")
	}

	// Ejecutar el comando unmount y capturar los mensajes importantes en el buffer
	err := commandUnmount(cmd, &outputBuffer)
	if err != nil {
		return "", err
	}

	return outputBuffer.String(), nil
}

// commandUnmount ejecuta el comando unmount para desmontar la partición con el ID especificado
func commandUnmount(unmount *Unmount, outputBuffer *bytes.Buffer) error {
	fmt.Fprintln(outputBuffer, "========================== UNMOUNT ==========================")

	// Verificar si el ID de la partición existe en las particiones montadas globales
	mountedPath, exists := globals.MountedPartitions[unmount.id]
	if !exists {
		return fmt.Errorf("error: la partición con ID '%s' no está montada", unmount.id)
	}

	// Abrir el archivo del disco
	file, err := os.OpenFile(mountedPath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error abriendo el archivo del disco: %v", err)
	}
	defer file.Close()

	// Leer el MBR del disco
	var mbr structures.MBR
	err = mbr.Decode(file)
	if err != nil {
		return fmt.Errorf("error deserializando el MBR: %v", err)
	}

	// Buscar la partición en el MBR que tiene el ID especificado
	found := false
	for i := range mbr.MbrPartitions {
		partition := &mbr.MbrPartitions[i] // Obtener referencia a la partición
		partitionID := strings.TrimSpace(string(partition.Part_id[:]))
		if partitionID == unmount.id {
			// Desmontar la partición: Cambiar el valor del correlativo a 0
			err = partition.MountPartition(0, "")
			if err != nil {
				return fmt.Errorf("error desmontando la partición: %v", err)
			}

			// Actualizar el MBR en el archivo después del desmontaje
			err = mbr.Encode(file)
			if err != nil {
				return fmt.Errorf("error al actualizar el MBR en el disco: %v", err)
			}

			found = true
			break
		}
	}

	// Si no se encontró la partición con el ID, devolver error
	if !found {
		return fmt.Errorf("error: no se encontró la partición con ID '%s' en el disco", unmount.id)
	}

	// Remover el ID de la partición de la lista de particiones montadas
	delete(globals.MountedPartitions, unmount.id)

	// Imprimir el estado después del desmontaje
	fmt.Fprintf(outputBuffer, "Partición con ID '%s' desmontada exitosamente.\n", unmount.id)
	fmt.Fprintln(outputBuffer, "\n=== Particiones Montadas ===")
	for id, path := range globals.MountedPartitions {
		fmt.Fprintf(outputBuffer, "ID: %s | Path: %s\n", id, path)
	}
	fmt.Fprintln(outputBuffer, "===========================================================")

	return nil
}
