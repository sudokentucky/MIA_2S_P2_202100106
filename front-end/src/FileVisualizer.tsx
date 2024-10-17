import { useState } from "react";
import { useDisks } from "../hooks/useDisks"; // Importamos el hook personalizado

function FileVisualizer() {
  const { disks, loading, error, readDisks } = useDisks(); // Usamos el hook personalizado
  const [selectedFiles, setSelectedFiles] = useState<FileList | null>(null); // Estado para los archivos seleccionados

  // Función para manejar la selección de archivos
  const handleFileSelect = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSelectedFiles(e.target.files);
  };

  // Función para leer las particiones del disco seleccionado
  const handleReadDisks = () => {
    if (selectedFiles) {
      readDisks(selectedFiles);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gray-100">
      <div className="w-full max-w-4xl bg-white p-8 rounded-lg shadow-lg">
        <h1 className="text-3xl font-bold mb-6">Visualizador del Sistema de Archivos</h1>

        {/* Mostrar error si ocurre */}
        {error && <div className="text-red-500 mb-4">Error: {error}</div>}

        {/* Selector de archivos */}
        <div className="mb-4">
          <label className="block text-gray-700 text-sm font-bold mb-2">Seleccionar Archivos de Disco</label>
          <input
            type="file"
            multiple
            onChange={handleFileSelect}
            className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
          />
        </div>

        <button
          onClick={handleReadDisks}
          className={`bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline ${
            !selectedFiles || loading ? "opacity-50 cursor-not-allowed" : ""
          }`}
          disabled={!selectedFiles || loading}
        >
          {loading ? "Cargando..." : "Leer Discos"}
        </button>

        {/* Mostrar la información de los discos seleccionados */}
        {disks.length > 0 && (
          <div className="mt-8">
            <h2 className="text-2xl font-bold mb-4">Información de los Discos</h2>
            {disks.map((disk, index) => (
              <div key={index} className="mb-4">
                <h3 className="text-xl font-bold">Disco: {disk.fileName}</h3>
                <pre>{JSON.stringify(disk, null, 2)}</pre>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
}

export default FileVisualizer;
