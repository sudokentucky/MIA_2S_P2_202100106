import { useState, useEffect } from "react";
import { useLogin } from "../hooks/useLogin"; // Importa el hook personalizado
import { useNavigate } from "react-router-dom";
import Message from "./Message"; // Importar el componente Message

function Login() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [userId, setUserId] = useState(""); // ID del usuario
  const [error, setError] = useState<string | null>(null); // Estado para errores en el frontend
  const [checkingUser, setCheckingUser] = useState(true);  // Estado para saber si se está comprobando si ya hay un usuario logueado
  const { login, loading, backendMessage, messageType } = useLogin(); // Usa el hook personalizado
  const navigate = useNavigate();

  // Función para verificar si hay un usuario ya logueado
  const checkUserLoggedIn = async () => {
    try {
      const response = await fetch("http://localhost:3000/users/logged-in"); // Llama al endpoint que comprueba si hay usuario logueado
      const data = await response.json();
      if (data.status === "success") {
        setError(`Ya hay un usuario logueado: ${data.user}`);
        setTimeout(() => {
          navigate("/user-management"); // Redirige a la gestión de usuarios si hay un usuario logueado
        }, 2000); // Espera 2 segundos antes de redirigir para que el usuario vea el mensaje
      }
    } catch (err) {
      setError(null); // No hay usuario logueado, permite continuar con el login normal
    } finally {
      setCheckingUser(false); // Termina la comprobación
    }
  };

  // Llamar a checkUserLoggedIn cuando el componente se monta
  useEffect(() => {
    checkUserLoggedIn();
  }, []);

  // Manejar el envío del formulario de login
  const handleLogin = (e: React.FormEvent) => {
    e.preventDefault();

    // Validar si los campos están vacíos antes de llamar a login
    if (!username || !password || !userId) {
      setError("Todos los campos son obligatorios: ID de usuario, nombre de usuario y contraseña."); // Mostrar mensaje de error
      return;
    }

    // Si los campos están completos, se procede con el login
    setError(null); // Limpiar cualquier error anterior
    login(username, password, userId); // Llamamos a la función login desde el hook
  };

  // Si está verificando si hay un usuario logueado, muestra un mensaje de carga
  if (checkingUser) {
    return <div className="text-center text-dracula-300">Verificando si hay un usuario logueado...</div>;
  }

  return (
    <div className="flex justify-center items-center min-h-screen">
      <div className="w-full max-w-sm p-6 bg-nosferatu-900 text-light-50 rounded-lg shadow-lg">
        <h1 className="text-3xl font-bold mb-6 text-center text-dracula-500">Iniciar Sesión</h1>
        
        <form onSubmit={handleLogin}>
          <div className="mb-4">
            <label className="block text-sm font-medium text-light-200 mb-2">ID del Usuario</label>
            <input
              type="text"
              className="w-full p-2 border border-nosferatu-700 rounded-md focus:outline-none focus:ring-2 focus:ring-dracula-300 focus:border-dracula-500 bg-nosferatu-800 text-light-50"
              value={userId}
              onChange={(e) => setUserId(e.target.value)}
              placeholder="Ingrese su ID"
            />
          </div>

          <div className="mb-4">
            <label className="block text-sm font-medium text-light-200 mb-2">Nombre de usuario</label>
            <input
              type="text"
              className="w-full p-2 border border-nosferatu-700 rounded-md focus:outline-none focus:ring-2 focus:ring-dracula-300 focus:border-dracula-500 bg-nosferatu-800 text-light-50"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="Ingrese su usuario"
            />
          </div>

          <div className="mb-4">
            <label className="block text-sm font-medium text-light-200 mb-2">Contraseña</label>
            <input
              type="password"
              className="w-full p-2 border border-nosferatu-700 rounded-md focus:outline-none focus:ring-2 focus:ring-dracula-300 focus:border-dracula-500 bg-nosferatu-800 text-light-50"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              placeholder="Ingrese su contraseña"
            />
          </div>

          {/* Mostrar error si falta algún campo o si ya hay un usuario logueado */}
          {error && <div className="text-center text-marcelin-400 mb-4">{error}</div>}

          {loading && <div className="text-center text-blue-500 mb-4">Iniciando sesión...</div>}
          
          {/* Componente Message para mostrar mensaje del backend */}
          <Message text={backendMessage ?? ""} type={messageType} />

          <button
            type="submit"
            className={`w-full bg-dracula-500 text-white p-2 rounded-md hover:bg-dracula-600 transition-all shadow-lg ${
              loading ? "opacity-50 cursor-not-allowed" : ""
            }`}
            disabled={loading}
          >
            Iniciar Sesión
          </button>
        </form>

        <button
          onClick={() => navigate("/")}
          className="w-full mt-4 bg-nosferatu-700 text-light-50 p-2 rounded-md hover:bg-nosferatu-800 transition-all shadow-md"
        >
          Regresar
        </button>
      </div>
    </div>
  );
}

export default Login;
