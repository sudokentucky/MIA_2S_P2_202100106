import { useParams, useNavigate } from "react-router-dom";
import partitionIcon from "/public/partition.svg"; 

interface PartitionVisualizerProps {
  disks: any[];
}

const PartitionVisualizer = ({ disks }: PartitionVisualizerProps) => {
  const { diskIndex } = useParams<{ diskIndex: string }>(); // Obtener el índice del disco de la URL
  const navigate = useNavigate(); // Hook para navegar a otra ruta
  const disk = disks[parseInt(diskIndex || "0", 10)]; // Obtener el disco seleccionado

  // Función para regresar a la lista de discos
  const goBack = () => {
    navigate("/"); // Navegar de regreso al FileVisualizer
  };

  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-gray-100">
      <div className="w-full max-w-4xl bg-white p-8 rounded-lg shadow-lg">
        <h1 className="text-3xl font-bold mb-6">Particiones del Disco: {disk?.fileName}</h1>

        <div className="mb-4">
          <button
            onClick={goBack}
            className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline"
          >
            Regresar a los Discos
          </button>
        </div>

        {/* Mostrar particiones del disco */}
        <div className="mt-8">
          <h2 className="text-2xl font-bold mb-4">Particiones</h2>

          {disk?.partitions?.length > 0 ? (
            disk.partitions.map((partition: any, index: number) => (
              <div key={index} className="flex items-center mb-4">
                <img src={partitionIcon} alt="Partición" className="w-8 h-8 mr-3" />
                <h3 className="text-xl font-bold">Partición {index + 1}: {partition.name}</h3>
              </div>
            ))
          ) : (
            <p className="text-gray-500">No se encontraron particiones.</p>
          )}
        </div>
      </div>
    </div>
  );
};

export default PartitionVisualizer;
