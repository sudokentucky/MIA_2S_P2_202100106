package structs

import (
	"backend/utils" // Asegúrate de ajustar el path del package "utils"
	"fmt"
	"os"
	"time"
)

type Inode struct {
	I_uid   int32     //UID del usuario propietario del archivo
	I_gid   int32     //GID del grupo propietario del archivo
	I_size  int32     //Tamaño del archivo en bytes
	I_atime float32   //Último acceso al archivo
	I_ctime float32   //Último cambio de permisos
	I_mtime float32   //Última modificación del archivo
	I_block [15]int32 // 12 bloques directos, 1 indirecto simple, 1 indirecto doble, 1 indirecto triple
	I_type  [1]byte   //Indica si es archivo o carpeta 1=archivo, 0=carpeta
	I_perm  [3]byte   //Guarda los permisos del archivo
	// Total: 88 bytes
}

func (inode *Inode) Encode(file *os.File, offset int64) error {
	// Utilizamos la función WriteToFile del paquete utils
	err := utils.WriteToFile(file, offset, inode)
	if err != nil {
		return fmt.Errorf("error writing Inode to file: %w", err)
	}
	return nil
}

func (inode *Inode) Decode(file *os.File, offset int64) error {
	// Utilizamos la función ReadFromFile del paquete utils
	err := utils.ReadFromFile(file, offset, inode)
	if err != nil {
		return fmt.Errorf("error reading Inode from file: %w", err)
	}
	return nil
}

// Crear y serializar un inodo, actualizando el bitmap de inodos
func (inode *Inode) CreateInode(
	file *os.File, // Archivo del sistema de archivos
	sb *Superblock, // Superbloque
	inodeType byte, // Tipo de inodo (0 para carpeta, 1 para archivo)
	size int32, // Tamaño del archivo o directorio
	blocks [15]int32, // Bloques asignados al inodo
	permissions [3]byte, // Permisos del inodo
) error {
	// Asignar un nuevo inodo usando AssignNewInode
	inodeIndex, err := sb.AssignNewInode(file)
	if err != nil {
		return fmt.Errorf("error asignando nuevo inodo: %w", err)
	}

	// Establecer los valores del inodo
	inode.I_uid = 1
	inode.I_gid = 1
	inode.I_size = size
	inode.I_atime = float32(time.Now().Unix())
	inode.I_ctime = float32(time.Now().Unix())
	inode.I_mtime = float32(time.Now().Unix())
	inode.I_block = blocks
	inode.I_type = [1]byte{inodeType}
	inode.I_perm = permissions

	// Serializar el inodo en la ubicación correcta
	inodeOffset := int64(sb.S_inode_start + (inodeIndex * sb.S_inode_size))
	err = inode.Encode(file, inodeOffset)
	if err != nil {
		return fmt.Errorf("error serializando el inodo en la ubicación %d: %w", inodeOffset, err)
	}

	return nil
}

func (inode *Inode) UpdateAtime() {
	inode.I_atime = float32(time.Now().Unix())
}

func (inode *Inode) UpdateMtime() {
	inode.I_mtime = float32(time.Now().Unix())
}

func (inode *Inode) UpdateCtime() {
	inode.I_ctime = float32(time.Now().Unix())
}

// Print imprime los atributos del inodo
func (inode *Inode) Print() {
	atime := time.Unix(int64(inode.I_atime), 0)
	ctime := time.Unix(int64(inode.I_ctime), 0)
	mtime := time.Unix(int64(inode.I_mtime), 0)

	fmt.Printf("I_uid: %d\n", inode.I_uid)
	fmt.Printf("I_gid: %d\n", inode.I_gid)
	fmt.Printf("I_size: %d\n", inode.I_size)
	fmt.Printf("I_atime: %s\n", atime.Format(time.RFC3339))
	fmt.Printf("I_ctime: %s\n", ctime.Format(time.RFC3339))
	fmt.Printf("I_mtime: %s\n", mtime.Format(time.RFC3339))
	fmt.Printf("I_block: %v\n", inode.I_block)
	fmt.Printf("I_type: %s\n", string(inode.I_type[:]))
	fmt.Printf("I_perm: %s\n", string(inode.I_perm[:]))
}
