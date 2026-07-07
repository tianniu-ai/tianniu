package model

type KVData struct {
	Key   string `gorm:"primaryKey"`
	Value string `gorm:"not null"`
}
