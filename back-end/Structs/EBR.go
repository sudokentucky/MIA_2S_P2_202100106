package structs

import (
	utilidades "backend/utils" // Importa el paquete utils
	"fmt"
	"os"
)

// EBR representa el Extended Boot Record
type EBR struct {
	Ebr_mount [1]byte  // Indica si la partición está montada o no
	Ebr_fit   [1]byte  // BF = Best Fit, FF = First Fit, WF = Worst Fit
	Ebr_start int32    // Byte donde inicia la partición
	Ebr_size  int32    // Tamaño de la partición en bytes
	Ebr_next  int32    // Byte donde inicia el siguiente EBR, -1 si no hay siguiente
	Ebr_name  [16]byte // Nombre de la partición
}

// Encode serializa la estructura EBR en un archivo en la posición especificada
func (e *EBR) Encode(file *os.File, position int64) error {
	return utilidades.WriteToFile(file, position, e)
}

// Decode deserializa la estructura EBR desde un archivo en la posición especificada
func (e *EBR) Decode(file *os.File, position int64) error {
	// Verificar que la posición no sea negativa y esté dentro del rango del archivo
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("error al obtener información del archivo: %v", err)
	}
	if position < 0 || position >= fileInfo.Size() {
		return fmt.Errorf("posición inválida para EBR: %d", position)
	}

	err = utilidades.ReadFromFile(file, position, e)
	if err != nil {
		return err
	}

	fmt.Printf("EBR decodificado con éxito desde la posición %d.\n", position)
	return nil
}

// SetEBR establece los valores del EBR
func (e *EBR) SetEBR(fit byte, size int32, start int32, next int32, name string) {
	fmt.Println("Estableciendo valores del EBR:")
	fmt.Printf("Fit: %c | Size: %d | Start: %d | Next: %d | Name: %s\n", fit, size, start, next, name)

	e.Ebr_mount[0] = '1' // Created
	e.Ebr_fit[0] = fit
	e.Ebr_start = start
	e.Ebr_size = size
	e.Ebr_next = next

	// Copiar el nombre al array Ebr_name y rellenar el resto con ceros
	copy(e.Ebr_name[:], name)
	for i := len(name); i < len(e.Ebr_name); i++ {
		e.Ebr_name[i] = 0 // Rellenar con ceros
	}
}

// CreateAndWriteEBR crea un nuevo EBR y lo escribe en el archivo de disco
func CreateAndWriteEBR(start int32, size int32, fit byte, name string, file *os.File) error {
	fmt.Printf("Creando y escribiendo EBR en la posición: %d\n", start)

	ebr := &EBR{}
	ebr.SetEBR(fit, size, start, -1, name) // Establecer los valores del EBR

	return ebr.Encode(file, int64(start))
}

// Print imprime los valores del EBR en una sola línea
func (e *EBR) Print() {
	fmt.Printf("Mount: %c | Fit: %c | Start: %d | Size: %d | Next: %d | Name: %s\n",
		e.Ebr_mount[0], e.Ebr_fit[0], e.Ebr_start, e.Ebr_size, e.Ebr_next, string(e.Ebr_name[:]))
}

// CalculateNextEBRStart calcula la posición de inicio del próximo EBR
func (e *EBR) CalculateNextEBRStart(extendedPartitionStart int32, extendedPartitionSize int32) (int32, error) {
	fmt.Printf("Calculando el inicio del siguiente EBR...\nEBR Actual - Start: %d, Size: %d, Next: %d\n",
		e.Ebr_start, e.Ebr_size, e.Ebr_next)

	if e.Ebr_size <= 0 {
		return -1, fmt.Errorf("EBR size is invalid or zero")
	}

	if e.Ebr_start < extendedPartitionStart {
		return -1, fmt.Errorf("EBR start position is invalid")
	}

	nextStart := e.Ebr_start + e.Ebr_size

	// Asegurarse de que nextStart esté dentro del rango de la partición extendida
	if nextStart <= e.Ebr_start || nextStart >= extendedPartitionStart+extendedPartitionSize {
		return -1, fmt.Errorf("error: el siguiente EBR está fuera de los límites de la partición extendida")
	}

	fmt.Printf("Inicio del siguiente EBR calculado con éxito: %d\n", nextStart)
	return nextStart, nil
}

func FindLastEBR(start int32, file *os.File) (*EBR, error) {
	fmt.Printf("Buscando el último EBR a partir de la posición: %d\n", start)

	// Crear la primera instancia de EBR
	currentEBR := &EBR{}

	// Decodificar el EBR en la posición inicial
	err := currentEBR.Decode(file, int64(start))
	if err != nil {
		return nil, err
	}

	// Recorrer la cadena de EBRs hasta encontrar el último
	for currentEBR.Ebr_next != -1 {
		if currentEBR.Ebr_next < 0 {
			// Evitar leer una posición negativa
			return currentEBR, nil
		}
		fmt.Printf("EBR encontrado - Start: %d, Next: %d\n", currentEBR.Ebr_start, currentEBR.Ebr_next)

		// Crear una nueva instancia de EBR para el siguiente
		nextEBR := &EBR{}
		err = nextEBR.Decode(file, int64(currentEBR.Ebr_next))
		if err != nil {
			return nil, err
		}
		currentEBR = nextEBR
	}

	fmt.Printf("Último EBR encontrado en la posición: %d\n", currentEBR.Ebr_start)
	return currentEBR, nil
}

// Overwrite sobrescribe el espacio de la partición lógica (EBR) con ceros
func (e *EBR) Overwrite(file *os.File) error {
	// Verificar si el EBR tiene un tamaño válido
	if e.Ebr_size <= 0 {
		return fmt.Errorf("el tamaño del EBR es inválido o cero")
	}

	// Posicionarse en el inicio del EBR (donde comienza la partición lógica)
	_, err := file.Seek(int64(e.Ebr_start), 0)
	if err != nil {
		return fmt.Errorf("error al mover el puntero del archivo a la posición del EBR: %v", err)
	}

	// Crear un buffer de ceros del tamaño de la partición lógica
	zeroes := make([]byte, e.Ebr_size)

	// Escribir los ceros en el archivo
	_, err = file.Write(zeroes)
	if err != nil {
		return fmt.Errorf("error al sobrescribir el espacio del EBR: %v", err)
	}

	fmt.Printf("Espacio de la partición lógica (EBR) en posición %d sobrescrito con ceros.\n", e.Ebr_start)
	return nil
}

// SetNextEBR establece el apuntador al siguiente EBR en la lista enlazada de EBRs
func (e *EBR) SetNextEBR(newNext int32) {
	fmt.Printf("Estableciendo el siguiente EBR: Actual Start: %d, Nuevo Next: %d\n", e.Ebr_start, newNext)
	e.Ebr_next = newNext
}
