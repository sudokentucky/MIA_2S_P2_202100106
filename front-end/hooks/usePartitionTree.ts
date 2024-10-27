import { useState, useEffect } from "react";

// Definimos la interfaz para representar cada archivo o carpeta
interface FileOrFolder {
  name: string;       // Nombre del archivo o carpeta
  type: "file" | "folder";  // Tipo: "file" (archivo) o "folder" (carpeta)
  children?: FileOrFolder[]; // Lista de hijos si es una carpeta
}

// Definimos la interfaz que refleja la estructura del árbol recibido desde el backend
interface TreeNode {
  name: string;         // Nombre del archivo o carpeta
  isDir: boolean;       // Si es una carpeta (true) o un archivo (false)
  children?: TreeNode[];  // Lista de hijos si es un directorio
}

// Hook personalizado para obtener el árbol de archivos y carpetas de una partición
export const usePartitionTree = (diskPath: string, partitionName: string) => {
  const [tree, setTree] = useState<FileOrFolder | null>(null);  // Mantener el árbol anidado como un solo objeto
  const [loading, setLoading] = useState(false);         // Estado de carga
  const [error, setError] = useState<string | null>(null); // Estado de error en caso de que ocurra

  // Efecto que se ejecuta cada vez que cambian el diskPath o partitionName
  useEffect(() => {
    if (diskPath && partitionName) {
      fetchPartitionTree(diskPath, partitionName); // Llama a la función para obtener el árbol
    }
  }, [diskPath, partitionName]);

  /**
   * Función que obtiene el árbol de archivos/carpetas del backend basado en el diskPath y partitionName.
   * @param {string} diskPath - Ruta del disco en el sistema.
   * @param {string} partitionName - Nombre de la partición que se desea explorar.
   */
  const fetchPartitionTree = async (diskPath: string, partitionName: string) => {
    setLoading(true);
    setError(null);

    try {
      // Solicitud al backend para obtener el árbol de archivos/carpetas
      const apiUrl = import.meta.env.VITE_API_URL;
      const response = await fetch(`${apiUrl}/api/disk/partition/tree`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ diskPath, partitionName }), // Enviar los datos necesarios en el cuerpo de la solicitud
      });

      if (!response.ok) {
        throw new Error("Error al obtener el árbol de archivos de la partición.");
      }

      // Convertir la respuesta a JSON
      const data = await response.json();

      // Convertir el árbol anidado recibido en la estructura correcta
      const formattedTree = formatTree(data.tree);
      setTree(formattedTree); // Actualizamos el estado con el árbol anidado
    } catch (err: any) {
      setError(err.message); // Guardamos el mensaje de error en caso de fallo
    } finally {
      setLoading(false); // Finalizamos el estado de carga
    }
  };

  /**
   * Función recursiva para convertir el árbol anidado en la estructura correcta de FileOrFolder.
   * @param {TreeNode} node - Nodo actual del árbol.
   * @returns {FileOrFolder} - Nodo de archivo o carpeta formateado.
   */
  const formatTree = (node: TreeNode): FileOrFolder => {
    const formattedNode: FileOrFolder = {
      name: node.name,
      type: node.isDir ? "folder" : "file",
      children: node.isDir ? node.children?.map(formatTree) : undefined,
    };

    return formattedNode; // Retornar el nodo formateado (con hijos si es directorio)
  };

  return { tree, loading, error }; // Devolver el árbol anidado, la carga y el error
};
