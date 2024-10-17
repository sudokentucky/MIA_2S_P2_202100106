import React, { useState } from "react";
import useCommandExecution from "../hooks/useCommandExecution";

const CreateUserForm = ({ refresh, setMessage, setMessageType }: { refresh: () => void; setMessage: (msg: string) => void; setMessageType: (type: "success" | "error") => void; }) => {
  const [userName, setUserName] = useState("");
  const [userPassword, setUserPassword] = useState("");
  const [userGroup, setUserGroup] = useState("");
  const { setInputText, handleExecute, loading } = useCommandExecution();

  const handleCreateUserSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!userName.trim() || !userPassword.trim() || !userGroup.trim()) {
      alert("Por favor, complete todos los campos para crear un usuario.");
      return;
    }

    const command = `mkusr -user=${userName} -pass=${userPassword} -grp=${userGroup}`;
    setInputText(command);
    try {
      await handleExecute();   // Ejecutar el comando
      setMessage("Usuario creado exitosamente");  // Mensaje de éxito
      setMessageType("success");
      refresh();               // Refrescar la lista de usuarios y grupos
    } catch (error) {
      setMessage("Error al crear el usuario");    // Mensaje de error
      setMessageType("error");
    }
  };

  return (
    <form onSubmit={handleCreateUserSubmit} className="bg-nosferatu-800 p-6 rounded-lg shadow-lg transition-all hover:shadow-xl duration-300">
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
        <label htmlFor="userPassword" className="block text-dracula-300 text-sm font-bold mb-2">
          Contraseña
        </label>
        <input
          type="password"
          id="userPassword"
          value={userPassword}
          onChange={(e) => setUserPassword(e.target.value)}
          className="shadow appearance-none border border-gray-600 bg-gray-700 rounded w-full py-2 px-3 text-gray-100 leading-tight focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="Ingrese la contraseña"
          disabled={loading}
        />
      </div>

      <div className="mb-4">
        <label htmlFor="userGroup" className="block text-dracula-300 text-sm font-bold mb-2">
          Grupo del Usuario
        </label>
        <input
          type="text"
          id="userGroup"
          value={userGroup}
          onChange={(e) => setUserGroup(e.target.value)}
          className="shadow appearance-none border border-gray-600 bg-gray-700 rounded w-full py-2 px-3 text-gray-100 leading-tight focus:outline-none focus:ring-2 focus:ring-blue-500"
          placeholder="Ingrese el grupo del usuario"
          disabled={loading}
        />
      </div>

      <button
        type="submit"
        className={`bg-green-500 hover:bg-green-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:ring-2 focus:ring-green-500 transition-all ${
          loading ? "opacity-50 cursor-not-allowed" : ""
        }`}
        disabled={loading}
      >
        {loading ? "Creando..." : "Crear Usuario"}
      </button>
    </form>
  );
};

export default CreateUserForm;
