import { useState } from "react";

export const useDisks = () => {
  const [disks, setDisks] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const readDisk = async (filePath: string) => {
    setLoading(true);
    setError(null);

    try {
      const response = await fetch("http://localhost:3000/api/disk", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          path: filePath,  // Enviar la ruta absoluta del archivo
        }),
      });

      if (!response.ok) {
        throw new Error("Error al leer el disco");
      }

      const data = await response.json();
      
      // Formatear los discos y particiones correctamente
      setDisks((prevDisks) => [
        ...prevDisks,
        { fileName: filePath, partitions: data.partitions }, // Añadir nuevo disco con particiones
      ]);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return { disks, loading, error, readDisk };
};
