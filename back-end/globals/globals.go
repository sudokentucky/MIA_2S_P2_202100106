package globals

import (
	structures "backend/Structs"
	"errors"
	"os"
)

// Mi carnet
const Carnet string = "06" // 202100106
var (
	// UsuarioActual guarda la información del usuario logueado actualmente
	UsuarioActual     *structures.User  = nil
	MountedPartitions map[string]string = make(map[string]string)
)

// GetMountedPartitionSuperblock obtiene el SuperBlock de la partición montada con el id especificado
func GetMountedPartitionSuperblock(id string) (*structures.Superblock, *structures.Partition, string, error) {
	// Obtener el path de la partición montada
	path := MountedPartitions[id]
	if path == "" {
		return nil, nil, "", errors.New("la partición no está montada")
	}
	//Abrir el archivo para leer el MBR
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, "", err
	}
	// Crear una instancia de MBR
	var mbr structures.MBR

	// Deserializar la estructura MBR desde un archivo binario
	err = mbr.Decode(file)
	if err != nil {
		return nil, nil, "", err
	}

	// Buscar la partición con el id especificado
	partition, err := mbr.GetPartitionByID(id)
	if partition == nil {
		return nil, nil, "", err
	}

	// Crear una instancia de SuperBlock
	var sb structures.Superblock

	// Deserializar la estructura SuperBlock desde un archivo binario
	err = sb.Decode(file, int64(partition.Part_start))
	if err != nil {
		return nil, nil, "", err
	}

	return &sb, partition, path, nil
}

// GetMountedPartition obtiene la partición montada con el id especificado
func GetMountedPartition(id string) (*structures.Partition, string, error) {
	// Obtener el path de la partición montada
	path := MountedPartitions[id]
	if path == "" {
		return nil, "", errors.New("la partición no está montada")
	}

	// Abrir el archivo para leer el MBR
	file, err := os.Open(path)
	if err != nil {
		return nil, "", err
	}
	defer file.Close()

	// Crear una instancia de MBR
	var mbr structures.MBR

	// Deserializar la estructura MBR desde un archivo binario
	err = mbr.Decode(file)
	if err != nil {
		return nil, "", err
	}

	// Buscar la partición con el id especificado
	partition, err := mbr.GetPartitionByID(id)
	if partition == nil {
		return nil, "", err
	}

	return partition, path, nil
}

// GetMountedPartitionRep obtiene el MBR y el SuperBlock de la partición montada con el id especificado
func GetMountedPartitionRep(id string) (*structures.MBR, *structures.Superblock, string, error) {
	// Obtener el path de la partición montada
	path := MountedPartitions[id]
	if path == "" {
		return nil, nil, "", errors.New("la partición no está montada")
	}

	// Abrir el archivo para leer el MBR
	file, err := os.Open(path)
	if err != nil {
		return nil, nil, "", err
	}
	defer file.Close()

	// Crear una instancia de MBR
	var mbr structures.MBR

	// Deserializar la estructura MBR desde el archivo
	err = mbr.Decode(file)
	if err != nil {
		return nil, nil, "", err
	}

	// Buscar la partición con el id especificado
	partition, err := mbr.GetPartitionByID(id)
	if err != nil {
		return nil, nil, "", err
	}

	// Crear una instancia de SuperBlock
	var sb structures.Superblock

	// Deserializar la estructura SuperBlock desde el archivo
	err = sb.Decode(file, int64(partition.Part_start))
	if err != nil {
		return nil, nil, "", err
	}

	return &mbr, &sb, path, nil
}

// IsLoggedIn verifica si hay un usuario logueado actualmente
func IsLoggedIn() bool {
	return UsuarioActual != nil && UsuarioActual.Status
}

// Logout cierra la sesión del usuario actual
func Logout() {
	if UsuarioActual != nil {
		UsuarioActual.Status = false // Cambiar el estado del usuario a no logueado
		UsuarioActual = nil          // Limpiar la información del usuario
	}
}
