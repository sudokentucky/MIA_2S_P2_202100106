package structs

import (
	"backend/utils"
	"fmt" // Importamos fmt para los mensajes de depuración
	"os"
	"strings"
	"time"
)

func (sb *Superblock) createFileInInode(
	file *os.File,
	inodeIndex int32,
	parentsDir []string,
	destFile string,
	fileSize int,
	fileContent []string) error {

	fmt.Printf("Intentando crear archivo '%s' en inodo index %d\n", destFile, inodeIndex)

	// Deserializar el inodo
	inode, err := sb.loadInode(file, inodeIndex)
	if err != nil {
		return fmt.Errorf("error al deserializar inodo %d: %v", inodeIndex, err)
	}

	// Verificar si el inodo es de tipo carpeta
	if inode.I_type[0] == '1' {
		fmt.Printf("El inodo %d es una carpeta, omitiendo.\n", inodeIndex)
		return nil
	}

	// Buscar la carpeta padre si hay directorios intermedios
	if len(parentsDir) > 0 {
		for _, blockIndex := range inode.I_block {
			if blockIndex == -1 {
				break
			}

			block, err := sb.loadFolderBlock(file, blockIndex)
			if err != nil {
				return fmt.Errorf("error al deserializar bloque %d: %v", blockIndex, err)
			}

			for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
				content := block.B_content[indexContent]
				if content.B_inodo == -1 {
					continue
				}

				parentDir, err := utils.First(parentsDir)
				if err != nil {
					return err
				}

				contentName := strings.Trim(string(content.B_name[:]), "\x00 ")
				if strings.EqualFold(contentName, parentDir) {
					return sb.createFileInInode(file, content.B_inodo, utils.RemoveElement(parentsDir, 0), destFile, fileSize, fileContent)
				}
			}
		}
	}

	// Si estamos en el directorio destino, buscar un espacio disponible para el archivo
	for _, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			continue
		}

		block, err := sb.loadFolderBlock(file, blockIndex)
		if err != nil {
			return fmt.Errorf("error al deserializar bloque %d: %v", blockIndex, err)
		}

		for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
			content := block.B_content[indexContent]

			// Si encontramos un espacio libre, asignar un nuevo inodo y guardar el archivo
			if content.B_inodo == -1 {
				return sb.createAndStoreFile(file, inodeIndex, block, indexContent, destFile, fileSize, fileContent)
			}
		}
	}

	// Si no hay espacio en los bloques actuales, asignar un nuevo bloque al inodo
	newBlockIndex, err := sb.AssignNewBlock(file, inode, 0)
	if err != nil {
		return fmt.Errorf("error asignando nuevo bloque para el archivo '%s': %v", destFile, err)
	}

	// Crear un nuevo bloque de carpeta y almacenar el archivo
	block := &FolderBlock{}
	err = sb.createAndStoreFile(file, inodeIndex, block, 2, destFile, fileSize, fileContent)
	if err != nil {
		return fmt.Errorf("error al crear y almacenar el archivo '%s': %v", destFile, err)
	}

	// Guardar el nuevo bloque
	err = block.Encode(file, int64(sb.S_block_start+(newBlockIndex*sb.S_block_size)))
	if err != nil {
		return fmt.Errorf("error al guardar el nuevo bloque %d: %v", newBlockIndex, err)
	}

	return nil
}

// createAndStoreFile asigna un nuevo inodo para un archivo y guarda su contenido
func (sb *Superblock) createAndStoreFile(
	file *os.File,
	parentInodeIndex int32,
	block *FolderBlock,
	indexContent int,
	destFile string,
	fileSize int,
	fileContent []string) error {

	// Asignar un nuevo inodo para el archivo
	newInodeIndex, err := sb.AssignNewInode(file)
	if err != nil {
		return fmt.Errorf("error asignando nuevo inodo para '%s': %v", destFile, err)
	}

	// Actualizar el contenido del bloque con el nombre del archivo y el inodo
	copy(block.B_content[indexContent].B_name[:], []byte(destFile))
	block.B_content[indexContent].B_inodo = newInodeIndex

	// Crear el inodo del archivo
	fileInode := &Inode{
		I_uid:   1,
		I_gid:   1,
		I_size:  int32(fileSize),
		I_atime: float32(time.Now().Unix()),
		I_ctime: float32(time.Now().Unix()),
		I_mtime: float32(time.Now().Unix()),
		I_block: [15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		I_type:  [1]byte{'0'}, // '0' indica que es un archivo
		I_perm:  [3]byte{'6', '6', '4'},
	}

	// Dividir el contenido en bloques utilizando SplitContent
	blocks, err := SplitContent(strings.Join(fileContent, ""))
	if err != nil {
		return fmt.Errorf("error al dividir el contenido del archivo: %v", err)
	}

	// Asignar y serializar los bloques de contenido
	for i, fb := range blocks {
		blockIndex, err := sb.AssignNewBlock(file, fileInode, i)
		if err != nil {
			return fmt.Errorf("error asignando bloque para archivo: %v", err)
		}

		// Serializar el bloque
		err = fb.Encode(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
		if err != nil {
			return fmt.Errorf("error al serializar bloque de archivo: %v", err)
		}

		fmt.Printf("Bloque de archivo '%s' serializado correctamente en el bloque %d.\n", destFile, blockIndex)
	}

	// Serializar el inodo del archivo
	err = fileInode.Encode(file, sb.CalculateInodeOffset(newInodeIndex))
	if err != nil {
		return fmt.Errorf("error al serializar inodo del archivo: %v", err)
	}

	fmt.Printf("Inodo del archivo '%s' serializado correctamente en %d.\n", destFile, newInodeIndex)
	return nil
}

// CreateFolderinExt3 crea una carpeta en el sistema de archivos
func (sb *Superblock) createFileInInodeExt3(file *os.File, inodeIndex int32, parentsDir []string, destFile string, fileSize int, fileContent []string) error {
	// Deserializar el inodo actual
	inode := &Inode{}
	err := inode.Decode(file, int64(sb.S_inode_start+(inodeIndex*sb.S_inode_size)))
	if err != nil {
		return err
	}

	// Verificar si el inodo es de tipo carpeta
	if inode.I_type[0] == '1' {
		return nil
	}

	// Iterar sobre cada bloque del inodo (apuntadores)
	for _, blockIndex := range inode.I_block {
		// Si el bloque no existe, salir
		if blockIndex == -1 {
			break
		}

		// Crear un nuevo bloque de carpeta
		block := &FolderBlock{}

		// Deserializar el bloque
		err := block.Decode(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size))) // 64 porque es el tamaño de un bloque
		if err != nil {
			return err
		}

		// Iterar sobre cada contenido del bloque, desde el index 2 porque los primeros dos son . y ..
		for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
			// Obtener el contenido del bloque
			content := block.B_content[indexContent]

			// Si las carpetas padre no están vacías, debemos buscar la carpeta padre más cercana
			if len(parentsDir) != 0 {
				// Si el contenido está vacío, salir
				if content.B_inodo == -1 {
					break
				}

				// Obtenemos la carpeta padre más cercana
				parentDir, err := utils.First(parentsDir)
				if err != nil {
					return err
				}

				// Convertir B_name a string y eliminar los caracteres nulos
				contentName := strings.Trim(string(content.B_name[:]), "\x00 ")
				parentDirName := strings.Trim(parentDir, "\x00 ")

				// Si el nombre del contenido coincide con el nombre de la carpeta padre
				if strings.EqualFold(contentName, parentDirName) {
					// Llamar recursivamente para procesar la siguiente carpeta en la jerarquía
					err := sb.createFileInInodeExt3(file, content.B_inodo, utils.RemoveElement(parentsDir, 0), destFile, fileSize, fileContent)
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

				// Actualizar el contenido del bloque con el nuevo archivo
				copy(content.B_name[:], []byte(destFile))
				content.B_inodo, err = sb.FindNextFreeInode(file)
				if err != nil {
					return fmt.Errorf("error al buscar el primer inodo libre: %w", err)
				}

				// Actualizar el bloque con la nueva entrada
				block.B_content[indexContent] = content

				// Serializar el bloque actualizado
				err = block.Encode(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
				if err != nil {
					return fmt.Errorf("error al serializar el bloque: %w", err)
				}

				// Registrar la creación del archivo en el journal
				journal := &Journal{}

				// Calcular el journaling_start como el inicio del superbloque más su tamaño
				journaling_start := int64(sb.S_inode_start) - int64(sb.S_inode_size)

				err = journal.SaveJournalEntry(
					file,
					journaling_start,              // Iniciar el journaling después del superbloque
					"mkfile",                      // Tipo de operación: crear archivo
					"/"+destFile,                  // Ruta del archivo
					strings.Join(fileContent, ""), // El contenido del archivo
				)
				if err != nil {
					return fmt.Errorf("error al guardar la entrada en el journal: %w", err)
				}

				// Crear el inodo del archivo
				fileInode := &Inode{}
				err = fileInode.CreateInode(
					file,
					sb,
					'1',             // Tipo de inodo: archivo
					int32(fileSize), // Tamaño del archivo
					[15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1}, // Inicializar los bloques del archivo
					[3]byte{'7', '7', '7'}, // Permisos del archivo
				)
				if err != nil {
					return fmt.Errorf("error al crear el inodo del archivo: %w", err)
				}

				// Crear los bloques del archivo y asignar los bloques correspondientes
				for i, content := range fileContent {
					// Encontrar un bloque libre para el archivo
					fileBlockIndex, err := sb.FindNextFreeBlock(file)
					if err != nil {
						return fmt.Errorf("error al buscar el primer bloque libre: %w", err)
					}

					// Actualizar el inodo del archivo
					fileInode.I_block[i] = fileBlockIndex

					// Crear el bloque del archivo
					fileBlock := &FileBlock{
						B_content: [64]byte{},
					}
					copy(fileBlock.B_content[:], content)

					// Serializar el bloque del archivo
					err = fileBlock.Encode(file, int64(sb.S_block_start+fileBlockIndex))
					if err != nil {
						return fmt.Errorf("error al serializar el bloque del archivo: %w", err)
					}

					// Actualizar el bitmap de bloques
					err = sb.UpdateBitmapBlock(file, fileBlockIndex, true)
					if err != nil {
						return fmt.Errorf("error al actualizar el bitmap de bloques: %w", err)
					}

					// Actualizar el superbloque
					sb.UpdateSuperblockAfterBlockAllocation()
				}

				// Serializar el inodo del archivo
				err = fileInode.Encode(file, int64(sb.S_inode_start+content.B_inodo*sb.S_inode_size))
				if err != nil {
					return fmt.Errorf("error al serializar el inodo del archivo: %w", err)
				}

				// Actualizar el bitmap de inodos
				err = sb.UpdateBitmapInode(file, content.B_inodo, true)
				if err != nil {
					return fmt.Errorf("error al actualizar el bitmap de inodos: %w", err)
				}

				// Actualizar el superbloque después de asignar el inodo
				sb.UpdateSuperblockAfterInodeAllocation()

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
