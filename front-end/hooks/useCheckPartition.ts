import { useState, useEffect } from 'react';
import { usePart } from './usePart'; // Importamos el hook de Zustand

export const useCheckPartition = () => {
  const [partitionStatus, setPartitionStatus] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);
  
  // Accedemos a las funciones `mount` y `unmount` del hook de Zustand
  const { mount, unmount } = usePart();

  useEffect(() => {
    const checkPartition = async () => {
      try {
        const response = await fetch('http://localhost:3000/check-partition');
        const data = await response.json();

        if (data.status === 'success') {
          setPartitionStatus('success');
          mount(); // Si hay éxito, marcamos la partición como montada en Zustand
        } else {
          setPartitionStatus(data.message); // Mensaje de error del backend
          unmount(); // Si falla, desmarcamos la partición como montada
        }
      } catch (error) {
        setPartitionStatus('Error al verificar particiones montadas');
        unmount(); // En caso de error, también desmarcamos
      } finally {
        setLoading(false); // Terminamos la verificación
      }
    };

    checkPartition();
  }, [mount, unmount]); // Aseguramos que las funciones `mount` y `unmount` estén disponibles en el efecto

  return { partitionStatus, loading };
};
