package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

func main() {
	app := fiber.New()

	// Tambahan CORS: Biar aman pas nanti diakses dari Make.com / web lain
	app.Use(cors.New())

	// --- ROUTE TES (Biar lu bisa cek langsung di Browser) ---
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("🔥 Server Proteksi Ebook Aktif dan Berjalan Mantap!")
	})

	// --- ROUTE UTAMA ---
	app.Post("/protect", func(c *fiber.Ctx) error {
		// 1. Tangkap data JSON dari Postman / Make.com
		type Request struct {
			Email       string `json:"email"`
			ProductName string `json:"product_name"`
		}
		req := new(Request)
		if err := c.BodyParser(req); err != nil {
			log.Println("Error parsing JSON:", err)
			return c.Status(400).JSON(fiber.Map{"error": "Format data JSON tidak valid"})
		}

		// 2. Bikin folder "output" otomatis di server biar nggak error
		if err := os.MkdirAll("./output", os.ModePerm); err != nil {
			log.Println("Gagal bikin folder output:", err)
		}

		// 3. Tentukan path file
		inputPath := fmt.Sprintf("./master/%s.pdf", req.ProductName)
		outputPath := fmt.Sprintf("./output/%s_protected_%s.pdf", req.ProductName, req.Email)

		// 4. Cek apakah file ebook master beneran ada
		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			log.Printf("⚠️ File master tidak ditemukan: %s\n", inputPath)
			// Gua ganti error code jadi 400 biar beda sama 404 salah alamat URL
			return c.Status(400).JSON(fiber.Map{
				"error":  "Ebook tidak ditemukan di folder master",
				"dicari": inputPath, // Biar lu tau sistem nyari file dengan nama apa
			})
		}

		// 5. Konfigurasi Proteksi
		conf := model.NewAESConfiguration(req.Email, "admin_super_secret", 256)
		conf.Permissions = model.PermissionsNone

		// 6. Eksekusi Enkripsi
		err := api.EncryptFile(inputPath, outputPath, conf)
		if err != nil {
			log.Println("Gagal enkripsi file PDF:", err)
			return c.Status(500).JSON(fiber.Map{"error": "Server gagal mengunci file PDF"})
		}

		log.Printf("✅ Sukses: %s untuk %s\n", req.ProductName, req.Email)

		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "File berhasil diproteksi",
			"file":    outputPath,
		})
	})

	// --- SETUP PORT UNTUK RAILWAY ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000" // Default kalau lu tes di laptop
	}

	log.Printf("🚀 Menjalankan server di port %s...\n", port)
	log.Fatal(app.Listen(":" + port))
}
