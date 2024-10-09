package structs

import (
	"encoding/binary"
	"fmt"
	"os"
)

const (
	FreeBlockBit     = 0
	OccupiedBlockBit = 1
)

// CreateBitMaps crea los Bitmaps de inodos y bloques en el archivo especificado
func (sb *Superblock) CreateBitMaps(file *os.File) error {
	// Crear el bitmap de inodos
	err := sb.createBitmap(file, sb.S_bm_inode_start, sb.S_inodes_count+sb.S_free_inodes_count, false)
	if err != nil {
		return fmt.Errorf("error creando bitmap de inodos: %w", err)
	}

	// Crear el bitmap de bloques
	err = sb.createBitmap(file, sb.S_bm_block_start, sb.S_blocks_count+sb.S_free_blocks_count, false)
	if err != nil {
		return fmt.Errorf("error creando bitmap de bloques: %w", err)
	}

	return nil
}

// Cada bloque o inodo está representado por un bit
func (sb *Superblock) createBitmap(file *os.File, start int32, count int32, occupied bool) error {
	_, err := file.Seek(int64(start), 0)
	if err != nil {
		return fmt.Errorf("error buscando el inicio del bitmap: %w", err)
	}

	// Calcular el número de bytes necesarios (cada byte tiene 8 bits)
	byteCount := (count + 7) / 8

	// Crear el buffer de bytes con todos los bits en 0 (libres) o 1 (ocupados)
	fillByte := byte(0x00) // 00000000 (todos los bloques libres)
	if occupied {
		fillByte = 0xFF // 11111111 (todos los bloques ocupados)
	}

	buffer := make([]byte, byteCount)
	for i := range buffer {
		buffer[i] = fillByte
	}

	// Escribir el buffer en el archivo
	err = binary.Write(file, binary.LittleEndian, buffer)
	if err != nil {
		return fmt.Errorf("error escribiendo el bitmap: %w", err)
	}

	return nil
}

// UpdateBitmapInode actualiza el bitmap de inodos
func (sb *Superblock) UpdateBitmapInode(file *os.File, position int32, occupied bool) error {
	return sb.updateBitmap(file, sb.S_bm_inode_start, position, occupied)
}

// UpdateBitmapBlock actualiza el bitmap de bloques
func (sb *Superblock) UpdateBitmapBlock(file *os.File, position int32, occupied bool) error {
	return sb.updateBitmap(file, sb.S_bm_block_start, position, occupied)
}

// updateBitmap es una función auxiliar que actualiza un bit en un bitmap
func (sb *Superblock) updateBitmap(file *os.File, start int32, position int32, occupied bool) error {
	// Calcular el byte y el bit dentro de ese byte
	byteIndex := position / 8
	bitOffset := position % 8

	// Mover el puntero al byte correspondiente
	_, err := file.Seek(int64(start)+int64(byteIndex), 0)
	if err != nil {
		return fmt.Errorf("error buscando la posición en el bitmap: %w", err)
	}

	// Leer el byte actual
	var byteVal byte
	err = binary.Read(file, binary.LittleEndian, &byteVal)
	if err != nil {
		return fmt.Errorf("error leyendo el byte del bitmap: %w", err)
	}

	// Actualizar el bit correspondiente dentro del byte
	if occupied {
		byteVal |= (1 << bitOffset) // Poner el bit a 1 (ocupado)
	} else {
		byteVal &= ^(1 << bitOffset) // Poner el bit a 0 (libre)
	}

	// Mover el puntero de nuevo al byte correspondiente
	_, err = file.Seek(int64(start)+int64(byteIndex), 0)
	if err != nil {
		return fmt.Errorf("error buscando la posición en el bitmap para escribir: %w", err)
	}

	// Escribir el byte actualizado de vuelta en el archivo
	err = binary.Write(file, binary.LittleEndian, &byteVal)
	if err != nil {
		return fmt.Errorf("error escribiendo el byte actualizado del bitmap: %w", err)
	}

	return nil
}

// isBlockFree verifica si un bloque en el bitmap está libre
func (sb *Superblock) isBlockFree(file *os.File, start int32, position int32) (bool, error) {
	fmt.Printf("=== isBlockFree: Verificando si el bloque en la posición %d está libre ===\n", position)

	// Calcular el byte y el bit dentro del byte
	byteIndex := position / 8
	bitOffset := position % 8
	fmt.Printf("Calculando byteIndex: %d y bitOffset: %d\n", byteIndex, bitOffset)

	// Mover el puntero al byte correspondiente
	_, err := file.Seek(int64(start)+int64(byteIndex), 0)
	if err != nil {
		return false, fmt.Errorf("error buscando la posición en el bitmap en byteIndex %d: %w", byteIndex, err)
	}
	fmt.Printf("Puntero movido a la posición %d en el archivo.\n", int64(start)+int64(byteIndex))

	// Leer el byte actual
	var byteVal byte
	err = binary.Read(file, binary.LittleEndian, &byteVal)
	if err != nil {
		return false, fmt.Errorf("error leyendo el byte en byteIndex %d del bitmap: %w", byteIndex, err)
	}
	fmt.Printf("Byte leído: %08b\n", byteVal)

	// Verificar si el bit está libre (0) o ocupado (1)
	isFree := (byteVal & (1 << bitOffset)) == 0
	if isFree {
		fmt.Printf("El bloque en la posición %d está libre.\n", position)
	} else {
		fmt.Printf("El bloque en la posición %d está ocupado.\n", position)
	}
	return isFree, nil
}

// isInodeFree verifica si un inodo en el bitmap está libre
func (sb *Superblock) isInodeFree(file *os.File, start int32, position int32) (bool, error) {
	fmt.Printf("=== isInodeFree: Verificando si el inodo en la posición %d está libre ===\n", position)

	// Calcular el byte y el bit dentro del byte
	byteIndex := position / 8
	bitOffset := position % 8
	fmt.Printf("Calculando byteIndex: %d y bitOffset: %d\n", byteIndex, bitOffset)

	// Mover el puntero al byte correspondiente
	_, err := file.Seek(int64(start)+int64(byteIndex), 0)
	if err != nil {
		return false, fmt.Errorf("error buscando la posición en el bitmap en byteIndex %d: %w", byteIndex, err)
	}
	fmt.Printf("Puntero movido a la posición %d en el archivo.\n", int64(start)+int64(byteIndex))

	// Leer el byte actual
	var byteVal byte
	err = binary.Read(file, binary.LittleEndian, &byteVal)
	if err != nil {
		return false, fmt.Errorf("error leyendo el byte en byteIndex %d del bitmap de inodos: %w", byteIndex, err)
	}
	fmt.Printf("Byte leído: %08b\n", byteVal)

	// Verificar si el bit correspondiente está en 0 (libre)
	isFree := (byteVal & (1 << bitOffset)) == 0
	if isFree {
		fmt.Printf("El inodo en la posición %d está libre.\n", position)
	} else {
		fmt.Printf("El inodo en la posición %d está ocupado.\n", position)
	}
	return isFree, nil
}

// FreeBlock libera un bloque específico y lo marca como libre en el bitmap
func (sb *Superblock) FreeBlock(file *os.File, blockIndex int32) error {
	fmt.Printf("Liberando el bloque en la posición %d.\n", blockIndex)

	// Marcar el bloque como libre en el bitmap
	err := sb.UpdateBitmapBlock(file, blockIndex, false)
	if err != nil {
		return fmt.Errorf("error al liberar el bloque %d: %w", blockIndex, err)
	}

	// Actualizar el superbloque después de liberar el bloque
	sb.UpdateSuperblockAfterBlockDeallocation()
	fmt.Printf("Bloque %d liberado y actualizado en el superbloque.\n", blockIndex)

	return nil
}

// FreeInode libera un inodo específico y lo marca como libre en el bitmap
func (sb *Superblock) FreeInode(file *os.File, inodeIndex int32) error {
	fmt.Printf("Liberando el inodo en la posición %d.\n", inodeIndex)

	// Marcar el inodo como libre en el bitmap
	err := sb.UpdateBitmapInode(file, inodeIndex, false)
	if err != nil {
		return fmt.Errorf("error al liberar el inodo %d: %w", inodeIndex, err)
	}

	// Actualizar el superbloque después de liberar el inodo
	sb.UpdateSuperblockAfterInodeDeallocation()
	fmt.Printf("Inodo %d liberado y actualizado en el superbloque.\n", inodeIndex)

	return nil
}
