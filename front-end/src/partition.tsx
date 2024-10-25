import { useNavigate, useLocation } from "react-router-dom";
import partitionIcon from "/public/partition.svg";
import { usePartitions } from "../hooks/usePartition";

const PartitionVisualizer = () => {
  const navigate = useNavigate();
  const location = useLocation();

  const diskPath = location.state?.diskPath;

  const { partitions, loading, error } = usePartitions(diskPath);

  const goBack = () => {
    navigate("/");
  };

  const handleViewPartitionTree = (partitionName: string) => {
    navigate("/partition-tree", { state: { diskPath, partitionName } }); // Pasamos diskPath y partitionName
  };

  return (
    <div className="min-h-screen flex flex-col items-center justify-center bg-gray-100">
      <div className="w-full max-w-4xl bg-white p-8 rounded-lg shadow-lg">
        <h1 className="text-3xl font-bold mb-6">Particiones del Disco: {diskPath}</h1>

        <div className="mb-4">
          <button
            onClick={goBack}
            className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline"
          >
            Regresar a los Discos
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
                onClick={() => handleViewPartitionTree(partition.name)} // Manejamos el clic en la partición
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
