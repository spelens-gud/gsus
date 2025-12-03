CREATE TABLE `orders` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '订单ID',
    `user_id` bigint NOT NULL COMMENT '用户ID',
    `order_no` varchar(50) NOT NULL COMMENT '订单号',
    `total_amount` decimal(10,2) NOT NULL COMMENT '总金额',
    `status` enum('pending','paid','shipped','completed') DEFAULT 'pending' COMMENT '订单状态',
    `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `order_no` (`order_no`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_order_no` (`order_no`),
    CONSTRAINT `orders_ibfk_1` FOREIGN KEY (`user_id`) REFERENCES `users` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='订单表'

CREATE TABLE `products` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '商品ID',
    `name` varchar(200) NOT NULL COMMENT '商品名称',
    `description` text COMMENT '商品描述',
    `price` decimal(10,2) NOT NULL COMMENT '价格',
    `stock` int DEFAULT '0' COMMENT '库存',
    `category` varchar(50) DEFAULT NULL COMMENT '分类',
    `status` enum('active','inactive') DEFAULT 'active' COMMENT '状态',
    `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_name` (`name`),
    KEY `idx_category` (`category`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='商品表'

CREATE TABLE `users` (
    `id` bigint NOT NULL AUTO_INCREMENT COMMENT '用户ID',
    `username` varchar(50) NOT NULL COMMENT '用户名',
    `email` varchar(100) DEFAULT NULL COMMENT '邮箱',
    `password` varchar(255) NOT NULL COMMENT '密码',
    `age` int DEFAULT NULL COMMENT '年龄',
    `balance` decimal(10,2) DEFAULT '0.00' COMMENT '余额',
    `is_active` tinyint(1) DEFAULT '1' COMMENT '是否激活',
    `created_at` datetime DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    `updated_at` datetime DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `email` (`email`),
    KEY `idx_username` (`username`),
    KEY `idx_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='用户表'
