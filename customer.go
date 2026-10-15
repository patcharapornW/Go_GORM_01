package main

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// ==========================================
// 1. Customer Creation Services
// ==========================================

// createCustomerDB เพิ่มข้อมูลลูกค้าใหม่ลงฐานข้อมูล
func createCustomerDB(db *gorm.DB, customer *Customer) error {
	if strings.TrimSpace(customer.Name) == "" {
		return errors.New("ชื่อลูกค้าต้องไม่เป็นค่าว่าง")
	}
	if strings.TrimSpace(customer.Email) == "" {
		return errors.New("อีเมลต้องไม่เป็นค่าว่าง")
	}
	if strings.TrimSpace(customer.Phone) == "" {
		return errors.New("เบอร์โทรศัพท์ต้องไม่เป็นค่าว่าง")
	}

	if err := db.Create(customer).Error; err != nil {
		return fmt.Errorf("ไม่สามารถสร้างข้อมูลลูกค้าได้: %w", err)
	}
	return nil
}

// ==========================================
// 2. Customer Query Services
// ==========================================

// getCustomerByEmailDB ค้นหาข้อมูลลูกค้าด้วย Email
func getCustomerByEmailDB(db *gorm.DB, email string) (*Customer, error) {
	var customer Customer
	if err := db.Where("email = ?", strings.TrimSpace(email)).First(&customer).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("ไม่พบข้อมูลลูกค้าสำหรับอีเมล: %s", email)
		}
		return nil, fmt.Errorf("เกิดข้อผิดพลาดในการค้นหาลูกค้า: %w", err)
	}
	return &customer, nil
}

// getCustomerByIDDB ค้นหาข้อมูลลูกค้าด้วย ID
func getCustomerByIDDB(db *gorm.DB, id uint) (*Customer, error) {
	var customer Customer
	if err := db.First(&customer, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("ไม่พบลูกค้า ID: %d", id)
		}
		return nil, fmt.Errorf("เกิดข้อผิดพลาดในการค้นหาลูกค้า: %w", err)
	}
	return &customer, nil
}

// getAllCustomersDB ดึงข้อมูลลูกค้าทั้งหมด
func getAllCustomersDB(db *gorm.DB) ([]Customer, error) {
	var customers []Customer
	if err := db.Find(&customers).Error; err != nil {
		return nil, fmt.Errorf("ไม่สามารถดึงข้อมูลลูกค้าทั้งหมดได้: %w", err)
	}
	return customers, nil
}

// ==========================================
// 3. Customer Update Services
// ==========================================

// updateCustomerByIDDB แก้ไขข้อมูลลูกค้าตาม ID
func updateCustomerByIDDB(db *gorm.DB, id uint, updates map[string]interface{}) error {
	var customer Customer
	if err := db.First(&customer, id).Error; err != nil {
		return fmt.Errorf("ไม่พบลูกค้า ID %d: %w", id, err)
	}

	if err := db.Model(&customer).Updates(updates).Error; err != nil {
		return fmt.Errorf("ไม่สามารถอัปเดตข้อมูลลูกค้าได้: %w", err)
	}
	return nil
}

// ==========================================
// 4. Customer Deletion & Constraints
// ==========================================

// deleteCustomerByIDDB ลบข้อมูลลูกค้าตาม ID
// กฎความปลอดภัยทางธุรกิจ: "ห้ามลบลูกค้าที่มีประวัติคำสั่งซื้อในระบบ"
func deleteCustomerByIDDB(db *gorm.DB, id uint) error {
	// 1. ตรวจสอบว่ามีลูกค้ารายนี้หรือไม่
	var customer Customer
	if err := db.First(&customer, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("ไม่พบลูกค้า ID %d", id)
		}
		return fmt.Errorf("เกิดข้อผิดพลาดในการค้นหาลูกค้า: %w", err)
	}

	// 2. ตรวจสอบประวัติคำสั่งซื้อของลูกค้า
	var orderCount int64
	if err := db.Model(&Order{}).Where("customer_id = ?", id).Count(&orderCount).Error; err != nil {
		return fmt.Errorf("เกิดข้อผิดพลาดในการตรวจสอบประวัติคำสั่งซื้อ: %w", err)
	}

	if orderCount > 0 {
		return fmt.Errorf("ไม่อนุญาตให้ลบลูกค้า ID %d เนื่องจากมีประวัติคำสั่งซื้ออยู่ %d รายการ", id, orderCount)
	}

	// 3. ทำการลบข้อมูลลูกค้า
	if err := db.Delete(&customer).Error; err != nil {
		return fmt.Errorf("ไม่สามารถลบลูกค้าได้: %w", err)
	}

	return nil
}
