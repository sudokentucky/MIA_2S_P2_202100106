package main

import (
	analyzer "backend/Analyzer" // Importa el paquete "analyzer" desde el directorio "backend/analyzer"
	structs "backend/Structs"
	commands "backend/commands"
	usercommands "backend/commands/Users"
	"backend/globals"
	"encoding/binary"
	"fmt"
	"log" // Importa el paquete "log" para registrar mensajes de error
	"os"
	"strings" // Importa el paquete "strings" para manipulación de cadenas

	"github.com/gofiber/fiber/v2"                 // Importa el paquete Fiber para crear la API
	"github.com/gofiber/fiber/v2/middleware/cors" // Importa el middleware CORS para manejar CORS
)

var diskManager = commands.NewDiskManager() // Crear una nueva instancia de DiskManager
func main() {
	// Crear una nueva instancia de Fiber
	app := fiber.New()

	// Configurar el middleware CORS
	app.Use(cors.New())

	// Definir la ruta POST para recibir el comando del usuario
	app.Post("/analyze", func(c *fiber.Ctx) error {
		// Estructura para recibir el JSON
		type Request struct {
			Command string `json:"command"`
		}

		// Crear una instancia de Request
		var req Request

		// Parsear el cuerpo de la solicitud como JSON
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid JSON",
			})
		}

		// Obtener el comando del cuerpo de la solicitud
		input := req.Command
		fmt.Println("input: ", input)

		// Separar el comando en líneas
		lines := strings.Split(input, "\n")

		// Lista para acumular los resultados de salida
		var results []string

		// Analizar cada línea
		for _, line := range lines {
			// Ignorar líneas vacías
			if strings.TrimSpace(line) == "" {
				continue
			}

			// Llamar a la función Analyzer del paquete analyzer para analizar la línea
			result, err := analyzer.Analyzer(line)
			if err != nil {
				// Si hay un error, almacenar el mensaje de error en lugar del resultado
				result = fmt.Sprintf("Error: %s", err.Error())
			}

			// Acumular los resultados
			results = append(results, result)
		}

		// Devolver una respuesta JSON con la lista de resultados
		return c.JSON(fiber.Map{
			"results": results,
		})
	})

	// Endpoint para obtener las particiones de un disco
	app.Post("/api/disk/partitions", func(c *fiber.Ctx) error {
		type DiskRequest struct {
			Path string `json:"path"`
		}

		var req DiskRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Formato JSON inválido",
			})
		}

		if req.Path == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Debe proporcionar una ruta válida al archivo de disco",
			})
		}

		// Cargar el disco usando DiskManager
		err := diskManager.LoadDisk(req.Path)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Error al cargar el disco: %v", err),
			})
		}

		// Obtener el MBR del disco cargado
		mbr, exists := diskManager.PartitionMBRs[req.Path]
		if !exists {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "No se encontró el MBR para el disco proporcionado",
			})
		}

		// Obtener la lista de particiones usando ListPartitions
		partitions := mbr.ListPartitions()

		return c.JSON(fiber.Map{
			"partitions": partitions,
		})
	})

	app.Post("/api/disk/partition/tree", func(c *fiber.Ctx) error {
		// Estructura para recibir el JSON
		type PartitionRequest struct {
			DiskPath      string `json:"diskPath"`
			PartitionName string `json:"partitionName"`
		}

		// Parsear el cuerpo de la solicitud
		var req PartitionRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Formato JSON inválido",
			})
		}

		// Verificar que se proporcionó la ruta del disco y el nombre de la partición
		if req.DiskPath == "" || req.PartitionName == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Debe proporcionar 'diskPath' y 'partitionName'",
			})
		}

		// Verificar si el archivo existe
		_, err := os.Stat(req.DiskPath)
		if os.IsNotExist(err) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "El archivo especificado no existe",
			})
		} else if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Error al verificar el archivo: %v", err),
			})
		}

		// Crear una nueva instancia de DirectoryTreeService
		treeService, err := commands.NewDirectoryTreeService()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Error al inicializar el servicio del árbol de directorios: %v", err),
			})
		}
		defer treeService.Close()

		// Obtener el árbol de directorios de la partición
		tree, err := treeService.GetDirectoryTree("/")
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Error al obtener el árbol de directorios: %v", err),
			})
		}

		// Devolver el árbol en formato JSON
		return c.JSON(fiber.Map{
			"tree": tree,
		})
	})

	// Definir la ruta POST para el inicio de sesión de los usuarios
	app.Post("/users/login", func(c *fiber.Ctx) error {
		type LoginRequest struct {
			Username string `json:"username"`
			Password string `json:"password"`
			ID       string `json:"id"`
		}

		var req LoginRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": "Invalid JSON format",
			})
		}

		// Crear el comando de login sin verificar nada en este nivel
		loginCommand := fmt.Sprintf("login -user=%s -pass=%s -id=%s", req.Username, req.Password, req.ID)

		// Pasar los parámetros directamente al comando ParserLogin
		result, _ := usercommands.ParserLogin(strings.Split(loginCommand, " "))

		// Retornar directamente el resultado del comando, que ya maneja los errores
		return c.JSON(result)
	})

	// Definir la ruta GET para retornar la lista de usuarios y grupos
	app.Get("/list-users-groups", func(c *fiber.Ctx) error {
		// Verificar si hay una sesión activa
		if !globals.IsLoggedIn() {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "No hay ninguna sesión activa",
			})
		}

		// Verificar que la partición esté montada
		_, path, err := globals.GetMountedPartition(globals.UsuarioActual.Id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": fmt.Sprintf("No se puede encontrar la partición montada: %v", err),
			})
		}

		// Abrir el archivo de la partición
		file, err := os.OpenFile(path, os.O_RDWR, 0755)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": fmt.Sprintf("No se puede abrir el archivo de la partición: %v", err),
			})
		}
		defer file.Close()

		// Cargar el Superblock y la partición
		_, sb, _, err := globals.GetMountedPartitionRep(globals.UsuarioActual.Id)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": fmt.Sprintf("No se pudo cargar el Superblock: %v", err),
			})
		}

		// Leer el inodo de users.txt (asumimos que es el segundo inodo)
		var inode structs.Inode
		inodeOffset := int64(sb.S_inode_start + int32(binary.Size(inode))) // Calcular el offset del inodo de users.txt
		err = inode.Decode(file, inodeOffset)                              // Decodificar el inodo
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": fmt.Sprintf("Error leyendo el inodo de users.txt: %v", err),
			})
		}

		// Obtener la lista de usuarios y grupos desde users.txt
		data, err := globals.ListUsersAndGroups(file, sb, &inode) // Pasar el puntero de inode
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"status":  "error",
				"message": fmt.Sprintf("Error listando usuarios y grupos: %v", err),
			})
		}

		// Devolver la lista de usuarios y grupos en formato JSON
		return c.JSON(data)
	})

	// Definir la ruta GET para verificar si hay un usuario logueado
	app.Get("/users/logged-in", func(c *fiber.Ctx) error {
		// Verificar si hay una sesión activa
		if globals.UsuarioActual == nil || !globals.UsuarioActual.Status {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "No hay ningún usuario logueado.",
			})
		}

		// Si hay un usuario logueado, devolver su información
		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Usuario logueado.",
			"user":    globals.UsuarioActual.Name,
			"id":      globals.UsuarioActual.Id,
		})
	})

	//Verificar si al menos una partición está montada
	app.Get("/check-partition", func(c *fiber.Ctx) error {
		// Verificar si hay al menos una partición montada
		if len(globals.MountedPartitions) == 0 {
			return c.Status(fiber.StatusOK).JSON(fiber.Map{
				"status":  "error",
				"message": "No hay particiones montadas. Debe montar una partición antes de iniciar sesión.",
			})
		}

		// Si hay particiones montadas, devolver un mensaje de éxito
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"status":  "success",
			"message": fmt.Sprintf("Hay %d particiones montadas. Puede proceder con el login.", len(globals.MountedPartitions)),
		})
	})

	// Iniciar el servidor en el puerto 3000
	log.Fatal(app.Listen(":3000"))
}
