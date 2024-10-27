import { create } from "zustand";

interface Disk {
  fileName: string;
  filePath: string;
}

interface DiskState {
  disks: Disk[];
  loading: boolean;
  error: string | null;
  addDisk: (filePath: string) => void;
  addDisksFromFolder: (filePaths: string[]) => void;
  setLoading: (isLoading: boolean) => void;
  clearDisks: () => void; // Agregamos la función para limpiar discos
}

export const useDisksStore = create<DiskState>((set, get) => ({
  disks: JSON.parse(localStorage.getItem("disks") || "[]"),
  loading: false,
  error: null,
  
  addDisk: (filePath) => {
    const { disks } = get();
    const fileName = filePath.split("/").pop() || filePath;
    if (disks.some((disk) => disk.filePath === filePath)) {
      set({ error: "El disco ya ha sido agregado." });
      return;
    }

    const newDisks = [...disks, { fileName, filePath }];
    set({ disks: newDisks, error: null });
    localStorage.setItem("disks", JSON.stringify(newDisks));
  },

  addDisksFromFolder: (filePaths) => {
    const { disks } = get();
    set({ loading: true });

    const newDisks = filePaths
      .filter((filePath) => !disks.some((disk) => disk.filePath === filePath))
      .map((filePath) => {
        const fileName = filePath.split("/").pop() || filePath;
        return { fileName, filePath };
      });

    const updatedDisks = [...disks, ...newDisks];
    set({ disks: updatedDisks, loading: false });
    localStorage.setItem("disks", JSON.stringify(updatedDisks));
  },

  setLoading: (isLoading) => set({ loading: isLoading }),

  clearDisks: () => {
    set({ disks: [], error: null });
    localStorage.removeItem("disks");
  },
}));
