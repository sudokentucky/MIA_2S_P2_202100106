package structs

import (
	"backend/utils"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"
)

func (sb *Superblock) createFolderInInode(file *os.File, inodeIndex int32, parentsDir []string, destDir string) error {
	// Deserializar el inodo
	inode := &Inode{}
	fmt.Printf("Deserializando inodo %d\n", inodeIndex) // Depuración
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

	// Iterar sobre los bloques del inodo
	for blockIdx := 0; blockIdx < len(inode.I_block); blockIdx++ {
		blockIndex := inode.I_block[blockIdx]

		// Si el bloque no está asignado, asignar uno nuevo
		if blockIndex == -1 {
			fmt.Printf("Inodo %d no tiene un bloque en la posición %d, asignando uno nuevo.\n", inodeIndex, blockIdx) // Depuración
			blockIndex, err = sb.AssignNewBlock(file, inode, blockIdx)
			if err != nil {
				return fmt.Errorf("error al asignar nuevo bloque al inodo %d: %v", inodeIndex, err)
			}

			// Serializar el inodo actualizado con el nuevo bloque
			err = inode.Encode(file, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
			if err != nil {
				return fmt.Errorf("error al serializar el inodo %d después de asignar bloque: %v", inodeIndex, err)
			}
			fmt.Printf("Inodo %d actualizado con el nuevo bloque %d.\n", inodeIndex, blockIndex)
		}

		// Deserializar el bloque existente
		fmt.Printf("Deserializando bloque %d del inodo %d\n", blockIndex, inodeIndex) // Depuración
		block := &FolderBlock{}
		err = block.Decode(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size))) // Calcular la posición del bloque
		if err != nil {
			return fmt.Errorf("error al deserializar bloque %d: %v", blockIndex, err)
		}
		fmt.Printf("Bloque %d del inodo %d deserializado correctamente\n", blockIndex, inodeIndex) // Depuración

		// Iterar sobre el contenido del bloque
		for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
			content := block.B_content[indexContent]

			// Si estamos manejando carpetas padre
			if len(parentsDir) != 0 {
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

				if strings.EqualFold(contentName, parentDirName) {
					fmt.Printf("Carpeta padre '%s' encontrada en inodo %d. Recursión para crear el siguiente directorio.\n", parentDirName, content.B_inodo) // Depuración
					// Llamada recursiva para seguir creando carpetas
					err := sb.createFolderInInode(file, content.B_inodo, utils.RemoveElement(parentsDir, 0), destDir)
					if err != nil {
						return err
					}
					return nil
				}
			} else { // Llegamos al directorio destino (destDir)
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

				sb.UpdateSuperblockAfterInodeAllocation()

				// Crear el bloque de la nueva carpeta
				folderBlock := NewFolderBlock(sb.S_inodes_count, inodeIndex, nil)
				err = folderBlock.Encode(file, int64(sb.S_first_blo))
				if err != nil {
					return fmt.Errorf("error al serializar el bloque del directorio '%s': %v", destDir, err)
				}

				// Actualizar el bitmap de bloques
				err = sb.UpdateBitmapBlock(file, sb.S_blocks_count, true)
				if err != nil {
					return fmt.Errorf("error al actualizar el bitmap de bloques para el directorio '%s': %v", destDir, err)
				}

				sb.UpdateSuperblockAfterBlockAllocation()

				fmt.Printf("Directorio '%s' creado correctamente en inodo %d.\n", destDir, sb.S_inodes_count) // Depuración
				return nil
			}
		}

		// Si el bloque está lleno, asignar uno nuevo
		if block.IsFull() {
			fmt.Printf("El bloque %d del inodo %d está lleno, asignando un nuevo bloque.\n", blockIndex, inodeIndex) // Depuración
			blockIndex, err = sb.AssignNewBlock(file, inode, blockIdx)
			if err != nil {
				return fmt.Errorf("error al asignar nuevo bloque al inodo %d: %v", inodeIndex, err)
			}

			err = inode.Encode(file, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
			if err != nil {
				return fmt.Errorf("error al serializar el inodo %d después de asignar nuevo bloque: %v", inodeIndex, err)
			}
			fmt.Printf("Inodo %d actualizado con nuevo bloque %d.\n", inodeIndex, blockIndex)
		}
	}

	fmt.Printf("No se encontraron bloques disponibles para crear la carpeta '%s' en inodo %d\n", destDir, inodeIndex) // Depuración
	return nil
}

func (sb *Superblock) createFolderInInodeExt3(file *os.File, inodeIndex int32, parentsDir []string, destDir string) error {
	// Deserializar el inodo actual
	inode := &Inode{}
	err := inode.Decode(file, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil {
		return err
	}

	// Verificar si el inodo es de tipo archivo en lugar de carpeta
	if inode.I_type[0] == '1' {
		return nil // Si es un archivo, no hacemos nada
	}

	// Iterar sobre los bloques del inodo
	for _, blockIndex := range inode.I_block {
		// Si el bloque no existe, salir
		if blockIndex == -1 {
			break
		}

		// Crear un bloque de carpeta
		block := &FolderBlock{}

		// Deserializar el bloque
		err := block.Decode(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
		if err != nil {
			return err
		}

		// Iterar sobre el contenido del bloque de carpeta
		for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
			content := block.B_content[indexContent]

			// Si las carpetas padre no están vacías, debemos buscar la carpeta padre más cercana
			if len(parentsDir) != 0 {
				// Si el contenido está vacío, salir
				if content.B_inodo == -1 {
					break
				}

				// Obtener el nombre de la carpeta padre
				parentDir, err := utils.First(parentsDir)
				if err != nil {
					return err
				}

				// Verificar si el contenido coincide con la carpeta padre
				contentName := strings.Trim(string(content.B_name[:]), "\x00 ")
				parentDirName := strings.Trim(parentDir, "\x00 ")

				if strings.EqualFold(contentName, parentDirName) {
					// Llamar recursivamente para procesar la siguiente carpeta en la jerarquía
					err := sb.createFolderInInodeExt3(file, content.B_inodo, utils.RemoveElement(parentsDir, 0), destDir)
					if err != nil {
						return err
					}
					return nil
				}
			} else {
				// Si el apuntador al inodo está ocupado, continuar con el siguiente
				if content.B_inodo != -1 {
					continue
				}

				// Actualizar el contenido del bloque con la nueva carpeta
				copy(content.B_name[:], destDir)
				content.B_inodo, err = sb.FindNextFreeInode(file)
				if err != nil {
					return fmt.Errorf("error al buscar el primer inodo libre: %w", err)
				}

				block.B_content[indexContent] = content

				// Serializar el bloque actualizado
				err = block.Encode(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
				if err != nil {
					return fmt.Errorf("error al serializar el bloque: %w", err)
				}

				// Registrar la creación de la carpeta en el journal
				journal := &Journal{}
				err = journal.SaveJournalEntry(
					file,
					int64(sb.S_block_start+sb.S_block_size), // Calcular el journaling_start
					"mkdir",
					"/"+destDir,
					"", // No hay contenido ya que es una carpeta
				)
				if err != nil {
					return fmt.Errorf("error al guardar la entrada en el journal: %w", err)
				}

				// Encontrar un bloque libre para la nueva carpeta
				folderBlockIndex, err := sb.FindNextFreeBlock(file)
				if err != nil {
					return fmt.Errorf("error al buscar el primer bloque libre para la carpeta: %w", err)
				}

				// Crear el inodo de la nueva carpeta
				folderInode := &Inode{}
				err = folderInode.CreateInode(
					file,
					sb,
					'0', // Tipo de inodo: carpeta
					0,   // Tamaño de la carpeta (inicialmente 0)
					[15]int32{folderBlockIndex, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, // Bloques de la carpeta
					[3]byte{'7', '7', '7'}, // Permisos de la carpeta
				)
				if err != nil {
					return fmt.Errorf("error al crear el inodo de la carpeta: %w", err)
				}

				// Crear el bloque de la carpeta
				folderBlock := &FolderBlock{
					B_content: [4]FolderContent{
						{B_name: [12]byte{'.'}, B_inodo: content.B_inodo}, // Apunta al propio inodo
						{B_name: [12]byte{'.', '.'}, B_inodo: inodeIndex}, // Apunta al inodo padre
						{B_name: [12]byte{'-'}, B_inodo: -1},              // Espacio vacío
						{B_name: [12]byte{'-'}, B_inodo: -1},              // Espacio vacío
					},
				}

				// Serializar el bloque de la nueva carpeta
				err = folderBlock.Encode(file, int64(sb.S_block_start+folderBlockIndex))
				if err != nil {
					return fmt.Errorf("error al serializar el bloque de la carpeta: %w", err)
				}

				// Actualizar el bitmap de bloques para la nueva carpeta
				err = sb.UpdateBitmapBlock(file, folderBlockIndex, true)
				if err != nil {
					return fmt.Errorf("error al actualizar el bitmap de bloques: %w", err)
				}

				// Actualizar el superbloque
				sb.UpdateSuperblockAfterBlockAllocation()
				return nil
			}
		}
	}
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
