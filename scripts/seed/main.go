package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/argon2"
)

type seedUser struct {
	Email, Password, Name, Role string
}

type seedProposal struct {
	ApplicantEmail, Title, Description, Organization, Status string
	NominalAmount                                            float64
}

type seedReview struct {
	ProposalTitle string
	ReviewerEmail string
	Score         int
	Comment       string
	Approve       bool
	Reject        bool
}

type seedDocument struct {
	ProposalTitle string
	Filename      string
	MimeType      string
}

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:postgres@localhost:5432/smart_grant?sslmode=disable"
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		fmt.Println("Error connecting to database:", err)
		os.Exit(1)
	}
	defer pool.Close()

	fmt.Println("Connected to PostgreSQL")
	fmt.Println("Seeding database...")

	clearAll(ctx, pool)

	users := seedUsers(ctx, pool)
	fmt.Printf("  users: %d created\n", len(users))

	proposals := seedProposals(ctx, pool, users)
	fmt.Printf("  proposals: %d created\n", len(proposals))

	docs := seedDocuments(ctx, pool, proposals)
	fmt.Printf("  documents: %d created\n", len(docs))

	reviews := seedReviews(ctx, pool, proposals, users)
	fmt.Printf("  reviews: %d created\n", len(reviews))

	totalScore := seedRiskScores(ctx, pool, proposals)
	fmt.Printf("  risk_scores: %d created\n", totalScore)

	fmt.Println("Done. Database seeded successfully.")
}

func clearAll(ctx context.Context, pool *pgxpool.Pool) {
	tables := []string{"notifications", "audit_logs", "risk_scores", "proposal_versions", "proposal_documents", "reviews", "proposals", "users"}
	for _, t := range tables {
		pool.Exec(ctx, "DELETE FROM "+t)
	}
}

func seedUsers(ctx context.Context, pool *pgxpool.Pool) map[string]string {
	list := []seedUser{
		{Email: "admin@example.com", Password: "password123", Name: "Admin Utama", Role: "admin"},
		{Email: "reviewer1@example.com", Password: "password123", Name: "Budi Santoso", Role: "reviewer"},
		{Email: "reviewer2@example.com", Password: "password123", Name: "Siti Rahayu", Role: "reviewer"},
		{Email: "reviewer3@example.com", Password: "password123", Name: "Rudi Hermawan", Role: "reviewer"},
		{Email: "applicant1@example.com", Password: "password123", Name: "Andi Pratama", Role: "applicant"},
		{Email: "applicant2@example.com", Password: "password123", Name: "Dewi Lestari", Role: "applicant"},
		{Email: "applicant3@example.com", Password: "password123", Name: "Bambang Susilo", Role: "applicant"},
		{Email: "applicant4@example.com", Password: "password123", Name: "Ratna Sari", Role: "applicant"},
	}

	insertSQL := `INSERT INTO users (email, password_hash, name, role) VALUES ($1, $2, $3, $4) RETURNING id`
	result := make(map[string]string)

	for _, u := range list {
		hash := hashPassword(u.Password)
		var id string
		if err := pool.QueryRow(ctx, insertSQL, u.Email, hash, u.Name, u.Role).Scan(&id); err != nil {
			fmt.Println("  Error creating user", u.Email, ":", err)
			continue
		}
		result[u.Email] = id
	}
	return result
}

func seedProposals(ctx context.Context, pool *pgxpool.Pool, users map[string]string) []seedProposal {
	list := []seedProposal{
		{ApplicantEmail: "applicant1@example.com", Title: "Pengembangan Sistem AI untuk Deteksi Dini Penyakit", Description: "Pengembangan sistem kecerdasan buatan untuk mendeteksi penyakit secara dini menggunakan citra medis dan data pasien. Proyek ini mencakup pengembangan model deep learning, dataset pelatihan, dan platform berbasis web.", Organization: "AI Health Lab", Status: "submitted", NominalAmount: 750000000},
		{ApplicantEmail: "applicant1@example.com", Title: "Peningkatan Kapasitas Peternakan Sapi Organik", Description: "Program peningkatan kapasitas peternakan sapi organik berbasis masyarakat di 5 desa. Mencakup pelatihan, pembangunan infrastruktur kandang, dan sistem distribusi.", Organization: "Ternak Maju Bersama", Status: "approved", NominalAmount: 350000000},
		{ApplicantEmail: "applicant2@example.com", Title: "Program Literasi Digital untuk 1000 Pelajar", Description: "Program literasi digital untuk 1000 pelajar sekolah menengah di daerah terpencil. Termasuk pengadaan perangkat, pelatihan guru, dan modul pembelajaran.", Organization: "Edukasi Kita Foundation", Status: "draft", NominalAmount: 150000000},
		{ApplicantEmail: "applicant2@example.com", Title: "Riset Material Baterai Ramah Lingkungan", Description: "Penelitian dan pengembangan material baterai berbasis natrium sebagai alternatif lithium yang lebih ramah lingkungan dan murah. Meliputi sintesis material, pengujian, dan prototyping.", Organization: "Green Energy Research", Status: "submitted", NominalAmount: 1200000000},
		{ApplicantEmail: "applicant1@example.com", Title: "Pembangunan Infrastruktur Air Bersih Desa", Description: "Pembangunan sistem penyediaan air bersih untuk 3 desa di daerah kering. Mencakup pengeboran sumur dalam, jaringan pipa, dan penampungan air.", Organization: "Desa Sejahtera Foundation", Status: "rejected", NominalAmount: 2500000000},
		{ApplicantEmail: "applicant3@example.com", Title: "Platform E-Commerce untuk UMKM Lokal", Description: "Pengembangan platform e-commerce khusus untuk UMKM lokal dengan fitur manajemen inventori, pembayaran digital, dan logistik terintegrasi.", Organization: "UMKM Go Digital", Status: "approved", NominalAmount: 500000000},
		{ApplicantEmail: "applicant1@example.com", Title: "Pelatihan Coding untuk 500 Remaja Putus Sekolah", Description: "Program pelatihan coding dan keterampilan digital untuk 500 remaja putus sekolah. Kurikulum mencakup web development, mobile apps, dan dasar-dasar AI.", Organization: "Yayasan Edukasi Digital", Status: "submitted", NominalAmount: 250000000},
		{ApplicantEmail: "applicant2@example.com", Title: "Pengolahan Limbah Plastik Menjadi Bahan Bakar", Description: "Proyek pengolahan limbah plastik menjadi bahan bakar alternatif menggunakan teknologi pirolisis. Termasuk pembangunan fasilitas pengolahan dan sosialisasi masyarakat.", Organization: "Eco Waste Solution", Status: "in_review", NominalAmount: 1800000000},
		{ApplicantEmail: "applicant4@example.com", Title: "Aplikasi Telemedicine untuk Daerah Terpencil", Description: "Pengembangan aplikasi telemedicine yang dapat diakses tanpa koneksi internet stabil. Fitur meliputi konsultasi dokter, resep digital, dan monitoring kesehatan.", Organization: "TeleMedika Nusantara", Status: "submitted", NominalAmount: 900000000},
		{ApplicantEmail: "applicant3@example.com", Title: "Sistem Irigasi Cerdas Berbasis IoT", Description: "Pengembangan sistem irigasi otomatis berbasis IoT yang memonitor kelembaban tanah, cuaca, dan kebutuhan air tanaman secara real-time.", Organization: "AgriTech Indonesia", Status: "draft", NominalAmount: 400000000},
		{ApplicantEmail: "applicant4@example.com", Title: "Budidaya Rumput Laut Berkelanjutan", Description: "Program budidaya rumput laut berbasis masyarakat pesisir dengan teknik berkelanjutan. Mencakup pelatihan, bibit unggul, dan fasilitas pengolahan.", Organization: "Laut Lestari", Status: "submitted", NominalAmount: 200000000},
		{ApplicantEmail: "applicant3@example.com", Title: "Pemasangan Solar Panel untuk 10 Desa Terpencil", Description: "Pemasangan panel surya untuk elektrifikasi 10 desa terpencil yang belum terjangkau jaringan listrik nasional. Kapasitas total 500 kWp.", Organization: "Energi Hijau Nusantara", Status: "rejected", NominalAmount: 3000000000},
		{ApplicantEmail: "applicant2@example.com", Title: "Produksi Film Edukasi untuk 1000 Sekolah", Description: "Produksi 20 film edukasi pendek tentang sains, teknologi, dan lingkungan untuk didistribusikan ke 1000 sekolah dasar.", Organization: "Edukasi Kita Foundation", Status: "draft", NominalAmount: 100000000},
		{ApplicantEmail: "applicant1@example.com", Title: "Robotik untuk Pendidikan STEM", Description: "Pengembangan kit robotik edukasi untuk pembelajaran STEM di sekolah menengah. Termasuk modul, sensor, dan platform coding visual.", Organization: "STEM Edu Lab", Status: "submitted", NominalAmount: 600000000},
		{ApplicantEmail: "applicant4@example.com", Title: "Aplikasi Bank Sampah Digital", Description: "Platform digital untuk mengelola bank sampah berbasis komunitas. Fitur: pencatatan setor, penukaran poin, dan penjadwalan jemput sampah.", Organization: "Lingkungan Bersih", Status: "approved", NominalAmount: 180000000},
		{ApplicantEmail: "applicant3@example.com", Title: "Pengembangan Wisata Alam Berkelanjutan", Description: "Pengembangan destinasi wisata alam berbasis masyarakat dengan konsep ekowisata berkelanjutan. Infrastruktur dasar, pelatihan pemandu, dan promosi digital.", Organization: "Eco Tourism ID", Status: "submitted", NominalAmount: 850000000},
		{ApplicantEmail: "applicant1@example.com", Title: "Aplikasi Belajar Bahasa Daerah untuk Anak", Description: "Aplikasi pembelajaran interaktif untuk 10 bahasa daerah Indonesia yang terancam punah. Target pengguna: anak usia 5-12 tahun.", Organization: "Yayasan Edukasi Digital", Status: "draft", NominalAmount: 120000000},
		{ApplicantEmail: "applicant2@example.com", Title: "Pengolahan Sampah Organik Menjadi Kompos", Description: "Pembangunan fasilitas pengolahan sampah organik skala kota menjadi pupuk kompos berkualitas tinggi. Kapasitas 10 ton/hari.", Organization: "Eco Waste Solution", Status: "submitted", NominalAmount: 450000000},
		{ApplicantEmail: "applicant4@example.com", Title: "Perpustakaan Keliling Digital", Description: "Program perpustakaan keliling dengan tablet dan buku digital untuk anak-anak di daerah terpencil. Mencakup 50 titik layanan.", Organization: "Lingkungan Bersih", Status: "approved", NominalAmount: 75000000},
		{ApplicantEmail: "applicant3@example.com", Title: "Pusat Inovasi dan Kreativitas Desa", Description: "Pembangunan pusat inovasi desa yang dilengkapi maker space, laboratorium komputer, dan ruang kolaborasi untuk pemuda desa.", Organization: "Energi Hijau Nusantara", Status: "submitted", NominalAmount: 1500000000},
	}

	insertSQL := `INSERT INTO proposals (applicant_id, title, description, nominal_amount, organization, status) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	var result []seedProposal

	for _, p := range list {
		userID := users[p.ApplicantEmail]
		var id string
		now := time.Now().Add(-time.Duration(len(result)) * time.Hour)
		if err := pool.QueryRow(ctx, insertSQL, userID, p.Title, p.Description, p.NominalAmount, p.Organization, p.Status).Scan(&id); err != nil {
			fmt.Println("  Error creating proposal:", err)
			continue
		}
		if p.Status != "draft" {
			pool.Exec(ctx, "UPDATE proposals SET created_at = $1, updated_at = $1 WHERE id = $2", now, id)
		}
		p.Title = id
		result = append(result, p)
	}
	return result
}

func seedDocuments(ctx context.Context, pool *pgxpool.Pool, proposals []seedProposal) []seedDocument {
	docConfigs := map[string][]string{
		"Pengembangan Sistem AI untuk Deteksi Dini Penyakit":     {"proposal.pdf", "anggaran.xlsx", "timeline.docx"},
		"Peningkatan Kapasitas Peternakan Sapi Organik":           {"proposal.pdf", "rab.xlsx"},
		"Riset Material Baterai Ramah Lingkungan":                 {"proposal.pdf", "laporan_teknis.pdf"},
		"Pelatihan Coding untuk 500 Remaja Putus Sekolah":         {"proposal.pdf"},
		"Aplikasi Telemedicine untuk Daerah Terpencil":            {"proposal.pdf", "spesifikasi_teknis.pdf"},
		"Platform E-Commerce untuk UMKM Lokal":                    {"proposal.pdf", "analisis_pasar.pdf"},
		"Robotik untuk Pendidikan STEM":                           {"proposal.pdf", "kurikulum.pdf"},
		"Pengolahan Sampah Organik Menjadi Kompos":                {"proposal.pdf"},
	}

	insertSQL := `INSERT INTO proposal_documents (proposal_id, filename, file_url, mime_type, file_size) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	var result []seedDocument

	for _, p := range proposals {
		files, ok := docConfigs[shortTitle(p.Title)]
		if !ok {
			continue
		}
		for _, f := range files {
			url := fmt.Sprintf("http://localhost:9000/smart-grant/proposals/%s/%s", p.Title, f)
			mime := "application/pdf"
			if strings.HasSuffix(f, ".xlsx") {
				mime = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
			} else if strings.HasSuffix(f, ".docx") {
				mime = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
			}
			var docID string
			if err := pool.QueryRow(ctx, insertSQL, p.Title, f, url, mime, int64(500*1024)).Scan(&docID); err == nil {
				result = append(result, seedDocument{})
			}
		}
	}

	return result
}

func seedReviews(ctx context.Context, pool *pgxpool.Pool, proposals []seedProposal, users map[string]string) []seedReview {
	list := []seedReview{
		{ProposalTitle: "Pengembangan Sistem AI untuk Deteksi Dini Penyakit", ReviewerEmail: "reviewer1@example.com", Score: 85, Comment: "Proposal sangat baik, metodologi jelas, dan dampak yang diharapkan terukur. Tim memiliki pengalaman yang relevan.", Approve: false},
		{ProposalTitle: "Peningkatan Kapasitas Peternakan Sapi Organik", ReviewerEmail: "reviewer1@example.com", Score: 90, Comment: "Program tepat sasaran, anggaran realistis, dan memiliki potensi keberlanjutan. Rekomendasi: APPROVE.", Approve: true},
		{ProposalTitle: "Riset Material Baterai Ramah Lingkungan", ReviewerEmail: "reviewer2@example.com", Score: 72, Comment: "Ide bagus dengan potensi besar, namun detail teknis perlu dilengkapi. Anggaran agak membengkak di sisi peralatan.", Approve: false},
		{ProposalTitle: "Platform E-Commerce untuk UMKM Lokal", ReviewerEmail: "reviewer3@example.com", Score: 88, Comment: "Layak didanai. Tim solid, pasar jelas, dan dampak ekonomi langsung terasa.", Approve: true},
		{ProposalTitle: "Pelatihan Coding untuk 500 Remaja Putus Sekolah", ReviewerEmail: "reviewer1@example.com", Score: 78, Comment: "Program bagus dengan dampak sosial tinggi. Perlu diperjelas mekanisme evaluasi keberhasilan.", Approve: false},
		{ProposalTitle: "Pengolahan Limbah Plastik Menjadi Bahan Bakar", ReviewerEmail: "reviewer2@example.com", Score: 60, Comment: "Konsep menarik, tetapi analisis risiko teknis dan lingkungan masih kurang mendalam.", Approve: false},
		{ProposalTitle: "Aplikasi Telemedicine untuk Daerah Terpencil", ReviewerEmail: "reviewer3@example.com", Score: 82, Comment: "Sangat relevan dengan kebutuhan saat ini. Perlu dipastikan konektivitas di daerah target.", Approve: false},
		{ProposalTitle: "Budidaya Rumput Laut Berkelanjutan", ReviewerEmail: "reviewer1@example.com", Score: 75, Comment: "Proposal cukup baik, perlu tambahan data tentang potensi pasar.", Approve: false},
		{ProposalTitle: "Pemasangan Solar Panel untuk 10 Desa Terpencil", ReviewerEmail: "reviewer2@example.com", Score: 45, Comment: "Biaya per desa terlalu tinggi. Perlu evaluasi ulang desain teknis.", Approve: false, Reject: true},
		{ProposalTitle: "Robotik untuk Pendidikan STEM", ReviewerEmail: "reviewer3@example.com", Score: 80, Comment: "Proposal inovatif dengan kurikulum yang terstruktur. Layak didanai.", Approve: false},
		{ProposalTitle: "Aplikasi Bank Sampah Digital", ReviewerEmail: "reviewer1@example.com", Score: 92, Comment: "Salah satu proposal terbaik. Dampak lingkungan dan sosial sangat signifikan.", Approve: true},
		{ProposalTitle: "Pengembangan Wisata Alam Berkelanjutan", ReviewerEmail: "reviewer2@example.com", Score: 70, Comment: "Potensi bagus, namun perlu kajian dampak lingkungan yang lebih komprehensif.", Approve: false},
	}

	insertSQL := `INSERT INTO reviews (proposal_id, reviewer_id, score, comment, status) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	var result []seedReview

	for _, r := range list {
		proposalID := findProposalID(proposals, r.ProposalTitle)
		if proposalID == "" {
			continue
		}
		reviewerID := users[r.ReviewerEmail]
		status := "pending"
		if r.Approve || r.Reject {
			status = "approved"
		}
		var id string
		if err := pool.QueryRow(ctx, insertSQL, proposalID, reviewerID, r.Score, r.Comment, status).Scan(&id); err != nil {
			continue
		}
		if r.Approve {
			pool.Exec(ctx, "UPDATE proposals SET status = 'approved', updated_at = now() WHERE id = $1", proposalID)
		}
		if r.Reject {
			pool.Exec(ctx, "UPDATE proposals SET status = 'rejected', updated_at = now() WHERE id = $1", proposalID)
		}
		result = append(result, r)
	}

	return result
}

func seedRiskScores(ctx context.Context, pool *pgxpool.Pool, proposals []seedProposal) int {
	type riskData struct {
		level  string
		conf   float64
		freq   float64
		comp   float64
	}
	riskMap := map[string]riskData{
		"Pengembangan Sistem AI untuk Deteksi Dini Penyakit":     {"low", 0.85, 0, 1.0},
		"Peningkatan Kapasitas Peternakan Sapi Organik":          {"low", 0.80, 1, 0.8},
		"Program Literasi Digital untuk 1000 Pelajar":           {"low", 0.75, 0, 1.0},
		"Riset Material Baterai Ramah Lingkungan":                {"medium", 0.70, 2, 0.6},
		"Pembangunan Infrastruktur Air Bersih Desa":             {"high", 0.65, 3, 0.3},
		"Platform E-Commerce untuk UMKM Lokal":                   {"low", 0.90, 0, 1.0},
		"Pelatihan Coding untuk 500 Remaja Putus Sekolah":        {"low", 0.82, 0, 0.7},
		"Pengolahan Limbah Plastik Menjadi Bahan Bakar":          {"medium", 0.68, 2, 0.5},
		"Aplikasi Telemedicine untuk Daerah Terpencil":          {"low", 0.78, 0, 0.9},
		"Sistem Irigasi Cerdas Berbasis IoT":                     {"low", 0.85, 0, 1.0},
		"Budidaya Rumput Laut Berkelanjutan":                     {"low", 0.80, 0, 0.8},
		"Pemasangan Solar Panel untuk 10 Desa Terpencil":        {"high", 0.70, 1, 0.2},
		"Produksi Film Edukasi untuk 1000 Sekolah":               {"low", 0.75, 0, 1.0},
		"Robotik untuk Pendidikan STEM":                          {"medium", 0.72, 2, 0.6},
		"Aplikasi Bank Sampah Digital":                           {"low", 0.88, 0, 1.0},
		"Pengembangan Wisata Alam Berkelanjutan":                 {"medium", 0.74, 1, 0.5},
		"Aplikasi Belajar Bahasa Daerah untuk Anak":              {"low", 0.80, 0, 1.0},
		"Pengolahan Sampah Organik Menjadi Kompos":               {"low", 0.82, 1, 0.9},
		"Perpustakaan Keliling Digital":                          {"low", 0.90, 0, 1.0},
		"Pusat Inovasi dan Kreativitas Desa":                     {"medium", 0.71, 2, 0.4},
	}

	insertSQL := `INSERT INTO risk_scores (proposal_id, risk_level, confidence, features, details, model_version) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	count := 0

	for _, p := range proposals {
		rd, ok := riskMap[shortTitle(p.Title)]
		if !ok {
			continue
		}
		features := fmt.Sprintf(`{"nominal_amount":%.0f,"funding_frequency_30d":%.0f,"document_completeness":%.1f}`, p.NominalAmount, rd.freq, rd.comp)
		details := fmt.Sprintf(`{"tree_depth":4}`)
		var scoreID string
		if err := pool.QueryRow(ctx, insertSQL, p.Title, rd.level, rd.conf, features, details, "c4.5-v2").Scan(&scoreID); err == nil {
			count++
		}
	}

	return count
}

func findProposalID(proposals []seedProposal, title string) string {
	for _, p := range proposals {
		if shortTitle(p.Title) == shortTitle(title) {
			return p.Title
		}
	}
	return ""
}

func shortTitle(title string) string {
	return title
}

func hashPassword(password string) string {
	salt := make([]byte, 16)
	rand.Read(salt)
	hash := argon2.IDKey([]byte(password), salt, 1, 64*1024, 4, 32)
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)
	return fmt.Sprintf("$argon2id$v=%d$m=65536,t=1,p=4$%s$%s", argon2.Version, b64Salt, b64Hash)
}
