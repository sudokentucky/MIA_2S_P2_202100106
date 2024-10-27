import React from "react";
import folderIcon from "../src/assets/folder-flat.svg"; // Importa el icono de carpeta

interface FolderProps {
  name: string;
  onClick?: () => void;  // Añadir la propiedad onClick como opcional
}

const Folder: React.FC<FolderProps> = ({ name, onClick }) => {
  return (
    <div className="flex flex-col items-center cursor-pointer" onClick={onClick}>
      <img src={folderIcon} alt="Folder" className="w-16 h-16" />
      <p className="mt-2 text-center">{name}</p>
    </div>
  );
};

export default Folder;
