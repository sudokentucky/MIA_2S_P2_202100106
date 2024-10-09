package structs

import (
	"backend/utils"
	"fmt"
	"os"
	"strings"
	"time"
)

// createFolderInInode crea una carpeta en un inodo específico
func (sb *Superblock) createFolderInInode(file *os.File, inodeIndex int32, parentsDir []string, destDir string) error {
	inode, err := sb.loadInode(file, inodeIndex)
	if err != nil {
		return fmt.Errorf("error al deserializar inodo %d: %v", inodeIndex, err)
	}
	//Se verifica si el inodo es de tipo archivo
	if inode.I_type[0] != '0' {
		return fmt.Errorf("inodo %d no es una carpeta", inodeIndex)
	}
	//se recorre los bloques del inodo
	for _, blockIndex := range inode.I_block {
		if blockIndex == -1 { //si el bloque esta
			break
		}
		//se carga el bloque de carpeta
		block, err := sb.loadFolderBlock(file, blockIndex)
		if err != nil {
			return fmt.Errorf("error al deserializar bloque %d: %v", blockIndex, err)
		}

		for indexContent := 2; indexContent < len(block.B_content); indexContent++ {
			content := block.B_content[indexContent]

			if len(parentsDir) > 0 {
				if content.B_inodo == -1 {
					break
				}

				parentDir, err := utils.First(parentsDir)
				if err != nil {
					return err
				}

				contentName := strings.Trim(string(content.B_name[:]), "\x00 ")
				if strings.EqualFold(contentName, parentDir) {
					return sb.createFolderInInode(file, content.B_inodo, utils.RemoveElement(parentsDir, 0), destDir)
				}
			} else {
				if content.B_inodo == -1 {
					newInode, err := sb.AssignNewInode(file)
					if err != nil {
						return fmt.Errorf("error asignando nuevo inodo para '%s': %v", destDir, err)
					}

					copy(content.B_name[:], destDir)
					content.B_inodo = newInode
					block.B_content[indexContent] = content

					err = sb.saveFolderBlock(file, blockIndex, block)
					if err != nil {
						return fmt.Errorf("error al guardar el bloque %d: %v", blockIndex, err)
					}

					err = sb.createFolderStructure(file, newInode, inodeIndex)
					if err != nil {
						return fmt.Errorf("error al crear estructura de carpeta '%s': %v", destDir, err)
					}

					return nil
				}
			}
		}
	}

	return fmt.Errorf("no se encontraron bloques disponibles para la carpeta '%s'", destDir)
}

// loadInode deserializa un inodo desde un archivo
func (sb *Superblock) loadInode(file *os.File, inodeIndex int32) (*Inode, error) {
	inode := &Inode{}
	err := inode.Decode(file, sb.CalculateInodeOffset(inodeIndex))
	if err != nil {
		return nil, err
	}
	return inode, nil
}

// loadFolderBlock deserializa un bloque de carpeta desde un archivo
func (sb *Superblock) loadFolderBlock(file *os.File, blockIndex int32) (*FolderBlock, error) {
	block := &FolderBlock{}
	err := block.Decode(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
	if err != nil {
		return nil, err
	}
	return block, nil
}

// saveFolderBlock guarda un bloque de carpeta en el archivo
func (sb *Superblock) saveFolderBlock(file *os.File, blockIndex int32, block *FolderBlock) error {
	return block.Encode(file, int64(sb.S_block_start+(blockIndex*sb.S_block_size)))
}

// createFolderStructure crea la estructura básica para un nuevo directorio
func (sb *Superblock) createFolderStructure(file *os.File, newInode int32, parentInodeIndex int32) error {
	inode := &Inode{
		I_uid:   1,
		I_gid:   1,
		I_size:  0,
		I_atime: float32(time.Now().Unix()),
		I_ctime: float32(time.Now().Unix()),
		I_mtime: float32(time.Now().Unix()),
		I_block: [15]int32{-1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1, -1},
		I_type:  [1]byte{'0'},
		I_perm:  [3]byte{'6', '6', '4'},
	}

	blockIndex, err := sb.AssignNewBlock(file, inode, 0)
	if err != nil {
		return fmt.Errorf("error al asignar nuevo bloque: %v", err)
	}

	folderBlock := &FolderBlock{
		B_content: [4]FolderContent{
			{B_name: [12]byte{'.'}, B_inodo: newInode},
			{B_name: [12]byte{'.', '.'}, B_inodo: parentInodeIndex},
			{B_name: [12]byte{'-'}, B_inodo: -1},
			{B_name: [12]byte{'-'}, B_inodo: -1},
		},
	}

	err = sb.saveFolderBlock(file, blockIndex, folderBlock)
	if err != nil {
		return fmt.Errorf("error al serializar el bloque: %v", err)
	}

	err = inode.Encode(file, sb.CalculateInodeOffset(newInode))
	if err != nil {
		return fmt.Errorf("error al serializar el inodo: %v", err)
	}

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
