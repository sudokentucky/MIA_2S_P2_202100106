import React, { useState } from "react";
import useCommandExecution from "../hooks/useCommandExecution";

const CreateGroupForm = ({ refresh, setMessage, setMessageType }: { refresh: () => void; setMessage: (msg: string) => void; setMessageType: (type: "success" | "error") => void; }) => {
  const [groupName, setGroupName] = useState("");
  const { setInputText, handleExecute, loading } = useCommandExecution();

  const handleCreateGroupSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!groupName.trim()) {
      alert("Por favor, ingrese un nombre de grupo válido.");
      return;
    }

    const command = `mkgrp -name=${groupName}`;
    setInputText(command);
    try {
      await handleExecute();   // Ejecutar el comando
      setMessage("Grupo creado exitosamente");  // Mensaje de éxito
      setMessageType("success");
      refresh();               // Refrescar la lista después de la creación
    } catch (error) {
      setMessage("Error al crear el grupo");    // Mensaje de error
      setMessageType("error");
    }
  };

  return (
    <form onSubmit={handleCreateGroupSubmit} className="bg-nosferatu-800 p-6 rounded-lg shadow-lg transition-all hover:shadow-xl duration-300">
      <div className="mb-4">
        <label htmlFor="groupName" className="block text-dracula-300 text-sm font-bold mb-2">
          Nombre del Grupo
        </label>
        <input
          type="text"
          id="groupName"
          value={groupName}
          onChange={(e) => setGroupName(e.target.value)}
          className="shadow appearance-none border border-gray-600 bg-gray-700 rounded w-full py-2 px-3 text-gray-100 leading-tight focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="Ingrese el nombre del grupo"
          disabled={loading}
        />
      </div>
      <button
        type="submit"
        className={`bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:ring-2 focus:ring-blue-500 transition-all ${
          loading ? "opacity-50 cursor-not-allowed" : ""
        }`}
        disabled={loading}
      >
        {loading ? "Creando..." : "Crear Grupo"}
      </button>
    </form>
  );
};

export default CreateGroupForm;
