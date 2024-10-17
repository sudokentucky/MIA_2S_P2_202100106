import { useState, useEffect } from "react";

// Hook personalizado para obtener la lista de usuarios y grupos
function useUsersGroups() {
  const [users, setUsers] = useState<string[]>([]);         // Estado para almacenar la lista de usuarios
  const [groups, setGroups] = useState<string[]>([]);       // Estado para almacenar la lista de grupos
  const [loading, setLoading] = useState(true);   // Estado para controlar el estado de carga
  const [error, setError] = useState<string | null>(null);  // Estado para manejar los errores

  // Función para obtener la lista de usuarios y grupos desde el servidor
  const fetchUsersGroups = async () => {
    setLoading(true);
    try {
      const response = await fetch("http://localhost:3000/list-users-groups");
      if (!response.ok) {
        throw new Error("Error al obtener la lista de usuarios y grupos");
      }

      const data = await response.json();
      setUsers(data.users || []);
      setGroups(data.groups || []);
    } catch (err) {
      // Verificar si `err` es una instancia de Error
      if (err instanceof Error) {
        setError(err.message); // Acceder al mensaje de error
      } else {
        setError("Error desconocido");
      }
    } finally {
      setLoading(false);
    }
  };

  // useEffect para llamar a la API cuando el componente se monta
  useEffect(() => {
    fetchUsersGroups();
  }, []);  // El array vacío asegura que se llame solo al montar el componente

  return {
    users,
    groups,
    loading,
    error,
    refresh: fetchUsersGroups,  // Función para refrescar los datos manualmente
  };
}

export default useUsersGroups;
