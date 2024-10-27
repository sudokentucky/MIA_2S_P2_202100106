import { useRef } from "react";
import useCommandExecution from "../hooks/useCommandExecution";
import useLineCounter from "../hooks/useLineCounter";
import FileInput from "./FileInput";
import Message from "./Message";


// Definir el tipo de las props que recibe CommandExecution
interface CommandExecutionProps {
  updateNavbar: () => void;  // Tipar correctamente updateNavbar como una función
}

function CommandExecution({ updateNavbar }: CommandExecutionProps) {  // Recibe updateNavbar como prop
  const { 
    inputText, 
    setInputText, 
    outputText, 
    loading, 
    handleExecute, 
    handleReset, 
    message,        // Recibiendo mensaje del hook
    messageType     // Recibiendo tipo de mensaje del hook
  } = useCommandExecution();

  const { lineCount } = useLineCounter(inputText);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const lineCounterRef = useRef<HTMLDivElement>(null);

  const syncScroll = () => {
    if (textareaRef.current && lineCounterRef.current) {
      lineCounterRef.current.scrollTop = textareaRef.current.scrollTop;
    }
  };

  const executeCommand = async () => {
    await handleExecute(); // Ejecuta el comando
    
    // Llama a updateNavbar después de ejecutar con éxito
    updateNavbar();
  };

  return (
<div className="flex flex-col flex-grow  p-4 rounded-md shadow-md">
  <div className="flex-grow flex items-center justify-center">
    <div className="w-full max-w-5xl p-6 bg-cullen-100 rounded-lg shadow-lg transition-all hover:shadow-xl transform hover:scale-105 duration-300 ease-in-out">
      <h1 className="text-3xl font-bold mb-6 text-center text-dracula-500 shadow-md">
        Ejecución de Comandos
      </h1>
      
      <div className="mb-4 relative">
        <label className="block text-base font-medium text-darker-700 mb-2" htmlFor="input-text">
          Entrada de comando o archivo de texto
        </label>
        <div className="flex">
          <div 
            ref={lineCounterRef} 
            className="line-numbers bg-darker-100 p-2 rounded-l-md text-sm text-right overflow-hidden scrollbar-thin" 
            style={{ minHeight: '200px', maxHeight: '300px' }}>
            {Array.from({ length: lineCount }, (_, i) => i + 1).map((line) => (
              <div key={line}>{line}</div>
            ))}
          </div>
          <textarea
            id="input-text"
            ref={textareaRef}
            className="w-full p-3 border border-gray-300 rounded-r-md bg-light-50 shadow-sm focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none text-sm overflow-y-auto scrollbar-thin"
            value={inputText}
            onChange={(e) => setInputText(e.target.value)}
            onScroll={syncScroll}
            placeholder="Ingrese comandos aquí..."
            disabled={loading}
            style={{ minHeight: '200px', fontSize: '13px', whiteSpace: 'pre' }}
          />
        </div>
      </div>

      <div className="mb-4">
        <label className="block text-base font-medium text-darker-700 mb-2" htmlFor="output-text">
          Resultado de la ejecución
        </label>
        <textarea
          id="output-text"
          className="w-full p-3 border border-gray-300 rounded-md bg-light-50 shadow-sm focus:outline-none focus:ring-2 focus:ring-green-500 text-sm"
          value={outputText}
          readOnly
          placeholder="Resultado de la ejecución aparecerá aquí..."
          style={{ fontFamily: '"Courier New", monospace', fontSize: '13px', minHeight: '200px' }}
        />
      </div>

      <div className="flex justify-between items-center space-x-4">
        <FileInput onFileChange={setInputText} showMessage={() => {}} loading={loading} />
        <button 
          onClick={executeCommand}
          className={`px-4 py-2 rounded-md text-white border-2 ${
            loading ? "bg-gray-400 border-gray-400" : "bg-green-500 border-green-500 hover:bg-green-600 hover:border-green-600"
          } transition-all`} 
          disabled={loading}>
          {loading ? "Ejecutando..." : "Ejecutar"}
        </button>
        <button 
          onClick={handleReset} 
          className="px-4 py-2 bg-red-500 text-white border-2 border-red-500 rounded-md hover:bg-red-600 hover:border-red-600 transition-all" 
          disabled={loading}>
          🧹 Limpiar
        </button>
      </div>

      {/* Mostrar el componente Message cuando haya mensaje */}
      <Message text={message} type={messageType} />

      {loading && (
        <div className="mt-4 flex justify-center items-center">
          <div className="loader ease-linear rounded-full border-4 border-t-4 border-gray-200 h-6 w-6 mb-4"></div>
          <span className="text-blue-500 ml-2">Procesando...</span>
        </div>
      )}
    </div>
  </div>
</div>

  );
}

export default CommandExecution;
