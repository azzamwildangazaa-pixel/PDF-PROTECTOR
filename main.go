package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

func main() {
	app := fiber.New()

	// Endpoint untuk memproses proteksi PDF
	app.Post("/protect", func(c *fiber.Ctx) error {
		// 1. Tangkap data dari request (dikirim oleh Make.com nanti)
		type Request struct {
			Email       string `json:"email"`
			ProductName string `json:"product_name"` // Harus sama dengan nama file di folder master
		}
		req := new(Request)
		if err := c.BodyParser(req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Data tidak valid"})
		}

		// 2. Tentukan path file
		inputPath := fmt.Sprintf("./master/%s.pdf", req.ProductName)
		outputPath := fmt.Sprintf("./output/%s_protected_%s.pdf", req.ProductName, req.Email)

		// Cek apakah file master ada
		if _, err := os.Stat(inputPath); os.IsNotExist(err) {
			return c.Status(404).JSON(fiber.Map{"error": "Ebook tidak ditemukan di folder master"})
		}

		// 3. Konfigurasi Proteksi
		// Password User = Email Pembeli
		// Password Admin = "rahasia123" (ganti sesukamu)
		conf := model.NewAESConfiguration(req.Email, "admin_super_secret", 256)

		// Batasi akses (biar gak bisa di-copy atau di-print)
		conf.Permissions = model.PermissionsNone

		// 4. Eksekusi Enkripsi
		err := api.EncryptFile(inputPath, outputPath, conf)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "Gagal mengunci file"})
		}

		fmt.Printf("✅ Berhasil memproses: %s untuk %s\n", req.ProductName, req.Email)

		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "File berhasil diproteksi",
			"file":    outputPath,
		})
	})

	// Jalankan server di port 3000
	log.Fatal(app.Listen(":3000"))
}
