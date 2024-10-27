package commands

import (
	structures "backend/Structs"
	global "backend/globals"
	"backend/utils"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"
	"time"
)

// MKFS estructura que representa el comando mkfs con sus parámetros
type MKFS struct {
	id  string // ID del disco
	typ string // Tipo de formato (full)
	fs  string // Tipo de sistema de archivos (2fs o 3fs)
}

func ParserMkfs(tokens []string) (string, error) {
	var outputBuffer bytes.Buffer
	cmd := &MKFS{}

	args := strings.Join(tokens, " ")
	// Modificado para que acepte también -fs
	re := regexp.MustCompile(`-id=[^\s]+|-type=[^\s]+|-fs=[^\s]+`)
	matches := re.FindAllString(args, -1)

	for _, match := range matches {
		kv := strings.SplitN(match, "=", 2)
		if len(kv) != 2 {
			return "", fmt.Errorf("formato de parámetro inválido: %s", match)
		}
		key, value := strings.ToLower(kv[0]), kv[1]
		if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			value = strings.Trim(value, "\"")
		}

		switch key {
		case "-id":
			if value == "" {
				return "", errors.New("el id no puede estar vacío")
			}
			cmd.id = value
		case "-type":
			if value != "full" {
				return "", errors.New("el tipo debe ser full")
			}
			cmd.typ = value
		case "-fs":
			if value != "2fs" && value != "3fs" {
				return "", errors.New("el sistema de archivos debe ser 2fs o 3fs")
			}
			cmd.fs = value
		default:
			return "", fmt.Errorf("parámetro desconocido: %s", key)
		}
	}

	if cmd.id == "" {
		return "", errors.New("faltan parámetros requeridos: -id")
	}

	// Definir fs por defecto si no se especifica
	if cmd.fs == "" {
		cmd.fs = "2fs" // Por defecto EXT2
	}

	if cmd.typ == "" {
		cmd.typ = "full"
	}

	err := commandMkfs(cmd, &outputBuffer)
	if err != nil {
		fmt.Println("Error:", err)
		return "", err
	}

	return outputBuffer.String(), nil
}

func commandMkfs(mkfs *MKFS, outputBuffer *bytes.Buffer) error {
	fmt.Fprintf(outputBuffer, "========================== MKFS ==========================\n")

	// Obtener la partición montada
	mountedPartition, partitionPath, err := global.GetMountedPartition(mkfs.id)
	if err != nil {
		return fmt.Errorf("error al obtener la partición montada con ID %s: %v", mkfs.id, err)
	}

	// Abrir el archivo de la partición
	file, err := os.OpenFile(partitionPath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("error abriendo el archivo de la partición en %s: %v", partitionPath, err)
	}
	defer file.Close()

	fmt.Fprintf(outputBuffer, "Partición montada correctamente en %s.\n", partitionPath)
	fmt.Println("\nPartición montada:")
	mountedPartition.Print()

	// Calcular el valor de n
	n := calculateN(mountedPartition, mkfs.fs)
	fmt.Println("\nValor de n:", n)

	// Crear el superblock basado en el sistema de archivos especificado (EXT2 o EXT3)
	superBlock := createSuperBlock(mountedPartition, n, mkfs.fs)
	fmt.Println("\nSuperBlock:")
	superBlock.Print()

	// Crear bitmaps
	err = superBlock.CreateBitMaps(file)
	if err != nil {
		return fmt.Errorf("error creando bitmaps: %v", err)
	}
	fmt.Fprintln(outputBuffer, "Bitmaps creados correctamente.")

	// Crear el archivo users.txt (para EXT3 incluye journaling)
	if mkfs.fs == "3fs" {
		err = superBlock.CreateUsersFileExt3(file, int64(mountedPartition.Part_start+int32(binary.Size(structures.Superblock{}))))
	} else {
		err = superBlock.CreateUsersFile(file)
	}
	if err != nil {
		return fmt.Errorf("error creando el archivo users.txt: %v", err)
	}
	fmt.Fprintln(outputBuffer, "Archivo users.txt creado correctamente.")

	// Serializar el superbloque
	err = utils.WriteToFile(file, int64(mountedPartition.Part_start), superBlock)
	if err != nil {
		return fmt.Errorf("error escribiendo el superbloque en el disco: %v", err)
	}
	fmt.Fprintln(outputBuffer, "Superbloque escrito correctamente en el disco.")
	fmt.Fprintln(outputBuffer, "===========================================================")

	return nil
}

func calculateN(partition *structures.Partition, fs string) int32 {
	// Numerador: tamaño de la partición menos el tamaño del superblock
	numerator := int(partition.Part_size) - binary.Size(structures.Superblock{})

	// Denominador base: 4 + tamaño de inodos + 3 * tamaño de bloques de archivo
	baseDenominator := 4 + binary.Size(structures.Inode{}) + 3*binary.Size(structures.FileBlock{})

	// Si el sistema de archivos es "3fs", se añade el tamaño del journaling al denominador
	temp := 0
	if fs == "3fs" {
		temp = binary.Size(structures.Journal{})
	}

	// Denominador final
	denominator := baseDenominator + temp

	// Calcular n
	n := math.Floor(float64(numerator) / float64(denominator))

	return int32(n)
}

func createSuperBlock(partition *structures.Partition, n int32, fs string) *structures.Superblock {
	// Calcular punteros de las estructuras
	journal_start, bm_inode_start, bm_block_start, inode_start, block_start := calculateStartPositions(partition, fs, n)

	fmt.Println("\nInicio del SuperBlock:", partition.Part_start)
	fmt.Println("\nFin del SuperBlock:", partition.Part_start+int32(binary.Size(structures.Superblock{})))
	fmt.Println("\nInicio del Journal:", journal_start)
	fmt.Println("\nFin del Journal:", journal_start+int32(binary.Size(structures.Journal{})))
	fmt.Println("\nInicio del Bitmap de Inodos:", bm_inode_start)
	fmt.Println("\nFin del Bitmap de Inodos:", bm_inode_start+n)
	fmt.Println("\nInicio del Bitmap de Bloques:", bm_block_start)
	fmt.Println("\nFin del Bitmap de Bloques:", bm_block_start+(3*n))
	fmt.Println("\nInicio de Inodos:", inode_start)

	// Tipo de sistema de archivos
	var fsType int32

	if fs == "2fs" {
		fsType = 2
	} else {
		fsType = 3
	}

	// Crear un nuevo superbloque
	superBlock := &structures.Superblock{
		S_filesystem_type:   fsType,
		S_inodes_count:      0,
		S_blocks_count:      0,
		S_free_inodes_count: int32(n),
		S_free_blocks_count: int32(n * 3),
		S_mtime:             float64(time.Now().Unix()),
		S_umtime:            float64(time.Now().Unix()),
		S_mnt_count:         1,
		S_magic:             0xEF53,
		S_inode_size:        int32(binary.Size(structures.Inode{})),
		S_block_size:        int32(binary.Size(structures.FileBlock{})),
		S_first_ino:         inode_start,
		S_first_blo:         block_start,
		S_bm_inode_start:    bm_inode_start,
		S_bm_block_start:    bm_block_start,
		S_inode_start:       inode_start,
		S_block_start:       block_start,
	}
	return superBlock
}

func calculateStartPositions(partition *structures.Partition, fs string, n int32) (int32, int32, int32, int32, int32) {
	superblockSize := int32(binary.Size(structures.Superblock{}))
	journalSize := int32(binary.Size(structures.Journal{}))
	inodeSize := int32(binary.Size(structures.Inode{}))

	// Inicializar posiciones
	// EXT2
	journalStart := int32(0)
	bmInodeStart := partition.Part_start + superblockSize
	bmBlockStart := bmInodeStart + n
	inodeStart := bmBlockStart + (3 * n)
	blockStart := inodeStart + (inodeSize * n)

	// Ajustar para EXT3
	if fs == "3fs" {
		journalStart = partition.Part_start + superblockSize
		bmInodeStart = journalStart + (journalSize * n)
		bmBlockStart = bmInodeStart + n
		inodeStart = bmBlockStart + (3 * n)
		blockStart = inodeStart + (inodeSize * n)
	}

	return journalStart, bmInodeStart, bmBlockStart, inodeStart, blockStart
}
