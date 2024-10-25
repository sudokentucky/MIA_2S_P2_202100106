package structs

import (
	"backend/utils"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"
)

func (sb *Superblock) createFileInInode(file *os.File, inodeIndex int32, parentsDir []string, destFile string, fileSize int, fileContent []string) error {
	fmt.Printf("Intentando crear archivo '%s' en inodo index %d\n", destFile, inodeIndex) // Depuración
	//Deserializar el inodo
	inode := &Inode{}
	err := inode.Decode(file, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil {
		return fmt.Errorf("error al deserializar inodo %d: %v", inodeIndex, err)
	}
	// Verificar si el inodo es de tipo carpeta
	if inode.I_type[0] == '1' {
		fmt.Printf("El inodo %d es una carpeta, omitiendo.\n", inodeIndex) // Depuración
		return nil
	}

	// Iterar sobre cada bloque del inodo (apuntadores)
	for _, blockIndex := range inode.I_block {
		// Si el bloque no existe, salir
		if blockIndex == -1 {
			fmt.Printf("El inodo %d no tiene más bloques, saliendo.\n", inodeIndex) // Depuración
			break
		}

		// Crear un nuevo bloque de carpeta
		block := &FolderBlock{}

		// Deserializar el bloque
		err := block.Decode(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size))) // posición del bloque
		if err != nil {
			return fmt.Errorf("error al deserializar bloque %d: %v", blockIndex, err)
		}

		// Iterar sobre cada contenido del bloque, desde el index 2 porque los primeros dos son . y ..
		for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
			content := block.B_content[indexContent]

			// Si hay carpetas padres, buscar la carpeta más cercana
			if len(parentsDir) != 0 {
				if content.B_inodo == -1 {
					fmt.Printf("No se encontró carpeta padre en el inodo %d, saliendo.\n", inodeIndex) // Depuración
					break
				}

				parentDir, err := utils.First(parentsDir)
				if err != nil {
					return err
				}

				contentName := strings.Trim(string(content.B_name[:]), "\x00 ")
				parentDirName := strings.Trim(parentDir, "\x00 ")

				if strings.EqualFold(contentName, parentDirName) {
					fmt.Printf("Encontrada carpeta padre '%s' en inodo %d\n", parentDirName, content.B_inodo) // Depuración
					err := sb.createFileInInode(file, content.B_inodo, utils.RemoveElement(parentsDir, 0), destFile, fileSize, fileContent)
					if err != nil {
						return err
					}
					return nil
				}
			} else {
				// Si el apuntador al inodo está ocupado, continuar con el siguiente
				if content.B_inodo != -1 {
					fmt.Printf("El inodo %d ya está ocupado, continuando.\n", content.B_inodo) // Depuración
					continue
				}

				// Actualizar el contenido del bloque
				copy(content.B_name[:], []byte(destFile))
				content.B_inodo = sb.S_inodes_count
				block.B_content[indexContent] = content

				// Serializar el bloque
				err = block.Encode(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
				if err != nil {
					return fmt.Errorf("error al serializar bloque %d: %v", blockIndex, err)
				}

				fmt.Printf("Bloque actualizado para el archivo '%s' en el inodo %d\n", destFile, sb.S_inodes_count) // Depuración

				// Verificación adicional para realizar el journaling si el sistema de archivos lo requiere
				if sb.S_filesystem_type == 3 { // Aquí usamos el tipo de sistema de archivos (por ejemplo, ext3)
					FileJournal := &Journal{
						J_count: sb.S_inodes_count,
					}
					var superblockSize int64 = int64(binary.Size(sb))
					journaling_start := superblockSize
					err := FileJournal.SaveJournalEntry(
						file,
						journaling_start,              // Iniciar el journaling después del superbloque
						"mkfile",                      // Tipo de operación: crear archivo
						"/"+destFile,                  // Ruta del archivo
						strings.Join(fileContent, ""), // El contenido del archivo
					)
					if err != nil {
						return fmt.Errorf("error al guardar la entrada en el journal: %w", err)
					}

					//Codificar el journal
					err = FileJournal.Encode(file, journaling_start)
					if err != nil {
						return fmt.Errorf("error al codificar el journal: %w", err)
					}
					fmt.Println("Journal creado para el archivo:", destFile)
					FileJournal.Print()
				}

				// Crear el inodo del archivo
				fileInode := &Inode{
					I_uid:   1,
					I_gid:   1,
					I_size:  int32(fileSize),
					I_atime: float32(time.Now().Unix()),
					I_ctime: float32(time.Now().Unix()),
					I_mtime: float32(time.Now().Unix()),
					I_block: [15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
					I_type:  [1]byte{'1'},
					I_perm:  [3]byte{'6', '6', '4'},
				}

				// Combinar todo el contenido en un string
				contentStr := strings.Join(fileContent, "")

				// Dividir el contenido en bloques de tamaño BlockSize
				blocks, err := SplitContent(contentStr)
				if err != nil {
					return fmt.Errorf("error al dividir el contenido en bloques: %v", err)
				}

				blockIndex := 0
				for _, fileBlock := range blocks {
					if blockIndex >= len(fileInode.I_block) {
						return fmt.Errorf("se alcanzó el límite máximo de bloques del inodo")
					}

					if fileInode.I_block[blockIndex] == -1 {
						newBlockIndex, err := sb.AssignNewBlock(file, fileInode, blockIndex)
						if err != nil {
							return fmt.Errorf("error asignando nuevo bloque: %v", err)
						}
						fileInode.I_block[blockIndex] = newBlockIndex
					}

					// Calcular el offset del bloque en el archivo
					blockOffset := int64(sb.S_block_start + fileInode.I_block[blockIndex]*int32(sb.S_block_size))

					// Escribir el contenido del bloque en el archivo
					err = fileBlock.Encode(file, blockOffset)
					if err != nil {
						return fmt.Errorf("error escribiendo bloque %d: %v", fileInode.I_block[blockIndex], err)
					}

					fmt.Printf("Bloque de archivo '%s' serializado correctamente en el bloque %d.\n", destFile, fileInode.I_block[blockIndex]) // Depuración

					// Actualizar el bitmap de bloques
					err = sb.UpdateBitmapBlock(file, fileInode.I_block[blockIndex], true)
					if err != nil {
						return fmt.Errorf("error al actualizar bitmap de bloque: %v", err)
					}

					sb.UpdateSuperblockAfterBlockAllocation()

					blockIndex++
				}

				// Actualizar el tamaño del archivo en el inodo
				fileInode.I_size = int32(fileSize)

				// Actualizar los tiempos de modificación y creación
				fileInode.UpdateMtime()
				fileInode.UpdateCtime()

				// Serializar el inodo
				err = fileInode.Encode(file, int64(sb.S_inode_start+(sb.S_inodes_count*sb.S_inode_size)))
				if err != nil {
					return fmt.Errorf("error al serializar inodo del archivo: %v", err)
				}

				fmt.Printf("Inodo del archivo '%s' serializado correctamente.\n", destFile) // Depuración

				// Actualizar el bitmap de inodos
				err = sb.UpdateBitmapInode(file, sb.S_inodes_count, true)
				if err != nil {
					return fmt.Errorf("error al actualizar bitmap de inodo: %v", err)
				}

				// Actualizar el superbloque
				sb.UpdateSuperblockAfterInodeAllocation()

				fmt.Printf("Archivo '%s' creado correctamente en el inodo %d.\n", destFile, sb.S_inodes_count) // Depuración

				return nil
			}
		}
	}
	return nil
}

// CreateFile crea un archivo en el sistema de archivos
func (sb *Superblock) CreateFile(file *os.File, parentsDir []string, destFile string, size int, cont []string) error {
	fmt.Printf("Creando archivo '%s' con tamaño %d\n", destFile, size) // Depuración

	// Si parentsDir está vacío, solo trabajar con el primer inodo que sería el raíz "/"
	if len(parentsDir) == 0 {
		return sb.createFileInInode(file, 0, parentsDir, destFile, size, cont)
	}

	// Iterar sobre cada inodo ya que se necesita buscar el inodo padre
	for i := int32(0); i < sb.S_inodes_count; i++ {
		err := sb.createFileInInode(file, i, parentsDir, destFile, size, cont)
		if err != nil {
			return err
		}
	}

	fmt.Printf("Archivo '%s' creado exitosamente.\n", destFile) // Depuración
	return nil
}

// deleteFileInInode elimina un archivo en un inodo específico
func (sb *Superblock) deleteFileInInode(file *os.File, inodeIndex int32, fileName string) error {
	// Deserializar el inodo
	inode := &Inode{}
	err := inode.Decode(file, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil {
		return fmt.Errorf("error al deserializar inodo %d: %v", inodeIndex, err)
	}

	// Verificar que el inodo sea una carpeta (I_type[0] == '0')
	if inode.I_type[0] != '0' {
		return fmt.Errorf("el inodo %d no es una carpeta", inodeIndex)
	}

	// Iterar sobre los bloques del inodo (contenidos del directorio)
	for _, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			break // No hay más bloques
		}

		// Deserializar el bloque de la carpeta
		block := &FolderBlock{}
		err = block.Decode(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
		if err != nil {
			return fmt.Errorf("error al deserializar bloque %d: %v", blockIndex, err)
		}

		// Buscar el archivo en los contenidos del bloque
		for _, content := range block.B_content {
			contentName := strings.Trim(string(content.B_name[:]), "\x00 ")
			if content.B_inodo != -1 && strings.EqualFold(contentName, fileName) {
				fmt.Printf("Archivo '%s' encontrado en inodo %d, eliminando.\n", fileName, content.B_inodo)

				// Deserializar el inodo del archivo
				fileInode := &Inode{}
				err := fileInode.Decode(file, int64(sb.S_inode_start+(content.B_inodo*sb.S_inode_size)))
				if err != nil {
					return fmt.Errorf("error al deserializar inodo del archivo %d: %v", content.B_inodo, err)
				}

				// Liberar los bloques asignados al archivo
				for _, fileBlockIndex := range fileInode.I_block {
					if fileBlockIndex != -1 {
						err = sb.FreeBlock(file, fileBlockIndex)
						if err != nil {
							return fmt.Errorf("error al liberar bloque %d: %v", fileBlockIndex, err)
						}
					}
				}

				// Liberar el inodo del archivo
				err = sb.FreeInode(file, content.B_inodo)
				if err != nil {
					return fmt.Errorf("error al liberar inodo %d: %v", content.B_inodo, err)
				}

				// Eliminar la referencia al archivo en el bloque de la carpeta
				content.B_inodo = -1
				copy(content.B_name[:], "")

				// Serializar el bloque actualizado
				err = block.Encode(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
				if err != nil {
					return fmt.Errorf("error al serializar bloque %d después de eliminar archivo: %v", blockIndex, err)
				}

				fmt.Printf("Archivo '%s' eliminado correctamente.\n", fileName)
				return nil
			}
		}
	}

	return fmt.Errorf("archivo '%s' no encontrado en inodo %d", fileName, inodeIndex)
}

// DeleteFile elimina un archivo del sistema de archivos
func (sb *Superblock) DeleteFile(file *os.File, parentsDir []string, fileName string) error {
	fmt.Printf("Intentando eliminar archivo '%s'\n", fileName) // Depuración

	// Si no hay directorios padres, trabajar desde el inodo raíz
	if len(parentsDir) == 0 {
		return sb.deleteFileInInode(file, 0, fileName)
	}

	// Iterar sobre cada inodo para encontrar el archivo
	for i := int32(0); i < sb.S_inodes_count; i++ {
		// Deserializar el inodo
		inode := &Inode{}
		err := inode.Decode(file, int64(sb.S_inode_start+(i*sb.S_inode_size)))
		if err != nil {
			return fmt.Errorf("error al deserializar inodo %d: %v", i, err)
		}

		// Verificar que sea una carpeta
		if inode.I_type[0] == '0' {
			// Iterar sobre los bloques de la carpeta
			for _, blockIndex := range inode.I_block {
				if blockIndex == -1 {
					break
				}

				block := &FolderBlock{}
				err := block.Decode(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
				if err != nil {
					return fmt.Errorf("error al deserializar bloque %d: %v", blockIndex, err)
				}

				// Buscar la carpeta que contiene el archivo
				for _, content := range block.B_content {
					contentName := strings.Trim(string(content.B_name[:]), "\x00 ")
					if content.B_inodo != -1 && strings.EqualFold(contentName, parentsDir[0]) {
						fmt.Printf("Carpeta '%s' encontrada, buscando archivo '%s' en inodo %d.\n", parentsDir[0], fileName, content.B_inodo)

						// Llamada recursiva para eliminar el archivo dentro de la carpeta
						return sb.deleteFileInInode(file, content.B_inodo, fileName)
					}
				}
			}
		}
	}
	return fmt.Errorf("archivo '%s' no encontrado en ninguna carpeta", fileName)
}
