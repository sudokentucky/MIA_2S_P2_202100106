package structs

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	utilidades "backend/utils" // Importa el paquete utils
)

// Estructura que representa un MBR
type MBR struct {
	MbrSize          int32        // Tamaño del MBR
	MbrCreacionDate  float32      // Fecha de creación del MBR
	MbrDiskSignature int32        // Número de serie del disco (random)
	MbrDiskFit       [1]byte      // BF = Best Fit, FF = First Fit, WF = Worst Fit
	MbrPartitions    [4]Partition // Particiones del MBR (4 particiones)
}

// Encode serializa la estructura MBR en un archivo
func (mbr *MBR) Encode(file *os.File) error {
	return utilidades.WriteToFile(file, 0, mbr) // Escribe el MBR en el inicio del archivo
}

// Decode deserializa la estructura MBR desde un archivo
func (mbr *MBR) Decode(file *os.File) error {
	return utilidades.ReadFromFile(file, 0, mbr) // Lee el MBR desde el inicio del archivo
}

// Método para obtener la primera partición disponible
func (mbr *MBR) GetFirstAvailablePartition() (*Partition, int, int) {
	// Calcular el offset para el start de la partición
	offset := binary.Size(mbr) // Tamaño del MBR en bytes

	// Recorrer las particiones del MBR
	for i := 0; i < len(mbr.MbrPartitions); i++ {
		if mbr.MbrPartitions[i].Part_start == -1 { // -1 disponible
			return &mbr.MbrPartitions[i], offset, i
		} else {
			// Calcular el nuevo offset para la siguiente partición, es decir, sumar el tamaño de la partición
			offset += int(mbr.MbrPartitions[i].Part_size)
		}
	}
	return nil, -1, -1
}

// Método para obtener una partición por nombre
func (mbr *MBR) GetPartitionByName(name string) (*Partition, int) {
	for i, partition := range mbr.MbrPartitions {
		partitionName := strings.Trim(string(partition.Part_name[:]), "\x00 ")
		inputName := strings.Trim(name, "\x00 ")
		// Si el nombre de la partición coincide, devolver la partición y el índice
		if strings.EqualFold(partitionName, inputName) {
			return &partition, i
		}
	}
	return nil, -1
}

// Función para obtener una partición por ID
func (mbr *MBR) GetPartitionByID(id string) (*Partition, error) {
	for i := 0; i < len(mbr.MbrPartitions); i++ {
		partitionID := strings.Trim(string(mbr.MbrPartitions[i].Part_id[:]), "\x00 ")
		inputID := strings.Trim(id, "\x00 ")
		// Si el nombre de la partición coincide, devolver la partición
		if strings.EqualFold(partitionID, inputID) {
			return &mbr.MbrPartitions[i], nil
		}
	}
	return nil, errors.New("partición no encontrada")
}

// HasExtendedPartition verifica si ya existe una partición extendida en el MBR
func (mbr *MBR) HasExtendedPartition() bool {
	for _, partition := range mbr.MbrPartitions {
		// Verificar si la partición es extendida
		if partition.Part_type[0] == 'E' {
			return true // Devolver verdadero si se encuentra una partición extendida
		}
	}
	return false
}

// CalculateAvailableSpace calcula el espacio disponible en el disco.
func (mbr *MBR) CalculateAvailableSpace() (int32, error) {
	totalSize := mbr.MbrSize
	usedSpace := int32(binary.Size(MBR{})) // Tamaño del MBR

	partitions := mbr.MbrPartitions[:] // Obtener todas las particiones
	for _, part := range partitions {
		if part.Part_size != 0 { // Si la partición está ocupada
			usedSpace += part.Part_size
		}
	}

	if usedSpace >= totalSize {
		return 0, fmt.Errorf("there is no available space on the disk")
	}

	return totalSize - usedSpace, nil
}

// Método para imprimir los valores del MBR
func (mbr *MBR) Print() {
	creationTime := time.Unix(int64(mbr.MbrCreacionDate), 0)
	diskFit := rune(mbr.MbrDiskFit[0])
	fmt.Printf("MBR Size: %d | Creation Date: %s | Disk Signature: %d | Disk Fit: %c\n",
		mbr.MbrSize, creationTime.Format(time.RFC3339), mbr.MbrDiskSignature, diskFit)
}

// Método para imprimir las particiones del MBR
func (mbr *MBR) PrintPartitions() {
	for i, partition := range mbr.MbrPartitions {
		partStatus := rune(partition.Part_status[0])
		partType := rune(partition.Part_type[0])
		partFit := rune(partition.Part_fit[0])
		partName := strings.TrimSpace(string(partition.Part_name[:]))
		partID := strings.TrimSpace(string(partition.Part_id[:]))

		// Imprimir en una sola línea la información de cada partición
		fmt.Printf("Partition %d: Status: %c | Type: %c | Fit: %c | Start: %d | Size: %d | Name: %s | Correlative: %d | ID: %s\n",
			i+1, partStatus, partType, partFit, partition.Part_start, partition.Part_size, partName, partition.Part_correlative, partID)
	}
}

// Método que aplica un ajuste a las particiones del MBR (First Fit, Best Fit, Worst Fit)
func (mbr *MBR) ApplyFit(partitionSize int32) (*Partition, error) {
	availableSpace, err := mbr.CalculateAvailableSpace()
	if err != nil {
		return nil, err
	}

	if availableSpace < partitionSize {
		return nil, fmt.Errorf("no hay suficiente espacio en el disco")
	}

	switch rune(mbr.MbrDiskFit[0]) {
	case 'F': // First Fit
		return mbr.ApplyFirstFit(partitionSize)
	case 'B': // Best Fit
		return mbr.ApplyBestFit(partitionSize)
	case 'W': // Worst Fit
		return mbr.ApplyWorstFit(partitionSize)
	default:
		return nil, fmt.Errorf("tipo de ajuste inválido")
	}
}

// First Fit: Encuentra el primer espacio disponible que sea mayor o igual al tamaño de la partición
func (mbr *MBR) ApplyFirstFit(partitionSize int32) (*Partition, error) {
	fmt.Println("Iniciando First Fit...")

	offset := binary.Size(mbr)
	for i := 0; i < len(mbr.MbrPartitions); i++ {
		partition := &mbr.MbrPartitions[i]

		fmt.Printf("Evaluando partición %d: Inicio %d, Tamaño %d, Estado %c\n", i, partition.Part_start, partition.Part_size, partition.Part_status[0])

		// Verificar si la partición está disponible y si el tamaño es suficiente
		if partition.Part_start == -1 && (partition.Part_size == -1 || partition.Part_size >= partitionSize) {
			fmt.Printf("Partición %d es adecuada para First Fit: Inicio en %d, Tamaño %d\n", i, offset, partitionSize)

			// Asignar el inicio y el tamaño de la partición
			partition.Part_start = int32(offset)
			partition.Part_size = partitionSize
			return partition, nil
		}
		// Actualizar el offset con el tamaño de la partición actual
		offset += int(partition.Part_size)
	}

	fmt.Println("No se encontró espacio suficiente con First Fit.")
	return nil, fmt.Errorf("no se encontró espacio suficiente con First Fit")
}

// Best Fit: Encuentra el espacio disponible más pequeño que sea mayor o igual al tamaño de la partición
func (mbr *MBR) ApplyBestFit(partitionSize int32) (*Partition, error) {
	fmt.Println("Iniciando Best Fit...")

	bestFit := int32(-1)
	bestPartition := -1
	offset := binary.Size(mbr)

	for i := 0; i < len(mbr.MbrPartitions); i++ {
		partition := &mbr.MbrPartitions[i]

		fmt.Printf("Evaluando partición %d: Inicio %d, Tamaño %d, Estado %c\n", i, partition.Part_start, partition.Part_size, partition.Part_status[0])

		if partition.Part_start == -1 && (partition.Part_size == -1 || partition.Part_size >= partitionSize) {
			fmt.Printf("Partición %d es candidata para Best Fit: Tamaño actual %d\n", i, partition.Part_size)

			if bestFit == -1 || partition.Part_size < bestFit {
				bestFit = partition.Part_size
				bestPartition = i
				fmt.Printf("Partición %d es la mejor opción actual para Best Fit con tamaño %d\n", i, bestFit)
			}
		}
		offset += int(partition.Part_size)
	}

	if bestPartition == -1 {
		fmt.Println("No se encontró espacio suficiente con Best Fit.")
		return nil, fmt.Errorf("no se encontró espacio suficiente con Best Fit")
	}

	partition := &mbr.MbrPartitions[bestPartition]
	partition.Part_start = int32(offset)
	partition.Part_size = partitionSize
	fmt.Printf("Partición %d seleccionada para Best Fit: Inicio en %d, Tamaño %d\n", bestPartition, offset, partitionSize)
	return partition, nil
}

// Worst Fit: Encuentra el espacio disponible más grande que sea mayor o igual al tamaño de la partición
func (mbr *MBR) ApplyWorstFit(partitionSize int32) (*Partition, error) {
	fmt.Println("Iniciando Worst Fit...")

	worstFit := int32(-1)
	worstPartition := -1
	offset := binary.Size(mbr)

	for i := 0; i < len(mbr.MbrPartitions); i++ {
		partition := &mbr.MbrPartitions[i]

		fmt.Printf("Evaluando partición %d: Inicio %d, Tamaño %d, Estado %c\n", i, partition.Part_start, partition.Part_size, partition.Part_status[0])

		if partition.Part_start == -1 && (partition.Part_size == -1 || partition.Part_size >= partitionSize) {
			fmt.Printf("Partición %d es candidata para Worst Fit: Tamaño actual %d\n", i, partition.Part_size)

			if worstFit == -1 || partition.Part_size > worstFit {
				worstFit = partition.Part_size
				worstPartition = i
				fmt.Printf("Partición %d es la peor opción actual para Worst Fit con tamaño %d\n", i, worstFit)
			}
		}
		offset += int(partition.Part_size)
	}

	if worstPartition == -1 {
		fmt.Println("No se encontró espacio suficiente con Worst Fit.")
		return nil, fmt.Errorf("no se encontró espacio suficiente con Worst Fit")
	}

	partition := &mbr.MbrPartitions[worstPartition]
	partition.Part_start = int32(offset)
	partition.Part_size = partitionSize
	fmt.Printf("Partición %d seleccionada para Worst Fit: Inicio en %d, Tamaño %d\n", worstPartition, offset, partitionSize)
	return partition, nil
}

// Unificar la creación de la partición y el ajuste (fit)
func (mbr *MBR) CreatePartitionWithFit(partSize int32, partType, partName string) error {
	// Verificar si hay espacio suficiente en el disco
	availableSpace, err := mbr.CalculateAvailableSpace()
	if err != nil {
		return fmt.Errorf("error calculando el espacio disponible: %v", err)
	}
	if availableSpace < partSize {
		return fmt.Errorf("no hay suficiente espacio en el disco para la nueva partición")
	}

	// Aplicar el ajuste (fit) almacenado en el MBR
	partition, err := mbr.ApplyFit(partSize)
	if err != nil {
		return fmt.Errorf("error al aplicar el ajuste: %v", err)
	}

	// Ahora unificamos la creación de la partición directamente aquí
	partition.Part_status[0] = '1' // Activar partición (1 = Activa)
	partition.Part_size = partSize

	// Asignar el tipo de partición (Primaria o Extendida)
	if len(partType) > 0 {
		partition.Part_type[0] = partType[0]
	}

	// Asignar el tipo de ajuste (fit) basado en el MBR
	switch mbr.MbrDiskFit[0] {
	case 'B', 'F', 'W': // Verificamos si es Best Fit, First Fit o Worst Fit
		partition.Part_fit[0] = mbr.MbrDiskFit[0]
	default:
		return fmt.Errorf("ajuste inválido en el MBR: %c. Debe ser BF (Best Fit), FF (First Fit) o WF (Worst Fit)", mbr.MbrDiskFit[0])
	}

	// Asignar el nombre de la partición
	copy(partition.Part_name[:], partName)

	fmt.Printf("Partición '%s' creada exitosamente con el ajuste '%c'.\n", partName, mbr.MbrDiskFit[0])
	return nil
}
