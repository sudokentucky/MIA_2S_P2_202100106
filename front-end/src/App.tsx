import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import CommandExecution from "./CommandExecution";
import Login from "./Login";
import UserManagement from "./UserManagement";
import DiskVisualizer from "./DiskVisualizer"; // Importa el componente de visualización de archivos
import PartitionVisualizer from "./partition"; // Importa el nuevo componente
import FileSystemTree from "./FileView";
import Navbar from "./NavBar";
import { useState } from "react";

function App() {
  const [navbarUpdated, setNavbarUpdated] = useState(false);

  return (
    <Router>
      <div className="min-h-screen flex flex-col">
        <Navbar />
        <div className="flex-grow p-4" style={{ backgroundColor: "#E6E6FA" }}>
          <div className="container mx-auto">
            <Routes>
              <Route
                path="/"
                element={
                  <CommandExecution updateNavbar={() => setNavbarUpdated(!navbarUpdated)} />
                }
              />
              <Route path="/login" element={<Login />} />
              <Route path="/user-management" element={<UserManagement />} />
              <Route path="/file-visualizer" element={<DiskVisualizer />} />
              {/* Nueva ruta para visualizar particiones */}
              <Route path="/partitions/:diskIndex" element={<PartitionVisualizer />} />
              {/* Nueva ruta para visualizar el árbol de archivos */}
              <Route path="/partition-tree" element={<FileSystemTree />} />
            </Routes>
          </div>
        </div>
      </div>
    </Router>
  );
}

export default App;
