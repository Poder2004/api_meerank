package models

// User struct สำหรับเก็บข้อมูลใน Firestore
type User struct {
	// เปลี่ยน UID เป็น ID ชนิด string และใช้ firestore tag
	ID           string  `firestore:"-" json:"id"` // ID ของ Document จะไม่ถูกเก็บใน field ของ document เอง
	Name         string  `firestore:"name" json:"name"`
	Phone        *string `firestore:"phone,omitempty" json:"phone,omitempty"`
	Age          *int    `firestore:"age,omitempty" json:"age,omitempty"`
	Gender       *string `firestore:"gender,omitempty" json:"gender,omitempty"`
	Minute       int     `firestore:"minute" json:"minute"`
	Score        int     `firestore:"score" json:"score"`
	NumberTree   int     `firestore:"number_tree" json:"number_tree"`
	TreeProgress int     `firestore:"tree_progress" json:"tree_progress"`
	Role         string  `firestore:"role" json:"role"`
	// Password   string  `firestore:"-" json:"-"` // ไม่มี password อีกต่อไป
}

// --- Constants ---

// ชื่อ Collection ใน Firestore
const (
	CollectionUsers = "users"
)

// Constants สำหรับ Role ยังคงใช้งานได้เหมือนเดิม
const (
	RoleMember = "member"
	RoleAdmin  = "admin"
)
