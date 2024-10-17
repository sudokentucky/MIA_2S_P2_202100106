import { useState } from "react";

export const useDisks = () => {
  const [disks, setDisks] = useState<any[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const readDisks = async (files: FileList) => {
    setLoading(true);
    setError(null);

    const paths = Array.from(files).map((file) => file.name);

    try {
      const response = await fetch("http://localhost:3000/disk/read", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          paths: paths,
          isEncrypted: false,
          key: 0, // Cambia si usas encriptación
        }),
      });

      if (!response.ok) {
        throw new Error("Error al leer los discos");
      }

      const data = await response.json();
      setDisks(data.disks);
    } catch (err: any) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return { disks, loading, error, readDisks };
};
