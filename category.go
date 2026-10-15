package main

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// createCategoryDB เพิ่มประเภทสินค้าใหม่ลงฐานข้อมูล
func createCategoryDB(db *gorm.DB, category *Category) error {
	if strings.TrimSpace(category.Name) == "" {
		return errors.New("ชื่อประเภทสินค้าต้องไม่เป็นค่าว่าง")
	}

	result := db.Create(category)
	if result.Error != nil {
		return fmt.Errorf("ไม่สามารถสร้างประเภทสินค้าได้: %w", result.Error)
	}

	return nil
}

// getAllCategoriesDB ดึงรายการประเภทสินค้าทั้งหมด
func getAllCategoriesDB(db *gorm.DB) ([]Category, error) {
	var categories []Category
	if err := db.Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("เกิดข้อผิดพลาดในการดึงประเภทสินค้า: %w", err)
	}
	return categories, nil
}

// deleteCategoryByIDDB ลบประเภทสินค้าตาม ID
// กฎความสัมพันธ์ทางธุรกิจ: ตรวจสอบความสัมพันธ์ก่อนลบ "ห้ามลบประเภทสินค้าที่ยังมีสินค้า"
func deleteCategoryByIDDB(db *gorm.DB, id uint) error {
	// 1. ตรวจสอบว่ามีประเภทสินค้านี้อยู่จริงหรือไม่
	var category Category
	if err := db.First(&category, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("ไม่พบประเภทสินค้ารหัส %d", id)
		}
		return fmt.Errorf("เกิดข้อผิดพลาดในการค้นหาประเภทสินค้า: %w", err)
	}

	// 2. ตรวจสอบว่ามีสินค้าสังกัดประเภทนี้อยู่หรือไม่
	var productCount int64
	if err := db.Model(&Product{}).Where("category_id = ?", id).Count(&productCount).Error; err != nil {
		return fmt.Errorf("เกิดข้อผิดพลาดในการตรวจสอบสินค้าในประเภท: %w", err)
	}

	if productCount > 0 {
		return fmt.Errorf("ไม่อนุญาตให้ลบประเภทสินค้า ID %d เนื่องจากยังมีสินค้าอยู่ในประเภทนี้จำนวน %d รายการ", id, productCount)
	}

	// 3. ทำการลบประเภทสินค้า
	if err := db.Delete(&category).Error; err != nil {
		return fmt.Errorf("ไม่สามารถลบประเภทสินค้าได้: %w", err)
	}

	return nil
}
