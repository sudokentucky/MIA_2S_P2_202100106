import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import CommandExecution from "./CommandExecution";
import Login from "./Login";
import UserManagement from "./UserManagement";
import FileVisualizer from "./FileVisualizer"; // Importa el componente de visualización de archivos
import Navbar from "./NavBar";
import { useState } from "react";

function App() {
  const [navbarUpdated, setNavbarUpdated] = useState(false);

  return (
    <Router>
      <div className="min-h-screen flex flex-col">
        <Navbar navbarUpdated={navbarUpdated} />
        <div className="flex-grow p-4" style={{ backgroundColor: "#E6E6FA" }}>
          <div className="container mx-auto">
            <Routes>
              <Route path="/" element={<CommandExecution updateNavbar={() => setNavbarUpdated(!navbarUpdated)} />} />
              <Route path="/login" element={<Login />} />
              <Route path="/user-management" element={<UserManagement />} />
              <Route path="/file-visualizer" element={<FileVisualizer />} /> {/* Nueva ruta */}
            </Routes>
          </div>
        </div>
      </div>
    </Router>
  );
}

export default App;
