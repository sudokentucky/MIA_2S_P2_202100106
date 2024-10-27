import { useState } from "react";

export const useLogin = () => {
  const [loading, setLoading] = useState(false);
  const [backendMessage, setBackendMessage] = useState<string | null>(null); // Mensaje devuelto por el backend
  const [messageType, setMessageType] = useState<"success" | "error" | "info" | "">(""); // Tipo de mensaje

  const login = async (username: string, password: string, userId: string): Promise<string> => {
    setLoading(true);
    setBackendMessage(null); // Reiniciar el mensaje del backend
    setMessageType(""); // Reiniciar el tipo de mensaje

    try {

      // Realizar la solicitud al endpoint de login
      const apiUrl = import.meta.env.VITE_API_URL;
      const response = await fetch(`${apiUrl}/users/login`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ username, password, id: userId }),
      });

      const data = await response.json(); // Recibe la respuesta del backend

      if (data.status === "error") {
        // Mostrar el mensaje de error del backend
        setBackendMessage(data.message);
        setMessageType("error");
        return "error"; // Retornamos "error" si falla
      } else if (data.status === "success") {
        // Si el login es exitoso, redirigir y mostrar un mensaje de éxito
        setBackendMessage(data.message);
        setMessageType("success");
        return "success"; // Retornamos "success" si es exitoso
      }

      return "error"; // Retornamos "error" por defecto si no hay status
    } catch (error) {
      if (error instanceof Error) {
        setBackendMessage(`Error: ${error.message}`);
        setMessageType("error");
      } else {
        setBackendMessage("Ocurrió un error desconocido.");
        setMessageType("error");
      }
      return "error"; // Retornamos "error" si ocurre algún error
    } finally {
      setLoading(false); // Terminar el estado de carga
    }
  };

  return {
    loading,
    backendMessage,
    messageType,
    login,
  };
};
