package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
)

// AllowedOrderStatuses รายการสถานะคำสั่งซื้อที่อนุญาตในระบบ
var AllowedOrderStatuses = map[string]bool{
	"Pending":   true,
	"Completed": true,
	"Cancelled": true,
}

// ==========================================
// 1. Order Transaction Services
// ==========================================

// createOrderDB สร้างคำสั่งซื้อใหม่
// ขั้นตอนการประมวลผลธุรกรรม (Transaction Workflow):
// 1. ตรวจสอบลูกค้าและสินค้าว่ามีอยู่จริงในระบบ
// 2. ตรวจสอบสต็อกคงเหลือ (ต้องเพียงพอกับจำนวนที่สั่ง)
// 3. คำนวณ TotalPrice = Price * Quantity อัตโนมัติ
// 4. ลดสต็อกสินค้า
// 5. ใช้ Transaction ให้การเพิ่มคำสั่งซื้อกับการลดสต็อกสำเร็จหรือยกเลิกพร้อมกัน (Atomic)
func createOrderDB(db *gorm.DB, order *Order) error {
	if order.Quantity <= 0 {
		return errors.New("จำนวนสินค้าที่สั่งซื้อต้องมากกว่า 0")
	}

	// กำหนดสถานะเริ่มต้นหากไม่ได้ระบุ
	if strings.TrimSpace(order.Status) == "" {
		order.Status = "Pending"
	} else if !AllowedOrderStatuses[order.Status] {
		return fmt.Errorf("สถานะคำสั่งซื้อไม่ถูกต้อง: %s (ต้องเป็น Pending, Completed หรือ Cancelled)", order.Status)
	}

	// ดำเนินการผ่าน Database Transaction
	return db.Transaction(func(tx *gorm.DB) error {
		// 1. ตรวจสอบข้อมูลลูกค้า
		var customer Customer
		if err := tx.First(&customer, order.CustomerID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("ไม่พบลูกค้ารหัส ID: %d", order.CustomerID)
			}
			return fmt.Errorf("เกิดข้อผิดพลาดในการตรวจสอบลูกค้า: %w", err)
		}

		// 2. ตรวจสอบข้อมูลสินค้าและสต็อก (ใช้ Lock For Update เพื่อความปลอดภัยในการตัดสต็อกพร้อมกัน)
		var product Product
		if err := tx.Set("gorm:query_option", "FOR UPDATE").First(&product, order.ProductID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("ไม่พบสินค้ารหัส ID: %d", order.ProductID)
			}
			return fmt.Errorf("เกิดข้อผิดพลาดในการตรวจสอบสินค้า: %w", err)
		}

		// ตรวจสอบสต็อก
		if product.Quantity < order.Quantity {
			return fmt.Errorf("สินค้า '%s' สต็อกไม่เพียงพอ (มีคงเหลือ %d ชิ้น, ต้องการสั่ง %d ชิ้น)",
				product.Name, product.Quantity, order.Quantity)
		}

		// 3. คำนวณยอดรวม TotalPrice = Price * Quantity
		order.TotalPrice = product.Price * float64(order.Quantity)

		// 4. ลดจำนวนสต็อกสินค้า
		newQuantity := product.Quantity - order.Quantity
		if err := tx.Model(&product).Update("quantity", newQuantity).Error; err != nil {
			return fmt.Errorf("ไม่สามารถลดสต็อกสินค้าได้: %w", err)
		}

		// 5. บันทึกคำสั่งซื้อ
		if err := tx.Create(order).Error; err != nil {
			return fmt.Errorf("ไม่สามารถสร้างคำสั่งซื้อได้: %w", err)
		}

		return nil
	})
}

// ==========================================
// 2. Order Query Services
// ==========================================

// getOrdersByCustomerDB ดึงคำสั่งซื้อของลูกค้าตาม CustomerID
func getOrdersByCustomerDB(db *gorm.DB, customerID uint) ([]Order, error) {
	var orders []Order
	if err := db.Preload("Product").
		Where("customer_id = ?", customerID).
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("เกิดข้อผิดพลาดในการดึงคำสั่งซื้อของลูกค้า: %w", err)
	}
	return orders, nil
}

// getOrdersByStatusDB ดึงคำสั่งซื้อตามสถานะ (เช่น Pending, Completed, Cancelled)
func getOrdersByStatusDB(db *gorm.DB, status string) ([]Order, error) {
	var orders []Order
	if err := db.Preload("Customer").Preload("Product").
		Where("status = ?", status).
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("เกิดข้อผิดพลาดในการดึงคำสั่งซื้อตามสถานะ: %w", err)
	}
	return orders, nil
}

// getOrdersByDateRangeDB ดึงคำสั่งซื้อตามช่วงวันที่
func getOrdersByDateRangeDB(db *gorm.DB, startDate, endDate time.Time) ([]Order, error) {
	var orders []Order
	if err := db.Preload("Customer").Preload("Product").
		Where("created_at BETWEEN ? AND ?", startDate, endDate).
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("เกิดข้อผิดพลาดในการดึงคำสั่งซื้อตามช่วงวันที่: %w", err)
	}
	return orders, nil
}

// getOrdersWithDetailsDB ดึงคำสั่งซื้อทั้งหมดพร้อมข้อมูลลูกค้าและสินค้า (Preload)
func getOrdersWithDetailsDB(db *gorm.DB) ([]Order, error) {
	var orders []Order
	if err := db.Preload("Customer").Preload("Product").
		Order("id ASC").
		Find(&orders).Error; err != nil {
		return nil, fmt.Errorf("เกิดข้อผิดพลาดในการดึงคำสั่งซื้อพร้อมรายละเอียด: %w", err)
	}
	return orders, nil
}

// ==========================================
// 3. Order Status & Lifecycle Services
// ==========================================

// updateOrderStatusDB เปลี่ยนสถานะคำสั่งซื้อ
// กฎทางธุรกิจ: สถานะต้องเป็น Pending, Completed หรือ Cancelled เท่านั้น
// ความสมบูรณ์ของระบบ: หากเปลี่ยนสถานะเป็น Cancelled จะคืนสต็อกสินค้าให้โดยอัตโนมัติผ่าน Transaction
func updateOrderStatusDB(db *gorm.DB, id uint, newStatus string) error {
	newStatus = strings.TrimSpace(newStatus)
	if !AllowedOrderStatuses[newStatus] {
		return fmt.Errorf("สถานะไม่ถูกต้อง: '%s' (สถานะต้องเป็น Pending, Completed หรือ Cancelled เท่านั้น)", newStatus)
	}

	return db.Transaction(func(tx *gorm.DB) error {
		var order Order
		if err := tx.First(&order, id).Error; err != nil {
			return fmt.Errorf("ไม่พบคำสั่งซื้อ ID %d: %w", id, err)
		}

		oldStatus := order.Status
		if oldStatus == newStatus {
			return nil // ไม่มีการเปลี่ยนแปลง
		}

		// หากคำสั่งซื้อเดิมไม่ใช่ Cancelled แต่เปลี่ยนเป็น Cancelled ให้คืนสต็อกสินค้า
		if oldStatus != "Cancelled" && newStatus == "Cancelled" {
			if err := tx.Model(&Product{}).Where("id = ?", order.ProductID).
				Update("quantity", gorm.Expr("quantity + ?", order.Quantity)).Error; err != nil {
				return fmt.Errorf("เกิดข้อผิดพลาดในการคืนสต็อกสินค้า: %w", err)
			}
		}

		// หากเดิมยกเลิกไปแล้ว จะเปลี่ยนกลับมาเป็น Pending หรือ Completed ให้ตัดสต็อกคืน
		if oldStatus == "Cancelled" && newStatus != "Cancelled" {
			var product Product
			if err := tx.First(&product, order.ProductID).Error; err != nil {
				return fmt.Errorf("ไม่พบข้อมูลสินค้า: %w", err)
			}
			if product.Quantity < order.Quantity {
				return fmt.Errorf("สต็อกสินค้า '%s' ไม่พอสำหรับการเปิดใช้งานคำสั่งซื้อนี้อีกครั้ง", product.Name)
			}
			if err := tx.Model(&product).Update("quantity", product.Quantity-order.Quantity).Error; err != nil {
				return fmt.Errorf("เกิดข้อผิดพลาดในการตัดสต็อกสินค้า: %w", err)
			}
		}

		// อัปเดตสถานะ
		if err := tx.Model(&order).Update("status", newStatus).Error; err != nil {
			return fmt.Errorf("ไม่สามารถอัปเดตสถานะคำสั่งซื้อได้: %w", err)
		}

		return nil
	})
}
