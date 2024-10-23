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
	inode := &structs.Inode{}
	err := inode.Decode(dts.file, int64(dts.partitionSuperblock.S_inode_start+(inodeIndex*dts.partitionSuperblock.S_inode_size)))
	if err != nil {
		return nil, fmt.Errorf("error al deserializar el inodo %d: %v", inodeIndex, err)
	}

	tree := &DirectoryTree{
		Name:  currentPath,
		IsDir: inode.I_type[0] == '0',
	}

	// Si no es un directorio, devolverlo directamente
	if !tree.IsDir {
		return tree, nil
	}

	// Iterar sobre los bloques del inodo del directorio
	for _, blockIndex := range inode.I_block {
		if blockIndex == -1 {
			break // No hay más bloques asignados
		}

		block := &structs.FolderBlock{}
		err := block.Decode(dts.file, int64(dts.partitionSuperblock.S_block_start+(blockIndex*dts.partitionSuperblock.S_block_size)))
		if err != nil {
			return nil, fmt.Errorf("error al deserializar el bloque %d: %v", blockIndex, err)
		}

		for _, content := range block.B_content {
			if content.B_inodo == -1 {
				continue
			}

			contentName := strings.Trim(string(content.B_name[:]), "\x00 ")
			if contentName == "." || contentName == ".." {
				continue
			}

			childPath := currentPath + "/" + contentName
			childNode, err := dts.buildDirectoryTree(content.B_inodo, childPath)
			if err != nil {
				return nil, err
			}

			tree.Children = append(tree.Children, childNode)
		}
	}

	return tree, nil
}
