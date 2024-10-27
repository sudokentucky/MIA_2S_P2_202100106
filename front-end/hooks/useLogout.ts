import { useState } from "react";
import { useAuth } from "../hooks/useAuth"; // Importamos el hook de Zustand para manejar el estado de autenticación
import { useDisksStore } from "../hooks/useDiskStore"; // Importamos el hook para manejar el estado de los discos

export const useLogout = () => {
  const [loading, setLoading] = useState(false); // Estado para controlar si se está realizando el logout
  const [backendMessage, setBackendMessage] = useState<string | null>(null); // Mensaje devuelto por el backend
  const [messageType, setMessageType] = useState<"success" | "error" | "info" | "">(""); // Tipo de mensaje
  const { Logout } = useAuth(); // Traemos la función Logout de Zustand para actualizar el estado global
  const { clearDisks } = useDisksStore(); // Traemos la función clearDisks de Zustand para limpiar la lista de discos

  const logout = async () => {
    setLoading(true);
    setBackendMessage(null); // Limpiar mensaje anterior
    setMessageType(""); // Limpiar el tipo de mensaje anterior

    try {
      // Realizar la solicitud al backend para enviar el comando logout
      const apiUrl = import.meta.env.VITE_API_URL;
      const response = await fetch(`${apiUrl}/analyze`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ command: "logout" }), // Enviar el comando "logout"
      });

      const data = await response.json(); // Recibe la respuesta del backend

      if (response.ok && data.results && data.results.length > 0) {
        // Mostrar el primer mensaje de éxito o error del backend
        setBackendMessage(data.results[0]);
        setMessageType("success");
        Logout(); // Actualizamos el estado de autenticación en Zustand (cambia isLogged a false)
        clearDisks(); // Limpiar la lista de discos al cerrar sesión
      } else {
        // Mostrar un mensaje de error si no se recibe respuesta adecuada
        setBackendMessage("Error: No se pudo cerrar la sesión.");
        setMessageType("error");
      }
    } catch (error) {
      if (error instanceof Error) {
        setBackendMessage(`Error: ${error.message}`);
        setMessageType("error");
      } else {
        setBackendMessage("Ocurrió un error desconocido.");
        setMessageType("error");
      }
    } finally {
      setLoading(false); // Terminar el estado de carga
    }
  };

  return {
    loading,        // Indica si el logout está en proceso
    backendMessage, // Mensaje del backend
    messageType,    // Tipo de mensaje (éxito o error)
    logout,         // Función para ejecutar el logout
  };
};
