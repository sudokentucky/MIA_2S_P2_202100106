import React, { useState } from "react";
import CreateGroupForm from "../interface/CreateGroupForm";
import DeleteGroupForm from "../interface/DeleteGroupForm";
import CreateUserForm from "../interface/CreateUserForm";
import DeleteUserForm from "../interface/DeleteUserForm";
import ChangeUserGroupForm from "../interface/ChangeUserGroupForm";
import useUsersGroups from "../hooks/useUsersGroups";

function UserManagement() {
  const { users, groups, loading: loadingUsersGroups, error, refresh } = useUsersGroups();
  const [message, setMessage] = useState("");           // Estado para mostrar mensajes
  const [messageType, setMessageType] = useState<"success" | "error" | "">(""); // Estado para tipo de mensaje

  return (
    <div className="min-h-screen flex items-center justify-center"> {/* Fondo más oscuro */}
      <div className="w-full max-w-4xl p-8 bg-aro-900 rounded-lg shadow-lg transition-all transform hover:shadow-xl duration-300">
        <h1 className="text-3xl font-bold mb-6 text-center text-dracula-300">Gestión de Usuarios y Grupos</h1>

        {/* Mostrar mensajes de éxito o error */}
        {message && (
          <div className={`mt-4 p-4 rounded-lg ${messageType === "error" ? "bg-red-500" : "bg-green-500"} text-white`}>
            {message}
          </div>
        )}

        {/* Gestión de Grupos */}
        <div className="mt-8">
          <h2 className="text-2xl font-bold mb-4 text-dracula-200">Gestión de Grupos</h2>

          <h3 className="text-xl font-bold mb-2 text-dracula-300">Crear Grupo</h3>
          <CreateGroupForm refresh={refresh} setMessage={setMessage} setMessageType={setMessageType} />
          
          <hr className="my-6 border-t-2 border-dracula-400" /> {/* Línea divisoria */}

          <h3 className="text-xl font-bold mb-2 text-dracula-300">Eliminar Grupo</h3>
          <DeleteGroupForm refresh={refresh} setMessage={setMessage} setMessageType={setMessageType} />
        </div>

        {/* Gestión de Usuarios */}
        <div className="mt-8">
          <h2 className="text-2xl font-bold mb-4 text-dracula-200">Gestión de Usuarios</h2>

          <h3 className="text-xl font-bold mb-2 text-dracula-300">Crear Usuario</h3>
          <CreateUserForm refresh={refresh} setMessage={setMessage} setMessageType={setMessageType} />
          
          <hr className="my-6 border-t-2 border-dracula-400" /> {/* Línea divisoria */}

          <h3 className="text-xl font-bold mb-2 text-dracula-300">Eliminar Usuario</h3>
          <DeleteUserForm refresh={refresh} setMessage={setMessage} setMessageType={setMessageType} />

          <hr className="my-6 border-t-2 border-dracula-400" /> {/* Línea divisoria */}

          <h3 className="text-xl font-bold mb-2 text-dracula-300">Cambiar Grupo de Usuario</h3>
          <ChangeUserGroupForm refresh={refresh} setMessage={setMessage} setMessageType={setMessageType} />
        </div>

        {/* Mostrar lista de usuarios y grupos */}
        <div className="mt-8">
          <h2 className="text-2xl font-bold mb-4 text-dracula-200">Listado de Usuarios y Grupos</h2>
          {loadingUsersGroups ? (
            <p className="text-dracula-400">Cargando usuarios y grupos...</p>
          ) : error ? (
            <p className="text-red-500">Error: {error}</p>
          ) : (
            <>
              <h3 className="text-xl font-bold mb-2 text-dracula-300">Grupos</h3>
              <ul className="mb-4 text-gray-300">
                {groups.map((group) => (
                  <li key={group}>{group}</li>
                ))}
              </ul>
              <h3 className="text-xl font-bold mb-2 text-dracula-300">Usuarios</h3>
              <ul className="text-gray-300">
                {users.map((user) => (
                  <li key={user}>{user}</li>
                ))}
              </ul>
            </>
          )}
        </div>
      </div>
    </div>
  );
}

export default UserManagement;
