package commands

import (
	structs "backend/Structs"
	"backend/globals"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// DiskManager maneja el acceso a discos, MBR y particiones
type DiskManager struct {
	disks         map[string]*os.File     // Mapa de discos abiertos (ruta -> archivo)
	PartitionMBRs map[string]*structs.MBR // Mapa de MBRs leídos (disco -> MBR)
}

// NewDiskManager crea un nuevo gestor de discos
func NewDiskManager() *DiskManager {
	return &DiskManager{
		disks:         make(map[string]*os.File),
		PartitionMBRs: make(map[string]*structs.MBR),
	}
}

// LoadDisk abre un archivo de disco binario y carga su MBR
func (dm *DiskManager) LoadDisk(diskPath string) error {
	file, err := os.OpenFile(diskPath, os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("error al abrir el disco: %w", err)
	}

	// Leer el MBR del disco
	mbr := &structs.MBR{}
	err = mbr.Decode(file)
	if err != nil {
		file.Close() // Si falla, cerramos el archivo
		return fmt.Errorf("error al leer el MBR del disco: %w", err)
	}

	// Guardar el archivo y el MBR en las estructuras del DiskManager
	dm.disks[diskPath] = file
	dm.PartitionMBRs[diskPath] = mbr
	fmt.Printf("Disco '%s' cargado exitosamente.\n", diskPath)
	return nil
}

// CloseDisk cierra un disco abierto
func (dm *DiskManager) CloseDisk(diskPath string) error {
	if file, exists := dm.disks[diskPath]; exists {
		file.Close()
		delete(dm.disks, diskPath)
		delete(dm.PartitionMBRs, diskPath)
		fmt.Printf("Disco '%s' cerrado exitosamente.\n", diskPath)
		return nil
	}
	return fmt.Errorf("disco no encontrado: %s", diskPath)
}

// MountPartition monta una partición por nombre en un disco
func (dm *DiskManager) MountPartition(diskPath string, partitionName string) (*structs.Partition, error) {
	mbr, exists := dm.PartitionMBRs[diskPath]
	if !exists {
		return nil, fmt.Errorf("MBR no encontrado para el disco '%s'", diskPath)
	}

	// Buscar la partición por nombre
	partition, index := mbr.GetPartitionByName(partitionName)
	if partition == nil {
		return nil, fmt.Errorf("partición '%s' no encontrada en el disco '%s'", partitionName, diskPath)
	}

	// Verificar si la partición ya está montada
	if id, isMounted := dm.CheckPartitionMounted(diskPath, partitionName); isMounted {
		return partition, fmt.Errorf("error: la partición '%s' ya está montada con ID: %s", partitionName, id)
	}

	// Generar ID único para la partición
	idPartition := fmt.Sprintf("%s%d", partitionName, index+1)

	// Marcar partición como montada y actualizar el mapa global
	partition.MountPartition(index, idPartition)
	globals.MountedPartitions[idPartition] = diskPath

	// Guardar los cambios en el MBR de vuelta en el disco
	file := dm.disks[diskPath]
	if err := mbr.Encode(file); err != nil {
		return nil, fmt.Errorf("error serializando el MBR de vuelta al disco: %v", err)
	}

	fmt.Printf("Partición '%s' montada correctamente con ID: %s\n", partitionName, idPartition)
	return partition, nil
}

// CheckPartitionMounted verifica si la partición ya está montada
func (dm *DiskManager) CheckPartitionMounted(diskPath string, partitionName string) (string, bool) {
	for id, mountedPath := range globals.MountedPartitions {
		if mountedPath == diskPath && strings.Contains(id, partitionName) {
			return id, true
		}
	}
	return "", false
}

// GetMountedPartitionSuperblock accede a una partición ya montada y devuelve el superblock y demás detalles
func (dm *DiskManager) GetMountedPartitionSuperblock(id string) (*structs.Superblock, *structs.Partition, string, error) {
	return globals.GetMountedPartitionSuperblock(id) // Reutilizar la función de globals
}

// GetMountedPartition obtiene la partición montada por su ID
func (dm *DiskManager) GetMountedPartition(id string) (*structs.Partition, string, error) {
	return globals.GetMountedPartition(id) // Reutilizar la función de globals
}

// PrintPartitionTree imprime el árbol de directorios de una partición montada
func (dm *DiskManager) PrintPartitionTree(diskPath string, partitionName string, outputBuffer *bytes.Buffer) error {
	tree, err := dm.GetPartitionTree(diskPath, partitionName)
	if err != nil {
		return fmt.Errorf("error obteniendo el árbol de directorios: %v", err)
	}

	// Imprimir el árbol en formato JSON
	treeJSON, err := json.MarshalIndent(tree, "", "  ")
	if err != nil {
		return fmt.Errorf("error al serializar el árbol de directorios a JSON: %v", err)
	}

	outputBuffer.WriteString(string(treeJSON))
	return nil
}

// GetPartitionTree genera el árbol de ficheros de una partición
func (dm *DiskManager) GetPartitionTree(diskPath string, partitionName string) (*DirectoryTree, error) {
	_, exists := dm.disks[diskPath]
	if !exists {
		return nil, fmt.Errorf("disco '%s' no está cargado", diskPath)
	}

	// Obtener la partición
	partition, err := dm.MountPartition(diskPath, partitionName)
	if err != nil {
		return nil, err
	}

	// Usar el `DirectoryTreeService` para construir el árbol del sistema de archivos
	treeService, err := NewDirectoryTreeService()
	if err != nil {
		return nil, fmt.Errorf("error inicializando el servicio de árbol de directorios: %v", err)
	}
	defer treeService.Close()

	// Obtener el árbol de directorios desde el inicio de la partición
	tree, err := treeService.GetDirectoryTree(fmt.Sprintf("/partition/%s", partition.Part_name))
	if err != nil {
		return nil, fmt.Errorf("error obteniendo el árbol de directorios: %v", err)
	}

	return tree, nil
}
