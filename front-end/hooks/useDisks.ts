import { useState, useEffect } from "react";

export const useDisks = () => {
  const [disks, setDisks] = useState<{ fileName: string; filePath: string }[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Cargar los discos desde localStorage al iniciar el hook
  useEffect(() => {
    const storedDisks = localStorage.getItem("disks");
    if (storedDisks) {
      setDisks(JSON.parse(storedDisks));
    }
  }, []); // Esto debería cargar los discos almacenados cada vez que `useDisks` se use.
  
  // Guardar los discos en localStorage cada vez que cambian
  useEffect(() => {
    localStorage.setItem("disks", JSON.stringify(disks));
  }, [disks]);

  const addDisk = (filePath: string) => {
    const fileName = filePath.split("/").pop() || filePath;
    if (disks.some((disk) => disk.filePath === filePath)) {
      setError("El disco ya ha sido agregado.");
      return;
    }
    setDisks((prevDisks) => [...prevDisks, { fileName, filePath }]);
  };

  const addDisksFromFolder = (filePaths: string[]) => {
    setLoading(true);
    const newDisks = filePaths
      .filter((filePath) => !disks.some((disk) => disk.filePath === filePath))
      .map((filePath) => {
        const fileName = filePath.split("/").pop() || filePath;
        return { fileName, filePath };
      });
    setDisks((prevDisks) => [...prevDisks, ...newDisks]);
    setLoading(false);
  };

  return { disks, loading, error, addDisk, addDisksFromFolder };
};
