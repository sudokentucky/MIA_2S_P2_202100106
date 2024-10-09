package commands

import (
	structs "backend/Structs"
	globals "backend/globals"
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// MKUSR : Estructura para el comando MKUSR
type MKUSR struct {
	User string
	Pass string
	Grp  string
}

// validateParamLength : Valida que los parámetros no excedan una longitud máxima
func validateParamLength(param string, maxLength int, paramName string) error {
	if len(param) > maxLength {
		return fmt.Errorf("%s debe tener un máximo de %d caracteres", paramName, maxLength)
	}
	return nil
}

// ParserMkusr : Parseo de argumentos para el comando mkusr y captura de los mensajes importantes
func ParserMkusr(tokens []string) (string, error) {
	var outputBuffer bytes.Buffer // Buffer para capturar los mensajes importantes para el usuario

	// Inicializar el comando MKUSR
	cmd := &MKUSR{}

	// Expresión regular para encontrar los parámetros -user, -pass, -grp
	reUser := regexp.MustCompile(`-user=[^\s]+`)
	rePass := regexp.MustCompile(`-pass=[^\s]+`)
	reGrp := regexp.MustCompile(`-grp=[^\s]+`)

	// Buscar los parámetros
	matchesUser := reUser.FindString(strings.Join(tokens, " "))
	matchesPass := rePass.FindString(strings.Join(tokens, " "))
	matchesGrp := reGrp.FindString(strings.Join(tokens, " "))

	// Verificar que se proporcionen todos los parámetros
	if matchesUser == "" {
		return "", fmt.Errorf("falta el parámetro -user")
	}
	if matchesPass == "" {
		return "", fmt.Errorf("falta el parámetro -pass")
	}
	if matchesGrp == "" {
		return "", fmt.Errorf("falta el parámetro -grp")
	}

	// Extraer los valores de los parámetros
	cmd.User = strings.SplitN(matchesUser, "=", 2)[1]
	cmd.Pass = strings.SplitN(matchesPass, "=", 2)[1]
	cmd.Grp = strings.SplitN(matchesGrp, "=", 2)[1]

	// Validar longitudes de los parámetros
	if err := validateParamLength(cmd.User, 10, "Usuario"); err != nil {
		return "", err
	}
	if err := validateParamLength(cmd.Pass, 10, "Contraseña"); err != nil {
		return "", err
	}
	if err := validateParamLength(cmd.Grp, 10, "Grupo"); err != nil {
		return "", err
	}

	// Ejecutar la lógica del comando mkusr
	err := commandMkusr(cmd, &outputBuffer)
	if err != nil {
		return "", err
	}

	// Retornar los mensajes importantes capturados en el buffer
	return outputBuffer.String(), nil
}

// commandMkusr : Ejecuta el comando MKUSR con captura de mensajes importantes en el buffer
func commandMkusr(mkusr *MKUSR, outputBuffer *bytes.Buffer) error {
	fmt.Fprintln(outputBuffer, "======================= MKUSR =======================")
	// Verificar si hay una sesión activa y si el usuario es root
	if !globals.IsLoggedIn() {
		return fmt.Errorf("no hay ninguna sesión activa")
	}
	if globals.UsuarioActual.Name != "root" {
		return fmt.Errorf("solo el usuario root puede ejecutar este comando")
	}

	// Verificar que la partición esté montada
	_, path, err := globals.GetMountedPartition(globals.UsuarioActual.Id)
	if err != nil {
		return fmt.Errorf("no se puede encontrar la partición montada: %v", err)
	}

	// Abrir el archivo de la partición
	file, err := os.OpenFile(path, os.O_RDWR, 0755)
	if err != nil {
		return fmt.Errorf("no se puede abrir el archivo de la partición: %v", err)
	}
	defer file.Close()

	// Cargar el Superblock y la partición utilizando la función GetMountedPartitionRep
	mbr, sb, _, err := globals.GetMountedPartitionRep(globals.UsuarioActual.Id)
	if err != nil {
		return fmt.Errorf("no se pudo cargar el Superblock: %v", err)
	}

	// Obtener la partición montada
	partition, err := mbr.GetPartitionByID(globals.UsuarioActual.Id)
	if err != nil {
		return fmt.Errorf("no se pudo obtener la partición: %v", err)
	}

	// Leer el inodo de users.txt
	var usersInode structs.Inode
	inodeOffset := int64(sb.S_inode_start + int32(binary.Size(usersInode))) //ubicación de los bloques de users.txt
	err = usersInode.Decode(file, inodeOffset)                              // Usar el descriptor de archivo
	if err != nil {
		return fmt.Errorf("error leyendo el inodo de users.txt: %v", err)
	}

	// Verificar si el grupo existe en el archivo
	_, err = globals.FindInUsersFile(file, sb, &usersInode, mkusr.Grp, "G")
	if err != nil {
		return fmt.Errorf("el grupo '%s' no existe", mkusr.Grp)
	}

	// Verificar si el usuario ya existe
	_, err = globals.FindInUsersFile(file, sb, &usersInode, mkusr.User, "U")
	if err == nil {
		return fmt.Errorf("el usuario '%s' ya existe", mkusr.User)
	}

	// Crear un nuevo objeto de tipo User
	usuario := structs.NewUser(fmt.Sprintf("%d", sb.S_inodes_count+1), mkusr.Grp, mkusr.User, mkusr.Pass)
	fmt.Println(usuario.ToString())

	// Insertar la nueva entrada en el archivo users.txt
	err = globals.InsertIntoUsersFile(file, sb, &usersInode, usuario.ToString())
	if err != nil {
		return fmt.Errorf("error insertando el usuario '%s': %v", mkusr.User, err)
	}

	// Actualizar el inodo de users.txt
	err = usersInode.Encode(file, inodeOffset)
	if err != nil {
		return fmt.Errorf("error actualizando inodo de users.txt: %v", err)
	}

	// Guardar el Superblock usando el Part_start como el offset
	err = sb.Encode(file, int64(partition.Part_start))
	if err != nil {
		return fmt.Errorf("error guardando el Superblock: %v", err)
	}

	// Mostrar mensaje de éxito
	fmt.Fprintf(outputBuffer, "Usuario '%s' agregado exitosamente al grupo '%s'\n", mkusr.User, mkusr.Grp)
	fmt.Println("\nSuperblock")
	sb.Print()
	fmt.Println("\nInodos")
	sb.PrintInodes(file.Name())
	fmt.Println("\nBloques")
	sb.PrintBlocks(file.Name())
	fmt.Fprintf(outputBuffer, "=====================================================")

	return nil
}
