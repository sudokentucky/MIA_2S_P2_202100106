package structs

import (
	"fmt"
	"os"
)

func (sb *Superblock) CreateUsersFile(file *os.File) error {
	// ----------- Crear Inodo Raíz -----------

	// Encontrar el primer bloque libre para la raíz
	rootBlockIndex, err := sb.FindNextFreeBlock(file)
	if err != nil {
		return fmt.Errorf("error al encontrar el primer bloque libre para la raíz: %w", err)
	}

	// Crear los bloques asignados para la raíz
	rootBlocks := [15]int32{rootBlockIndex, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}

	// Crear el inodo raíz utilizando la función CreateInode
	rootInode := &Inode{}
	err = rootInode.CreateInode(
		file,
		sb,
		'0',                    // Tipo de inodo: '0' para directorio
		0,                      // Tamaño del directorio (0 al inicio)
		rootBlocks,             // Bloques asignados para la raíz
		[3]byte{'7', '7', '7'}, // Permisos para el directorio raíz
	)
	if err != nil {
		return fmt.Errorf("error al crear el inodo raíz: %w", err)
	}

	// Crear el bloque de la carpeta raíz
	rootBlock := &FolderBlock{
		B_content: [4]FolderContent{
			{B_name: [12]byte{'.'}, B_inodo: 0},      // Apunta a sí mismo
			{B_name: [12]byte{'.', '.'}, B_inodo: 0}, // Apunta al padre
			{B_name: [12]byte{'-'}, B_inodo: -1},     // Bloque vacío
			{B_name: [12]byte{'-'}, B_inodo: -1},     // Bloque vacío
		},
	}

	// Actualizar el bitmap de bloques y serializar el bloque raíz
	err = sb.UpdateBitmapBlock(file, rootBlockIndex, true) // true indica que el bloque está ocupado
	if err != nil {
		return fmt.Errorf("error actualizando el bitmap de bloques: %w", err)
	}
	err = rootBlock.Encode(file, int64(sb.S_block_start+rootBlockIndex))
	if err != nil {
		return fmt.Errorf("error serializando el bloque raíz: %w", err)
	}

	// Actualizar el contador de bloques en el superbloque
	sb.UpdateSuperblockAfterBlockAllocation()

	// ----------- Crear Inodo para /users.txt -----------

	// Preparar el contenido del archivo /users.txt
	rootGroup := NewGroup("1", "root")
	rootUser := NewUser("1", "root", "root", "123")
	usersText := fmt.Sprintf("%s\n%s\n", rootGroup.ToString(), rootUser.ToString())

	// Encontrar el primer bloque libre para /users.txt
	usersBlockIndex, err := sb.FindNextFreeBlock(file)
	if err != nil {
		return fmt.Errorf("error al encontrar el primer bloque libre para /users.txt: %w", err)
	}

	// Crear los bloques asignados para el archivo /users.txt
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

	// Crear el bloque para users.txt
	usersBlock := &FileBlock{}
	usersBlock.AppendContent(usersText)

	// Serializar el bloque de /users.txt
	err = usersBlock.Encode(file, int64(sb.S_block_start+usersInode.I_size))
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
	err = rootBlock.Encode(file, int64(sb.S_block_start+0))
	if err != nil {
		return fmt.Errorf("error serializando el bloque raíz actualizado: %w", err)
	}

	// Actualizar el contador de bloques en el superbloque después de la asignación
	sb.UpdateSuperblockAfterBlockAllocation()

	// Mostrar el resultado final
	fmt.Println("Archivo users.txt creado correctamente en EXT2.")
	fmt.Println("Superbloque después de la creación de users.txt:")
	sb.Print()
	fmt.Println("\nBloques:")
	sb.PrintBlocks(file.Name())
	fmt.Println("\nInodos:")
	sb.PrintInodes(file.Name())

	return nil
}
