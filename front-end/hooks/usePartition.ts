import { useState, useEffect } from "react";

interface Partition {
  name: string;
  size: number;
  type: string;
  fit: string;
  start: number;
  status: string;
  // Agrega más campos si es necesario
}

export const usePartitions = (diskPath: string) => {
  const [partitions, setPartitions] = useState<Partition[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (diskPath) {
      fetchPartitions(diskPath);
    }
  }, [diskPath]);

  const fetchPartitions = async (path: string) => {
    setLoading(true);
    setError(null);

    try {
      const apiUrl = import.meta.env.VITE_API_URL;
      const response = await fetch(`${apiUrl}/api/disk/partitions`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ path }),
      });

      if (!response.ok) {
        throw new Error("Error al obtener las particiones del disco.");
      }

      const data = await response.json();
      setPartitions(data.partitions);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return { partitions, loading, error };
};
