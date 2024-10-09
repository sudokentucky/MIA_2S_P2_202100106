import { useNavigate } from "react-router-dom";
import { useState, useEffect } from "react";
import { useCheckPartition } from "../hooks/useCheckPartition"; // Hook para verificar particiones
import Message from "./Message"; // Importar el componente Message

// Tipamos la prop navbarUpdated como boolean
interface NavbarProps {
  navbarUpdated: boolean;
}

function Navbar({ navbarUpdated }: NavbarProps) {  // Tipamos la prop como boolean
  const { partitionStatus } = useCheckPartition(); // Usar el hook para verificar particiones
  const navigate = useNavigate(); // Hook para redirigir
  const [checkingPartition, setCheckingPartition] = useState(false); // Estado para mostrar cargando
  const [message, setMessage] = useState<string>(""); // Estado para el mensaje
  const [messageType, setMessageType] = useState<"success" | "error" | "info" | "">(""); // Estado para el tipo de mensaje

  // Verificar partición cada vez que navbarUpdated cambie
  useEffect(() => {
    if (partitionStatus !== "success") {
    } else {
      setMessage("Partición montada correctamente");
      setMessageType("success");
    }
  }, [navbarUpdated, partitionStatus]); // Dependencia de navbarUpdated

  // Temporizador para eliminar el mensaje automáticamente después de 5 segundos
  useEffect(() => {
    if (message) {
      const timer = setTimeout(() => {
        setMessage("");
        setMessageType("");
      }, 5000); // Elimina el mensaje después de 5 segundos

      // Limpieza del temporizador cuando el mensaje cambia o se desmonta el componente
      return () => clearTimeout(timer);
    }
  }, [message]); // Se ejecuta cada vez que el mensaje cambie

  // Función para manejar el click en el botón de inicio de sesión
  const handleLoginClick = () => {
    if (checkingPartition) return;

    setCheckingPartition(true);

    // Verificar partición
    if (partitionStatus === "success") {
      // Si hay partición montada, redirigir al login
      navigate("/login");
    } else {
      // Mostrar mensaje de error si no hay partición montada
      setMessage("Debe de haber al menos una partición montada para iniciar sesión.");
      setMessageType("error");
    }

    setCheckingPartition(false);
  };

  return (
    <>
  <nav className="bg-nosferatu-900 text-dracula-300 shadow-md py-4">
  <div className="container mx-auto px-4">
    <div className="flex justify-between items-center">
      {/* Logo o título */}
      <div className="text-2xl font-bold">
        <a href="/">Sistema de Archivos ext2</a>
      </div>
      
      {/* Enlaces de navegación */}
      <div className="space-x-4">
        <a href="/" className="hover:bg-dracula-100 hover:text-nosferatu-50 px-3 py-2 rounded-md">
          Ejecutar Comandos
        </a>

        {/* Botón de "Inicio de Sesión" */}
        <button
          onClick={handleLoginClick}
          className={`bg-dracula-500 hover:bg-dracula-600 text-nosferatu-50 px-3 py-2 rounded-md ${checkingPartition ? "opacity-50 cursor-wait" : ""}`}
          disabled={checkingPartition}
        >
          {checkingPartition ? "Verificando..." : "Inicio de Sesión"}
        </button>
      </div>
    </div>
  </div>
</nav>


      {/* Mostrar el mensaje si existe */}
      {message && <Message text={message} type={messageType} />}
    </>
  );
}

export default Navbar;
