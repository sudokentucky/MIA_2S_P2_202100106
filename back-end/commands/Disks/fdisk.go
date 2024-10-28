package commands

import (
	structures "backend/Structs"
	utils "backend/utils"
	"bytes"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// Fdisk estructura que representa el comando fdisk con sus parámetros
// Fdisk estructura que representa el comando fdisk con sus parámetros
type Fdisk struct {
	size   int    // Tamaño de la partición (solo para crear particiones)
	unit   string // Unidad de medida del tamaño (K o M)
	fit    string // Tipo de ajuste (BF, FF, WF)
	path   string // Ruta del archivo del disco
	typ    string // Tipo de partición (P, E, L)
	name   string // Nombre de la partición
	add    int    // Espacio a agregar o quitar (solo para agregar o quitar espacio)
	delete string // Método de eliminación (fast o full)
}

// ParserFdisk parsea el comando fdisk y devuelve los mensajes generados
func ParserFdisk(tokens []string) (string, error) {
	var outputBuffer bytes.Buffer
	cmd := &Fdisk{}

	args := strings.Join(tokens, " ")
	re := regexp.MustCompile(`-size=\d+|-unit=[bBkKmM]|-fit=[bBfFwfW]{2}|-path="[^"]+"|-path=[^\s]+|-type=[pPeElL]|-name="[^"]+"|-name=[^\s]+|-add=[+-]?\d+|-delete=(fast|full)`)
	matches := re.FindAllString(args, -1)

	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return "", fmt.Errorf("formato de parámetro inválido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {
		case "-size":
			size, err := strconv.Atoi(value)
			if err != nil || size <= 0 {
				return "", errors.New("el tamaño debe ser un número entero positivo")
			}
			cmd.size = size
		case "-unit":
			value = strings.ToUpper(value)
			if value != "B" && value != "K" && value != "M" {
				return "", errors.New("la unidad debe ser B, K o M")
			}
			cmd.unit = value
		case "-fit":
			value = strings.ToUpper(value)
			if value != "BF" && value != "FF" && value != "WF" {
				return "", errors.New("el ajuste debe ser BF, FF o WF")
			}
			cmd.fit = value
		case "-path":
			if value == "" {
				return "", errors.New("el path no puede estar vacío")
			}
			cmd.path = value
		case "-type":
			value = strings.ToUpper(value)
			if value != "P" && value != "E" && value != "L" {
				return "", errors.New("el tipo debe ser P, E o L")
			}
			cmd.typ = value
		case "-name":
			if value == "" {
				return "", errors.New("el nombre no puede estar vacío")
			}
			cmd.name = value
		case "-add":
			add, err := strconv.Atoi(value)
			if err != nil {
				return "", errors.New("el valor de -add debe ser un número entero")
			}
			cmd.add = add
		case "-delete":
			value = strings.ToLower(value)
			if value != "fast" && value != "full" {
				return "", errors.New("el valor de -delete debe ser 'fast' o 'full'")
			}
			cmd.delete = value
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	// Identificar el tipo de operación: add, delete o crear partición
	if cmd.delete != "" {
		// Operación de eliminación de partición
		if cmd.path == "" {
			return "", errors.New("falta el parámetro requerido: -path")
		}
		if cmd.name == "" {
			return "", errors.New("falta el parámetro requerido: -name")
		}
		return processDeletePartition(cmd, &outputBuffer)
	}

	if cmd.add != 0 {
		// Operación de agregar/quitar espacio
		if cmd.path == "" {
			return "", errors.New("falta el parámetro requerido: -path")
		}
		if cmd.name == "" {
			return "", errors.New("falta el parámetro requerido: -name")
		}
		return processAddPartition(cmd, &outputBuffer)
	}

	// Operación de crear partición (requiere -size, -path, -name)
	if cmd.size == 0 {
		return "", errors.New("faltan parámetros requeridos: -size")
	}
	if cmd.path == "" {
		return "", errors.New("falta el parámetro requerido: -path")
	}
	if cmd.name == "" {
		return "", errors.New("falta el parámetro requerido: -name")
	}

	if cmd.unit == "" {
		cmd.unit = "K" // Valor por defecto
	}

	if cmd.fit == "" {
		cmd.fit = "WF" // Valor por defecto
	}

	if cmd.typ == "" {
		cmd.typ = "P" // Valor por defecto
	}

	// Ejecutar el comando fdisk para crear la partición
	err := commandFdisk(cmd, &outputBuffer)
	if err != nil {
		return "", fmt.Errorf("error al crear la partición: %v", err)
	}

	return outputBuffer.String(), nil
}

// processDeletePartition maneja la eliminación de particiones
func processDeletePartition(cmd *Fdisk, outputBuffer *bytes.Buffer) (string, error) {
	fmt.Fprintf(outputBuffer, "========================== DELETE ==========================\n")
	fmt.Fprintf(outputBuffer, "Eliminando partición con nombre '%s' usando el método %s...\n", cmd.name, cmd.delete)

	// Abrir el archivo del disco
	file, err := os.OpenFile(cmd.path, os.O_RDWR, 0644)
	if err != nil {
		return "", fmt.Errorf("error abriendo el archivo del disco: %v", err)
	}
	defer file.Close()

	// Leer el MBR del archivo
	var mbr structures.MBR
	err = mbr.Decode(file)
	if err != nil {
		return "", fmt.Errorf("error al deserializar el MBR: %v", err)
	}

	// Buscar la partición por nombre y eliminarla
	partition, _ := mbr.GetPartitionByName(cmd.name)
	if partition == nil {
		return "", fmt.Errorf("la partición '%s' no existe", cmd.name)
	}

	// Verificar si es extendida para eliminar particiones lógicas
	isExtended := partition.Part_type[0] == 'E'
	err = partition.Delete(cmd.delete, file, isExtended)
	if err != nil {
		return "", fmt.Errorf("error al eliminar la partición: %v", err)
	}

	// Actualizar el MBR en el archivo después de la eliminación
	err = mbr.Encode(file)
	if err != nil {
		return "", fmt.Errorf("error al actualizar el MBR en el disco: %v", err)
	}

	// Mensaje de éxito
	fmt.Fprintf(outputBuffer, "Partición '%s' eliminada exitosamente.\n", cmd.name)
	fmt.Fprintf(outputBuffer, "===========================================================\n")

	// Imprimir las particiones restantes
	fmt.Fprintf(outputBuffer, "========================== PARTICIONES ==========================\n")
	printPartitions(&mbr, outputBuffer)
	fmt.Fprintf(outputBuffer, "===========================================================\n")

	return outputBuffer.String(), nil
}

// processAddPartition maneja el agregar o quitar espacio a particiones
func processAddPartition(cmd *Fdisk, outputBuffer *bytes.Buffer) (string, error) {
	fmt.Fprintf(outputBuffer, "========================== ADD ==========================\n")
	fmt.Fprintf(outputBuffer, "Modificando partición '%s', ajustando %d unidades...\n", cmd.name, cmd.add)

	// Abrir el archivo del disco
	file, err := os.OpenFile(cmd.path, os.O_RDWR, 0644)
	if err != nil {
		return "", fmt.Errorf("error abriendo el archivo del disco: %v", err)
	}
	defer file.Close()

	// Leer el MBR del archivo
	var mbr structures.MBR
	err = mbr.Decode(file)
	if err != nil {
		return "", fmt.Errorf("error al deserializar el MBR: %v", err)
	}

	// Buscar la partición por nombre
	partition, _ := mbr.GetPartitionByName(cmd.name)
	if partition == nil {
		return "", fmt.Errorf("la partición '%s' no existe", cmd.name)
	}

	// Convertir cmd.add a bytes según la unidad especificada
	addBytes, err := utils.ConvertToBytes(cmd.add, cmd.unit)
	if err != nil {
		return "", fmt.Errorf("error al convertir las unidades de -add: %v", err)
	}

	// Calcular espacio disponible si se está agregando espacio
	var availableSpace int32 = 0
	if addBytes > 0 {
		availableSpace, err = mbr.CalculateAvailableSpaceForPartition(partition)
		if err != nil {
			return "", fmt.Errorf("error al calcular el espacio disponible para la partición '%s': %v", cmd.name, err)
		}
	}

	// Modificar el tamaño de la partición
	err = partition.ModifySize(int32(addBytes), availableSpace)
	if err != nil {
		return "", fmt.Errorf("error al modificar el tamaño de la partición: %v", err)
	}

	// Actualizar el MBR en el archivo después de la modificación
	err = mbr.Encode(file)
	if err != nil {
		return "", fmt.Errorf("error al actualizar el MBR en el disco: %v", err)
	}

	// Mensaje de éxito
	fmt.Fprintf(outputBuffer, "Espacio en la partición '%s' modificado exitosamente.\n", cmd.name)
	fmt.Fprintf(outputBuffer, "===========================================================\n")

	// Imprimir las particiones después de modificar el espacio
	fmt.Fprintf(outputBuffer, "========================== PARTICIONES ==========================\n")
	printPartitions(&mbr, outputBuffer)
	fmt.Fprintf(outputBuffer, "===========================================================\n")

	return outputBuffer.String(), nil
}

// printPartitions imprime las particiones actuales del MBR
func printPartitions(mbr *structures.MBR, outputBuffer *bytes.Buffer) {
	for i, partition := range mbr.MbrPartitions {
		if partition.Part_start != -1 {
			fmt.Fprintf(outputBuffer, "Partición %d: Nombre: %s | Inicio: %d | Tamaño: %d bytes | Tipo: %c | Estado: %c\n",
				i+1,
				strings.TrimSpace(string(partition.Part_name[:])),
				partition.Part_start,
				partition.Part_size,
				partition.Part_type[0],
				partition.Part_status[0],
			)
		} else {
			fmt.Fprintf(outputBuffer, "Partición %d: (Vacía)\n", i+1)
		}
	}
}

// commandFdisk ejecuta el comando fdisk con los parámetros especificados
func commandFdisk(fdisk *Fdisk, outputBuffer *bytes.Buffer) error {
	fmt.Fprintf(outputBuffer, "========================== FDISK ==========================\n")
	fmt.Fprintf(outputBuffer, "Creando partición con nombre '%s' y tamaño %d %s...\n", fdisk.name, fdisk.size, fdisk.unit)
	fmt.Println("Detalles internos de la creación de partición:", fdisk.size, fdisk.unit, fdisk.fit, fdisk.path, fdisk.typ, fdisk.name)

	// Abrir el archivo del disco
	file, err := os.OpenFile(fdisk.path, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error abriendo el archivo del disco: %v", err)
	}
	defer file.Close()

	sizeBytes, err := utils.ConvertToBytes(fdisk.size, fdisk.unit)
	if err != nil {
		fmt.Println("Error converting size:", err) // Mensaje de depuración
		return err
	}

	if fdisk.typ == "P" {
		err = createPrimaryPartition(file, fdisk, sizeBytes, outputBuffer)
		if err != nil {
			fmt.Println("Error creando partición primaria:", err) // Mensaje de depuración
			return err
		}
	} else if fdisk.typ == "E" {
		fmt.Println("Creando partición extendida...") // Mensaje de depuración
		err = createExtendedPartition(file, fdisk, sizeBytes, outputBuffer)
		if err != nil {
			fmt.Println("Error creando partición extendida:", err) // Mensaje de depuración
			return err
		}
	} else if fdisk.typ == "L" {
		fmt.Println("Creando partición lógica...") // Mensaje de depuración
		err = createLogicalPartition(file, fdisk, sizeBytes, outputBuffer)
		if err != nil {
			fmt.Println("Error creando partición lógica:", err) // Mensaje de depuración
			return err
		}
	}

	fmt.Fprintln(outputBuffer, "Partición creada exitosamente.") // Mensaje importante para el usuario
	fmt.Fprintln(outputBuffer, "===========================================================")
	return nil
}

// Crear una partición primaria en el MBR
func createPrimaryPartition(file *os.File, fdisk *Fdisk, sizeBytes int, outputBuffer *bytes.Buffer) error {
	fmt.Fprintf(outputBuffer, "Creando partición primaria con tamaño %d %s...\n", fdisk.size, fdisk.unit)

	var mbr structures.MBR
	err := mbr.Decode(file)
	if err != nil {
		return fmt.Errorf("error al deserializar el MBR: %v", err)
	}
	availableSpace, err := mbr.CalculateAvailableSpace()
	if err != nil {
		fmt.Println("Error calculando el espacio disponible:", err)
	} else {
		fmt.Println("Espacio disponible en el disco:", availableSpace)
	}
	// Llamar al método del MBR para crear la partición con el ajuste correspondiente
	err = mbr.CreatePartitionWithFit(int32(sizeBytes), fdisk.typ, fdisk.name)
	if err != nil {
		return fmt.Errorf("error al crear la partición primaria: %v", err)
	}

	// Actualizar el MBR en el archivo del disco
	err = mbr.Encode(file)
	if err != nil {
		return fmt.Errorf("error al actualizar el MBR en el disco: %v", err)
	}

	fmt.Fprintln(outputBuffer, "Partición primaria creada exitosamente.")
	return nil
}

// Crear una partición extendida
func createExtendedPartition(file *os.File, fdisk *Fdisk, sizeBytes int, outputBuffer *bytes.Buffer) error {
	fmt.Fprintf(outputBuffer, "Creando partición extendida con tamaño %d %s...\n", fdisk.size, fdisk.unit)

	var mbr structures.MBR
	err := mbr.Decode(file)
	if err != nil {
		return fmt.Errorf("error al deserializar el MBR: %v", err)
	}

	// Verificar si ya existe una partición extendida
	if mbr.HasExtendedPartition() {
		return errors.New("ya existe una partición extendida en este disco")
	}

	// Usar el método del MBR para crear la partición con el ajuste correspondiente
	err = mbr.CreatePartitionWithFit(int32(sizeBytes), "E", fdisk.name)
	if err != nil {
		return fmt.Errorf("error al crear la partición extendida: %v", err)
	}

	// Crear el primer EBR dentro de la partición extendida
	extendedPartition, _ := mbr.GetPartitionByName(fdisk.name)
	err = structures.CreateAndWriteEBR(extendedPartition.Part_start, 0, fdisk.fit[0], fdisk.name, file)
	if err != nil {
		return fmt.Errorf("error al crear el primer EBR en la partición extendida: %v", err)
	}

	// Actualizar el MBR
	err = mbr.Encode(file)
	if err != nil {
		return fmt.Errorf("error al actualizar el MBR en el disco: %v", err)
	}

	fmt.Fprintln(outputBuffer, "Partición extendida creada exitosamente.")
	return nil
}

// Crear una partición lógica
func createLogicalPartition(file *os.File, fdisk *Fdisk, sizeBytes int, outputBuffer *bytes.Buffer) error {
	fmt.Fprintf(outputBuffer, "Creando partición lógica con tamaño %d %s...\n", fdisk.size, fdisk.unit)

	var mbr structures.MBR
	err := mbr.Decode(file)
	if err != nil {
		return fmt.Errorf("error al deserializar el MBR: %v", err)
	}

	// Verificar si existe una partición extendida utilizando HasExtendedPartition
	if !mbr.HasExtendedPartition() {
		return errors.New("no se encontró una partición extendida en el disco")
	}

	// Identificar la partición extendida específica
	var extendedPartition *structures.Partition
	for i := range mbr.MbrPartitions {
		if mbr.MbrPartitions[i].Part_type[0] == 'E' {
			extendedPartition = &mbr.MbrPartitions[i]
			break
		}
	}

	// Buscar el último EBR en la partición extendida
	lastEBR, err := structures.FindLastEBR(extendedPartition.Part_start, file)
	if err != nil {
		return fmt.Errorf("error al buscar el último EBR: %v", err)
	}

	// Verificar si es el primer EBR
	if lastEBR.Ebr_size == 0 {
		fmt.Println("Detectado EBR inicial vacío, asignando tamaño a la nueva partición lógica.")
		lastEBR.Ebr_size = int32(sizeBytes)
		copy(lastEBR.Ebr_name[:], fdisk.name)

		err = lastEBR.Encode(file, int64(lastEBR.Ebr_start))
		if err != nil {
			return fmt.Errorf("error al escribir el primer EBR con la nueva partición lógica: %v", err)
		}

		fmt.Fprintln(outputBuffer, "Primera partición lógica creada exitosamente.")
		return nil
	}

	// Calcular el inicio del nuevo EBR
	newEBRStart, err := lastEBR.CalculateNextEBRStart(extendedPartition.Part_start, extendedPartition.Part_size)
	if err != nil {
		return fmt.Errorf("error calculando el inicio del nuevo EBR: %v", err)
	}

	availableSize := extendedPartition.Part_size - (newEBRStart - extendedPartition.Part_start)
	if availableSize < int32(sizeBytes) {
		return errors.New("no hay suficiente espacio en la partición extendida para una nueva partición lógica")
	}

	// Crear el nuevo EBR
	newEBR := structures.EBR{}
	newEBR.SetEBR(fdisk.fit[0], int32(sizeBytes), newEBRStart, -1, fdisk.name)

	// Escribir el nuevo EBR en el disco
	err = newEBR.Encode(file, int64(newEBRStart))
	if err != nil {
		return fmt.Errorf("error al escribir el nuevo EBR en el disco: %v", err)
	}

	// Actualizar el último EBR para que apunte al nuevo
	lastEBR.SetNextEBR(newEBRStart)
	err = lastEBR.Encode(file, int64(lastEBR.Ebr_start))
	if err != nil {
		return fmt.Errorf("error al actualizar el EBR anterior: %v", err)
	}

	fmt.Fprintln(outputBuffer, "Partición lógica creada exitosamente.")
	return nil
}
