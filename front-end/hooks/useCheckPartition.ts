import { useState, useEffect } from 'react';

export const useCheckPartition = () => {
  const [partitionStatus, setPartitionStatus] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const checkPartition = async () => {
      try {
        const response = await fetch('http://localhost:3000/check-partition');
        const data = await response.json();

        if (data.status === 'success') {
          setPartitionStatus('success');
        } else {
          setPartitionStatus(data.message); // Mensaje de error del backend
        }
      } catch (error) {
        setPartitionStatus('Error al verificar particiones montadas');
      } finally {
        setLoading(false);
      }
    };

    checkPartition();
  }, []);

  return { partitionStatus, loading };
};
