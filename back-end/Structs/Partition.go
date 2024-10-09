package structs

import (
	"fmt"
	"os"
	"strings"
)

// Constantes para los valores de fit
const (
	FitBF = 'B' // Best Fit
	FitFF = 'F' // First Fit
	FitWF = 'W' // Worst Fit
)

// Estructura que representa una partición
type Partition struct {
	Part_status      [1]byte  // Indica si la partición está activa o no
	Part_type        [1]byte  // P = Primaria, E = Extendida
	Part_fit         [1]byte  // B = Best Fit, F = First Fit, W = Worst Fit
	Part_start       int32    // Byte donde inicia la partición
	Part_size        int32    // Tamaño de la partición en bytes
	Part_name        [16]byte // Nombre de la partición
	Part_correlative int32    // Inicialmente 0, se asigna cuando se monta la partición
	Part_id          [4]byte  // Inicialmente vacío, se asigna cuando se monta la partición
}

// Método que modifica el tamaño de una partición
func (p *Partition) ModifySize(addSize int32, availableSpace int32) error {
	newSize := p.Part_size + addSize

	// Validar que el tamaño nuevo no sea negativo
	if newSize < 0 {
		return fmt.Errorf("el tamaño de la partición no puede ser negativo")
	}

	// Si estamos aumentando el tamaño, validar que haya suficiente espacio disponible
	if addSize > 0 && availableSpace < addSize {
		return fmt.Errorf("no hay suficiente espacio disponible para agregar a la partición")
	}

	// Modificar el tamaño de la partición
	p.Part_size = newSize
	fmt.Printf("El tamaño de la partición '%s' ha sido modificado. Nuevo tamaño: %d bytes.\n", string(p.Part_name[:]), p.Part_size)
	return nil
}

// Método que elimina la partición y sus particiones lógicas si es extendida (sin confirmación)
func (p *Partition) Delete(deleteType string, file *os.File, isExtended bool) error {
	// Si es una partición extendida, eliminar las particiones lógicas
	if isExtended {
		err := p.deleteLogicalPartitions(file)
		if err != nil {
			return fmt.Errorf("error al eliminar las particiones lógicas dentro de la partición extendida: %v", err)
		}
	}

	// Marcar la partición como vacía (eliminarla de la tabla de particiones)
	p.Part_start = -1
	p.Part_size = -1
	p.Part_name = [16]byte{}

	// Si el tipo de eliminación es 'full', sobrescribir el espacio
	if deleteType == "full" {
		err := p.Overwrite(file)
		if err != nil {
			return fmt.Errorf("error al sobrescribir la partición: %v", err)
		}
	}

	fmt.Printf("La partición '%s' ha sido eliminada (%s).\n", strings.TrimSpace(string(p.Part_name[:])), deleteType)
	return nil
}

// Método que sobrescribe el espacio de la partición con \0 (para eliminación Full)
func (p *Partition) Overwrite(file *os.File) error {
	// Mover el puntero del archivo al inicio de la partición
	_, err := file.Seek(int64(p.Part_start), 0)
	if err != nil {
		return err
	}

	// Crear un buffer de ceros del tamaño de la partición
	zeroes := make([]byte, p.Part_size)

	// Escribir los ceros en el archivo
	_, err = file.Write(zeroes)
	if err != nil {
		return fmt.Errorf("error al sobrescribir el espacio de la partición: %v", err)
	}

	fmt.Printf("Espacio de la partición sobrescrito con ceros.\n")
	return nil
}

// Método para eliminar todas las particiones lógicas dentro de una partición extendida
func (p *Partition) deleteLogicalPartitions(file *os.File) error {
	fmt.Println("Eliminando particiones lógicas dentro de la partición extendida...")

	// Variables para recorrer los EBRs (Extended Boot Records)
	var currentEBR EBR
	start := p.Part_start

	// Recorrer las particiones lógicas (EBRs)
	for {
		// Leer el EBR actual
		err := currentEBR.Decode(file, int64(start))
		if err != nil {
			return fmt.Errorf("error al leer el EBR: %v", err)
		}

		// Si no hay más EBRs (end of chain)
		if currentEBR.Ebr_start == -1 {
			break
		}

		// Eliminar la partición lógica actual marcando el EBR como vacío
		currentEBR.Ebr_start = -1
		currentEBR.Ebr_size = -1
		copy(currentEBR.Ebr_name[:], "")

		// Sobrescribir el espacio si es eliminación Full
		err = currentEBR.Overwrite(file)
		if err != nil {
			return fmt.Errorf("error al sobrescribir el EBR: %v", err)
		}

		// Mover al siguiente EBR en la cadena
		start = currentEBR.Ebr_next
	}

	fmt.Println("Particiones lógicas eliminadas exitosamente.")
	return nil
}

// Método que monta una partición
func (p *Partition) MountPartition(correlative int, id string) error {
	p.Part_correlative = int32(correlative)
	copy(p.Part_id[:], id)
	return nil
}

// Imprimir los valores de la partición en una sola línea
func (p *Partition) Print() {
	fmt.Printf("Status: %c | Type: %c | Fit: %c | Start: %d | Size: %d | Name: %s | Correlative: %d | ID: %s\n",
		p.Part_status[0], p.Part_type[0], p.Part_fit[0], p.Part_start, p.Part_size,
		string(p.Part_name[:]), p.Part_correlative, string(p.Part_id[:]))
}
