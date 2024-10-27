// useDisks.tsx o useDisks.js
import { useState } from "react";

export const useDisks = () => {
  const [disks, setDisks] = useState<{ fileName: string; filePath: string }[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

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
