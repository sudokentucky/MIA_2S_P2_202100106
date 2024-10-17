import { useNavigate } from "react-router-dom";
import { useState, useEffect } from "react";
import { useCheckPartition } from "../hooks/useCheckPartition"; // Hook para verificar particiones
import Message from "./Message"; // Importar el componente Message

interface NavbarProps {
  navbarUpdated: boolean;
}

function Navbar({ navbarUpdated }: NavbarProps) {
  const { partitionStatus } = useCheckPartition(); // Usar el hook para verificar particiones
  const navigate = useNavigate(); // Hook para redirigir
  const [checkingPartition, setCheckingPartition] = useState(false); // Estado para mostrar cargando
  const [message, setMessage] = useState<string>(""); // Estado para el mensaje
  const [messageType, setMessageType] = useState<"success" | "error" | "info" | "">(""); // Estado para el tipo de mensaje
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false); // Estado de autenticación

  useEffect(() => {
    if (partitionStatus === "success") {
      setMessage("Partición montada correctamente");
      setMessageType("success");
    }
  }, [navbarUpdated, partitionStatus]);

  useEffect(() => {
    if (message) {
      const timer = setTimeout(() => {
        setMessage("");
        setMessageType("");
      }, 5000); // Elimina el mensaje después de 5 segundos
      return () => clearTimeout(timer);
    }
  }, [message]);

  const handleLoginClick = () => {
    if (checkingPartition) return;

    setCheckingPartition(true);

    // Verificar partición
    if (partitionStatus === "success") {
      if (isAuthenticated) {
        navigate("/user-management");
      } else {
        navigate("/login");
      }
    } else {
      setMessage("Debe de haber al menos una partición montada para iniciar sesión.");
      setMessageType("error");
    }

    setCheckingPartition(false);
  };

  // Función para manejar el click en el botón de visualizador de archivos
  const handleFileVisualizerClick = () => {
    navigate("/file-visualizer");
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


              {/* Botón para el Visualizador de Archivos */}
              <button
                onClick={handleFileVisualizerClick}
                className="bg-dracula-500 hover:bg-dracula-600 text-nosferatu-50 px-3 py-2 rounded-md"
              >
                Visualizador de Archivos
              </button>

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

      {message && <Message text={message} type={messageType} />}
    </>
  );
}

export default Navbar;
