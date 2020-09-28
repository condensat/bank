package model

type RoleName String

const (
	RoleNameDefault RoleName = "default"
	RoleNameAdmin   RoleName = "admin"
)

type UserRole struct {
	UserID UserID   `gorm:"unique_index:idx_user_role;index;not null"` // [FK] Reference to User table
	Role   RoleName `gorm:"unique_index:idx_user_role;not null"`       // RoleName
}
