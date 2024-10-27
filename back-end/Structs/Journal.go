package structs

import (
	"backend/utils"
	"encoding/binary"
	"fmt"
	"os"
	"strings"
	"time"
)

// Clase que define el Journal en el sistema de archivos ex3, el journal es un log de operaciones que se realizan en el sistema de archivos
type Journal struct {
	J_count   int32       // 4 bytes
	J_content Information // 110 bytes
	// Total: 114 bytes
}

type Information struct {
	I_operation [10]byte // 10 bytes
	I_path      [32]byte // 32 bytes
	I_content   [64]byte // 64 bytes
	I_date      float32  // 4 bytes
	// Total: 110 bytes
}

// Codifica el Journal en un archivo binario usando WriteToFile de utils
func (journal *Journal) Encode(file *os.File, journaling_start int64) error {
	// Calcular la posición en el archivo
	offset := journaling_start + (int64(binary.Size(Journal{})) * int64(journal.J_count))

	// Utilizar WriteToFile de utils para escribir los datos del journal
	err := utils.WriteToFile(file, offset, journal)
	if err != nil {
		return fmt.Errorf("error al escribir el journal en el archivo: %w", err)
	}

	return nil
}

// Decode deserializa el Journal desde un archivo binario usando ReadFromFile de utils
func (journal *Journal) Decode(file *os.File, offset int64) error {
	// Utilizar ReadFromFile de utils para leer los datos del journal
	err := utils.ReadFromFile(file, offset, journal)
	if err != nil {
		return fmt.Errorf("error al leer el journal del archivo: %w", err)
	}

	return nil
}

// PrintJournal imprime en consola la estructura Journal
func (journal *Journal) Print() {
	// Convertir el tiempo de montaje a una fecha
	date := time.Unix(int64(journal.J_content.I_date), 0)

	fmt.Println("Journal:")
	fmt.Printf("J_count: %d\n", journal.J_count)
	fmt.Println("Information:")
	fmt.Printf("I_operation: %s\n", string(journal.J_content.I_operation[:]))
	fmt.Printf("I_path: %s\n", string(journal.J_content.I_path[:]))
	fmt.Printf("I_content: %s\n", string(journal.J_content.I_content[:]))
	fmt.Printf("I_date: %s\n", date.Format(time.RFC3339))
}

// CreateJournalEntry crea una nueva entrada en el journal
func (journal *Journal) CreateJournalEntry(operation string, path string, content string) {
	// Limpiar los buffers
	copy(journal.J_content.I_operation[:], []byte{}) // Limpiar el buffer
	copy(journal.J_content.I_path[:], []byte{})
	copy(journal.J_content.I_content[:], []byte{})

	// Copiar los valores a la estructura del Journal
	copy(journal.J_content.I_operation[:], []byte(operation)) // Copiar la operación
	copy(journal.J_content.I_path[:], []byte(path))           // Copiar la ruta
	copy(journal.J_content.I_content[:], []byte(content))     // Copiar el contenido
	journal.J_content.I_date = float32(time.Now().Unix())     // Copiar la fecha actual
}

// Crear una tabla en formato dot para el journal
func (journal *Journal) GenerateJournalTable(journalIndex int32) string {
	// Convertir fecha a string
	date := time.Unix(int64(journal.J_content.I_date), 0).Format(time.RFC3339)

	// Crear la tabla en formato DOT
	table := fmt.Sprintf(`journal_table_%d [label=<
		<TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0" CELLPADDING="4">
			<TR>
				<TD COLSPAN="2" BGCOLOR="#4CAF50"><FONT COLOR="#FFFFFF">Journal Entry %d</FONT></TD>
			</TR>
			<TR>
				<TD BGCOLOR="#FF7043">Operation:</TD>
				<TD>%s</TD>
			</TR>
			<TR>
				<TD BGCOLOR="#FF7043">Path:</TD>
				<TD>%s</TD>
			</TR>
			<TR>
				<TD BGCOLOR="#FF7043">Content:</TD>
				<TD>%s</TD>
			</TR>
			<TR>
				<TD BGCOLOR="#FF7043">Date:</TD>
				<TD>%s</TD>
			</TR>
		</TABLE>
	>];`, journalIndex, journalIndex, string(journal.J_content.I_operation[:]), string(journal.J_content.I_path[:]), string(journal.J_content.I_content[:]), date)

	return table
}

// GenerateGraph genera el contenido del grafo de las entradas del Journal en formato DOT
func (journal *Journal) GenerateGraph(journalStart int64, journalCount int32, file *os.File) (string, error) {
	dotContent := ""                           // Inicia el contenido del DOT
	entrySize := int64(binary.Size(Journal{})) // Tamaño de cada entrada de Journal

	fmt.Println("Generando grafo de Journal...")

	for i := int32(0); i < journalCount; i++ {
		offset := journalStart + int64(i)*entrySize
		fmt.Printf("Leyendo entrada del Journal en offset: %d\n", offset)

		// Leer el journal desde la posición correspondiente
		err := journal.Decode(file, offset)
		if err != nil {
			return "", fmt.Errorf("error al deserializar el journal %d en offset %d: %v", i, offset, err)
		}

		// Verificar si la entrada del journal es válida (si tiene una operación registrada)
		operation := strings.TrimSpace(string(journal.J_content.I_operation[:]))
		if operation == "" {
			fmt.Printf("Entrada de Journal vacía encontrada en índice %d, deteniendo la lectura.\n", i)
			break
		}

		fmt.Printf("Generando tabla para la entrada de Journal %d con operación: %s\n", i, operation)
		// Generar la tabla para cada entrada del journal utilizando el método GenerateJournalTable
		dotContent += journal.GenerateJournalTable(i)
	}

	return dotContent, nil
}

// SaveJournalEntry guarda una nueva entrada en el journal y la serializa en el archivo
func (journal *Journal) SaveJournalEntry(file *os.File, journaling_start int64, operation string, path string, content string) error {
	// Crear la nueva entrada en el Journal
	journal.CreateJournalEntry(operation, path, content)

	// Serializar el Journal en el archivo
	err := journal.Encode(file, journaling_start)
	if err != nil {
		return fmt.Errorf("error al guardar la entrada de journal: %w", err)
	}

	return nil
}
