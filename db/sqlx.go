package db

import (
	"gorm.io/gorm"
)

type ExtendDB struct {
	DB *gorm.DB
}

func NewExtendDB(db *gorm.DB) *ExtendDB {
	return &ExtendDB{db}
}

// Close 关闭数据库（注入到 bean 统一关闭）
func (edb *ExtendDB) Close() {
	// 创建 gorm.DB 对象的时候连接并没有被创建，在具体使用的时候才会创建。
	// gorm 内部，准确的说是 database/sql 内部会维护一个连接池，可以通过参数设置最大空闲连接数，连接最大空闲时间等。
	// 使用者不需要管连接的创建和关闭。
}
