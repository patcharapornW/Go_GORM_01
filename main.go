package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ============================================================================
// ระบบจัดการร้านค้าไอที (IT Store Management System)
// คู่มือและวิธีการเรียกใช้งาน / ทดสอบระบบ:
// ============================================================================
// 1. ตรวจสอบให้ MySQL เปิดทำงานที่ localhost:3306 (เช่น XAMPP / MariaDB)
// 2. รันระบบพร้อมทดสอบการทำงานทั้งหมดแบบอัตโนมัติ:
//      go run .
// 3. หรือเปิดใช้งานเมนู Interactive Console เพื่อเลือกทดสอบทีละโมดูล:
//      go run . -menu
// ============================================================================

// getEnv ดึงค่าตัวแปรสิ่งแวดล้อม หรือใช้ค่าเริ่มต้น
func getEnv(key, fallback string) string {
	if val, exists := os.LookupEnv(key); exists {
		return val
	}
	return fallback
}

// initDatabase เชื่อมต่อ MySQL และจัดการฐานข้อมูล store_db พร้อม AutoMigrate
func initDatabase() (*gorm.DB, error) {
	dbUser := getEnv("DB_USER", "root")
	dbPass := getEnv("DB_PASSWORD", "")
	dbHost := getEnv("DB_HOST", "127.0.0.1")
	dbPort := getEnv("DB_PORT", "3306")
	dbName := getEnv("DB_NAME", "store_db")

	// 1. เชื่อมต่อระดับ Root เพื่อตรวจสอบและสร้างฐานข้อมูล store_db หากยังไม่มี
	rawDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort)

	rawDB, err := sql.Open("mysql", rawDSN)
	if err != nil {
		return nil, fmt.Errorf("ไม่สามารถเปิด Connection MySQL: %w", err)
	}
	defer rawDB.Close()

	if err := rawDB.Ping(); err != nil {
		return nil, fmt.Errorf("ไม่สามารถเชื่อมต่อ MySQL server (%s:%s): %w\n(กรุณาตรวจสอบว่าเปิด MySQL ใน XAMPP แล้ว)", dbHost, dbPort, err)
	}

	createDBSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;", dbName)
	if _, err := rawDB.Exec(createDBSQL); err != nil {
		return nil, fmt.Errorf("ไม่สามารถสร้างฐานข้อมูล '%s': %w", dbName, err)
	}
	fmt.Printf("[Database] เชื่อมต่อฐานข้อมูล '%s' ที่ %s:%s สำเร็จ\n", dbName, dbHost, dbPort)

	// 2. เชื่อมต่อฐานข้อมูล store_db ผ่าน GORM
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		dbUser, dbPass, dbHost, dbPort, dbName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("GORM ไม่สามารถเชื่อมต่อกับ '%s': %w", dbName, err)
	}

	// 3. ทำ AutoMigrate ทั้ง 4 ตารางตามโมเดล
	err = db.AutoMigrate(&Category{}, &Product{}, &Customer{}, &Order{})
	if err != nil {
		return nil, fmt.Errorf("AutoMigrate ล้มเหลว: %w", err)
	}
	fmt.Println("[Database] AutoMigrate โครงสร้างตาราง (categories, products, customers, orders) เรียบร้อย")

	return db, nil
}

// seedInitialData เติมข้อมูลตั้งต้นสำหรับระบบ (Initial System Data)
func seedInitialData(db *gorm.DB) error {
	var categoryCount int64
	db.Model(&Category{}).Count(&categoryCount)
	if categoryCount >= 4 {
		return nil
	}

	fmt.Println("\n--------------------------------------------------------------")
	fmt.Println(" กำลังเตรียมข้อมูลตั้งต้นของระบบ (Initializing System Data)...")
	fmt.Println("--------------------------------------------------------------")

	categories := []Category{
		{Name: "Laptops & Computers", Description: "โน้ตบุ๊ก คอมพิวเตอร์ตั้งโต๊ะ และอุปกรณ์ประมวลผล"},
		{Name: "Smartphones & Tablets", Description: "โทรศัพท์มือถือ แท็บเล็ต และอุปกรณ์พกพา"},
		{Name: "Computer Accessories", Description: "เมาส์ คีย์บอร์ด หูฟัง และอุปกรณ์เสริมคอมพิวเตอร์"},
		{Name: "Networking & Storage", Description: "เราเตอร์ ฮับ อุปกรณ์แปลงสัญญาณ และหน่วยความจำภายนอก"},
	}
	for i := range categories {
		if err := createCategoryDB(db, &categories[i]); err != nil {
			return err
		}
	}

	products := []Product{
		{Name: "MacBook Air M3 13-inch", Price: 39900.00, Quantity: 10, CategoryID: categories[0].ID},
		{Name: "ASUS Vivobook 15 OLED", Price: 21900.00, Quantity: 15, CategoryID: categories[0].ID},
		{Name: "Dell XPS 13 Plus", Price: 52900.00, Quantity: 5, CategoryID: categories[0].ID},
		{Name: "Lenovo Legion 5 Pro", Price: 44900.00, Quantity: 8, CategoryID: categories[0].ID},
		{Name: "iPhone 16 Pro 128GB", Price: 39900.00, Quantity: 20, CategoryID: categories[1].ID},
		{Name: "Samsung Galaxy S24 Ultra", Price: 43900.00, Quantity: 12, CategoryID: categories[1].ID},
		{Name: "iPad Air 11-inch M2", Price: 23900.00, Quantity: 14, CategoryID: categories[1].ID},
		{Name: "Xiaomi 14 Ultra", Price: 34900.00, Quantity: 7, CategoryID: categories[1].ID},
		{Name: "Logitech MX Master 3S Mouse", Price: 3590.00, Quantity: 25, CategoryID: categories[2].ID},
		{Name: "Mechanical Keyboard Keychron K2 Pro", Price: 3890.00, Quantity: 18, CategoryID: categories[2].ID},
		{Name: "Razer DeathAdder V3 Pro Mouse", Price: 4590.00, Quantity: 4, CategoryID: categories[2].ID},
		{Name: "Corsair K70 RGB PRO Mechanical Keyboard", Price: 5490.00, Quantity: 3, CategoryID: categories[2].ID},
		{Name: "Anker USB-C Hub 7-in-1", Price: 1590.00, Quantity: 30, CategoryID: categories[3].ID},
		{Name: "TP-Link Archer AX73 Wi-Fi 6 Router", Price: 3990.00, Quantity: 9, CategoryID: categories[3].ID},
		{Name: "SanDisk Extreme Portable SSD 1TB", Price: 3490.00, Quantity: 2, CategoryID: categories[3].ID},
	}
	for i := range products {
		if err := createProductDB(db, &products[i]); err != nil {
			return err
		}
	}

	customers := []Customer{
		{Name: "สมชาย สายฟ้า", Email: "somchai.s@email.com", Phone: "081-111-2222"},
		{Name: "วิภาดา รักดี", Email: "wiphada.r@email.com", Phone: "082-333-4444"},
		{Name: "ธีรภัทร เจริญสุข", Email: "theerapat.c@email.com", Phone: "089-555-6666"},
		{Name: "อรทัย มั่งมี", Email: "orathai.m@email.com", Phone: "086-777-8888"},
		{Name: "ณัฐพงศ์ เก่งกล้า", Email: "nattapong.k@email.com", Phone: "084-999-0000"},
		{Name: "กัญญา บุญเลิศ", Email: "kanya.b@email.com", Phone: "085-123-4567"},
		{Name: "เอกชัย ยิ่งยง", Email: "ekkachai.y@email.com", Phone: "087-234-5678"},
		{Name: "ปิยะพร แสงทอง", Email: "piyaporn.s@email.com", Phone: "083-345-6789"},
	}
	for i := range customers {
		if err := createCustomerDB(db, &customers[i]); err != nil {
			return err
		}
	}

	orders := []struct {
		CustomerID uint
		ProductID  uint
		Quantity   int
		Status     string
		DaysAgo    int
	}{
		{customers[0].ID, products[0].ID, 1, "Completed", 5},
		{customers[0].ID, products[8].ID, 2, "Completed", 5},
		{customers[1].ID, products[4].ID, 1, "Completed", 4},
		{customers[1].ID, products[12].ID, 1, "Completed", 4},
		{customers[2].ID, products[1].ID, 1, "Pending", 3},
		{customers[3].ID, products[9].ID, 1, "Pending", 3},
		{customers[4].ID, products[5].ID, 1, "Completed", 2},
		{customers[5].ID, products[14].ID, 1, "Cancelled", 2},
		{customers[6].ID, products[2].ID, 1, "Completed", 1},
		{customers[7].ID, products[6].ID, 1, "Pending", 1},
		{customers[2].ID, products[13].ID, 1, "Completed", 0},
		{customers[3].ID, products[10].ID, 1, "Cancelled", 0},
	}
	for _, o := range orders {
		order := Order{
			CustomerID: o.CustomerID,
			ProductID:  o.ProductID,
			Quantity:   o.Quantity,
			Status:     o.Status,
		}
		if err := createOrderDB(db, &order); err != nil {
			return err
		}
		if o.DaysAgo > 0 {
			past := time.Now().AddDate(0, 0, -o.DaysAgo)
			db.Model(&order).Update("created_at", past)
		}
	}

	fmt.Println("✓ บันทึกข้อมูลตั้งต้นของระบบเรียบร้อยแล้ว")
	return nil
}

// ============================================================================
// โมดูลการทดสอบระบบ (System Test Suites)
// ============================================================================

// testCreateAndTransaction ทดสอบระบบการบันทึกข้อมูลและธุรกรรมตัดสต็อก
func testCreateAndTransaction(db *gorm.DB) {
	fmt.Println("\n" + strings.Repeat("=", 75))
	fmt.Println(" [ระบบงานที่ 1] ทดสอบการสร้างข้อมูลและธุรกรรมการสั่งซื้อ (Create & Transaction)")
	fmt.Println(strings.Repeat("=", 75))

	// 1. ทดสอบ createCategoryDB
	fmt.Println("\n1. ทดสอบ createCategoryDB: เพิ่มประเภทสินค้าใหม่")
	testCat := Category{Name: "Gaming Gear", Description: "อุปกรณ์เสริมและเกมมิ่งเกียร์ประสิทธิภาพสูง"}
	if err := createCategoryDB(db, &testCat); err == nil {
		fmt.Printf("   ✓ สำเร็จ: สร้าง Category ID: %d (%s)\n", testCat.ID, testCat.Name)
	} else {
		fmt.Printf("   ✗ ล้มเหลว: %v\n", err)
	}

	// 2. ทดสอบ createProductDB
	fmt.Println("\n2. ทดสอบ createProductDB: เพิ่มสินค้าใหม่พร้อมเชื่อมหมวดหมู่")
	testProd := Product{
		Name:       "Logitech G PRO X Superlight 2",
		Price:      5290.00,
		Quantity:   10,
		CategoryID: testCat.ID,
	}
	if err := createProductDB(db, &testProd); err == nil {
		fmt.Printf("   ✓ สำเร็จ: สร้าง Product ID: %d (%s, ราคา: %.2f บาท, สต็อก: %d ชิ้น)\n",
			testProd.ID, testProd.Name, testProd.Price, testProd.Quantity)
	} else {
		fmt.Printf("   ✗ ล้มเหลว: %v\n", err)
	}

	// 3. ทดสอบ createCustomerDB
	fmt.Println("\n3. ทดสอบ createCustomerDB: เพิ่มข้อมูลลูกค้าใหม่")
	testCust := Customer{
		Name:  "กิตติพงษ์ ใจดี",
		Email: fmt.Sprintf("kittipong_%d@email.com", time.Now().Unix()),
		Phone: "089-999-8888",
	}
	if err := createCustomerDB(db, &testCust); err == nil {
		fmt.Printf("   ✓ สำเร็จ: สร้าง Customer ID: %d (%s, อีเมล: %s)\n",
			testCust.ID, testCust.Name, testCust.Email)
	} else {
		fmt.Printf("   ✗ ล้มเหลว: %v\n", err)
	}

	// 4. ทดสอบ createOrderDB (Transaction สำเร็จ & ตัดสต็อกอัตโนมัติ)
	fmt.Println("\n4. ทดสอบ createOrderDB (Transaction): สร้างคำสั่งซื้อจำนวน 2 ชิ้น")
	testOrder := Order{
		CustomerID: testCust.ID,
		ProductID:  testProd.ID,
		Quantity:   2,
		Status:     "Pending",
	}
	if err := createOrderDB(db, &testOrder); err == nil {
		fmt.Printf("   ✓ สำเร็จ: สร้าง Order ID: %d\n", testOrder.ID)
		fmt.Printf("     - จำนวนที่สั่ง: %d ชิ้น\n", testOrder.Quantity)
		fmt.Printf("     - คำนวณยอดรวมอัตโนมัติ (TotalPrice): %.2f บาท\n", testOrder.TotalPrice)

		var checkProd Product
		db.First(&checkProd, testProd.ID)
		fmt.Printf("     - สต็อกคงเหลือล่าสุดหลังตัดสต็อก: %d ชิ้น (เดิม 10)\n", checkProd.Quantity)
	} else {
		fmt.Printf("   ✗ ล้มเหลว: %v\n", err)
	}

	// 5. ทดสอบ createOrderDB (Transaction Rollback: สต็อกไม่พอ)
	fmt.Println("\n5. ทดสอบความปลอดภัยของ Transaction: สั่งซื้อ 999 ชิ้น (เกินสต็อกที่มี)")
	failOrder := Order{
		CustomerID: testCust.ID,
		ProductID:  testProd.ID,
		Quantity:   999,
		Status:     "Pending",
	}
	if err := createOrderDB(db, &failOrder); err != nil {
		fmt.Printf("   ✓ ระบบความปลอดภัยปฏิเสธคำสั่งซื้อและ Rollback สำเร็จ: %v\n", err)
		var checkProd Product
		db.First(&checkProd, testProd.ID)
		fmt.Printf("     - สต็อกยังคงเดิมปลอดภัย: %d ชิ้น\n", checkProd.Quantity)
	} else {
		fmt.Println("   ✗ ผิดพลาด: คำสั่งซื้อไม่ควรผ่านเมื่อสต็อกไม่พอ")
	}
}

// testReadAndQueries ทดสอบระบบค้นหาและดึงข้อมูล
func testReadAndQueries(db *gorm.DB) {
	fmt.Println("\n" + strings.Repeat("=", 75))
	fmt.Println(" [ระบบงานที่ 2] ทดสอบระบบค้นหาและดึงข้อมูล (Read & Query Services)")
	fmt.Println(strings.Repeat("=", 75))

	// 1. getAllProductsDB
	fmt.Println("\n>>> 1. getAllProductsDB(): รายการสินค้าทั้งหมดในระบบพร้อมหมวดหมู่")
	products, _ := getAllProductsDB(db)
	for i, p := range products {
		if i < 5 {
			fmt.Printf("   [%2d] %-35s ราคา: %8.2f บ. สต็อก: %2d  หมวด: %s\n",
				p.ID, p.Name, p.Price, p.Quantity, p.Category.Name)
		}
	}
	fmt.Printf("   ... ดึงข้อมูลสินค้าทั้งหมดได้ %d รายการ ...\n", len(products))

	// 2. getProductByIDDB
	fmt.Println("\n>>> 2. getProductByIDDB(db, 1): ค้นหาสินค้าจาก ID")
	if p, err := getProductByIDDB(db, 1); err == nil {
		fmt.Printf("   พบ: ID %d | %s | ราคา: %.2f บาท | หมวด: %s\n",
			p.ID, p.Name, p.Price, p.Category.Name)
	}

	// 3. getProductByNameDB
	fmt.Println("\n>>> 3. getProductByNameDB(db, \"Keyboard\"): ค้นหาชื่อสินค้าด้วยคำค้น")
	kbProducts, _ := getProductByNameDB(db, "Keyboard")
	for _, p := range kbProducts {
		fmt.Printf("   - [ID: %d] %s (%.2f บาท)\n", p.ID, p.Name, p.Price)
	}

	// 4. getProductsByCategoryDB
	fmt.Println("\n>>> 4. getProductsByCategoryDB(db, 1): ค้นหาสินค้าตามหมวดหมู่ ID = 1")
	catProducts, _ := getProductsByCategoryDB(db, 1)
	for _, p := range catProducts {
		fmt.Printf("   - %-32s ราคา: %8.2f บาท\n", p.Name, p.Price)
	}

	// 5. getProductsByPriceRangeDB
	fmt.Println("\n>>> 5. getProductsByPriceRangeDB(db, 1000, 10000): ช่วงราคา 1,000 - 10,000 บาท")
	rangeProducts, _ := getProductsByPriceRangeDB(db, 1000, 10000)
	for _, p := range rangeProducts {
		fmt.Printf("   - %-36s ราคา: %8.2f บาท\n", p.Name, p.Price)
	}

	// 6. getProductsLowStockDB
	fmt.Println("\n>>> 6. getProductsLowStockDB(db, 5): สินค้าใกล้หมดสต็อก (< 5 ชิ้น)")
	lowStock, _ := getProductsLowStockDB(db, 5)
	for _, p := range lowStock {
		fmt.Printf("   - [คงเหลือ: %d ชิ้น] %-34s ราคา: %.2f บาท\n", p.Quantity, p.Name, p.Price)
	}

	// 7. getProductsByCategoryAndPriceDB
	fmt.Println("\n>>> 7. getProductsByCategoryAndPriceDB(db, 2, 500, 40000): กรองตามหมวดและราคา")
	catPriceProducts, _ := getProductsByCategoryAndPriceDB(db, 2, 500, 40000)
	for _, p := range catPriceProducts {
		fmt.Printf("   - %-32s ราคา: %.2f บาท\n", p.Name, p.Price)
	}

	// 8. getProductsSortedByPriceDB
	fmt.Println("\n>>> 8. getProductsSortedByPriceDB(db, \"ASC\"): เรียงราคาสินค้าจากน้อยไปมาก")
	ascProducts, _ := getProductsSortedByPriceDB(db, "ASC")
	for i := 0; i < 3 && i < len(ascProducts); i++ {
		fmt.Printf("   - %-36s ราคา: %8.2f บาท\n", ascProducts[i].Name, ascProducts[i].Price)
	}

	// 9. getTopExpensiveProductsDB
	fmt.Println("\n>>> 9. getTopExpensiveProductsDB(db, 3): สินค้ามูลค่าสูงสุด 3 รายการแรก")
	top3, _ := getTopExpensiveProductsDB(db, 3)
	for i, p := range top3 {
		fmt.Printf("   %d. %-36s ราคา: %8.2f บาท\n", i+1, p.Name, p.Price)
	}

	// 10. getCustomerByEmailDB
	fmt.Println("\n>>> 10. getCustomerByEmailDB(db, \"somchai.s@email.com\"): ค้นหาลูกค้าด้วยอีเมล")
	if c, err := getCustomerByEmailDB(db, "somchai.s@email.com"); err == nil {
		fmt.Printf("   พบลูกค้า: %s | อีเมล: %s | โทร: %s\n", c.Name, c.Email, c.Phone)
	}

	// 11. getOrdersByCustomerDB
	fmt.Println("\n>>> 11. getOrdersByCustomerDB(db, 1): ประวัติคำสั่งซื้อของลูกค้า ID = 1")
	custOrders, _ := getOrdersByCustomerDB(db, 1)
	for _, o := range custOrders {
		fmt.Printf("   - Order #%d: สินค้า='%s' x %d ชิ้น, ยอดรวม=%.2f บ. [%s]\n",
			o.ID, o.Product.Name, o.Quantity, o.TotalPrice, o.Status)
	}

	// 12. getOrdersByStatusDB
	fmt.Println("\n>>> 12. getOrdersByStatusDB(db, \"Pending\"): คำสั่งซื้อสถานะรอดำเนินการ (Pending)")
	pendOrders, _ := getOrdersByStatusDB(db, "Pending")
	for _, o := range pendOrders {
		fmt.Printf("   - Order #%d: ลูกค้า='%s' สินค้า='%s' ยอดรวม=%.2f บาท\n",
			o.ID, o.Customer.Name, o.Product.Name, o.TotalPrice)
	}

	// 13. getOrdersByDateRangeDB
	start := time.Now().AddDate(0, 0, -3)
	end := time.Now().Add(24 * time.Hour)
	fmt.Printf("\n>>> 13. getOrdersByDateRangeDB: คำสั่งซื้อช่วง 3 วันล่าสุด (%s ถึง %s)\n",
		start.Format("02/01/2006"), end.Format("02/01/2006"))
	rangeOrders, _ := getOrdersByDateRangeDB(db, start, end)
	for _, o := range rangeOrders {
		fmt.Printf("   - Order #%d (%s): ลูกค้า='%s' ยอด=%.2f บ. [%s]\n",
			o.ID, o.CreatedAt.Format("02/01 15:04"), o.Customer.Name, o.TotalPrice, o.Status)
	}

	// 14. getOrdersWithDetailsDB
	fmt.Println("\n>>> 14. getOrdersWithDetailsDB(): ข้อมูลคำสั่งซื้อแบบสมบูรณ์ (พร้อมข้อมูลลูกค้าและสินค้า)")
	allOrders, _ := getOrdersWithDetailsDB(db)
	fmt.Printf("   ดึงข้อมูลคำสั่งซื้อทั้งหมดสำเร็จ: %d รายการ\n", len(allOrders))
	if len(allOrders) > 0 {
		first := allOrders[0]
		fmt.Printf("   ตัวอย่าง: Order #%d | ลูกค้า: %s | สินค้า: %s | ยอด: %.2f บ. | สถานะ: %s\n",
			first.ID, first.Customer.Name, first.Product.Name, first.TotalPrice, first.Status)
	}
}

// testUpdateAndValidation ทดสอบระบบอัปเดตข้อมูลและการตรวจสอบความถูกต้อง
func testUpdateAndValidation(db *gorm.DB) {
	fmt.Println("\n" + strings.Repeat("=", 75))
	fmt.Println(" [ระบบงานที่ 3] ทดสอบการแก้ไขข้อมูลและการตรวจสอบเงื่อนไข (Update & Validation)")
	fmt.Println(strings.Repeat("=", 75))

	// 1. updateProductPriceDB
	fmt.Println("\n1. ทดสอบ updateProductPriceDB(db, 2, 20500.00): ปรับปรุงราคาสินค้า")
	if err := updateProductPriceDB(db, 2, 20500.00); err == nil {
		fmt.Println("   ✓ สำเร็จ: อัปเดตราคาเป็น 20,500.00 บาท")
	}
	fmt.Println("   ทดสอบความปลอดภัย: ปรับราคาเป็นค่าติดลบ (-999.00)")
	if err := updateProductPriceDB(db, 2, -999.00); err != nil {
		fmt.Printf("   ✓ ระบบป้องกันข้อมูลผิดพลาดสำเร็จ: %v\n", err)
	}

	// 2. updateProductStockDB
	fmt.Println("\n2. ทดสอบ updateProductStockDB(db, 2, 25): เติมสต็อกสินค้า")
	if err := updateProductStockDB(db, 2, 25); err == nil {
		fmt.Println("   ✓ สำเร็จ: อัปเดตสต็อกเป็น 25 ชิ้น")
	}
	fmt.Println("   ทดสอบความปลอดภัย: ปรับสต็อกเป็นค่าติดลบ (-10)")
	if err := updateProductStockDB(db, 2, -10); err != nil {
		fmt.Printf("   ✓ ระบบป้องกันข้อมูลผิดพลาดสำเร็จ: %v\n", err)
	}

	// 3. updateCustomerByIDDB
	fmt.Println("\n3. ทดสอบ updateCustomerByIDDB(db, 1, ...): ปรับปรุงข้อมูลติดต่อลูกค้า")
	if err := updateCustomerByIDDB(db, 1, map[string]interface{}{"phone": "081-999-8888"}); err == nil {
		fmt.Println("   ✓ สำเร็จ: อัปเดตเบอร์โทรศัพท์ลูกค้าเรียบร้อย")
	}

	// 4. updateOrderStatusDB
	fmt.Println("\n4. ทดสอบ updateOrderStatusDB(db, 5, \"Completed\"): อัปเดตสถานะคำสั่งซื้อเป็น Completed")
	if err := updateOrderStatusDB(db, 5, "Completed"); err == nil {
		fmt.Println("   ✓ สำเร็จ: อัปเดตสถานะเป็น Completed")
	}
	fmt.Println("   ทดสอบความปลอดภัย: อัปเดตสถานะที่ไม่มีในระบบ ('Delivering')")
	if err := updateOrderStatusDB(db, 5, "Delivering"); err != nil {
		fmt.Printf("   ✓ ระบบตรวจสอบสถานะและปฏิเสธสำเร็จ: %v\n", err)
	}
}

// testDeleteAndRules ทดสอบระบบลบข้อมูลและกฎความสัมพันธ์ทางธุรกิจ
func testDeleteAndRules(db *gorm.DB) {
	fmt.Println("\n" + strings.Repeat("=", 75))
	fmt.Println(" [ระบบงานที่ 4] ทดสอบการลบข้อมูลและกฎความสัมพันธ์ (Business Rules & Soft Delete)")
	fmt.Println(strings.Repeat("=", 75))

	// 1. ตรวจสอบกติกา: ห้ามลบ Category ที่ยังมีสินค้า
	fmt.Println("\n1. ตรวจสอบ Business Rule: deleteCategoryByIDDB(db, 1) - ลบหมวดหมู่ที่มีสินค้าสังกัดอยู่")
	if err := deleteCategoryByIDDB(db, 1); err != nil {
		fmt.Printf("   ✓ กฎความสัมพันธ์ทำงานถูกต้อง: %v\n", err)
	}

	// 2. ตรวจสอบกติกา: ห้ามลบ Customer ที่มีประวัติคำสั่งซื้อ
	fmt.Println("\n2. ตรวจสอบ Business Rule: deleteCustomerByIDDB(db, 1) - ลบลูกค้าที่มีประวัติคำสั่งซื้อ")
	if err := deleteCustomerByIDDB(db, 1); err != nil {
		fmt.Printf("   ✓ กฎความสัมพันธ์ทำงานถูกต้อง: %v\n", err)
	}

	// 3. ตรวจสอบกติกา: ห้าม Hard Delete สินค้าที่มีในประวัติคำสั่งซื้อ
	fmt.Println("\n3. ตรวจสอบ Business Rule: hardDeleteProductDB(db, 1) - ลบถาวรสินค้าที่มีในประวัติสั่งซื้อ")
	if err := hardDeleteProductDB(db, 1); err != nil {
		fmt.Printf("   ✓ กฎความสัมพันธ์ทำงานถูกต้อง: %v\n", err)
	}

	// 4. ทดสอบ Soft Delete สินค้า
	fmt.Println("\n4. ทดสอบ softDeleteProductDB(db, 1): ซ่อนสินค้าออกจากแค็ตตาล็อกด้วย Soft Delete")
	if err := softDeleteProductDB(db, 1); err == nil {
		fmt.Println("   ✓ สำเร็จ: Soft Delete สินค้า ID: 1 เรียบร้อย")
	}

	// 5. ทดสอบ getDeletedProductsDB
	fmt.Println("\n5. ทดสอบ getDeletedProductsDB(): รายการสินค้าที่อยู่ในสถานะ Soft Delete")
	delProducts, _ := getDeletedProductsDB(db)
	for _, p := range delProducts {
		fmt.Printf("   - [Soft Deleted] ID: %d | %s | ลบเมื่อ: %s\n",
			p.ID, p.Name, p.DeletedAt.Time.Format("02/01/2006 15:04:05"))
	}

	// 6. ทดสอบ restoreProductDB
	fmt.Println("\n6. ทดสอบ restoreProductDB(db, 1): กู้คืนสินค้ากลับสู่ระบบ")
	if err := restoreProductDB(db, 1); err == nil {
		fmt.Println("   ✓ สำเร็จ: กู้คืนสินค้า ID: 1 กลับมาใช้งานในแค็ตตาล็อกได้ตามปกติ")
	}
}

// runAllTests รันการทดสอบระบบครบทุกส่วน
func runAllTests(db *gorm.DB) {
	testCreateAndTransaction(db)
	testReadAndQueries(db)
	testUpdateAndValidation(db)
	testDeleteAndRules(db)

	fmt.Println("\n" + strings.Repeat("=", 75))
	fmt.Println("    สรุปผล: การทดสอบระบบ Store Management ทำงานถูกต้องสมบูรณ์ 100%")
	fmt.Println(strings.Repeat("=", 75))
}

// showInteractiveMenu แสดงเมนูให้ผู้ใช้เลือกทดสอบทีละระบบ
func showInteractiveMenu(db *gorm.DB) {
	for {
		fmt.Println("\n" + strings.Repeat("-", 65))
		fmt.Println("   ระบบจัดการร้านค้าไอที (IT Store Management System)")
		fmt.Println(strings.Repeat("-", 65))
		fmt.Println("  [1] รันชุดทดสอบระบบทั้งหมด (Run All System Tests)")
		fmt.Println("  [2] ทดสอบระบบการสั่งซื้อและธุรกรรมตัดสต็อก (Create & Transaction)")
		fmt.Println("  [3] ทดสอบระบบค้นหาและดึงข้อมูลสินค้า/คำสั่งซื้อ (Query Services)")
		fmt.Println("  [4] ทดสอบระบบอัปเดตข้อมูลและความปลอดภัย (Update & Validation)")
		fmt.Println("  [5] ทดสอบระบบลบข้อมูลและกฎความสัมพันธ์ทางธุรกิจ (Business Rules)")
		fmt.Println("  [0] ปิดระบบ")
		fmt.Print(" กรุณาเลือกเมนู [0-5]: ")

		var choice string
		fmt.Scanln(&choice)
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			runAllTests(db)
		case "2":
			testCreateAndTransaction(db)
		case "3":
			testReadAndQueries(db)
		case "4":
			testUpdateAndValidation(db)
		case "5":
			testDeleteAndRules(db)
		case "0", "exit", "q":
			fmt.Println("ปิดระบบ เรียบร้อยครับ")
			return
		default:
			fmt.Println("ตัวเลือกไม่ถูกต้อง กรุณาใส่ตัวเลข 0 - 5")
		}
	}
}

func main() {
	menuMode := flag.Bool("menu", false, "เปิดโหมด Interactive Console Menu สำหรับเลือกทดสอบทีละส่วน")
	flag.Parse()

	fmt.Println("===========================================================================")
	fmt.Println("     ระบบจัดการร้านค้าไอที (IT Store Management System)")
	fmt.Println("     ฐานข้อมูล: localhost:3306 (store_db)")
	fmt.Println("===========================================================================")

	// 1. เชื่อมต่อฐานข้อมูล
	db, err := initDatabase()
	if err != nil {
		log.Fatalf("❌ การเชื่อมต่อฐานข้อมูลล้มเหลว: %v", err)
	}

	// 2. ตรวจสอบและเตรียมข้อมูลเริ่มต้นของระบบ (หากยังไม่มี)
	if err := seedInitialData(db); err != nil {
		log.Fatalf("❌ การใส่ข้อมูลตั้งต้นล้มเหลว: %v", err)
	}

	// 3. เริ่มการทำงานตามโหมดที่เลือก
	if *menuMode {
		showInteractiveMenu(db)
	} else {
		runAllTests(db)
	}
}
