package main

import (
	analyzer "backend/Analyzer" // Importa el paquete "analyzer" desde el directorio "backend/analyzer"
	env "backend/Env"
	structs "backend/Structs"
	commands "backend/commands/Users"
	"backend/globals"
	"encoding/binary"
	"fmt"
	"log" // Importa el paquete "log" para registrar mensajes de error
	"os"
	"strings" // Importa el paquete "strings" para manipulación de cadenas

	"github.com/gofiber/fiber/v2"                 // Importa el paquete Fiber para crear la API
	"github.com/gofiber/fiber/v2/middleware/cors" // Importa el middleware CORS para manejar CORS
)

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

		// Validar campos vacíos
		if req.Username == "" || req.Password == "" || req.ID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"status":  "error",
				"message": "Todos los campos son obligatorios: usuario, contraseña e ID.",
			})
		}

		// Simular la lógica de autenticación (puedes adaptarlo a tu sistema)
		loginCommand := fmt.Sprintf("login -user=%s -pass=%s -id=%s", req.Username, req.Password, req.ID)
		result, err := commands.ParserLogin(strings.Split(loginCommand, " "))
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "error",
				"message": "Usuario o contraseña incorrectos.",
			})
		}

		// Si el login es exitoso
		return c.JSON(fiber.Map{
			"status":  "success",
			"message": result,
		})
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

	app.Post("/disk/read", func(c *fiber.Ctx) error {
		// Estructura para recibir el JSON
		type DiskRequest struct {
			Paths       []string `json:"paths"`       // Lista de rutas de archivos .mia
			IsEncrypted bool     `json:"isEncrypted"` // Indica si están encriptados
			Key         byte     `json:"key"`         // Clave para desencriptar si aplica
		}

		// Crear una instancia de DiskRequest para recibir los datos
		var req DiskRequest

		// Parsear el cuerpo de la solicitud como JSON
		if err := c.BodyParser(&req); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid JSON",
			})
		}

		// Validar que haya al menos un archivo
		if len(req.Paths) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "No se proporcionaron archivos de disco",
			})
		}

		// Lista para acumular la información de todos los discos
		diskInfos := make([]fiber.Map, 0)

		// Procesar cada archivo de disco
		for _, path := range req.Paths {
			// Validar si el archivo existe
			if _, err := os.Stat(path); os.IsNotExist(err) {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": fmt.Sprintf("Archivo %s no encontrado", path),
				})
			}

			// Crear un DiskReader
			diskReader, err := env.NewDiskReader(path, req.IsEncrypted, req.Key)
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": fmt.Sprintf("Error al abrir el disco %s: %v", path, err),
				})
			}
			defer diskReader.Close()

			// Leer el MBR
			mbr, err := diskReader.ReadMBR()
			if err != nil {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": fmt.Sprintf("Error al leer el MBR del disco %s: %v", path, err),
				})
			}

			// Crear un map para devolver la información del MBR y particiones
			mbrInfo := fiber.Map{
				"diskSignature": mbr.MbrDiskSignature,
				"diskSize":      mbr.MbrSize,
				"partitions":    []fiber.Map{},
				"fileName":      path,
			}

			// Recorrer las particiones del MBR
			for i, part := range mbr.MbrPartitions {
				if part.Part_start != -1 {
					partition := fiber.Map{
						"index": i + 1,
						"type":  string(part.Part_type[:]),
						"start": part.Part_start,
						"size":  part.Part_size,
						"name":  string(part.Part_name[:]),
					}

					// Si es extendida, leer particiones lógicas
					if part.Part_type[0] == 'E' {
						partition["logicalPartitions"], err = diskReader.ReadLogicalPartitions(part.Part_start)
						if err != nil {
							return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
								"error": fmt.Sprintf("Error al leer particiones lógicas del disco %s: %v", path, err),
							})
						}
					}

					// Agregar la partición al array de particiones
					mbrInfo["partitions"] = append(mbrInfo["partitions"].([]fiber.Map), partition)
				}
			}

			// Agregar la información del disco a la lista
			diskInfos = append(diskInfos, mbrInfo)
		}

		// Devolver la información de todos los discos procesados
		return c.JSON(fiber.Map{
			"status": "success",
			"disks":  diskInfos,
		})
	})

	// Iniciar el servidor en el puerto 3000
	log.Fatal(app.Listen(":3000"))
}
