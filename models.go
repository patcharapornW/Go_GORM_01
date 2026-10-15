package main

import (
	"gorm.io/gorm"
)

// Category โมเดลสำหรับประเภทสินค้า
// ความสัมพันธ์: 1 Category มีได้หลาย Product (1:N)
type Category struct {
	gorm.Model
	Name        string    `gorm:"size:255;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Products    []Product `gorm:"foreignKey:CategoryID" json:"products,omitempty"`
}

// Product โมเดลสำหรับสินค้า
// ความสัมพันธ์: สังกัด 1 Category (N:1) และเชื่อมต่อไปยัง Order (1:N)
type Product struct {
	gorm.Model
	Name       string   `gorm:"size:255;not null" json:"name"`
	Price      float64  `gorm:"type:decimal(10,2);not null" json:"price"`
	Quantity   int      `gorm:"not null;default:0" json:"quantity"`
	CategoryID uint     `gorm:"not null" json:"category_id"`
	Category   Category `gorm:"foreignKey:CategoryID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"category,omitempty"`
	Orders     []Order  `gorm:"foreignKey:ProductID" json:"orders,omitempty"`
}

// Customer โมเดลสำหรับข้อมูลลูกค้า
// ความสัมพันธ์: 1 Customer มีได้หลาย Order (1:N)
type Customer struct {
	gorm.Model
	Name   string  `gorm:"size:255;not null" json:"name"`
	Email  string  `gorm:"size:255;not null;uniqueIndex" json:"email"`
	Phone  string  `gorm:"size:50;not null" json:"phone"`
	Orders []Order `gorm:"foreignKey:CustomerID" json:"orders,omitempty"`
}

// Order โมเดลสำหรับคำสั่งซื้อ
// โดยหนึ่ง Order มีสินค้าเพียงหนึ่งรายการ
// ความสัมพันธ์: เชื่อมโยงกับ Customer (N:1) และ Product (N:1)
type Order struct {
	gorm.Model
	CustomerID uint     `gorm:"not null" json:"customer_id"`
	Customer   Customer `gorm:"foreignKey:CustomerID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"customer,omitempty"`
	ProductID  uint     `gorm:"not null" json:"product_id"`
	Product    Product  `gorm:"foreignKey:ProductID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"product,omitempty"`
	Quantity   int      `gorm:"not null" json:"quantity"`
	TotalPrice float64  `gorm:"type:decimal(12,2);not null" json:"total_price"`
	Status     string   `gorm:"size:50;not null;default:'Pending'" json:"status"` // ค่าที่ยอมรับได้: Pending, Completed, Cancelled
}
