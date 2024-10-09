package structs

import (
	utilidades "backend/utils" // Importa el paquete utils
	"encoding/binary"
	"fmt"
	"os"
	"time"
)

type Superblock struct {
	S_filesystem_type   int32   // Número que identifica el sistema de archivos usado
	S_inodes_count      int32   // Número total de inodos creados
	S_blocks_count      int32   // Número total de bloques creados
	S_free_blocks_count int32   // Número de bloques libres
	S_free_inodes_count int32   // Número de inodos libres
	S_mtime             float64 // Última fecha en que el sistema fue montado
	S_umtime            float64 // Última fecha en que el sistema fue desmontado
	S_mnt_count         int32   // Número de veces que se ha montado el sistema
	S_magic             int32   // Valor que identifica el sistema de archivos
	S_inode_size        int32   // Tamaño de la estructura inodo
	S_block_size        int32   // Tamaño de la estructura bloque
	S_first_ino         int32   // Primer inodo libre
	S_first_blo         int32   // Primer bloque libre
	S_bm_inode_start    int32   // Inicio del bitmap de inodos
	S_bm_block_start    int32   // Inicio del bitmap de bloques
	S_inode_start       int32   // Inicio de la tabla de inodos
	S_block_start       int32   // Inicio de la tabla de bloques
}

// Encode codifica la estructura Superblock en un archivo
func (sb *Superblock) Encode(file *os.File, offset int64) error {
	return utilidades.WriteToFile(file, offset, sb)
}

// Decode decodifica la estructura Superblock desde un archivo
func (sb *Superblock) Decode(file *os.File, offset int64) error {
	return utilidades.ReadFromFile(file, offset, sb)
}

// Print imprime los valores de la estructura SuperBlock
func (sb *Superblock) Print() {
	fmt.Printf("%-25s %-10s\n", "Campo", "Valor")
	fmt.Printf("%-25s %-10s\n", "-------------------------", "----------")
	fmt.Printf("%-25s %-10d\n", "S_filesystem_type:", sb.S_filesystem_type)
	fmt.Printf("%-25s %-10d\n", "S_inodes_count:", sb.S_inodes_count)
	fmt.Printf("%-25s %-10d\n", "S_blocks_count:", sb.S_blocks_count)
	fmt.Printf("%-25s %-10d\n", "S_free_blocks_count:", sb.S_free_blocks_count)
	fmt.Printf("%-25s %-10d\n", "S_free_inodes_count:", sb.S_free_inodes_count)
	fmt.Printf("%-25s %-10s\n", "S_mtime:", time.Unix(int64(sb.S_mtime), 0).Format("02/01/2006 15:04"))
	fmt.Printf("%-25s %-10s\n", "S_umtime:", time.Unix(int64(sb.S_umtime), 0).Format("02/01/2006 15:04"))
	fmt.Printf("%-25s %-10d\n", "S_mnt_count:", sb.S_mnt_count)
	fmt.Printf("%-25s %-10x\n", "S_magic:", sb.S_magic)
	fmt.Printf("%-25s %-10d\n", "S_inode_size:", sb.S_inode_size)
	fmt.Printf("%-25s %-10d\n", "S_block_size:", sb.S_block_size)
	fmt.Printf("%-25s %-10d\n", "S_first_ino:", sb.S_first_ino)
	fmt.Printf("%-25s %-10d\n", "S_first_blo:", sb.S_first_blo)
	fmt.Printf("%-25s %-10d\n", "S_bm_inode_start:", sb.S_bm_inode_start)
	fmt.Printf("%-25s %-10d\n", "S_bm_block_start:", sb.S_bm_block_start)
	fmt.Printf("%-25s %-10d\n", "S_inode_start:", sb.S_inode_start)
	fmt.Printf("%-25s %-10d\n", "S_block_start:", sb.S_block_start)
}

// PrintInodes imprime los inodos desde el archivo
func (sb *Superblock) PrintInodes(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer file.Close()

	fmt.Println("\nInodos\n----------------")
	inodes := make([]Inode, sb.S_inodes_count)

	// Deserializar todos los inodos en memoria
	for i := int32(0); i < sb.S_inodes_count; i++ {
		inode := &inodes[i]
		err := utilidades.ReadFromFile(file, int64(sb.S_inode_start+(i*int32(binary.Size(Inode{})))), inode)
		if err != nil {
			return fmt.Errorf("failed to decode inode %d: %w", i, err)
		}
	}

	// Imprimir los inodos
	for i, inode := range inodes {
		fmt.Printf("\nInodo %d:\n", i)
		inode.Print()
	}

	return nil
}

// PrintBlocks imprime los bloques desde el archivo
func (sb *Superblock) PrintBlocks(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", path, err)
	}
	defer file.Close()

	fmt.Println("\nBloques\n----------------")
	inodes := make([]Inode, sb.S_inodes_count)

	// Deserializar todos los inodos en memoria
	for i := int32(0); i < sb.S_inodes_count; i++ {
		inode := &inodes[i]
		err := utilidades.ReadFromFile(file, int64(sb.S_inode_start+(i*int32(binary.Size(Inode{})))), inode)
		if err != nil {
			return fmt.Errorf("failed to decode inode %d: %w", i, err)
		}
	}

	// Imprimir los bloques
	for _, inode := range inodes {
		for _, blockIndex := range inode.I_block {
			if blockIndex == -1 {
				break
			}
			if inode.I_type[0] == '0' {
				block := &FolderBlock{}
				err := utilidades.ReadFromFile(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size)), block)
				if err != nil {
					return fmt.Errorf("failed to decode folder block %d: %w", blockIndex, err)
				}
				fmt.Printf("\nBloque %d:\n", blockIndex)
				block.Print()
			} else if inode.I_type[0] == '1' {
				block := &FileBlock{}
				err := utilidades.ReadFromFile(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size)), block)
				if err != nil {
					return fmt.Errorf("failed to decode file block %d: %w", blockIndex, err)
				}
				fmt.Printf("\nBloque %d:\n", blockIndex)
				block.Print()
			}
		}
	}

	return nil
}

// FindNextFreeBlock busca el siguiente bloque libre y lo marca como ocupado
func (sb *Superblock) FindNextFreeBlock(file *os.File) (int32, error) {
	totalBlocks := sb.S_blocks_count + sb.S_free_blocks_count // Número total de bloques

	for position := int32(0); position < totalBlocks; position++ {
		isFree, err := sb.isBlockFree(file, sb.S_bm_block_start, position)
		if err != nil {
			return -1, fmt.Errorf("error buscando bloque libre: %w", err)
		}

		if isFree {
			// Marcar el bloque como ocupado
			err = sb.UpdateBitmapBlock(file, position, true)
			if err != nil {
				return -1, fmt.Errorf("error actualizando el bitmap del bloque: %w", err)
			}

			// Devolver el índice del bloque libre encontrado
			fmt.Println("Indice encontrado:", position)
			return position, nil
		}
	}

	// Si no hay bloques disponibles
	return -1, fmt.Errorf("no hay bloques disponibles")
}

// AssignNewBlock asigna un nuevo bloque al inodo en el índice especificado si es necesario
func (sb *Superblock) AssignNewBlock(file *os.File, inode *Inode, index int) (int32, error) {
	fmt.Println("=== Iniciando la asignación de un nuevo bloque ===")

	// Validar que el índice esté dentro del rango de bloques válidos
	if index < 0 || index >= len(inode.I_block) {
		return -1, fmt.Errorf("índice de bloque fuera de rango: %d", index)
	}

	// Verificar si ya hay un bloque asignado en ese índice
	if inode.I_block[index] != -1 {
		return -1, fmt.Errorf("bloque en el índice %d ya está asignado: %d", index, inode.I_block[index])
	}

	// Intentar encontrar un bloque libre
	newBlock, err := sb.FindNextFreeBlock(file)
	if err != nil {
		return -1, fmt.Errorf("error buscando nuevo bloque libre: %w", err)
	}

	// Verificar si se encontró un bloque libre
	if newBlock == -1 {
		return -1, fmt.Errorf("no hay bloques libres disponibles")
	}

	// Asignar el nuevo bloque en el índice especificado
	inode.I_block[index] = newBlock
	fmt.Printf("Nuevo bloque asignado: %d en I_block[%d]\n", newBlock, index)

	// Actualizar el Superblock después de asignar el bloque
	sb.UpdateSuperblockAfterBlockAllocation()

	// Retornar el nuevo bloque asignado
	return newBlock, nil
}

// FindNextFreeInode busca el siguiente inodo libre y lo marca como ocupado
func (sb *Superblock) FindNextFreeInode(file *os.File) (int32, error) {
	totalInodes := sb.S_inodes_count + sb.S_free_inodes_count // Número total de inodos

	for position := int32(0); position < totalInodes; position++ {
		isFree, err := sb.isInodeFree(file, sb.S_bm_inode_start, position)
		if err != nil {
			return -1, fmt.Errorf("error buscando inodo libre: %w", err)
		}

		if isFree {
			// Marcar el inodo como ocupado
			err = sb.UpdateBitmapInode(file, position, true)
			if err != nil {
				return -1, fmt.Errorf("error actualizando el bitmap de inodos: %w", err)
			}

			// Devolver el índice del inodo libre encontrado
			fmt.Println("Inodo libre encontrado:", position)
			return position, nil
		}
	}

	// Si no hay inodos disponibles
	return -1, fmt.Errorf("no hay inodos disponibles")
}

// AssignNewInode asigna un nuevo inodo y lo marca como ocupado
func (sb *Superblock) AssignNewInode(file *os.File) (int32, error) {
	fmt.Println("=== Iniciando la asignación de un nuevo inodo ===")

	// Intentar encontrar un inodo libre
	newInode, err := sb.FindNextFreeInode(file)
	if err != nil {
		return -1, fmt.Errorf("error buscando nuevo inodo libre: %w", err)
	}

	// Verificar si se encontró un inodo libre
	if newInode == -1 {
		return -1, fmt.Errorf("no hay inodos libres disponibles")
	}

	// Actualizar el Superblock después de asignar el inodo
	sb.UpdateSuperblockAfterInodeAllocation()

	// Retornar el nuevo inodo asignado
	fmt.Printf("Nuevo inodo asignado: %d\n", newInode)
	return newInode, nil
}

func (sb *Superblock) CalculateInodeOffset(inodeIndex int32) int64 {
	// Calcula el desplazamiento en el archivo basado en el índice del inodo
	return int64(sb.S_inode_start) + int64(inodeIndex)*int64(sb.S_inode_size)
}

// UpdateSuperblockAfterBlockAllocation actualiza el Superblock después de asignar un bloque
func (sb *Superblock) UpdateSuperblockAfterBlockAllocation() {
	// Incrementa el contador de bloques asignados
	sb.S_blocks_count++

	// Decrementa el contador de bloques libres
	sb.S_free_blocks_count--

	// Actualiza el puntero al primer bloque libre
	sb.S_first_blo += sb.S_block_size
}

// Funcion para regresar el contador de bloques una vez que se elimina un bloque
func (sb *Superblock) UpdateSuperblockAfterBlockDeallocation() {
	// Incrementa el contador de bloques asignados
	sb.S_blocks_count--

	// Decrementa el contador de bloques libres
	sb.S_free_blocks_count++

	// Actualiza el puntero al primer bloque libre
	sb.S_first_blo -= sb.S_block_size
}

// UpdateSuperblockAfterInodeAllocation actualiza el Superblock después de asignar un inodo
func (sb *Superblock) UpdateSuperblockAfterInodeAllocation() {
	// Incrementa el contador de inodos asignados
	sb.S_inodes_count++

	// Decrementa el contador de inodos libres
	sb.S_free_inodes_count--

	// Actualiza el puntero al primer inodo libre
	sb.S_first_ino += sb.S_inode_size
}

// UpdateSuperblockAfterInodeDeallocation actualiza el Superblock después de eliminar un inodo
func (sb *Superblock) UpdateSuperblockAfterInodeDeallocation() {
	// Incrementa el contador de inodos asignados
	sb.S_inodes_count--

	// Decrementa el contador de inodos libres
	sb.S_free_inodes_count++

	// Actualiza el puntero al primer inodo libre
	sb.S_first_ino -= sb.S_inode_size
}
