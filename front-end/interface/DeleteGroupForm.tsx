import React, { useState } from "react";
import useCommandExecution from "../hooks/useCommandExecution";

const DeleteGroupForm = ({ refresh, setMessage, setMessageType }: { refresh: () => void; setMessage: (msg: string) => void; setMessageType: (type: "success" | "error") => void; }) => {
  const [deleteGroupName, setDeleteGroupName] = useState("");
  const { setInputText, handleExecute, loading } = useCommandExecution();

  const handleDeleteGroupSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!deleteGroupName.trim()) {
      alert("Por favor, ingrese el nombre del grupo que desea eliminar.");
      return;
    }

    const command = `rmgrp -name=${deleteGroupName}`;
    setInputText(command);
    try {
      await handleExecute();   // Ejecutar el comando
      setMessage("Grupo eliminado exitosamente");  // Mensaje de éxito
      setMessageType("success");
      refresh();               // Refrescar la lista después de la eliminación
    } catch (error) {
      setMessage("Error al eliminar el grupo");    // Mensaje de error
      setMessageType("error");
    }
  };

  return (
    <form onSubmit={handleDeleteGroupSubmit} className="bg-nosferatu-800 p-6 rounded-lg shadow-lg transition-all hover:shadow-xl duration-300">
      <div className="mb-4">
        <label htmlFor="deleteGroupName" className="block text-dracula-300 text-sm font-bold mb-2">
          Nombre del Grupo a Eliminar
        </label>
        <input
          type="text"
          id="deleteGroupName"
          value={deleteGroupName}
          onChange={(e) => setDeleteGroupName(e.target.value)}
          className="shadow appearance-none border border-gray-600 bg-gray-700 rounded w-full py-2 px-3 text-gray-100 leading-tight focus:outline-none focus:ring-2 focus:ring-red-500"
          placeholder="Ingrese el nombre del grupo a eliminar"
          disabled={loading}
        />
      </div>
      <button
        type="submit"
        className={`bg-red-500 hover:bg-red-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:ring-2 focus:ring-red-500 transition-all ${
          loading ? "opacity-50 cursor-not-allowed" : ""
        }`}
        disabled={loading}
      >
        {loading ? "Eliminando..." : "Eliminar Grupo"}
      </button>
    </form>
  );
};

export default DeleteGroupForm;
