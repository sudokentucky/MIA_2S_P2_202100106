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

type RENAME struct {
	path string // Ruta del archivo
	name string // Nuevo nombre del archivo
}

// ParserRename parsea el comando rename y devuelve una instancia de RENAME
func ParserRename(tokens []string) (string, error) {
	cmd := &RENAME{}              // Crea una nueva instancia de RENAME
	var outputBuffer bytes.Buffer // Buffer para capturar mensajes importantes

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

	// Ejecutar el comando RENAME
	err := commandRename(cmd, &outputBuffer)
	if err != nil {
		return "", err
	}

	return outputBuffer.String(), nil
}

func commandRename(renameCmd *RENAME, outputBuffer *bytes.Buffer) error {
	fmt.Fprint(outputBuffer, "======================= RENAME =======================\n")

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

	// Desglosar el path en directorios y el archivo/carpeta a renombrar
	parentDirs, oldName := utils.GetParentDirectories(renameCmd.path)

	// Buscar el inodo del directorio donde está el archivo/carpeta
	inodeIndex, err := findFolderInode(file, partitionSuperblock, parentDirs)
	if err != nil {
		return fmt.Errorf("error al encontrar el directorio padre: %v", err)
	}

	// Cargar el FolderBlock del directorio padre
	folderBlock := &structs.FolderBlock{}
	err = folderBlock.Decode(file, int64(partitionSuperblock.S_block_start+(inodeIndex*partitionSuperblock.S_block_size)))
	if err != nil {
		return fmt.Errorf("error al deserializar el bloque de carpeta: %v", err)
	}

	// Verificar que no exista un archivo/carpeta con el nuevo nombre
	for _, content := range folderBlock.B_content {
		if strings.EqualFold(strings.Trim(string(content.B_name[:]), "\x00 "), renameCmd.name) {
			return fmt.Errorf("ya existe un archivo o carpeta con el nombre '%s'", renameCmd.name)
		}
	}

	// Renombrar el archivo/carpeta usando la función `RenameInFolderBlock`
	err = folderBlock.RenameInFolderBlock(oldName, renameCmd.name)
	if err != nil {
		return fmt.Errorf("error al renombrar el archivo o carpeta: %v", err)
	}

	// Guardar el bloque modificado de nuevo en el archivo
	err = folderBlock.Encode(file, int64(partitionSuperblock.S_block_start+(inodeIndex*partitionSuperblock.S_block_size)))
	if err != nil {
		return fmt.Errorf("error al guardar el bloque de carpeta modificado: %v", err)
	}

	fmt.Fprintf(outputBuffer, "Nombre cambiado exitosamente de '%s' a '%s'\n", oldName, renameCmd.name)
	fmt.Fprint(outputBuffer, "=====================================================\n")

	return nil
}
