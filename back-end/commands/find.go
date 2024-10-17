package commands

import (
	structs "backend/Structs"
	global "backend/globals"
	utils "backend/utils"
	"bytes"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// FIND estructura que representa el comando FIND con sus parámetros
type FIND struct {
	path string // Ruta inicial para iniciar la búsqueda
	name string // Nombre del archivo o carpeta a buscar (puede contener caracteres especiales)
}

// ParserFind parsea el comando find y devuelve una instancia de FIND
func ParserFind(tokens []string) (string, error) {
	cmd := &FIND{}
	var outputBuffer bytes.Buffer

	// Expresión regular para capturar los parámetros -path y -name
	re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+|-name="[^"]+"|-name=[^\s]+`)
	matches := re.FindAllString(strings.Join(tokens, " "), -1)

	// Verificar que se han proporcionado ambos parámetros
	if len(matches) != len(tokens) || len(matches) < 2 {
		return "", errors.New("faltan parámetros requeridos: -path o -name")
	}

	// Iterar sobre cada coincidencia y extraer los valores de -path y -name
	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		key := strings.ToLower(kv[0])
		value := strings.Trim(kv[1], "\"") // Eliminar comillas si existen

		// Asignar los valores de los parámetros
		switch key {
		case "-path":
			cmd.path = value
		case "-name":
			cmd.name = value
		}
	}

	// Verificar que ambos parámetros tengan valores
	if cmd.path == "" || cmd.name == "" {
		return "", errors.New("los parámetros -path y -name son obligatorios")
	}

	// Ejecutar el comando FIND
	err := commandFind(cmd, &outputBuffer)
	if err != nil {
		return "", err
	}

	return outputBuffer.String(), nil
}

func commandFind(findCmd *FIND, outputBuffer *bytes.Buffer) error {
	fmt.Fprint(outputBuffer, "======================= FIND =======================\n")

	// Verificar si hay un usuario logueado
	if !global.IsLoggedIn() {
		return fmt.Errorf("no hay un usuario logueado")
	}

	// Obtener el ID de la partición desde el usuario logueado
	idPartition := global.UsuarioActual.Id

	// Obtener la partición montada asociada al usuario logueado
	partitionSuperblock, _, partitionPath, err := global.GetMountedPartitionSuperblock(idPartition)
	if err != nil {
		return fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	// Abrir el archivo de partición para operar sobre él
	file, err := os.OpenFile(partitionPath, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo de partición: %w", err)
	}
	defer file.Close() // Cerrar el archivo cuando ya no sea necesario

	// Verificar si el path es la raíz "/"
	var rootInodeIndex int32
	if findCmd.path == "/" {
		rootInodeIndex = 0 // El inodo raíz es el inodo 0
	} else {
		// Desglosar el path en directorios y obtener el inodo del directorio inicial
		parentDirs, dirName := utils.GetParentDirectories(findCmd.path)
		rootInodeIndex, err = findFileInode(file, partitionSuperblock, parentDirs, dirName)
		if err != nil {
			return fmt.Errorf("error al encontrar el directorio inicial: %v", err)
		}
	}

	// Convertir el nombre con comodines a expresión regular
	pattern, err := wildcardToRegex(findCmd.name)
	if err != nil {
		return fmt.Errorf("error al convertir el patrón de búsqueda: %v", err)
	}

	// Iniciar la búsqueda recursiva
	err = searchRecursive(file, partitionSuperblock, rootInodeIndex, pattern, findCmd.path, outputBuffer)
	if err != nil {
		return fmt.Errorf("error durante la búsqueda: %v", err)
	}

	fmt.Fprint(outputBuffer, "=================================================\n")
	return nil
}

func searchRecursive(file *os.File, sb *structs.Superblock, inodeIndex int32, pattern *regexp.Regexp, currentPath string, outputBuffer *bytes.Buffer) error {
	// Deserializar el inodo del directorio actual
	inode := &structs.Inode{}
	err := inode.Decode(file, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil {
		return fmt.Errorf("error al deserializar el inodo %d: %v", inodeIndex, err)
	}

	// Verificar que el inodo sea un directorio
	if inode.I_type[0] != '0' {
		return nil // Si no es un directorio, no hacemos nada
	}

	// Iterar sobre los bloques del inodo del directorio
	for _, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			break // Si no hay más bloques asignados, terminamos la búsqueda
		}

		// Deserializar el bloque de directorio
		block := &structs.FolderBlock{}
		err := block.Decode(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
		if err != nil {
			return fmt.Errorf("error al deserializar el bloque %d: %v", blockIndex, err)
		}

		// Iterar sobre los contenidos del bloque
		for _, content := range block.B_content {
			if content.B_inodo == -1 {
				continue // Si no hay un inodo válido, lo saltamos
			}

			contentName := strings.Trim(string(content.B_name[:]), "\x00 ")

			// Evitar los enlaces "." y ".." que pueden causar loops infinitos
			if contentName == "." || contentName == ".." {
				continue
			}

			// Verificar si el nombre coincide con el patrón
			if pattern.MatchString(contentName) {
				fmt.Fprintf(outputBuffer, "%s/%s\n", currentPath, contentName)
			}

			// Si el contenido es un directorio, hacer una búsqueda recursiva
			newInodeIndex := content.B_inodo
			err = searchRecursive(file, sb, newInodeIndex, pattern, currentPath+"/"+contentName, outputBuffer)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func wildcardToRegex(pattern string) (*regexp.Regexp, error) {
	// Escapar caracteres especiales de expresiones regulares
	pattern = strings.ReplaceAll(pattern, ".", "\\.")
	pattern = strings.ReplaceAll(pattern, "?", ".")  // ? se convierte en un solo carácter
	pattern = strings.ReplaceAll(pattern, "*", ".*") // * se convierte en uno o más caracteres

	// Compilar la expresión regular
	return regexp.Compile("^" + pattern + "$")
}
