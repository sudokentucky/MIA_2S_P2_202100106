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

// EDIT estructura que representa el comando EDIT con sus parámetros
type EDIT struct {
	path      string // Ruta del archivo a editar
	contenido string // Ruta al archivo de contenido externo
}

// ParserEdit parsea el comando edit y devuelve una instancia de EDIT
func ParserEdit(tokens []string) (string, error) {
	cmd := &EDIT{}                // Crea una nueva instancia de EDIT
	var outputBuffer bytes.Buffer // Buffer para capturar mensajes importantes

	// Expresión regular para capturar los parámetros -path y -contenido
	re := regexp.MustCompile(`-path="[^"]+"|-path=[^\s]+|-contenido="[^"]+"|-contenido=[^\s]+`)
	matches := re.FindAllString(strings.Join(tokens, " "), -1)

	// Verificar que se han proporcionado ambos parámetros
	if len(matches) != len(tokens) || len(matches) < 2 {
		return "", errors.New("faltan parámetros requeridos: -path o -contenido")
	}

	// Iterar sobre cada coincidencia y extraer los valores de -path y -contenido
	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		key := strings.ToLower(kv[0])
		value := strings.Trim(kv[1], "\"") // Eliminar comillas si existen

		// Asignar los valores de los parámetros
		switch key {
		case "-path":
			cmd.path = value
		case "-contenido":
			cmd.contenido = value
		}
	}

	// Verificar que ambos parámetros tengan valores
	if cmd.path == "" || cmd.contenido == "" {
		return "", errors.New("los parámetros -path y -contenido son obligatorios")
	}

	// Ejecutar el comando EDIT
	err := commandEdit(cmd, &outputBuffer)
	if err != nil {
		return "", err
	}

	return outputBuffer.String(), nil
}

func commandEdit(editCmd *EDIT, outputBuffer *bytes.Buffer) error {
	fmt.Fprint(outputBuffer, "======================= EDIT =======================\n")

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

	// Desglosar el path en directorios y el archivo a editar
	parentDirs, fileName := utils.GetParentDirectories(editCmd.path)

	// Buscar el inodo del archivo a editar
	inodeIndex, err := findFileInode(file, partitionSuperblock, parentDirs, fileName)
	if err != nil {
		return fmt.Errorf("error al encontrar el archivo: %v", err)
	}

	// Buscar el archivo que contiene el nuevo contenido dentro del sistema de archivos recreado
	contentParentDirs, contentFileName := utils.GetParentDirectories(editCmd.contenido)
	contentInodeIndex, err := findFileInode(file, partitionSuperblock, contentParentDirs, contentFileName)
	if err != nil {
		return fmt.Errorf("error al encontrar el archivo de contenido: %v", err)
	}

	// Leer el contenido del archivo desde el sistema de archivos recreado
	newContent, err := readFileFromInode(file, partitionSuperblock, contentInodeIndex)
	if err != nil {
		return fmt.Errorf("error al leer el contenido del archivo: %v", err)
	}

	// Editar el contenido del archivo en el sistema de archivos
	err = editFileContent(file, partitionSuperblock, inodeIndex, []byte(newContent))
	if err != nil {
		return fmt.Errorf("error al editar el contenido del archivo: %v", err)
	}

	fmt.Fprintf(outputBuffer, "Contenido del archivo '%s' editado exitosamente\n", fileName)
	fmt.Fprint(outputBuffer, "=================================================\n")

	return nil
}

func editFileContent(file *os.File, sb *structs.Superblock, inodeIndex int32, newContent []byte) error {
	inode := &structs.Inode{}
	err := inode.Decode(file, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil {
		return fmt.Errorf("error al deserializar el inodo %d: %v", inodeIndex, err)
	}

	if inode.I_type[0] != '1' { // Asegurarse de que es un archivo
		return fmt.Errorf("el inodo %d no corresponde a un archivo", inodeIndex)
	}

	// Limpiar los bloques actuales (si deseas sobrescribir completamente el archivo)
	for _, blockIndex := range inode.I_block {
		if blockIndex != -1 {
			fileBlock := &structs.FileBlock{}
			fileBlock.ClearContent() // Limpia el contenido actual del bloque
			err := fileBlock.Encode(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
			if err != nil {
				return fmt.Errorf("error al limpiar el bloque %d: %v", blockIndex, err)
			}
		}
	}

	// Dividir el nuevo contenido en bloques de 64 bytes
	blocks, err := structs.SplitContent(string(newContent))
	if err != nil {
		return fmt.Errorf("error al dividir el contenido en bloques: %v", err)
	}

	// Escribir los nuevos bloques
	blockCount := len(blocks)
	for i := 0; i < blockCount; i++ {
		if i < len(inode.I_block) {
			// Si ya existe un bloque asignado, solo escribe en ese bloque
			blockIndex := inode.I_block[i]
			if blockIndex == -1 {
				// Si no hay bloque asignado, asignar un nuevo bloque
				blockIndex, err = sb.AssignNewBlock(file, inode, i)
				if err != nil {
					return fmt.Errorf("error asignando un nuevo bloque: %v", err)
				}
			}

			// Escribir el contenido en el bloque actual
			err := blocks[i].Encode(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
			if err != nil {
				return fmt.Errorf("error al escribir el bloque %d: %v", blockIndex, err)
			}
		} else {
			// Si se excede el espacio de I_block, usar un PointerBlock para manejar más bloques
			pointerBlockIndex := inode.I_block[len(inode.I_block)-1]
			if pointerBlockIndex == -1 {
				// Si no hay un bloque de punteros asignado, asignar uno nuevo
				pointerBlockIndex, err = sb.AssignNewBlock(file, inode, len(inode.I_block)-1)
				if err != nil {
					return fmt.Errorf("error asignando un nuevo bloque de apuntadores: %v", err)
				}
			}

			// Cargar el bloque de apuntadores
			pointerBlock := &structs.PointerBlock{}
			err := pointerBlock.Decode(file, int64(sb.S_block_start+(pointerBlockIndex*sb.S_block_size)))
			if err != nil {
				return fmt.Errorf("error al decodificar el bloque de apuntadores: %v", err)
			}

			// Buscar un puntero libre en el bloque de apuntadores
			freeIndex, err := pointerBlock.FindFreePointer()
			if err != nil {
				return fmt.Errorf("no hay apuntadores libres en el bloque de apuntadores: %v", err)
			}

			// Asignar un nuevo bloque para el contenido adicional
			newBlockIndex, err := sb.AssignNewBlock(file, inode, freeIndex)
			if err != nil {
				return fmt.Errorf("error asignando un nuevo bloque: %v", err)
			}

			// Actualizar el puntero en el bloque de apuntadores
			err = pointerBlock.SetPointer(freeIndex, int64(newBlockIndex))
			if err != nil {
				return fmt.Errorf("error actualizando el bloque de apuntadores: %v", err)
			}

			// Guardar el bloque de apuntadores actualizado
			err = pointerBlock.Encode(file, int64(sb.S_block_start+(pointerBlockIndex*sb.S_block_size)))
			if err != nil {
				return fmt.Errorf("error al guardar el bloque de apuntadores: %v", err)
			}

			// Escribir el contenido en el nuevo bloque
			err = blocks[i].Encode(file, int64(sb.S_block_start+(newBlockIndex*sb.S_block_size)))
			if err != nil {
				return fmt.Errorf("error al escribir el nuevo bloque %d: %v", newBlockIndex, err)
			}
		}
	}

	// Actualizar el tamaño del archivo en el inodo
	inode.I_size = int32(len(newContent))
	err = inode.Encode(file, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil {
		return fmt.Errorf("error al actualizar el inodo %d: %v", inodeIndex, err)
	}

	return nil
}
