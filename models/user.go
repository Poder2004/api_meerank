package model

// User struct maps to the users table in the database.
type User struct {
	UID          int64   `json:"uid" db:"uid" gorm:"primaryKey"`
	Name         string  `json:"name" db:"name"`
	Phone        *string `json:"phone,omitempty" db:"phone"`
	Password     string  `json:"-" db:"password"` // ซ่อนฟิลด์นี้จาก JSON response เสมอ
	Age          *int    `json:"age,omitempty" db:"age"`
	Gender       *string `json:"gender,omitempty" db:"gender"`
	Minute       int     `json:"minute" db:"minute"`
	Score        int     `json:"score" db:"score"`
	NumberTree   int     `json:"number_tree" db:"number_tree"`
	TreeProgress int     `json:"tree_progress" db:"tree_progress"`
	Role         string  `json:"role" db:"role"`
}

// --- Constants from your request ---

// Constants for the users table
const (
	TableUsers = "users" // ชื่อตาราง

	// Column names
	TableUsersUID        = "uid"
	TableUsersName       = "name"
	TableUsersPhone      = "phone"
	TableUsersPassword   = "password"
	TableUsersAge        = "age"
	TableUsersGender     = "gender"
	TableUsersMinute     = "minute"
	TableUsersScore      = "score"
	TableUsersNumberTree = "number_tree"
	TableUsersRole       = "role"
)

// Constants for user roles (Enum values)
const (
	RoleMember = "member"
	RoleAdmin  = "admin"
)
