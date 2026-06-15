package storage

import (
	"log"
	"server-management-service/internal/modules/identity/domain"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func SeedUsers(db *gorm.DB, adminEmail, adminPassword, userEmail, userPassword string) {
	if adminEmail == "" || adminPassword == "" {
		log.Println("Admin credentials not set, skipping admin seeder.")
	} else {
		seedSingleUser(db, adminEmail, adminPassword, domain.RoleCodeAdmin)
	}

	if userEmail == "" || userPassword == "" {
		log.Println("User credentials not set, skipping user seeder.")
	} else {
		seedSingleUser(db, userEmail, userPassword, domain.RoleCodeUser)
	}
}

func seedSingleUser(db *gorm.DB, email, password string, role domain.RoleCode) {
	var count int64
	db.Model(&domain.User{}).Where("email = ?", email).Count(&count)
	if count == 0 {
		log.Printf("Seeding default %s user...", role)
		hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		db.Create(&domain.User{Email: email, Password: string(hash), RoleCode: role})
	}
}
