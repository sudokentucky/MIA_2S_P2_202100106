import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useLogin } from "../hooks/useLogin";
import { useAuth } from "../hooks/useAuth";
import Message from "./Message";

function Login() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [userId, setUserId] = useState("");
  const [error, setError] = useState<string | null>(null);
  const { login, loading, backendMessage, messageType } = useLogin();
  const { Login } = useAuth();
  const navigate = useNavigate();

  const handleLogin = (e: React.FormEvent) => {
    e.preventDefault();
    if (!username || !password || !userId) {
      setError("Todos los campos son obligatorios: ID de usuario, nombre de usuario y contraseña.");
      return;
    }
    setError(null);
    login(username, password, userId).then((result) => {
      if (result === "success") {
        Login();
        navigate("/user-management");
      }
    });
  };

  return (
    <div className="min-h-screenpy-6 flex flex-col justify-center sm:py-12">
      <div className="relative py-3 sm:max-w-xl sm:mx-auto">
        <div className="absolute inset-0 bg-gradient-to-r from-dracula-500 to-marcelin-400 shadow-lg transform -skew-y-6 sm:skew-y-0 sm:-rotate-6 sm:rounded-3xl"></div>
        
        <div className="relative px-4 py-10 bg-nosferatu-700 shadow-lg sm:rounded-3xl sm:p-20">
          <div className="max-w-md mx-auto">
            <h1 className="text-2xl font-semibold text-center mb-6 text-light-50">Iniciar Sesión</h1>

            <form onSubmit={handleLogin} className="space-y-6">
              {/* Campo de ID de usuario */}
              <div className="relative">
                <input
                  type="text"
                  id="userId"
                  name="userId"
                  value={userId}
                  onChange={(e) => setUserId(e.target.value)}
                  placeholder="ID de Usuario"
                  className="peer placeholder-transparent h-10 w-full border-b-2 border-light-200 bg-nosferatu-700 text-light-50 focus:outline-none focus:border-dracula-500"
                  autoComplete="off"
                />
                <label
                  htmlFor="userId"
                  className="absolute left-0 -top-3.5 text-light-200 text-sm peer-placeholder-shown:top-2 peer-placeholder-shown:text-light-400 peer-placeholder-shown:text-base transition-all peer-focus:-top-3.5 peer-focus:text-dracula-500 peer-focus:text-sm"
                >
                  ID de Usuario
                </label>
              </div>

              {/* Campo de Nombre de Usuario */}
              <div className="relative">
                <input
                  type="text"
                  id="username"
                  name="username"
                  value={username}
                  onChange={(e) => setUsername(e.target.value)}
                  placeholder="Nombre de Usuario"
                  className="peer placeholder-transparent h-10 w-full border-b-2 border-light-200 bg-nosferatu-700 text-light-50 focus:outline-none focus:border-dracula-500"
                  autoComplete="off"
                />
                <label
                  htmlFor="username"
                  className="absolute left-0 -top-3.5 text-light-200 text-sm peer-placeholder-shown:top-2 peer-placeholder-shown:text-light-400 peer-placeholder-shown:text-base transition-all peer-focus:-top-3.5 peer-focus:text-dracula-500 peer-focus:text-sm"
                >
                  Nombre de Usuario
                </label>
              </div>

              {/* Campo de Contraseña */}
              <div className="relative">
                <input
                  type="password"
                  id="password"
                  name="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  placeholder="Contraseña"
                  className="peer placeholder-transparent h-10 w-full border-b-2 border-light-200 bg-nosferatu-700 text-light-50 focus:outline-none focus:border-dracula-500"
                  autoComplete="off"
                />
                <label
                  htmlFor="password"
                  className="absolute left-0 -top-3.5 text-light-200 text-sm peer-placeholder-shown:top-2 peer-placeholder-shown:text-light-400 peer-placeholder-shown:text-base transition-all peer-focus:-top-3.5 peer-focus:text-dracula-500 peer-focus:text-sm"
                >
                  Contraseña
                </label>
              </div>

              {/* Mostrar error */}
              {error && <div className="text-center text-marcelin-400">{error}</div>}
              {loading && <div className="text-center text-dracula-500">Iniciando sesión...</div>}
              
              {/* Mensaje del backend */}
              <Message text={backendMessage ?? ""} type={messageType} />

              {/* Botón de envío */}
              <button
                type="submit"
                className={`w-full bg-gradient-to-r from-dracula-500 to-marcelin-500 text-white py-2 rounded-md shadow-md hover:bg-gradient-to-l transition-all ${
                  loading ? "opacity-50 cursor-not-allowed" : ""
                }`}
                disabled={loading}
              >
                Iniciar Sesión
              </button>
            </form>

            {/* Botón Regresar */}
            <button
              onClick={() => navigate("/")}
              className="mt-4 w-full flex items-center justify-center bg-nosferatu-800 border border-nosferatu-700 rounded-md shadow-md py-2 text-sm font-medium text-light-50 hover:bg-nosferatu-700 focus:outline-none"
            >
              Regresar
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}

export default Login;
