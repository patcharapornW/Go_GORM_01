package main

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// ==========================================
// 1. Create Functions
// ==========================================

// createProductDB เพิ่มสินค้าพร้อมตรวจสอบประเภทสินค้า
func createProductDB(db *gorm.DB, product *Product) error {
	if strings.TrimSpace(product.Name) == "" {
		return errors.New("ชื่อสินค้าต้องไม่เป็นค่าว่าง")
	}
	if product.Price < 0 {
		return errors.New("ราคาสินค้าต้องไม่ติดลบ")
	}
	if product.Quantity < 0 {
		return errors.New("จำนวนสต็อกสินค้าต้องไม่ติดลบ")
	}

	// ตรวจสอบว่ามี CategoryID ที่ระบุอยู่จริงหรือไม่
	var category Category
	if err := db.First(&category, product.CategoryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("ไม่พบรหัสประเภทสินค้า CategoryID: %d ในระบบ", product.CategoryID)
		}
		return fmt.Errorf("เกิดข้อผิดพลาดในการตรวจสอบประเภทสินค้า: %w", err)
	}

	if err := db.Create(product).Error; err != nil {
		return fmt.Errorf("ไม่สามารถสร้างสินค้าได้: %w", err)
	}

	return nil
}

// ==========================================
// 2. Read & Query Services
// ==========================================

// getAllProductsDB ดึงรายการสินค้าทั้งหมดที่ไม่ถูก Soft Delete พร้อม Preload Category
func getAllProductsDB(db *gorm.DB) ([]Product, error) {
	var products []Product
	if err := db.Preload("Category").Find(&products).Error; err != nil {
		return nil, fmt.Errorf("ไม่สามารถดึงข้อมูลสินค้าทั้งหมดได้: %w", err)
	}
	return products, nil
}

// getProductByIDDB ค้นหาสินค้าจาก ID
func getProductByIDDB(db *gorm.DB, id uint) (*Product, error) {
	var product Product
	if err := db.Preload("Category").First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("ไม่พบสินค้ารหัส ID: %d", id)
		}
		return nil, fmt.Errorf("เกิดข้อผิดพลาดในการค้นหาสินค้า: %w", err)
	}
	return &product, nil
}

// getProductByNameDB ค้นหาชื่อสินค้าบางส่วนด้วย LIKE
func getProductByNameDB(db *gorm.DB, name string) ([]Product, error) {
	var products []Product
	query := "%" + strings.TrimSpace(name) + "%"
	if err := db.Preload("Category").Where("name LIKE ?", query).Find(&products).Error; err != nil {
		return nil, fmt.Errorf("เกิดข้อผิดพลาดในการค้นหาสินค้าด้วยชื่อ: %w", err)
	}
	return products, nil
}

// getProductsByCategoryDB ค้นหาสินค้าตาม CategoryID
func getProductsByCategoryDB(db *gorm.DB, categoryID uint) ([]Product, error) {
	var products []Product
	if err := db.Preload("Category").Where("category_id = ?", categoryID).Find(&products).Error; err != nil {
		return nil, fmt.Errorf("เกิดข้อผิดพลาดในการดึงสินค้าตามประเภท: %w", err)
	}
	return products, nil
}

// getProductsByPriceRangeDB ค้นหาสินค้าตามช่วงราคา Min - Max
func getProductsByPriceRangeDB(db *gorm.DB, minPrice, maxPrice float64) ([]Product, error) {
	var products []Product
	if err := db.Preload("Category").Where("price BETWEEN ? AND ?", minPrice, maxPrice).Order("price ASC").Find(&products).Error; err != nil {
		return nil, fmt.Errorf("เกิดข้อผิดพลาดในการค้นหาสินค้าตามช่วงราคา: %w", err)
	}
	return products, nil
}

// getProductsLowStockDB ค้นหาสินค้าที่มีจำนวนคงเหลือน้อยกว่าค่ากำหนด
func getProductsLowStockDB(db *gorm.DB, threshold int) ([]Product, error) {
	var products []Product
	if err := db.Preload("Category").Where("quantity < ?", threshold).Order("quantity ASC").Find(&products).Error; err != nil {
		return nil, fmt.Errorf("เกิดข้อผิดพลาดในการค้นหาสินค้าสต็อกต่ำ: %w", err)
	}
	return products, nil
}

// getProductsByCategoryAndPriceDB ค้นหาสินค้าตามประเภทและช่วงราคา
func getProductsByCategoryAndPriceDB(db *gorm.DB, categoryID uint, minPrice, maxPrice float64) ([]Product, error) {
	var products []Product
	if err := db.Preload("Category").
		Where("category_id = ? AND price BETWEEN ? AND ?", categoryID, minPrice, maxPrice).
		Order("price ASC").
		Find(&products).Error; err != nil {
		return nil, fmt.Errorf("เกิดข้อผิดพลาดในการค้นหาสินค้าตามประเภทและช่วงราคา: %w", err)
	}
	return products, nil
}

// getProductsSortedByPriceDB เรียงราคาสินค้าตามลำดับ ASC หรือ DESC
func getProductsSortedByPriceDB(db *gorm.DB, sortOrder string) ([]Product, error) {
	var products []Product
	orderClause := "price ASC"
	if strings.ToUpper(strings.TrimSpace(sortOrder)) == "DESC" {
		orderClause = "price DESC"
	}

	if err := db.Preload("Category").Order(orderClause).Find(&products).Error; err != nil {
		return nil, fmt.Errorf("เกิดข้อผิดพลาดในการเรียงลำดับราคาสินค้า: %w", err)
	}
	return products, nil
}

// getTopExpensiveProductsDB ค้นหาสินค้าที่แพงที่สุด N รายการแรก
func getTopExpensiveProductsDB(db *gorm.DB, limit int) ([]Product, error) {
	if limit <= 0 {
		limit = 5
	}
	var products []Product
	if err := db.Preload("Category").Order("price DESC").Limit(limit).Find(&products).Error; err != nil {
		return nil, fmt.Errorf("เกิดข้อผิดพลาดในการดึงสินค้าที่แพงที่สุด: %w", err)
	}
	return products, nil
}

// ==========================================
// 3. Update & Validation Services
// ==========================================

// updateProductByIDDB แก้ไขข้อมูลสินค้าตาม ID
func updateProductByIDDB(db *gorm.DB, id uint, updates map[string]interface{}) error {
	// ตรวจสอบเงื่อนไขราคาและสต็อกไม่ติดลบ หากมีการระบุใน updates
	if val, ok := updates["price"]; ok {
		if p, okPrice := val.(float64); okPrice && p < 0 {
			return errors.New("ราคาสินค้าต้องไม่ติดลบ")
		}
	}
	if val, ok := updates["quantity"]; ok {
		if q, okQty := val.(int); okQty && q < 0 {
			return errors.New("จำนวนสต็อกสินค้าต้องไม่ติดลบ")
		}
	}

	var product Product
	if err := db.First(&product, id).Error; err != nil {
		return fmt.Errorf("ไม่พบสินค้ารหัส %d: %w", id, err)
	}

	if err := db.Model(&product).Updates(updates).Error; err != nil {
		return fmt.Errorf("ไม่สามารถอัปเดตข้อมูลสินค้าได้: %w", err)
	}
	return nil
}

// updateProductPriceDB แก้ไขราคาสินค้า (ราคาต้องไม่ติดลบ)
func updateProductPriceDB(db *gorm.DB, id uint, newPrice float64) error {
	if newPrice < 0 {
		return errors.New("ราคาสินค้าต้องไม่ติดลบ (Price cannot be negative)")
	}

	var product Product
	if err := db.First(&product, id).Error; err != nil {
		return fmt.Errorf("ไม่พบสินค้ารหัส %d: %w", id, err)
	}

	if err := db.Model(&product).Update("price", newPrice).Error; err != nil {
		return fmt.Errorf("ไม่สามารถอัปเดตราคาสินค้าได้: %w", err)
	}
	return nil
}

// updateProductStockDB เพิ่มหรือลดสต็อกสินค้า (สต็อกต้องไม่ติดลบ)
func updateProductStockDB(db *gorm.DB, id uint, newStock int) error {
	if newStock < 0 {
		return errors.New("จำนวนสต็อกสินค้าต้องไม่ติดลบ (Stock cannot be negative)")
	}

	var product Product
	if err := db.First(&product, id).Error; err != nil {
		return fmt.Errorf("ไม่พบสินค้ารหัส %d: %w", id, err)
	}

	if err := db.Model(&product).Update("quantity", newStock).Error; err != nil {
		return fmt.Errorf("ไม่สามารถอัปเดตจำนวนสต็อกสินค้าได้: %w", err)
	}
	return nil
}

// ==========================================
// 4. Deletion & Recovery Services
// ==========================================

// softDeleteProductDB ทำการลบสินค้าแบบ Soft Delete (ตั้งค่า deleted_at)
func softDeleteProductDB(db *gorm.DB, id uint) error {
	var product Product
	if err := db.First(&product, id).Error; err != nil {
		return fmt.Errorf("ไม่พบสินค้ารหัส %d สำหรับ Soft Delete: %w", id, err)
	}

	if err := db.Delete(&product).Error; err != nil {
		return fmt.Errorf("เกิดข้อผิดพลาดในการ Soft Delete สินค้า: %w", err)
	}
	return nil
}

// hardDeleteProductDB ลบสินค้าถาวร (Unscoped Delete)
// กฎความปลอดภัยทางธุรกิจ (Business Integrity):
// ห้ามลบสินค้าถาวรหากมีประวัติคำสั่งซื้อแล้ว เพื่อป้องกันข้อมูลประวัติเสียหายและรักษา Foreign Key Integrity
func hardDeleteProductDB(db *gorm.DB, id uint) error {
	var product Product
	if err := db.Unscoped().First(&product, id).Error; err != nil {
		return fmt.Errorf("ไม่พบสินค้ารหัส %d สำหรับ Hard Delete: %w", id, err)
	}

	// ตรวจสอบประวัติใน Order
	var orderCount int64
	if err := db.Model(&Order{}).Where("product_id = ?", id).Count(&orderCount).Error; err != nil {
		return fmt.Errorf("เกิดข้อผิดพลาดในการตรวจสอบประวัติคำสั่งซื้อ: %w", err)
	}

	if orderCount > 0 {
		return fmt.Errorf("ไม่อนุญาตให้ลบสินค้าถาวร (Hard Delete) เนื่องจากสินค้ารหัส %d มีประวัติคำสั่งซื้ออยู่ %d รายการ (แนะนำให้ใช้ Soft Delete แทน)", id, orderCount)
	}

	if err := db.Unscoped().Delete(&product).Error; err != nil {
		return fmt.Errorf("เกิดข้อผิดพลาดในการ Hard Delete สินค้า: %w", err)
	}
	return nil
}

// getDeletedProductsDB ดึงรายการสินค้าทั้งหมดที่ถูก Soft Delete
func getDeletedProductsDB(db *gorm.DB) ([]Product, error) {
	var products []Product
	if err := db.Unscoped().Where("deleted_at IS NOT NULL").Preload("Category").Find(&products).Error; err != nil {
		return nil, fmt.Errorf("เกิดข้อผิดพลาดในการดึงสินค้าที่ถูก Soft Delete: %w", err)
	}
	return products, nil
}

// restoreProductDB กู้คืนสินค้าที่ถูก Soft Delete ให้กลับมาใช้งานได้ตามปกติ
func restoreProductDB(db *gorm.DB, id uint) error {
	var product Product
	if err := db.Unscoped().Where("id = ? AND deleted_at IS NOT NULL", id).First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("ไม่พบสินค้า Soft Deleted รหัส %d", id)
		}
		return fmt.Errorf("เกิดข้อผิดพลาดในการค้นหาสินค้า: %w", err)
	}

	// กู้คืนโดยตั้งค่า deleted_at เป็น NULL
	if err := db.Unscoped().Model(&product).Update("deleted_at", nil).Error; err != nil {
		return fmt.Errorf("ไม่สามารถกู้คืนสินค้าได้: %w", err)
	}
	return nil
}
