import { useState } from "react";
import { useNavigate } from "react-router-dom";

export const useLogin = () => {
  const [loading, setLoading] = useState(false);
  const [backendMessage, setBackendMessage] = useState<string | null>(null); // Mensaje devuelto por el backend
  const [messageType, setMessageType] = useState<"success" | "error" | "info" | "">(""); // Tipo de mensaje
  const navigate = useNavigate();

  const login = async (username: string, password: string, userId: string) => {
    setLoading(true);
    setBackendMessage(null); // Reiniciar el mensaje del backend
    setMessageType(""); // Reiniciar el tipo de mensaje
    try {
      // Realizar la solicitud al endpoint de login
      const response = await fetch("http://localhost:3000/users/login", {
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
      } else if (data.status === "success") {
        // Si el login es exitoso, redirigir y mostrar un mensaje de éxito
        setBackendMessage(data.message);
        setMessageType("success");
        navigate("/user-management"); // Redirigir a la gestión de usuarios
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
    loading,
    backendMessage,
    messageType, // Añadir el tipo de mensaje en el retorno
    login,
  };
};
