package structs

import (
	utilidades "backend/utils"
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

func (sb *Superblock) CreateUsersFile(file *os.File) error {
	// ----------- Crear Inodo Raíz -----------
	rootInode := &Inode{
		I_uid:   1,
		I_gid:   1,
		I_size:  0,
		I_atime: float32(time.Now().Unix()),
		I_ctime: float32(time.Now().Unix()),
		I_mtime: float32(time.Now().Unix()),
		I_block: [15]int32{sb.S_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		I_type:  [1]byte{'0'}, // Tipo carpeta
		I_perm:  [3]byte{'7', '7', '7'},
	}

	// Escribir el inodo raíz (inodo 0)
	err := utilidades.WriteToFile(file, int64(sb.S_inode_start), rootInode)
	if err != nil {
		return fmt.Errorf("error al escribir el inodo raíz: %w", err)
	}

	// Actualizar bitmap de inodos (índice 0)
	err = sb.UpdateBitmapInode(file, 0, true)
	if err != nil {
		return fmt.Errorf("error al actualizar bitmap de inodos: %w", err)
	}

	// Actualizar el contador de inodos y el puntero al primer inodo libre
	sb.UpdateSuperblockAfterInodeAllocation()

	// ----------- Crear Bloque Raíz (/ carpeta) -----------
	rootBlock := &FolderBlock{
		B_content: [4]FolderContent{
			{B_name: [12]byte{'.'}, B_inodo: 0},                                                         // Apunta a sí mismo
			{B_name: [12]byte{'.', '.'}, B_inodo: 0},                                                    // Apunta al padre
			{B_name: [12]byte{'u', 's', 'e', 'r', 's', '.', 't', 'x', 't'}, B_inodo: sb.S_inodes_count}, // Apunta a users.txt
			{B_name: [12]byte{'-'}, B_inodo: -1},                                                        // Vacío
		},
	}

	// Escribir el bloque raíz
	err = utilidades.WriteToFile(file, int64(sb.S_block_start), rootBlock)
	if err != nil {
		return fmt.Errorf("error al escribir el bloque raíz: %w", err)
	}

	// Actualizar bitmap de bloques (índice 0)
	err = sb.UpdateBitmapBlock(file, 0, true)
	if err != nil {
		return fmt.Errorf("error al actualizar el bitmap de bloques: %w", err)
	}

	// Actualizar el contador de bloques y el puntero al primer bloque libre
	sb.UpdateSuperblockAfterBlockAllocation()

	// ----------- Crear Inodo para /users.txt (inodo 1) -----------
	rootGroup := NewGroup("1", "root")
	rootUser := NewUser("1", "root", "root", "123")
	usersText := fmt.Sprintf("%s\n%s\n", rootGroup.ToString(), rootUser.ToString())

	usersInode := &Inode{
		I_uid:   1,
		I_gid:   1,
		I_size:  int32(len(usersText)),
		I_atime: float32(time.Now().Unix()),
		I_ctime: float32(time.Now().Unix()),
		I_mtime: float32(time.Now().Unix()),
		I_block: [15]int32{1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, // Apunta al bloque 1 (users.txt)
		I_type:  [1]byte{'1'},                                                         // Tipo archivo
		I_perm:  [3]byte{'7', '7', '7'},
	}

	// Escribir el inodo de users.txt (inodo 1)
	err = utilidades.WriteToFile(file, int64(sb.S_inode_start+int32(binary.Size(usersInode))), usersInode)
	if err != nil {
		return fmt.Errorf("error al escribir el inodo de users.txt: %w", err)
	}

	// Actualizar bitmap de inodos (índice 1)
	err = sb.UpdateBitmapInode(file, 1, true)
	if err != nil {
		return fmt.Errorf("error al actualizar bitmap de inodos para users.txt: %w", err)
	}

	// Actualizar el contador de inodos y el puntero al primer inodo libre
	sb.UpdateSuperblockAfterInodeAllocation()

	// ----------- Crear Bloque para users.txt (bloque 1) -----------
	usersBlock := &FileBlock{}
	copy(usersBlock.B_content[:], usersText)

	// Escribir el bloque de users.txt
	err = utilidades.WriteToFile(file, int64(sb.S_block_start+int32(binary.Size(usersBlock))), usersBlock)
	if err != nil {
		return fmt.Errorf("error al escribir el bloque de users.txt: %w", err)
	}

	// Actualizar el bitmap de bloques (índice 1)
	err = sb.UpdateBitmapBlock(file, 1, true)
	if err != nil {
		return fmt.Errorf("error al actualizar el bitmap de bloques para users.txt: %w", err)
	}

	// Actualizar el contador de bloques y el puntero al primer bloque libre
	sb.UpdateSuperblockAfterBlockAllocation()

	fmt.Println("Archivo users.txt creado correctamente.")
	fmt.Println("Superbloque después de la creación de users.txt:")
	sb.Print()
	fmt.Println("\nBloques:")
	sb.PrintBlocks(file.Name())
	fmt.Println("\nInodos:")
	sb.PrintInodes(file.Name())
	return nil
}
