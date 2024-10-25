package structs

import (
	"backend/utils"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"
)

// createFolderInInode crea una carpeta en un inodo específico
func (sb *Superblock) createFolderInInode(file *os.File, inodeIndex int32, parentsDir []string, destDir string) error {
	// Crear un nuevo inodo
	inode := &Inode{}
	fmt.Printf("Deserializando inodo %d\n", inodeIndex) // Depuración

	// Deserializar el inodo
	err := inode.Decode(file, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil {
		return fmt.Errorf("error al deserializar inodo %d: %v", inodeIndex, err)
	}
	fmt.Printf("Inodo %d deserializado. Tipo: %c\n", inodeIndex, inode.I_type[0]) // Depuración

	// Verificar si el inodo es de tipo carpeta
	if inode.I_type[0] != '0' {
		fmt.Printf("Inodo %d no es una carpeta, es de tipo: %c\n", inodeIndex, inode.I_type[0]) // Depuración
		return nil
	}

	// Iterar sobre cada bloque del inodo (apuntadores)
	for _, blockIndex := range inode.I_block {
		// Si el bloque no existe, salir
		if blockIndex == -1 {
			fmt.Printf("Inodo %d no tiene más bloques asignados, terminando la búsqueda.\n", inodeIndex) // Depuración
			break
		}

		fmt.Printf("Deserializando bloque %d del inodo %d\n", blockIndex, inodeIndex) // Depuración
		// Crear un nuevo bloque de carpeta
		block := &FolderBlock{}

		// Deserializar el bloque
		err := block.Decode(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size))) // Calcular la posición del bloque
		if err != nil {
			return fmt.Errorf("error al deserializar bloque %d: %v", blockIndex, err)
		}
		fmt.Printf("Bloque %d del inodo %d deserializado correctamente\n", blockIndex, inodeIndex) // Depuración

		// Iterar sobre cada contenido del bloque, desde el índice 2 (evitamos . y ..)
		for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
			content := block.B_content[indexContent]
			fmt.Printf("Verificando contenido en índice %d del bloque %d\n", indexContent, blockIndex) // Depuración

			// Si hay más carpetas padres en la ruta
			if len(parentsDir) != 0 {
				// Si el contenido está vacío, salir
				if content.B_inodo == -1 {
					fmt.Printf("No se encontró carpeta padre en inodo %d en la posición %d, terminando.\n", inodeIndex, indexContent) // Depuración
					break
				}

				// Obtener la carpeta padre más cercana
				parentDir, err := utils.First(parentsDir)
				if err != nil {
					return err
				}

				contentName := strings.Trim(string(content.B_name[:]), "\x00 ")
				parentDirName := strings.Trim(parentDir, "\x00 ")
				fmt.Printf("Comparando '%s' con el nombre de la carpeta padre '%s'\n", contentName, parentDirName) // Depuración

				// Si el nombre del contenido coincide con el nombre de la carpeta padre
				if strings.EqualFold(contentName, parentDirName) {
					fmt.Printf("Carpeta padre '%s' encontrada en inodo %d. Recursion para crear el siguiente directorio.\n", parentDirName, content.B_inodo) // Depuración
					// Llamada recursiva para seguir creando carpetas
					err := sb.createFolderInInode(file, content.B_inodo, utils.RemoveElement(parentsDir, 0), destDir)
					if err != nil {
						return err
					}
					return nil
				}
			} else { // Cuando llegamos al directorio destino (destDir)
				if content.B_inodo != -1 {
					fmt.Printf("El inodo %d ya está ocupado con otro contenido, saltando al siguiente.\n", content.B_inodo) // Depuración
					continue
				}

				fmt.Printf("Asignando el nombre del directorio '%s' al bloque en la posición %d\n", destDir, indexContent) // Depuración
				// Actualizar el contenido del bloque con el nuevo directorio
				copy(content.B_name[:], destDir)
				content.B_inodo = sb.S_inodes_count

				// Actualizar el bloque con el nuevo contenido
				block.B_content[indexContent] = content

				// Serializar el bloque
				err = block.Encode(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
				if err != nil {
					return fmt.Errorf("error al serializar el bloque %d: %v", blockIndex, err)
				}
				fmt.Printf("Bloque %d actualizado con éxito.\n", blockIndex) // Depuración

				// Verificación adicional para realizar el journaling si el sistema de archivos lo requiere
				if sb.S_filesystem_type == 3 { // Aquí usamos el tipo de sistema de archivos (ext3)
					FolderJournal := &Journal{
						J_count: sb.S_inodes_count,
					}
					var superblockSize int64 = int64(binary.Size(sb))
					journaling_start := superblockSize
					err = FolderJournal.SaveJournalEntry(
						file,
						journaling_start, // Iniciar el journaling después del superbloque
						"mkdir",          // Tipo de operación: crear carpeta
						"/"+destDir,      // Ruta del directorio
						"",               // No hay contenido asociado en este caso
					)
					if err != nil {
						return fmt.Errorf("error al guardar la entrada en el journal: %w", err)
					}

					//Codificar el journal
					err = FolderJournal.Encode(file, journaling_start)
					if err != nil {
						return fmt.Errorf("error al serializar el journal: %w", err)
					}

					fmt.Printf("Entrada de journal creada para la carpeta '%s'\n", destDir) // Depuración
					FolderJournal.Print()                                                   // Depuración
				}

				// Crear el inodo de la nueva carpeta
				folderInode := &Inode{
					I_uid:   1,
					I_gid:   1,
					I_size:  0,
					I_atime: float32(time.Now().Unix()),
					I_ctime: float32(time.Now().Unix()),
					I_mtime: float32(time.Now().Unix()),
					I_block: [15]int32{sb.S_blocks_count, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
					I_type:  [1]byte{'0'}, // Tipo carpeta
					I_perm:  [3]byte{'6', '6', '4'},
				}

				fmt.Printf("Serializando el inodo de la carpeta '%s' (inodo %d)\n", destDir, sb.S_inodes_count) // Depuración
				// Serializar el inodo de la nueva carpeta
				err = folderInode.Encode(file, int64(sb.S_first_ino))
				if err != nil {
					return fmt.Errorf("error al serializar el inodo del directorio '%s': %v", destDir, err)
				}

				// Actualizar el bitmap de inodos
				err = sb.UpdateBitmapInode(file, sb.S_inodes_count, true)
				if err != nil {
					return fmt.Errorf("error al actualizar el bitmap de inodos para el directorio '%s': %v", destDir, err)
				}

				// Actualizar el superbloque con los nuevos valores de inodos
				sb.UpdateSuperblockAfterInodeAllocation()

				// Crear el bloque para la nueva carpeta
				folderBlock := &FolderBlock{
					B_content: [4]FolderContent{
						{B_name: [12]byte{'.'}, B_inodo: content.B_inodo},
						{B_name: [12]byte{'.', '.'}, B_inodo: inodeIndex},
						{B_name: [12]byte{'-'}, B_inodo: -1},
						{B_name: [12]byte{'-'}, B_inodo: -1},
					},
				}

				fmt.Printf("Serializando el bloque de la carpeta '%s'\n", destDir) // Depuración
				// Serializar el bloque de la carpeta
				err = folderBlock.Encode(file, int64(sb.S_first_blo))
				if err != nil {
					return fmt.Errorf("error al serializar el bloque del directorio '%s': %v", destDir, err)
				}

				// Actualizar el bitmap de bloques
				err = sb.UpdateBitmapBlock(file, sb.S_blocks_count, true)
				if err != nil {
					return fmt.Errorf("error al actualizar el bitmap de bloques para el directorio '%s': %v", destDir, err)
				}

				// Actualizar el superbloque con los nuevos valores de bloques
				sb.UpdateSuperblockAfterBlockAllocation()

				fmt.Printf("Directorio '%s' creado correctamente en inodo %d.\n", destDir, sb.S_inodes_count) // Depuración
				return nil
			}
		}
	}

	fmt.Printf("No se encontraron bloques disponibles para crear la carpeta '%s' en inodo %d\n", destDir, inodeIndex) // Depuración
	return nil
}

// CreateFolder crea una carpeta en el sistema de archivos
func (sb *Superblock) CreateFolder(file *os.File, parentsDir []string, destDir string) error {
	// Si parentsDir está vacío, solo trabajar con el primer inodo que sería el raíz "/"
	if len(parentsDir) == 0 {
		return sb.createFolderInInode(file, 0, parentsDir, destDir)
	}

	// Iterar sobre cada inodo ya que se necesita buscar el inodo padre
	for i := int32(0); i < sb.S_inodes_count; i++ { //Desde el inodo 0
		err := sb.createFolderInInode(file, i, parentsDir, destDir)
		if err != nil {
			return err
		}
	}

	return nil
}

// CreateFolderRecursively crea carpetas recursivamente asegurando que cada directorio intermedio existe.
func (sb *Superblock) CreateFolderRecursively(file *os.File, path string) error {
	// Dividir el path en carpetas
	dirs := strings.Split(strings.Trim(path, "/"), "/")

	if len(dirs) == 0 {
		return fmt.Errorf("ruta inválida: %s", path)
	}

	// Llamar a la función recursiva empezando desde el inodo raíz
	return sb.createFolderRecursivelyInInode(file, 0, dirs)
}

// createFolderRecursivelyInInode se asegura de que cada carpeta en la lista exista o se cree
func (sb *Superblock) createFolderRecursivelyInInode(file *os.File, inodeIndex int32, dirs []string) error {
	if len(dirs) == 0 {
		return nil // No hay más carpetas que crear
	}

	currentDir := dirs[0]
	remainingDirs := dirs[1:]

	// Usar la función `createFolderInInode` para buscar o crear el directorio actual
	err := sb.createFolderInInode(file, inodeIndex, nil, currentDir)
	if err != nil {
		return err
	}

	// Después de crear el directorio actual, pasar al siguiente nivel recursivamente
	return sb.createFolderRecursivelyInInode(file, sb.S_inodes_count-1, remainingDirs)
}

// deleteFolderInInode elimina recursivamente el contenido de una carpeta en un inodo específico
func (sb *Superblock) deleteFolderInInode(file *os.File, inodeIndex int32) error {
	// Deserializar el inodo
	inode := &Inode{}
	err := inode.Decode(file, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil {
		return fmt.Errorf("error al deserializar inodo %d: %v", inodeIndex, err)
	}

	// Verificar que el inodo sea una carpeta
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

		// Eliminar los contenidos del bloque (recursivamente si son directorios)
		for _, content := range block.B_content {
			if content.B_inodo != -1 && string(content.B_name[:]) != "." && string(content.B_name[:]) != ".." {
				fmt.Printf("Eliminando contenido '%s' en inodo %d\n", content.B_name, content.B_inodo)

				// Deserializar el inodo para verificar si es archivo o carpeta
				childInode := &Inode{}
				err = childInode.Decode(file, int64(sb.S_inode_start+(content.B_inodo*sb.S_inode_size)))
				if err != nil {
					return fmt.Errorf("error al deserializar inodo hijo %d: %v", content.B_inodo, err)
				}

				if childInode.I_type[0] == '0' { // Si es carpeta
					// Llamada recursiva para eliminar la carpeta
					err = sb.deleteFolderInInode(file, content.B_inodo)
					if err != nil {
						return err
					}
				} else { // Si es archivo
					// Liberar los bloques del archivo
					for _, fileBlockIndex := range childInode.I_block {
						if fileBlockIndex != -1 {
							err = sb.FreeBlock(file, fileBlockIndex)
							if err != nil {
								return err
							}
						}
					}

					// Liberar el inodo del archivo
					err = sb.FreeInode(file, content.B_inodo)
					if err != nil {
						return err
					}
				}
			}
		}

		// Liberar el bloque del directorio
		err = sb.FreeBlock(file, blockIndex)
		if err != nil {
			return err
		}
	}

	// Finalmente, liberar el inodo de la carpeta
	err = sb.FreeInode(file, inodeIndex)
	if err != nil {
		return err
	}

	fmt.Printf("Carpeta en inodo %d eliminada correctamente.\n", inodeIndex)
	return nil
}

// DeleteFolder elimina un directorio y su contenido recursivamente en el sistema de archivos
func (sb *Superblock) DeleteFolder(file *os.File, parentsDir []string, folderName string) error {
	// Iterar sobre cada inodo para encontrar la carpeta
	for i := int32(0); i < sb.S_inodes_count; i++ {
		// Deserializar el inodo
		inode := &Inode{}
		err := inode.Decode(file, int64(sb.S_inode_start+(i*sb.S_inode_size)))
		if err != nil {
			return fmt.Errorf("error al deserializar inodo %d: %v", i, err)
		}

		// Verificar si es una carpeta
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

				// Buscar la carpeta a eliminar
				for _, content := range block.B_content {
					contentName := strings.Trim(string(content.B_name[:]), "\x00 ")
					if content.B_inodo != -1 && strings.EqualFold(contentName, folderName) {
						fmt.Printf("Carpeta '%s' encontrada, eliminando contenido recursivamente.\n", folderName)
						// Llamar a la función recursiva para eliminar el contenido de la carpeta
						err = sb.deleteFolderInInode(file, content.B_inodo)
						if err != nil {
							return err
						}

						// Eliminar la referencia a la carpeta en el bloque actual
						content.B_inodo = -1
						copy(content.B_name[:], "")

						// Actualizar el bloque después de eliminar la referencia
						err = block.Encode(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
						if err != nil {
							return fmt.Errorf("error al serializar el bloque %d después de eliminar la carpeta: %v", blockIndex, err)
						}

						return nil
					}
				}
			}
		}
	}

	return fmt.Errorf("carpeta '%s' no encontrada", folderName)
}
