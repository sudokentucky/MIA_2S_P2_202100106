import React, { useState } from "react";
import { useLocation, useNavigate } from "react-router-dom";
import Folder from "./Folder";
import File from "./File";
import { usePartitionTree } from "../hooks/usePartitionTree";
import goBackIcon from "../src/assets/goBack.svg"; // Importa el icono de retroceso

const FileSystemTree: React.FC = () => {
  const location = useLocation();
  const navigate = useNavigate();

  // Obtener el diskPath y partitionName del estado de la navegación
  const { diskPath, partitionName } = location.state || {};

  // Estado para el path actual
  const [currentPath, setCurrentPath] = useState("/");

  // Estado para el contenido del archivo
  const [fileContent, setFileContent] = useState<string | null>(null);

  // Usar el hook personalizado para obtener el árbol de archivos
  const { tree, loading, error } = usePartitionTree(diskPath, partitionName);

  // Si no hay diskPath o partitionName, redirigir al usuario
  if (!diskPath || !partitionName) {
    navigate("/partitions");
    return null;
  }

  // Obtener los segmentos del path actual
  const pathSegments = currentPath.split("/").filter(Boolean);

  // Función para regresar a la vista de particiones
  const handleReturnToPartitions = () => {
    navigate("/file-visualizer"); // Cambia a la ruta de particiones
  };

  // Función para navegar dentro de una carpeta
  const handleFolderClick = (folderName: string) => {
    setCurrentPath((prevPath) =>
      prevPath === "/" ? `/${folderName}` : `${prevPath}/${folderName}`
    );
    setFileContent(null);
  };

  // Función para obtener el contenido del archivo al hacer doble clic
  const handleFileDoubleClick = async (fileName: string) => {
    try {
      const command = `cat -file1="${currentPath}/${fileName}"`;
      const apiUrl = import.meta.env.VITE_API_URL;
      const response = await fetch(`${apiUrl}/analyze`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ command }),
      });

      if (!response.ok) {
        throw new Error("Error al obtener el contenido del archivo.");
      }

      const data = await response.json();
      setFileContent(data.results.join("\n"));
    } catch (err) {
      console.error(err);
      setFileContent("Error al obtener el contenido del archivo.");
    }
  };

  // Función para volver a un directorio anterior al hacer clic en el breadcrumb
  const handleBreadcrumbClick = (index: number) => {
    const newPath = "/" + pathSegments.slice(0, index + 1).join("/");
    setCurrentPath(newPath);
    setFileContent(null);
  };

  // Obtener el directorio actual en base al currentPath
  const findCurrentDirectory = (treeNode: any, path: string[]): any => {
    if (path.length === 0 || !treeNode) return treeNode;
    const [currentSegment, ...restPath] = path;
    const found = treeNode.children?.find((child: any) => child.name === currentSegment);
    return restPath.length > 0 ? findCurrentDirectory(found, restPath) : found;
  };
  const currentDirectory = tree ? findCurrentDirectory(tree, pathSegments) : undefined;

  return (
    <div className="flex-grow flex flex-col items-center justify-center p-6 bg-gray-100">
      <div className="w-full max-w-5xl bg-white rounded-lg shadow-lg p-8">
        {/* Botón de Regresar a Particiones */}
        <button
          onClick={handleReturnToPartitions}
          className="mb-4 flex items-center text-blue-600 hover:text-blue-800"
        >
          <img src={goBackIcon} alt="Regresar" className="w-4 h-4 mr-2" />
          Regresar a los discos
        </button>

        {/* Título Principal */}
        <h1 className="text-4xl font-bold mb-6 text-center text-gray-900">
          Explorador de Archivos <span className="text-blue-600">Ext2 | Ext3</span>
        </h1>

        {/* Subtítulo con el nombre de la partición */}
        <h2 className="text-2xl font-semibold mb-6 text-center text-gray-700">
          Partición: <span className="text-blue-600">{partitionName}</span>
        </h2>

        {/* Breadcrumb */}
        <nav
          className="flex justify-center px-5 py-3 text-gray-700 border border-gray-200 rounded-lg bg-gray-50"
          aria-label="Breadcrumb"
        >
          <ol className="inline-flex items-center space-x-1 md:space-x-2">
            <li className="inline-flex items-center">
              <span
                onClick={() => {
                  setCurrentPath("/");
                  setFileContent(null);
                }}
                className="inline-flex items-center text-sm font-medium text-gray-700 hover:text-blue-600 cursor-pointer"
              >
                <svg
                  className="w-3 h-3 mr-2.5"
                  fill="currentColor"
                  xmlns="http://www.w3.org/2000/svg"
                  viewBox="0 0 20 20"
                >
                  <path d="M19.707 9.293l-2-2-7-7a1 1 0 00-1.414 0l-7 7-2 2a1 1 0 001.414 1.414L2 10.414V18a2 2 0 002 2h3a1 1 0 001-1v-4a1 1 0 011-1h2a1 1 0 011 1v4a1 1 0 001 1h3a2 2 0 002-2v-7.586l.293.293a1 1 0 001.414-1.414z" />
                </svg>
                /
              </span>
            </li>

            {pathSegments.map((segment, index) => (
              <li key={index} className="flex items-center">
                <svg
                  className="w-3 h-3 text-gray-400 mx-1"
                  fill="none"
                  xmlns="http://www.w3.org/2000/svg"
                  viewBox="0 0 6 10"
                >
                  <path
                    stroke="currentColor"
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth="2"
                    d="M1 9l4-4-4-4"
                  />
                </svg>
                <span
                  onClick={() => handleBreadcrumbClick(index)}
                  className="text-sm font-medium text-gray-700 hover:text-blue-600 cursor-pointer"
                >
                  {segment}
                </span>
              </li>
            ))}
          </ol>
        </nav>

        {/* Mensajes de carga y error */}
        {loading && <p className="text-gray-500">Cargando archivos...</p>}
        {error && <p className="text-red-500">Error: {error}</p>}

        {/* Mostrar el contenido del archivo en un textarea si se selecciona un archivo */}
        {fileContent ? (
          <textarea
            className="w-full h-64 p-3 border rounded-md bg-gray-100"
            readOnly
            value={fileContent}
          />
        ) : (
          <div className="grid grid-cols-2 sm:grid-cols-3 gap-4 mt-4">
            {/* Mostrar los archivos y carpetas del directorio actual */}
            {currentDirectory && currentDirectory.children ? (
              currentDirectory.children.map((item: any, index: number) =>
                item.type === "folder" ? (
                  <Folder
                    key={index}
                    name={item.name}
                    onClick={() => handleFolderClick(item.name)}
                  />
                ) : (
                  <File
                    key={index}
                    name={item.name}
                    onDoubleClick={() => handleFileDoubleClick(item.name)}
                  />
                )
              )
            ) : (
              <p className="text-center text-gray-600 col-span-3">No hay archivos ni carpetas.</p>
            )}
          </div>
        )}
      </div>
    </div>
  );
};

export default FileSystemTree;
