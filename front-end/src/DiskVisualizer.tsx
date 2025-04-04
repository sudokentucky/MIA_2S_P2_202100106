import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useDisksStore } from "../hooks/useDiskStore"; // Asegúrate de ajustar la ruta correcta
import diskIcon from "../src/assets/disk.svg";

function DiskVisualizer() {
  const { disks, loading, error, addDisk, addDisksFromFolder } = useDisksStore();
  const [filePath, setFilePath] = useState<string>("");
  const navigate = useNavigate();

  const handleAddDisk = () => {
    if (filePath) {
      addDisk(filePath);
      setFilePath("");
    }
  };

  const handleFolderSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const files = e.target.files;
    if (!files) return;

    // Mapea los archivos para obtener sus rutas relativas.
    const filePaths = Array.from(files).map((file) => file.webkitRelativePath || file.name);

    // Asegúrate de que las rutas sean consistentes al enviarlas a addDisksFromFolder.
    addDisksFromFolder(filePaths);
  };

  const handleViewPartitions = (diskIndex: number) => {
    const selectedDisk = disks[diskIndex];
    navigate(`/partitions/${diskIndex}`, {
      state: {
        diskPath: selectedDisk.filePath,
        diskName: selectedDisk.fileName,
      },
    });
  };

  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-gray-100">
      <div className="w-full max-w-4xl bg-white p-8 rounded-lg shadow-lg">
        <h1 className="text-3xl font-bold mb-6">Visualizador del Sistema de Archivos</h1>

        {error && <div className="text-red-500 mb-4">Error: {error}</div>}

        <div className="mb-4">
          <label className="block text-gray-700 text-sm font-bold mb-2">
            Ingresar la Ruta Completa del Archivo de Disco en el Servidor
          </label>
          <input
            type="text"
            value={filePath}
            onChange={(e) => setFilePath(e.target.value)}
            placeholder="Ej: /home/user/disk.mia"
            className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
          />
        </div>

        <button
          onClick={handleAddDisk}
          className={`bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline ${
            !filePath || loading ? "opacity-50 cursor-not-allowed" : ""
          }`}
          disabled={!filePath || loading}
        >
          {loading ? "Cargando..." : "Agregar Disco"}
        </button>

        <div className="my-4">
          <label className="block text-gray-700 text-sm font-bold mb-2">
            O seleccionar una carpeta completa:
          </label>
          <input
            type="file"
            multiple
            onChange={handleFolderSelect}
            className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
            {...{ webkitdirectory: "true" } as any}
          />
        </div>

        <div className="mt-8">
          <h2 className="text-2xl font-bold mb-4">Discos</h2>

          {disks.length > 0 ? (
            disks.map((disk, index) => (
              <div
                key={index}
                className="mb-4 border border-gray-300 rounded-lg p-4 bg-gray-50 cursor-pointer"
                onClick={() => handleViewPartitions(index)}
              >
                <div className="flex items-center">
                  <img src={diskIcon} alt="Disco" className="w-8 h-8 mr-3" />
                  <div>
                    <h3 className="text-xl font-bold">Disco {index + 1}: {disk.fileName}</h3>
                    <p className="text-gray-600 text-sm">{disk.filePath}</p>
                  </div>
                </div>
              </div>
            ))
          ) : (
            <p className="text-gray-500">No se han agregado discos aún.</p>
          )}
        </div>
      </div>
    </div>
  );
}

export default DiskVisualizer;
