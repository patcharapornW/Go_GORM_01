-- =====================================================================
-- ระบบจัดการร้านค้าไอที (IT Store Management System)
-- โครงสร้างฐานข้อมูลและการจัดการข้อมูล: store_db (MySQL)
-- =====================================================================
-- รายละเอียด:
-- สคริปต์ SQL นี้สร้างฐานข้อมูล ตารางทั้ง 4 ตารางตามโครงสร้างโมเดลระบบ
-- พร้อมกำหนด Primary Key, Foreign Key, ดัชนี (Indexes),
-- และชุดข้อมูลตั้งต้นของระบบ (Categories, Products, Customers, Orders)
-- =====================================================================

-- -----------------------------------------------------
-- 1. ตรวจสอบและสร้างฐานข้อมูล store_db
-- -----------------------------------------------------
CREATE DATABASE IF NOT EXISTS `store_db`
    CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci;

USE `store_db`;

-- ปิดการตรวจสอบ Foreign Key ชั่วคราวเพื่อความสะดวกในการ Drop ตารางเก่า (หากมี)
SET FOREIGN_KEY_CHECKS = 0;

DROP TABLE IF EXISTS `orders`;
DROP TABLE IF EXISTS `products`;
DROP TABLE IF EXISTS `customers`;
DROP TABLE IF EXISTS `categories`;

SET FOREIGN_KEY_CHECKS = 1;

-- -----------------------------------------------------
-- 2. โครงสร้างตาราง (Tables Schema & Relations)
-- ใช้โครงสร้าง ID, CreatedAt, UpdatedAt, DeletedAt
-- -----------------------------------------------------

-- 1. ตาราง categories (ประเภทสินค้า)
CREATE TABLE `categories` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `created_at` DATETIME(3) NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at` DATETIME(3) NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    `deleted_at` DATETIME(3) NULL,
    `name` VARCHAR(255) NOT NULL,
    `description` TEXT NULL,
    PRIMARY KEY (`id`),
    INDEX `idx_categories_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 2. ตาราง products (สินค้า)
-- Foreign Key: category_id -> categories(id)
CREATE TABLE `products` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `created_at` DATETIME(3) NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at` DATETIME(3) NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    `deleted_at` DATETIME(3) NULL,
    `name` VARCHAR(255) NOT NULL,
    `price` DECIMAL(10, 2) NOT NULL,
    `quantity` INT NOT NULL DEFAULT 0,
    `category_id` BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (`id`),
    INDEX `idx_products_deleted_at` (`deleted_at`),
    INDEX `idx_products_category_id` (`category_id`),
    CONSTRAINT `fk_products_category`
        FOREIGN KEY (`category_id`)
        REFERENCES `categories` (`id`)
        ON UPDATE CASCADE
        ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 3. ตาราง customers (ลูกค้า)
CREATE TABLE `customers` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `created_at` DATETIME(3) NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at` DATETIME(3) NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    `deleted_at` DATETIME(3) NULL,
    `name` VARCHAR(255) NOT NULL,
    `email` VARCHAR(255) NOT NULL UNIQUE,
    `phone` VARCHAR(50) NOT NULL,
    PRIMARY KEY (`id`),
    INDEX `idx_customers_deleted_at` (`deleted_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- 4. ตาราง orders (คำสั่งซื้อ)
-- Foreign Keys: customer_id -> customers(id), product_id -> products(id)
CREATE TABLE `orders` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `created_at` DATETIME(3) NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at` DATETIME(3) NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    `deleted_at` DATETIME(3) NULL,
    `customer_id` BIGINT UNSIGNED NOT NULL,
    `product_id` BIGINT UNSIGNED NOT NULL,
    `quantity` INT NOT NULL,
    `total_price` DECIMAL(12, 2) NOT NULL,
    `status` VARCHAR(50) NOT NULL DEFAULT 'Pending',
    PRIMARY KEY (`id`),
    INDEX `idx_orders_deleted_at` (`deleted_at`),
    INDEX `idx_orders_customer_id` (`customer_id`),
    INDEX `idx_orders_product_id` (`product_id`),
    CONSTRAINT `fk_orders_customer`
        FOREIGN KEY (`customer_id`)
        REFERENCES `customers` (`id`)
        ON UPDATE CASCADE
        ON DELETE RESTRICT,
    CONSTRAINT `fk_orders_product`
        FOREIGN KEY (`product_id`)
        REFERENCES `products` (`id`)
        ON UPDATE CASCADE
        ON DELETE RESTRICT
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;


-- -----------------------------------------------------
-- 3. ข้อมูลตั้งต้นของระบบ (Initial Seed Data)
-- -----------------------------------------------------

-- ข้อมูลหมวดหมู่สินค้า (Categories)
INSERT INTO `categories` (`id`, `created_at`, `updated_at`, `name`, `description`) VALUES
(1, NOW(), NOW(), 'Laptops & Computers', 'โน้ตบุ๊ก คอมพิวเตอร์ตั้งโต๊ะ และอุปกรณ์ประมวลผล'),
(2, NOW(), NOW(), 'Smartphones & Tablets', 'โทรศัพท์มือถือ แท็บเล็ต และอุปกรณ์พกพา'),
(3, NOW(), NOW(), 'Computer Accessories', 'เมาส์ คีย์บอร์ด หูฟัง และอุปกรณ์เสริมคอมพิวเตอร์'),
(4, NOW(), NOW(), 'Networking & Storage', 'เราเตอร์ ฮับ อุปกรณ์แปลงสัญญาณ และหน่วยความจำภายนอก');

-- ข้อมูลสินค้า (Products)
INSERT INTO `products` (`id`, `created_at`, `updated_at`, `name`, `price`, `quantity`, `category_id`) VALUES
(1,  NOW(), NOW(), 'MacBook Air M3 13-inch', 39900.00, 10, 1),
(2,  NOW(), NOW(), 'ASUS Vivobook 15 OLED', 21900.00, 15, 1),
(3,  NOW(), NOW(), 'Dell XPS 13 Plus', 52900.00, 5, 1),
(4,  NOW(), NOW(), 'Lenovo Legion 5 Pro', 44900.00, 8, 1),
(5,  NOW(), NOW(), 'iPhone 16 Pro 128GB', 39900.00, 20, 2),
(6,  NOW(), NOW(), 'Samsung Galaxy S24 Ultra', 43900.00, 12, 2),
(7,  NOW(), NOW(), 'iPad Air 11-inch M2', 23900.00, 14, 2),
(8,  NOW(), NOW(), 'Xiaomi 14 Ultra', 34900.00, 7, 2),
(9,  NOW(), NOW(), 'Logitech MX Master 3S Mouse', 3590.00, 25, 3),
(10, NOW(), NOW(), 'Mechanical Keyboard Keychron K2 Pro', 3890.00, 18, 3),
(11, NOW(), NOW(), 'Razer DeathAdder V3 Pro Mouse', 4590.00, 4, 3),       -- สต็อกน้อย (< 5)
(12, NOW(), NOW(), 'Corsair K70 RGB PRO Mechanical Keyboard', 5490.00, 3, 3), -- สต็อกน้อย (< 5)
(13, NOW(), NOW(), 'Anker USB-C Hub 7-in-1', 1590.00, 30, 4),
(14, NOW(), NOW(), 'TP-Link Archer AX73 Wi-Fi 6 Router', 3990.00, 9, 4),
(15, NOW(), NOW(), 'SanDisk Extreme Portable SSD 1TB', 3490.00, 2, 4);     -- สต็อกน้อย (< 5)

-- ข้อมูลลูกค้า (Customer: 8 รายการ)
INSERT INTO `customers` (`id`, `created_at`, `updated_at`, `name`, `email`, `phone`) VALUES
(1, NOW(), NOW(), 'สมชาย สายฟ้า', 'somchai.s@email.com', '081-111-2222'),
(2, NOW(), NOW(), 'วิภาดา รักดี', 'wiphada.r@email.com', '082-333-4444'),
(3, NOW(), NOW(), 'ธีรภัทร เจริญสุข', 'theerapat.c@email.com', '089-555-6666'),
(4, NOW(), NOW(), 'อรทัย มั่งมี', 'orathai.m@email.com', '086-777-8888'),
(5, NOW(), NOW(), 'ณัฐพงศ์ เก่งกล้า', 'nattapong.k@email.com', '084-999-0000'),
(6, NOW(), NOW(), 'กัญญา บุญเลิศ', 'kanya.b@email.com', '085-123-4567'),
(7, NOW(), NOW(), 'เอกชัย ยิ่งยง', 'ekkachai.y@email.com', '087-234-5678'),
(8, NOW(), NOW(), 'ปิยะพร แสงทอง', 'piyaporn.s@email.com', '083-345-6789');

-- ข้อมูลคำสั่งซื้อ (Order: 12 รายการ)
-- สถานะ: Pending, Completed, Cancelled
INSERT INTO `orders` (`id`, `created_at`, `updated_at`, `customer_id`, `product_id`, `quantity`, `total_price`, `status`) VALUES
(1,  DATE_SUB(NOW(), INTERVAL 5 DAY), DATE_SUB(NOW(), INTERVAL 5 DAY), 1, 1,  1, 39900.00, 'Completed'),
(2,  DATE_SUB(NOW(), INTERVAL 5 DAY), DATE_SUB(NOW(), INTERVAL 5 DAY), 1, 9,  2,  7180.00, 'Completed'),
(3,  DATE_SUB(NOW(), INTERVAL 4 DAY), DATE_SUB(NOW(), INTERVAL 4 DAY), 2, 5,  1, 39900.00, 'Completed'),
(4,  DATE_SUB(NOW(), INTERVAL 4 DAY), DATE_SUB(NOW(), INTERVAL 4 DAY), 2, 13, 1,  1590.00, 'Completed'),
(5,  DATE_SUB(NOW(), INTERVAL 3 DAY), DATE_SUB(NOW(), INTERVAL 3 DAY), 3, 2,  1, 21900.00, 'Pending'),
(6,  DATE_SUB(NOW(), INTERVAL 3 DAY), DATE_SUB(NOW(), INTERVAL 3 DAY), 4, 10, 1,  3890.00, 'Pending'),
(7,  DATE_SUB(NOW(), INTERVAL 2 DAY), DATE_SUB(NOW(), INTERVAL 2 DAY), 5, 6,  1, 43900.00, 'Completed'),
(8,  DATE_SUB(NOW(), INTERVAL 2 DAY), DATE_SUB(NOW(), INTERVAL 2 DAY), 6, 15, 2,  6980.00, 'Cancelled'),
(9,  DATE_SUB(NOW(), INTERVAL 1 DAY), DATE_SUB(NOW(), INTERVAL 1 DAY), 7, 3,  1, 52900.00, 'Completed'),
(10, DATE_SUB(NOW(), INTERVAL 1 DAY), DATE_SUB(NOW(), INTERVAL 1 DAY), 8, 7,  1, 23900.00, 'Pending'),
(11, NOW(), NOW(), 3, 14, 1,  3990.00, 'Completed'),
(12, NOW(), NOW(), 4, 11, 1,  4590.00, 'Cancelled');


-- -----------------------------------------------------
-- 4. ตัวอย่างคำสั่ง SQL ค้นหาและรายงาน (Query & Analysis)
-- -----------------------------------------------------

-- ค้นหาสินค้าทั้งหมดที่ไม่ถูก Soft Delete
-- SELECT * FROM products WHERE deleted_at IS NULL;

-- ค้นหาตามชื่อสินค้าบางส่วน (LIKE)
-- SELECT * FROM products WHERE deleted_at IS NULL AND name LIKE '%Keyboard%';

-- ค้นหาตามช่วงราคา Min-Max (เช่น 1,000 - 10,000)
-- SELECT * FROM products WHERE deleted_at IS NULL AND price BETWEEN 1000 AND 10000;

-- หาสินค้าสต็อกน้อยกว่ากำหนด (เช่น น้อยกว่า 5 ชิ้น)
-- SELECT * FROM products WHERE deleted_at IS NULL AND quantity < 5;

-- หาสินค้าที่แพงที่สุด N รายการแรก (เช่น Top 3)
-- SELECT * FROM products WHERE deleted_at IS NULL ORDER BY price DESC LIMIT 3;

-- ดึงคำสั่งซื้อพร้อมข้อมูลลูกค้าและสินค้า (JOIN)
-- SELECT 
--     o.id AS order_id,
--     o.created_at,
--     c.name AS customer_name,
--     c.email AS customer_email,
--     p.name AS product_name,
--     o.quantity,
--     o.total_price,
--     o.status
-- FROM orders o
-- INNER JOIN customers c ON o.customer_id = c.id
-- INNER JOIN products p ON o.product_id = p.id
-- WHERE o.deleted_at IS NULL;
