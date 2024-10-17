import React, { useState } from "react";
import useCommandExecution from "../hooks/useCommandExecution";

const DeleteUserForm = ({ refresh, setMessage, setMessageType }: { refresh: () => void; setMessage: (msg: string) => void; setMessageType: (type: "success" | "error") => void; }) => {
  const [deleteUserName, setDeleteUserName] = useState("");
  const { setInputText, handleExecute, loading } = useCommandExecution();

  const handleDeleteUserSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!deleteUserName.trim()) {
      alert("Por favor, ingrese el nombre del usuario que desea eliminar.");
      return;
    }

    const command = `rmusr -user=${deleteUserName}`;
    setInputText(command);
    try {
      await handleExecute();   // Ejecutar el comando
      setMessage("Usuario eliminado exitosamente");  // Mensaje de éxito
      setMessageType("success");
      refresh();               // Refrescar la lista de usuarios y grupos
    } catch (error) {
      setMessage("Error al eliminar el usuario");    // Mensaje de error
      setMessageType("error");
    }
  };

  return (
    <form onSubmit={handleDeleteUserSubmit} className="bg-nosferatu-800 p-6 rounded-lg shadow-lg transition-all hover:shadow-xl duration-300">
      <div className="mb-4">
        <label htmlFor="deleteUserName" className="block text-dracula-300 text-sm font-bold mb-2">
          Nombre del Usuario a Eliminar
        </label>
        <input
          type="text"
          id="deleteUserName"
          value={deleteUserName}
          onChange={(e) => setDeleteUserName(e.target.value)}
          className="shadow appearance-none border border-gray-600 bg-gray-700 rounded w-full py-2 px-3 text-gray-100 leading-tight focus:outline-none focus:ring-2 focus:ring-red-500"
          placeholder="Ingrese el nombre del usuario a eliminar"
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
        {loading ? "Eliminando..." : "Eliminar Usuario"}
      </button>
    </form>
  );
};

export default DeleteUserForm;
