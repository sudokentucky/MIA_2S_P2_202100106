package commands

import (
	structs "backend/Structs"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strings"
)

type ListPartitions struct {
	path string // Ruta del archivo del disco
}

func ParserListPartitions(tokens []string) (string, error) {
	cmd := &ListPartitions{}
	var outputBuffer bytes.Buffer

	// Validar que el token contiene el path y extraerlo
	for _, token := range tokens {
		if strings.HasPrefix(strings.ToLower(token), "-path=") {
			cmd.path = strings.Trim(strings.SplitN(token, "=", 2)[1], "\"")
		}
	}

	// Verificar que el path no esté vacío
	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}

	// Ejecutar el comando para listar las particiones
	err := commandListPartitions(cmd, &outputBuffer)
	if err != nil {
		return "", fmt.Errorf("error al listar las particiones: %v", err)
	}

	return outputBuffer.String(), nil
}

func commandListPartitions(listCmd *ListPartitions, outputBuffer *bytes.Buffer) error {
	// Abrir el archivo del disco
	file, err := os.Open(listCmd.path)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo de disco: %v", err)
	}
	defer file.Close()

	// Leer el MBR del disco
	mbr := &structs.MBR{}
	err = mbr.Decode(file)
	if err != nil {
		return fmt.Errorf("error al leer el MBR del disco: %v", err)
	}

	fmt.Fprintln(outputBuffer, "===================== LISTA DE PARTICIONES =====================")
	fmt.Fprintf(outputBuffer, "	Disco: %s 	(Tamaño: %d 	bytes)\n", listCmd.path, mbr.MbrSize)
	fmt.Fprintln(outputBuffer, "-----------------------------------------------------------------")
	fmt.Fprintln(outputBuffer, "Tipo     	Nombre      	Inicio       	Tamaño       	Estado")

	// Recorrer las particiones y mostrarlas
	for _, part := range mbr.MbrPartitions {
		if part.Part_size > 0 { // Considerar solo particiones con tamaño válido
			partName := strings.TrimRight(string(part.Part_name[:]), "\x00")
			partType := "Desconocido"
			if part.Part_type[0] == 'P' {
				partType = "Primaria"
			} else if part.Part_type[0] == 'E' {
				partType = "Extendida"
			}

			partStatus := "Libre"
			if part.Part_status[0] != '9' {
				partStatus = "Ocupado"
			}

			fmt.Fprintf(outputBuffer, "%-8s %-10s %-12d %-12d %s\n", partType, partName, part.Part_start, part.Part_size, partStatus)

			// Si la partición es extendida, listar las particiones lógicas
			if part.Part_type[0] == 'E' {
				listLogicalPartitions(file, part.Part_start, outputBuffer)
			}
		}
	}

	fmt.Fprintln(outputBuffer, "=================================================================")
	return nil
}

func listLogicalPartitions(file *os.File, start int32, outputBuffer *bytes.Buffer) {
	ebrStart := start

	fmt.Fprintln(outputBuffer, "  Particiones lógicas dentro de la extendida:")
	for ebrStart != -1 {
		ebr := &structs.EBR{}
		err := ebr.Decode(file, int64(ebrStart))
		if err != nil {
			fmt.Fprintf(outputBuffer, "  Error al leer EBR en la posición %d: %v\n", ebrStart, err)
			return
		}

		// Extraer nombre y ajustar el formato
		ebrName := strings.TrimRight(string(ebr.Ebr_name[:]), "\x00")
		ebrFit := string(ebr.Ebr_fit[:])
		ebrMount := "No Montada"
		if ebr.Ebr_mount[0] == '1' {
			ebrMount = "Montada"
		}

		// Mostrar los datos del EBR
		fmt.Fprintf(outputBuffer, "  Lógica  %-10s %-12d %-12d %-6s %-10s Next: %d\n",
			ebrName, ebr.Ebr_start, ebr.Ebr_size, ebrFit, ebrMount, ebr.Ebr_next)

		// Mover al siguiente EBR
		ebrStart = ebr.Ebr_next
	}
}
