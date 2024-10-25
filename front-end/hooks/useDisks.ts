// useDisks.tsx o useDisks.js
import { useState } from "react";

export const useDisks = () => {
  const [disks, setDisks] = useState<{ fileName: string }[]>([]); // Solo almacenamos el fileName
  const [loading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const addDisk = (filePath: string) => {
    // Verificar si el disco ya está en la lista para evitar duplicados
    if (disks.some((disk) => disk.fileName === filePath)) {
      setError("El disco ya ha sido agregado.");
      return;
    }

    // Agregar el disco a la lista
    setDisks((prevDisks) => [...prevDisks, { fileName: filePath }]);
  };

  return { disks, loading, error, addDisk };
};
