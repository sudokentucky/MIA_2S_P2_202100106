import { useState, useEffect } from "react";

function useLineCounter(inputText: string) {
  const [lineCount, setLineCount] = useState(1);

  useEffect(() => {
    const lines = inputText.split("\n").length;
    setLineCount(lines);
  }, [inputText]);

  return { lineCount };
}

export default useLineCounter;
