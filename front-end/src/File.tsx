import React from "react";
import fileIcon from "../src/assets/file-flat.svg"; // Importa el icono de archivo

interface FileProps {
  name: string;
  content?: string;
  onDoubleClick?: () => void; // Añadir soporte para el evento onDoubleClick
}

const File: React.FC<FileProps> = ({ name, onDoubleClick }) => {
  return (
    <div className="flex flex-col items-center cursor-pointer" onDoubleClick={onDoubleClick}>
      <img src={fileIcon} alt="File" className="w-16 h-16" />
      <p className="mt-2 text-center">{name}</p>
    </div>
  );
};

export default File;
