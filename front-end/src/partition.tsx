import { useNavigate, useLocation } from "react-router-dom";
import partitionIcon from "../src/assets/partition.svg";
import { usePartitions } from "../hooks/usePartition";
import goBackIcon from "../src/assets/goBack.svg";

const PartitionVisualizer = () => {
  const navigate = useNavigate();
  const location = useLocation();

  const diskPath = location.state?.diskPath;
  const diskName = location.state?.diskName; // Obtener el nombre del disco

  const { partitions, loading, error } = usePartitions(diskPath);

  const goBack = () => {
    navigate("/file-visualizer");
  };

  const handleViewPartitionTree = (partitionName: string) => {
    navigate("/partition-tree", {
      state: { diskPath, partitionName } 
    });
  };

  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-gray-100">
      <div className="w-full max-w-4xl bg-white p-8 rounded-lg shadow-lg">
        <h1 className="text-3xl font-bold mb-6">Particiones del Disco: {diskName}</h1> {/* Mostrar diskName */}

        <div className="mb-4">
          <button
            onClick={goBack}
            className="mb-4 flex items-center text-blue-600 hover:text-blue-800"
          >
            <img src={goBackIcon} alt="Regresar" className="w-4 h-4 mr-2" />
            Regresar a los discos
          </button>
        </div>

        <div className="mt-8">
          <h2 className="text-2xl font-bold mb-4">Particiones</h2>

          {loading && <p>Cargando particiones...</p>}
          {error && <p className="text-red-500">{error}</p>}

          {partitions.length > 0 ? (
            partitions.map((partition, index) => (
              <div
                key={index}
                className="flex items-center mb-4 cursor-pointer"
                onClick={() => handleViewPartitionTree(partition.name)}
              >
                <img src={partitionIcon} alt="Partición" className="w-8 h-8 mr-3" />
                <h3 className="text-xl font-bold">
                  Partición {index + 1}: {partition.name}
                </h3>
              </div>
            ))
          ) : (
            !loading && <p className="text-gray-500">No se encontraron particiones.</p>
          )}
        </div>
      </div>
    </div>
  );
};

export default PartitionVisualizer;
