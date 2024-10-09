import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import CommandExecution from "./CommandExecution";
import Login from "./Login";
import UserManagement from "./UserManagement";
import Navbar from "./NavBar";
import { useState } from "react";

function App() {
  const [navbarUpdated, setNavbarUpdated] = useState(false); // Estado para actualizar el Navbar

  // Función que se pasa a CommandExecution para forzar la actualización
  const updateNavbar = () => {
    setNavbarUpdated((prev) => !prev); // Cambia el estado para forzar el renderizado
  };

  return (
    <Router>
      <div className="min-h-screen flex flex-col">
        {/* Incluimos el Navbar, que se renderizará cada vez que navbarUpdated cambie */}
        <Navbar navbarUpdated={navbarUpdated} />
        
        {/* Rutas principales */}
        <div className="flex-grow p-4" style={{ backgroundColor: "#E6E6FA" }}>
          <div className="container mx-auto">
            <Routes>
              <Route path="/" element={<CommandExecution updateNavbar={updateNavbar} />} />
              <Route path="/login" element={<Login />} />
              <Route path="/user-management" element={<UserManagement />} />
            </Routes>
          </div>
        </div>
      </div>
    </Router>
  );
}

export default App;
