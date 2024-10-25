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

// REMOVE estructura que representa el comando REMOVE con sus parámetros
type REMOVE struct {
	path string // Ruta del archivo o carpeta a eliminar
}

func ParserRemove(tokens []string) (string, error) {
	cmd := &REMOVE{}              // Crea una nueva instancia de REMOVE
	var outputBuffer bytes.Buffer // Buffer para capturar mensajes importantes

	// Expresión regular para capturar el parámetro -path="ruta"
	re := regexp.MustCompile(`-path=("[^"]+"|[^\s]+)`)
	matches := re.FindAllString(strings.Join(tokens, " "), -1)

	// Verificar si no se encontró la ruta
	if len(matches) == 0 {
		return "", errors.New("no se especificó una ruta para eliminar")
	}

	// Extraer el valor de la ruta
	kv := strings.SplitN(matches[0], "=", 2)
	if len(kv) == 2 {
		cmd.path = kv[1]
		// Si la ruta está entre comillas, eliminarlas
		if strings.HasPrefix(cmd.path, "\"") && strings.HasSuffix(cmd.path, "\"") {
			cmd.path = strings.Trim(cmd.path, "\"")
		}
	}

	// Ejecutar el comando REMOVE
	err := commandRemove(cmd, &outputBuffer)
	if err != nil {
		return "", err
	}

	return outputBuffer.String(), nil
}
func commandRemove(removeCmd *REMOVE, outputBuffer *bytes.Buffer) error {
	fmt.Fprint(outputBuffer, "====================== REMOVE ======================\n")

	// Verificar si hay un usuario logueado
	if !global.IsLoggedIn() {
		return fmt.Errorf("no hay un usuario logueado")
	}

	// Obtener la partición montada asociada al usuario logueado
	idPartition := global.UsuarioActual.Id
	_, _, partitionPath, err := global.GetMountedPartitionSuperblock(idPartition)
	if err != nil {
		return fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	// Abrir el archivo de partición para operar sobre él
	file, err := os.OpenFile(partitionPath, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo de partición: %w", err)
	}
	defer file.Close() // Cerrar el archivo cuando ya no sea necesario

	// Llamar a la función refactorizada para eliminar archivo/carpeta
	err = removeFileOrDirectory(removeCmd.path, file)
	if err != nil {
		return fmt.Errorf("error al eliminar archivo o carpeta: %v", err)
	}

	fmt.Fprintf(outputBuffer, "Archivo o carpeta '%s' eliminado exitosamente.\n", removeCmd.path)
	fmt.Fprint(outputBuffer, "====================================================\n")
	return nil
}

// removeFileOrDirectory elimina un archivo o carpeta dada la ruta
func removeFileOrDirectory(path string, file *os.File) error {
	// Obtener el Superblock y la partición montada asociada
	idPartition := global.UsuarioActual.Id
	partitionSuperblock, _, _, err := global.GetMountedPartitionSuperblock(idPartition)
	if err != nil {
		return fmt.Errorf("error al obtener la partición montada: %v", err)
	}

	// Convertir el path del archivo o carpeta en un array de carpetas
	parentDirs, fileName := utils.GetParentDirectories(path)

	// Intentar eliminar el archivo
	err = removeFile(partitionSuperblock, file, parentDirs, fileName)
	if err == nil {
		// Si el archivo se eliminó correctamente, regresar
		return nil
	}

	// Si no es un archivo, intentar eliminarlo como carpeta
	err = removeDirectory(partitionSuperblock, file, parentDirs, fileName)
	if err != nil {
		return fmt.Errorf("error al eliminar archivo o carpeta '%s': %v", path, err)
	}

	return nil
}

// removeFile intenta eliminar un archivo dado su path
func removeFile(sb *structs.Superblock, file *os.File, parentDirs []string, fileName string) error {
	// Buscar el inodo del archivo
	_, err := findFileInode(file, sb, parentDirs, fileName)
	if err != nil {
		// No se encontró el archivo
		return fmt.Errorf("archivo '%s' no encontrado: %v", fileName, err)
	}

	// Llamar a la función que elimina el archivo
	err = sb.DeleteFile(file, parentDirs, fileName)
	if err != nil {
		return fmt.Errorf("error al eliminar el archivo '%s': %v", fileName, err)
	}

	fmt.Println("Bloques")
	sb.PrintBlocks(file.Name())
	fmt.Printf("Archivo '%s' eliminado correctamente.\n", fileName)
	return nil
}

// removeDirectory intenta eliminar una carpeta dada su path
func removeDirectory(sb *structs.Superblock, file *os.File, parentDirs []string, dirName string) error {
	// Buscar el inodo de la carpeta
	_, err := findFolderInode(file, sb, parentDirs)
	if err != nil {
		// No se encontró la carpeta
		return fmt.Errorf("carpeta '%s' no encontrada: %v", dirName, err)
	}

	// Llamar a la función que elimina la carpeta
	err = sb.DeleteFolder(file, parentDirs, dirName)
	if err != nil {
		return fmt.Errorf("error al eliminar la carpeta '%s': %v", dirName, err)
	}

	fmt.Println("Bloques")
	sb.PrintBlocks(file.Name())
	fmt.Printf("Carpeta '%s' eliminada correctamente.\n", dirName)
	return nil
}
