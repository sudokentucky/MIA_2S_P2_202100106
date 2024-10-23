package structs

import (
	"fmt"
	"os"
)

func (sb *Superblock) CreateUsersFileExt3(file *os.File, journaling_start int64) error {
	// ----------- Creamos / (la raíz) -----------
	// Crear journal y registrar la operación de creación de la raíz en el journal
	rootJournal := &Journal{
		J_count: sb.S_inodes_count,
	}

	err := rootJournal.SaveJournalEntry(
		file,             // El archivo donde se serializa el journal
		journaling_start, // Inicio del journal en el archivo
		"mkdir",          // Tipo de operación (crear directorio)
		"/",              // Ruta de la operación (la raíz)
		"",               // Contenido vacío porque es un directorio
	)
	if err != nil {
		return fmt.Errorf("error al guardar la entrada de la raíz en el journal: %w", err)
	}
	//Codificar el journal
	err = rootJournal.Encode(file, journaling_start)
	if err != nil {
		return fmt.Errorf("error al codificar el journal: %w", err)
	}

	// Encontrar el primer bloque libre para la raíz
	rootBlockIndex, err := sb.FindNextFreeBlock(file)
	if err != nil {
		return fmt.Errorf("error al encontrar el primer bloque libre para la raíz: %w", err)
	}

	// Bloques para la raíz (directorios), solo ocupamos un bloque
	rootBlocks := [15]int32{rootBlockIndex, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}

	// Crear el inodo raíz utilizando la función genérica CreateInode
	rootInode := &Inode{}
	err = rootInode.CreateInode(
		file,
		sb,
		'0',                    // Tipo de inodo: '0' para directorio
		0,                      // Tamaño del directorio (0 al inicio)
		rootBlocks,             // Bloques asignados
		[3]byte{'7', '7', '7'}, // Permisos para el directorio raíz
	)
	if err != nil {
		return fmt.Errorf("error al crear el inodo raíz: %w", err)
	}

	// Creamos el bloque del Inodo Raíz
	rootBlock := &FolderBlock{
		B_content: [4]FolderContent{
			{B_name: [12]byte{'.'}, B_inodo: 0},      // El inodo raíz siempre es 0
			{B_name: [12]byte{'.', '.'}, B_inodo: 0}, // El inodo padre también es 0
			{B_name: [12]byte{'-'}, B_inodo: -1},     // Bloque vacío
			{B_name: [12]byte{'-'}, B_inodo: -1},     // Bloque vacío
		},
	}

	// Actualizar el bitmap de bloques
	err = sb.UpdateBitmapBlock(file, rootBlockIndex, true) // true indica que el bloque está ocupado
	if err != nil {
		return fmt.Errorf("error actualizando el bitmap de bloques: %w", err)
	}

	// Serializar el bloque de carpeta raíz
	err = rootBlock.Encode(file, int64(sb.S_first_blo))
	if err != nil {
		return fmt.Errorf("error serializando el bloque raíz: %w", err)
	}

	// Actualizar el contador de bloques en el superbloque
	sb.UpdateSuperblockAfterBlockAllocation()

	// ----------- Creamos /users.txt -----------
	rootGroup := NewGroup("1", "root")
	rootUser := NewUser("1", "root", "root", "123")
	usersText := fmt.Sprintf("%s\n%s\n", rootGroup.ToString(), rootUser.ToString())

	//Se crea el journal de users.txt
	FileJournal := &Journal{
		J_count: sb.S_inodes_count,
	}

	// Registrar la operación de creación de /users.txt en el journal
	err = FileJournal.SaveJournalEntry(
		file,             // El archivo donde se serializa el journal
		journaling_start, // Inicio del journal en el archivo
		"mkfile",         // Tipo de operación (crear archivo)
		"/users.txt",     // Ruta del archivo
		usersText,        // Contenido del archivo
	)
	if err != nil {
		return fmt.Errorf("error al guardar la entrada del archivo /users.txt en el journal: %w", err)
	}

	//Codificar el journal
	err = FileJournal.Encode(file, journaling_start)
	if err != nil {
		return fmt.Errorf("error al codificar el journal: %w", err)
	}
	// Encontrar el primer bloque libre para /users.txt
	usersBlockIndex, err := sb.FindNextFreeBlock(file)
	if err != nil {
		return fmt.Errorf("error al encontrar el primer bloque libre para /users.txt: %w", err)
	}

	// Bloques para el archivo /users.txt
	fileBlocks := [15]int32{usersBlockIndex, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}

	// Crear el inodo para /users.txt utilizando CreateInode
	usersInode := &Inode{}
	err = usersInode.CreateInode(
		file,                   // Archivo del sistema de archivos
		sb,                     // Superbloque
		'1',                    // Tipo de inodo: '1' para archivo
		int32(len(usersText)),  // Tamaño del archivo basado en el contenido
		fileBlocks,             // Bloques asignados para el archivo
		[3]byte{'7', '7', '7'}, // Permisos para el archivo
	)
	if err != nil {
		return fmt.Errorf("error al crear el inodo de /users.txt: %w", err)
	}

	// Creamos el bloque de users.txt
	usersBlock := &FileBlock{
		B_content: [64]byte{},
	}
	// Copiamos el texto de usuarios en el bloque
	usersBlock.AppendContent(usersText)

	// Serializar el bloque de users.txt
	//s_first_blo es el primer bloque de datos + el tamaño del inodo de users.txt = el inicio del bloque de users.txt
	err = usersBlock.Encode(file, int64(sb.S_first_blo+usersInode.I_size))
	if err != nil {
		return fmt.Errorf("error serializando el bloque de /users.txt: %w", err)
	}

	// Actualizar el bitmap de bloques para users.txt
	err = sb.UpdateBitmapBlock(file, usersBlockIndex, true)
	if err != nil {
		return fmt.Errorf("error actualizando el bitmap de bloques para /users.txt: %w", err)
	}

	// Actualizar el bloque del inodo raíz para apuntar a /users.txt
	rootBlock.B_content[2] = FolderContent{
		B_name:  [12]byte{'u', 's', 'e', 'r', 's', '.', 't', 'x', 't'}, // Nombre del archivo
		B_inodo: 1,                                                     // Inodo donde se encuentra /users.txt
	}

	// Serializar nuevamente el bloque raíz actualizado
	err = rootBlock.Encode(file, int64(sb.S_block_start+rootBlockIndex))
	if err != nil {
		return fmt.Errorf("error serializando el bloque raíz actualizado: %w", err)
	}

	// Actualizamos el superbloque
	sb.UpdateSuperblockAfterBlockAllocation()
	//mostar las estructuras
	fmt.Println("Bloques")
	sb.PrintBlocks(file.Name())
	fmt.Println("Journal")
	rootJournal.Print()
	FileJournal.Print()
	return nil
}
