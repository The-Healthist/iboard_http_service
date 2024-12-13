CREATE DATABASE IF NOT EXISTS iboard_db;
USE iboard_db;

-- 1超级管理员表 super_admins
CREATE TABLE super_admins (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 2楼宇管理员表 building_admins
CREATE TABLE building_admins (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 3楼宇表 buildings
CREATE TABLE buildings (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    ismart_id VARCHAR(100),
    password VARCHAR(255),
    remark TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

-- 4楼宇管理员与楼宇关联表 building_admins_buildings
CREATE TABLE building_admins_buildings (
    building_id BIGINT,
    building_admin_id BIGINT,
    PRIMARY KEY (building_id, building_admin_id),
    FOREIGN KEY (building_id) REFERENCES buildings(id),
    FOREIGN KEY (building_admin_id) REFERENCES building_admins(id)
);

-- 5文件表 files
CREATE TABLE files (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    size BIGINT NOT NULL,
    path VARCHAR(255) NOT NULL,
    mime_type VARCHAR(100),
    oss VARCHAR(255),
    uploader BIGINT NOT NULL,
    uploader_type ENUM('superadmin', 'buildingadmin') NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 6广告表 advertisements
CREATE TABLE advertisements (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(100) NOT NULL,
    description TEXT,
    type ENUM('video', 'image') NOT NULL,
    file_id BIGINT,
    active BOOLEAN DEFAULT true,
    duration INT,
    display ENUM('full', 'top', 'topfull') NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (file_id) REFERENCES files(id)
);

-- 7通知表 notices
CREATE TABLE notices (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    title VARCHAR(100) NOT NULL,
    description TEXT,
    type VARCHAR(50),
    file_id BIGINT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (file_id) REFERENCES files(id)
);

-- 8楼宇广告关联表 building_advertisements
CREATE TABLE building_advertisements (
    building_id BIGINT,
    advertisement_id BIGINT,
    PRIMARY KEY (building_id, advertisement_id),
    FOREIGN KEY (building_id) REFERENCES buildings(id),
    FOREIGN KEY (advertisement_id) REFERENCES advertisements(id)
);

-- 9楼宇通知关联表 building_notices
CREATE TABLE building_notices (
    building_id BIGINT,
    notice_id BIGINT,
    PRIMARY KEY (building_id, notice_id),
    FOREIGN KEY (building_id) REFERENCES buildings(id),
    FOREIGN KEY (notice_id) REFERENCES notices(id)
);