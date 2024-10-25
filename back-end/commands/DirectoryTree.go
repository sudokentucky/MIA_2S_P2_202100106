package commands

import (
	structs "backend/Structs"
	"backend/globals"
	"backend/utils"
	"fmt"
	"os"
	"strings"
)

// DirectoryTree representa el árbol de directorios
type DirectoryTree struct {
	Name     string           `json:"name"`
	Children []*DirectoryTree `json:"children,omitempty"`
	IsDir    bool             `json:"isDir"`
}

// DirectoryTreeService maneja la construcción y obtención del árbol de directorios
type DirectoryTreeService struct {
	partitionSuperblock *structs.Superblock
	partitionPath       string
	file                *os.File
}

// NewDirectoryTreeService inicializa un nuevo servicio de árbol de directorios
func NewDirectoryTreeService() (*DirectoryTreeService, error) {
	if !globals.IsLoggedIn() {
		return nil, fmt.Errorf("no hay un usuario logueado")
	}

	// Obtener la partición montada para el usuario logueado
	idPartition := globals.UsuarioActual.Id
	partitionSuperblock, _, partitionPath, err := globals.GetMountedPartitionSuperblock(idPartition)
	if err != nil {
		return nil, fmt.Errorf("error al obtener la partición montada: %w", err)
	}

	// Abrir el archivo de partición
	file, err := os.OpenFile(partitionPath, os.O_RDWR, 0666)
	if err != nil {
		return nil, fmt.Errorf("error al abrir el archivo de partición: %w", err)
	}

	return &DirectoryTreeService{
		partitionSuperblock: partitionSuperblock,
		partitionPath:       partitionPath,
		file:                file,
	}, nil
}

// Close se encarga de cerrar el archivo cuando ya no es necesario
func (dts *DirectoryTreeService) Close() {
	dts.file.Close()
}

// GetDirectoryTree obtiene el árbol de directorios desde un path dado
func (dts *DirectoryTreeService) GetDirectoryTree(path string) (*DirectoryTree, error) {
	// Verificar si el path es la raíz "/"
	var rootInodeIndex int32
	var err error

	if path == "/" {
		rootInodeIndex = 0 // El inodo raíz es el inodo 0
	} else {
		parentDirs, dirName := utils.GetParentDirectories(path)
		rootInodeIndex, err = findFileInode(dts.file, dts.partitionSuperblock, parentDirs, dirName)
		if err != nil {
			return nil, fmt.Errorf("error al encontrar el directorio inicial: %v", err)
		}
	}

	// Construir el árbol de directorios
	tree, err := dts.buildDirectoryTree(rootInodeIndex, path)
	if err != nil {
		return nil, err
	}

	return tree, nil
}

// buildDirectoryTree es la función recursiva que construye el árbol de directorios
func (dts *DirectoryTreeService) buildDirectoryTree(inodeIndex int32, currentPath string) (*DirectoryTree, error) {
	// Decodificar el inodo correspondiente al índice actual
	inode := &structs.Inode{}
	err := inode.Decode(dts.file, int64(dts.partitionSuperblock.S_inode_start+(inodeIndex*dts.partitionSuperblock.S_inode_size)))
	if err != nil {
		return nil, fmt.Errorf("error al deserializar el inodo %d: %v", inodeIndex, err)
	}

	// Asignar el nombre correctamente para la raíz
	var currentName string
	if currentPath == "/" {
		currentName = "/" // Si estamos en el root, mostrar como "/"
	} else {
		// Extraer el nombre del archivo/carpeta del currentPath
		pathSegments := strings.Split(strings.Trim(currentPath, "/"), "/")
		currentName = pathSegments[len(pathSegments)-1] // Tomar el último segmento como nombre
	}

	// Crear un nodo para el árbol de directorios, con el nombre actual y el tipo (directorio o archivo)
	tree := &DirectoryTree{
		Name:  currentName,            // Ahora usamos "/" para la raíz, y los nombres para los demás
		IsDir: inode.I_type[0] == '0', // Asumimos que '0' indica un directorio
	}

	// Si no es un directorio, devolver el nodo sin hijos
	if !tree.IsDir {
		return tree, nil
	}

	// Iterar sobre los bloques del inodo del directorio para procesar los archivos y subdirectorios
	for _, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			break // No hay más bloques asignados
		}

		// Decodificar el bloque de la carpeta
		block := &structs.FolderBlock{}
		err := block.Decode(dts.file, int64(dts.partitionSuperblock.S_block_start+(blockIndex*dts.partitionSuperblock.S_block_size)))
		if err != nil {
			return nil, fmt.Errorf("error al deserializar el bloque %d: %v", blockIndex, err)
		}

		// Iterar sobre los contenidos del bloque (entradas de directorios)
		for _, content := range block.B_content {
			if content.B_inodo == -1 {
				continue // No hay más entradas válidas
			}

			// Limpiar el nombre del archivo/carpeta y filtrar '.' y '..'
			contentName := strings.Trim(string(content.B_name[:]), "\x00 ")
			if contentName == "." || contentName == ".." {
				continue // Ignorar las entradas especiales '.' y '..'
			}

			// Crear la ruta completa del hijo (archivo o subdirectorio)
			childPath := currentPath + "/" + contentName

			// Llamada recursiva para construir el árbol del hijo
			childNode, err := dts.buildDirectoryTree(content.B_inodo, childPath)
			if err != nil {
				return nil, err
			}

			// Agregar el nodo hijo al directorio actual
			tree.Children = append(tree.Children, childNode)
		}
	}

	return tree, nil
}
