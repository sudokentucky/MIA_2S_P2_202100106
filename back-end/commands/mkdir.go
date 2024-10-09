package commands

import (
	structures "backend/Structs"
	global "backend/globals"
	utils "backend/utils"
	"bytes"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// MKDIR estructura que representa el comando mkdir con sus parámetros
type MKDIR struct {
	path string // Path del directorio
	p    bool   // Opción -p (crea directorios padres si no existen)
}

func ParserMkdir(tokens []string) (string, error) {
	cmd := &MKDIR{}               // Crea una nueva instancia de MKDIR
	var outputBuffer bytes.Buffer // Buffer para capturar mensajes importantes

	// Unir tokens en una sola cadena y luego dividir por espacios, respetando las comillas
	args := strings.Join(tokens, " ")
	// Expresión regular para encontrar los parámetros del comando mkdir
	re := regexp.MustCompile(`-path=[^\s]+|-p`)
	// Encuentra todas las coincidencias de la expresión regular en la cadena de argumentos
	matches := re.FindAllString(args, -1)

	// Verificar que todos los tokens fueron reconocidos por la expresión regular
	if len(matches) != len(tokens) {
		// Identificar el parámetro inválido
		for _, token := range tokens {
			if !re.MatchString(token) {
				return "", fmt.Errorf("parámetro inválido: %s", token)
			}
		}
	}

	// Itera sobre cada coincidencia encontrada
	for _, match := range matches {
		// Divide cada parte en clave y valor usando "=" como delimitador
		kv := strings.SplitN(match, "=", 2)
		key := strings.ToLower(kv[0])

		// Switch para manejar diferentes parámetros
		switch key {
		case "-path":
			if len(kv) != 2 {
				return "", fmt.Errorf("formato de parámetro inválido: %s", match)
			}
			value := kv[1]
			// Remove quotes from value if present
			if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
				value = strings.Trim(value, "\"")
			}
			cmd.path = value
		case "-p":
			cmd.p = true
		default:
			// Si el parámetro no es reconocido, devuelve un error
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	// Verifica que el parámetro -path haya sido proporcionado
	if cmd.path == "" {
		return "", errors.New("faltan parámetros requeridos: -path")
	}

	// Ejecutar el comando mkdir con captura de mensajes en el buffer
	err := commandMkdir(cmd, &outputBuffer)
	if err != nil {
		return "", err
	}

	// Retorna el contenido del buffer al usuario
	return outputBuffer.String(), nil
}

func commandMkdir(mkdir *MKDIR, outputBuffer *bytes.Buffer) error {
	// Verificar si hay un usuario logueado
	if !global.IsLoggedIn() {
		return fmt.Errorf("no hay un usuario logueado")
	}

	// Obtener el ID de la partición desde el usuario logueado
	idPartition := global.UsuarioActual.Id

	// Obtener la partición montada asociada al usuario logueado
	partitionSuperblock, mountedPartition, partitionPath, err := global.GetMountedPartitionSuperblock(idPartition)
	if err != nil {
		return fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	// Abrir el archivo de partición para operar sobre él
	file, err := os.OpenFile(partitionPath, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo de partición: %w", err)
	}
	defer file.Close() // Cerrar el archivo cuando ya no sea necesario

	// Capturar mensajes importantes en el buffer
	fmt.Fprintln(outputBuffer, "======================= MKDIR =======================")
	fmt.Fprintf(outputBuffer, "Creando directorio: %s\n", mkdir.path)

	// Crear el directorio usando el archivo abierto, pasando la opción -p
	err = createDirectory(mkdir.path, mkdir.p, partitionSuperblock, file, mountedPartition)
	if err != nil {
		return fmt.Errorf("error al crear el directorio: %w", err)
	}

	fmt.Fprintf(outputBuffer, "Directorio %s creado exitosamente\n", mkdir.path)
	fmt.Fprintln(outputBuffer, "=====================================================")

	return nil
}

func createDirectory(dirPath string, createParents bool, sb *structures.Superblock, file *os.File, mountedPartition *structures.Partition) error {
	// Si el parámetro -p está habilitado, crear los directorios intermedios recursivamente
	if createParents {
		// Utilizamos `CreateFolderRecursively` para crear los directorios si no existen
		err := sb.CreateFolderRecursively(file, dirPath)
		if err != nil {
			return fmt.Errorf("error al crear los directorios recursivamente: %w", err)
		}
	} else {
		// Si no se habilita el parámetro -p, asegurarse de que los directorios padres existan
		parentDirs, destDir := utils.GetParentDirectories(dirPath)

		// Verificar que todos los directorios padres existen
		err := verifyParentDirectoriesExist(sb, file, parentDirs)
		if err != nil {
			return err
		}

		// Crear el directorio final
		err = sb.CreateFolder(file, parentDirs, destDir)
		if err != nil {
			return fmt.Errorf("error al crear el directorio: %w", err)
		}
	}

	// Serializar el superbloque en el archivo de partición abierto
	err := sb.Encode(file, int64(mountedPartition.Part_start))
	if err != nil {
		return fmt.Errorf("error al serializar el superbloque: %w", err)
	}

	// Imprimir inodos y bloques en el buffer (para depuración)
	fmt.Println("\nInodos:") // Depuración
	sb.PrintInodes(file.Name())
	fmt.Println("\nBloques:") // Depuración
	sb.PrintBlocks(file.Name())

	return nil
}

func verifyParentDirectoriesExist(sb *structures.Superblock, file *os.File, parentDirs []string) error {
	// Comenzar desde el inodo raíz
	inodeIndex := int32(0) // Inodo raíz es el 0 en la mayoría de los casos

	// Iterar sobre los directorios padres para verificar que existan
	for _, dir := range parentDirs {
		exists, newInodeIndex, err := directoryExists(sb, file, inodeIndex, dir)
		if err != nil {
			return fmt.Errorf("error al verificar el directorio '%s': %w", dir, err)
		}
		if !exists {
			return fmt.Errorf("el directorio padre '%s' no existe", dir)
		}
		// Actualizar el inodo actual al inodo encontrado
		inodeIndex = newInodeIndex
	}

	return nil
}
