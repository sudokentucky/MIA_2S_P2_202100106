import { useState, useCallback } from "react";

function useCommandExecution() {
  const [inputText, setInputText] = useState("");
  const [outputText, setOutputText] = useState("");
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState(""); // Manejo de mensajes
  const [messageType, setMessageType] = useState<"success" | "error" | "info" | "">(""); // Tipo de mensaje

  // Función para mostrar mensajes
  const showMessage = useCallback((text: string, type: "success" | "error" | "info") => {
    setMessage(text);
    setMessageType(type);
    setTimeout(() => {
      setMessage("");
      setMessageType("");
    }, 5000); // El mensaje desaparece después de 5 segundos
  }, []);

  // Función para ejecutar el comando
  const handleExecute = useCallback(async () => {
    if (!inputText.trim()) {
      showMessage("El área de texto está vacía. Por favor, ingrese un comando o cargue un archivo.", "error");
      return;
    }

    setLoading(true);
    try {
      const response = await fetch("http://localhost:3000/analyze", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ command: inputText }),
      });

      if (!response.ok) {
        throw new Error("Error en la red o en la respuesta del servidor");
      }

      const data = await response.json();
      const results = data.results.join("\n");
      setOutputText(results);
      showMessage("Ejecución completada con éxito", "success");
    } catch (error) {
      if (error instanceof Error) {
        setOutputText(`Error: ${error.message}`);
        showMessage(`Error en la ejecución: ${error.message}`, "error");
      } else {
        setOutputText("Error desconocido");
        showMessage("Error en la ejecución: Error desconocido", "error");
      }
    } finally {
      setLoading(false);
    }
  }, [inputText, showMessage]);

  // Función para resetear los campos
  const handleReset = useCallback(() => {
    setInputText("");
    setOutputText("");
    showMessage("Campos limpiados correctamente", "info");
  }, [showMessage]);

  return {
    inputText,
    setInputText,
    outputText,
    loading,
    message,        // Retorna el mensaje actual
    messageType,    // Retorna el tipo de mensaje actual
    handleExecute,
    handleReset
  };
}

export default useCommandExecution;
