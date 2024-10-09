package env

import (
	structs "backend/Structs" // Importa tus estructuras MBR, EBR y Partition
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
)

type DiskReader struct {
	file        *os.File
	isEncrypted bool
	key         byte
}

// Nueva instancia de DiskReader
func NewDiskReader(path string, isEncrypted bool, key byte) (*DiskReader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error al abrir el archivo de disco: %v", err)
	}
	return &DiskReader{
		file:        file,
		isEncrypted: isEncrypted,
		key:         key,
	}, nil
}

// Método para cerrar el archivo
func (dr *DiskReader) Close() {
	if dr.file != nil {
		dr.file.Close()
	}
}

// Leer el MBR desde el archivo usando la función Decode
func (dr *DiskReader) ReadMBR() (*structs.MBR, error) {
	_, err := dr.file.Seek(0, 0)
	if err != nil {
		return nil, fmt.Errorf("error al reposicionar al inicio del archivo: %v", err)
	}

	// Decodificar el MBR
	mbr := &structs.MBR{}
	err = mbr.Decode(dr.file)
	if err != nil {
		return nil, fmt.Errorf("error al decodificar el MBR: %v", err)
	}

	return mbr, nil
}

// Leer particiones lógicas usando EBR
func (dr *DiskReader) ReadLogicalPartitions(extendedStart int32) ([]fiber.Map, error) {
	fmt.Printf("Leyendo particiones lógicas a partir de la partición extendida en %d...\n", extendedStart)

	logicalPartitions := []fiber.Map{}
	currentEBR := &structs.EBR{}

	for {
		err := currentEBR.Decode(dr.file, int64(extendedStart))
		if err != nil {
			return nil, fmt.Errorf("error al decodificar el EBR: %v", err)
		}

		// Agregar la partición lógica al resultado
		partition := fiber.Map{
			"start": currentEBR.Ebr_start,
			"size":  currentEBR.Ebr_size,
			"name":  string(currentEBR.Ebr_name[:]),
			"next":  currentEBR.Ebr_next,
		}
		logicalPartitions = append(logicalPartitions, partition)

		// Si no hay más EBRs, salir del bucle
		if currentEBR.Ebr_next == -1 {
			break
		}

		// Mover al siguiente EBR
		extendedStart = currentEBR.Ebr_next
	}

	return logicalPartitions, nil
}
