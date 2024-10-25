import { useEffect, useState } from "react";
import { useNavigate, useLocation, Link } from "react-router-dom"; // Cambiar a Link para el logo
import Message from "./Message"; // Importar el componente Message
import { useAuth } from "../hooks/useAuth"; // Importar Zustand para el estado de autenticación
import { useLogout } from "../hooks/useLogout"; // Importar el hook personalizado de logout

function Navbar() {
  const { isLogged } = useAuth(); // Obtener el estado de autenticación desde Zustand
  const navigate = useNavigate(); // Hook para redirigir
  const location = useLocation(); // Obtener la ruta actual
  const [message, setMessage] = useState<string>(""); // Estado para el mensaje
  const [messageType, setMessageType] = useState<"success" | "error" | "info" | "">(""); // Estado para el tipo de mensaje

  const { logout, loading: logoutLoading, backendMessage, messageType: logoutMessageType } = useLogout(); // Usar el hook personalizado de logout

  useEffect(() => {
    if (backendMessage) {
      setMessage(backendMessage);
      setMessageType(logoutMessageType);

      const timer = setTimeout(() => {
        setMessage("");
        setMessageType("");
      }, 5000); // Eliminar el mensaje después de 5 segundos

      return () => clearTimeout(timer);
    }
  }, [backendMessage, logoutMessageType]);

  const handleLoginClick = () => {
    navigate("/login"); // Redirigir directamente a la página de inicio de sesión
  };

  const handleFileVisualizerClick = () => {
    navigate("/file-visualizer"); // Redirigir a la página del visualizador de archivos
  };

  const handleLogoutClick = () => {
    logout(); // Ejecutar el proceso de logout
    navigate("/"); // Redirigir a la página principal
  };

  return (
    <>
      <nav className="bg-nosferatu-900 text-dracula-300 shadow-md py-4">
        <div className="container mx-auto px-4">
          <div className="flex justify-between items-center">
            {/* Logo o título */}
            <div className="text-2xl font-bold">
              {/* Cambiar <a> por <Link> para evitar recarga completa de la página */}
              <Link to="/">Sistema de Archivos Ext2 | Ext3</Link>
            </div>

            {/* Enlaces de navegación */}
            <div className="space-x-4">
              {/* Mostrar el botón de "Visualizador de Archivos" solo si el usuario está logueado */}
              {isLogged && location.pathname !== "/file-visualizer" && (
                <button
                  onClick={handleFileVisualizerClick}
                  className="bg-dracula-500 hover:bg-dracula-600 text-nosferatu-50 px-3 py-2 rounded-md"
                >
                  Discos
                </button>
              )}

              {/* Mostrar enlace a "Gestión de Usuarios" si el usuario está logueado */}
              {isLogged && (
                <button
                  onClick={() => navigate("/user-management")}
                  className="bg-dracula-500 hover:bg-dracula-600 text-nosferatu-50 px-3 py-2 rounded-md"
                >
                  Usuarios y Grupos
                </button>
              )}

              {/* Mostrar botón de "Inicio de Sesión" si el usuario no está logueado */}
              {!isLogged && location.pathname !== "/login" && (
                <button
                  onClick={handleLoginClick}
                  className="bg-dracula-500 hover:bg-dracula-600 text-nosferatu-50 px-3 py-2 rounded-md"
                >
                  Inicio de Sesión
                </button>
              )}

              {/* Mostrar botón de "Logout" si el usuario está logueado */}
              {isLogged && (
                <button
                  onClick={handleLogoutClick}
                  className="bg-red-500 hover:bg-red-600 text-nosferatu-50 px-3 py-2 rounded-md"
                  disabled={logoutLoading} // Deshabilitar si está procesando logout
                >
                  {logoutLoading ? "Cerrando sesión..." : "Cerrar Sesión"}
                </button>
              )}
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
