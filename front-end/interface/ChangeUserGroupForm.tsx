import React, { useState } from "react";
import useCommandExecution from "../hooks/useCommandExecution";

const ChangeUserGroupForm = ({ refresh, setMessage, setMessageType }: { refresh: () => void; setMessage: (msg: string) => void; setMessageType: (type: "success" | "error") => void; }) => {
  const [userName, setUserName] = useState("");
  const [newGroup, setNewGroup] = useState("");
  const { setInputText, handleExecute, loading } = useCommandExecution();

  const handleChangeUserGroupSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!userName.trim() || !newGroup.trim()) {
      alert("Por favor, complete ambos campos para cambiar el grupo del usuario.");
      return;
    }

    const command = `chgrp -user=${userName} -grp=${newGroup}`;
    setInputText(command);

    try {
      await handleExecute();   // Ejecutar el comando
      setMessage(`Grupo del usuario '${userName}' cambiado a '${newGroup}' exitosamente`);  // Mensaje de éxito
      setMessageType("success");
      refresh();               // Refrescar la lista de usuarios y grupos
    } catch (error) {
      setMessage("Error al cambiar el grupo del usuario");  // Mensaje de error
      setMessageType("error");
    }
  };

  return (
    <form onSubmit={handleChangeUserGroupSubmit} className="bg-nosferatu-800 p-6 rounded-lg shadow-lg transition-all hover:shadow-xl duration-300">
      <div className="mb-4">
        <label htmlFor="userName" className="block text-dracula-300 text-sm font-bold mb-2">
          Nombre de Usuario
        </label>
        <input
          type="text"
          id="userName"
          value={userName}
          onChange={(e) => setUserName(e.target.value)}
          className="shadow appearance-none border border-gray-600 bg-gray-700 rounded w-full py-2 px-3 text-gray-100 leading-tight focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="Ingrese el nombre del usuario"
          disabled={loading}
        />
      </div>

      <div className="mb-4">
        <label htmlFor="newGroup" className="block text-dracula-300 text-sm font-bold mb-2">
          Nuevo Grupo
        </label>
        <input
          type="text"
          id="newGroup"
          value={newGroup}
          onChange={(e) => setNewGroup(e.target.value)}
          className="shadow appearance-none border border-gray-600 bg-gray-700 rounded w-full py-2 px-3 text-gray-100 leading-tight focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="Ingrese el nuevo grupo"
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
        {loading ? "Cambiando..." : "Cambiar Grupo"}
      </button>
    </form>
  );
};

export default ChangeUserGroupForm;
