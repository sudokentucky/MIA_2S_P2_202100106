import { useState } from "react";
import { useNavigate } from "react-router-dom"; // Importar el hook de navegación
import { useDisks } from "../hooks/useDisks";
import diskIcon from "/public/disk.svg"; // Asegúrate de que la ruta sea correcta para tu proyecto

function FileVisualizer() {
  const { disks, loading, error, readDisk } = useDisks(); // Hook personalizado para leer discos
  const [filePath, setFilePath] = useState<string>(""); // Estado para almacenar la ruta completa del archivo
  const navigate = useNavigate(); // Hook para navegar a otra ruta

  // Función para leer las particiones del disco en la ruta ingresada
  const handleReadDisk = () => {
    if (filePath) {
      readDisk(filePath); // Pasamos la ruta completa del archivo al hook
      setFilePath(""); // Limpiar el campo de entrada después de enviar
    }
  };

  // Función para navegar al visualizador de particiones
  const handleViewPartitions = (diskIndex: number) => {
    // Navegar a la ruta del visualizador de particiones con el índice del disco
    navigate(`/partitions/${diskIndex}`);
  };

  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-gray-100">
      <div className="w-full max-w-4xl bg-white p-8 rounded-lg shadow-lg">
        <h1 className="text-3xl font-bold mb-6">Visualizador del Sistema de Archivos</h1>

        {/* Mostrar error si ocurre */}
        {error && <div className="text-red-500 mb-4">Error: {error}</div>}

        {/* Input para ingresar la ruta completa del archivo en el servidor */}
        <div className="mb-4">
          <label className="block text-gray-700 text-sm font-bold mb-2">
            Ingresar la Ruta Completa del Archivo de Disco en el Servidor
          </label>
          <input
            type="text"
            value={filePath}
            onChange={(e) => setFilePath(e.target.value)} // Actualizar el estado con la ruta ingresada
            placeholder="Ej: /ruta/completa/al/archivo.img"
            className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
          />
        </div>

        <button
          onClick={handleReadDisk}
          className={`bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline ${
            !filePath || loading ? "opacity-50 cursor-not-allowed" : ""
          }`}
          disabled={!filePath || loading} // Deshabilitar si no hay ruta o si está cargando
        >
          {loading ? "Cargando..." : "Leer Disco"}
        </button>

        {/* Mostrar la información de los discos seleccionados */}
        <div className="mt-8">
          <h2 className="text-2xl font-bold mb-4">Discos</h2>

          {/* Iterar sobre los discos y mostrar un panel por cada uno */}
          {disks.length > 0 ? (
            disks.map((disk, index) => (
              <div
                key={index}
                className="mb-4 border border-gray-300 rounded-lg p-4 bg-gray-50 cursor-pointer"
                onClick={() => handleViewPartitions(index)} // Navegar a la vista de particiones
              >
                <div className="flex items-center">
                  <img src={diskIcon} alt="Disco" className="w-8 h-8 mr-3" />
                  <h3 className="text-xl font-bold">Disco {index + 1}: {disk.fileName}</h3>
                </div>
              </div>
            ))
          ) : (
            <p className="text-gray-500">No se han cargado discos aún.</p>
          )}
        </div>
      </div>
    </div>
  );
}

export default FileVisualizer;
