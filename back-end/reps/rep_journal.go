package reps

import (
	structs "backend/Structs"
	"backend/utils"
	"encoding/binary"
	"fmt"
	"os"
	"os/exec"
)

// ReportJournal genera un reporte de las entradas del Journal y lo guarda en la ruta especificada
func ReportJournal(superblock *structs.Superblock, diskPath string, path string) error {
	// Calcular el inicio del Journal solo si es EXT3
	var journalStart int64
	var superblockSize int64 = int64(binary.Size(superblock))
	if superblock.S_filesystem_type == 3 {
		journalStart = superblockSize
	} else {
		return fmt.Errorf("el sistema de archivos no soporta Journaling (EXT2)")
	}

	// Crear las carpetas padre si no existen
	err := utils.CreateParentDirs(path)
	if err != nil {
		return fmt.Errorf("error al crear directorios: %v", err)
	}

	// Abrir el archivo de disco
	file, err := os.Open(diskPath)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo de disco: %v", err)
	}
	defer file.Close()

	// Obtener el nombre base del archivo sin la extensión
	dotFileName, outputImage := utils.GetFileNames(path)

	// Crear una instancia de Journal
	journal := &structs.Journal{}

	// Iniciar el contenido DOT
	dotContent := initJournalDotGraph()

	// Utilizar el método GenerateGraph del Journal
	dotGraph, err := journal.GenerateGraph(journalStart, journal.J_count, file)
	if err != nil {
		return err
	}

	// Añadir el contenido generado al DOT
	dotContent += dotGraph
	dotContent += "}" // Fin del Dot

	// Crear el archivo DOT
	err = writeDotFile(dotFileName, dotContent)
	if err != nil {
		return err
	}

	// Ejecutar Graphviz para generar la imagen
	err = generateJournalImage(dotFileName, outputImage)
	if err != nil {
		return err
	}

	fmt.Println("Imagen del Journal generada:", outputImage)
	return nil
}

// initJournalDotGraph inicializa el contenido básico del archivo DOT para el Journal
func initJournalDotGraph() string {
	return `digraph G {
		fontname="Helvetica,Arial,sans-serif"
		node [fontname="Helvetica,Arial,sans-serif", shape=plain, fontsize=12];
		edge [fontname="Helvetica,Arial,sans-serif", color="#FF7043", arrowsize=0.8];
		rankdir=TB;
		bgcolor="#FAFAFA";
		node [shape=plaintext];
		journalHeaderColor="#4CAF50"; 
		cellBackgroundColor="#FFFDE7";
		cellBorderColor="#EEEEEE";
		textColor="#263238";
	`
}

// generateJournalImage genera una imagen a partir del archivo DOT usando Graphviz
func generateJournalImage(dotFileName string, outputImage string) error {
	cmd := exec.Command("dot", "-Tpng", dotFileName, "-o", outputImage)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("error al ejecutar Graphviz: %v", err)
	}

	return nil
}
