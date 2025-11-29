package ds

type User struct {
	ID          uint   `json:"id"`
	Login       string `gorm:"type:varchar(30);unique;not null" json:"login"`
	Password    string `gorm:"type:varchar(100);not null" json:"-"`
	IsModerator bool   `gorm:"default:false;not null" json:"is_moderator"`

	GenerationRequests []GenerationRequest `gorm:"foreignKey:CreatedByID" json:"-"`
}

type CreateUser struct {
	Login    string `json:"login" binding:"required,min=3,max=30"`
	Password string `json:"password" binding:"required,min=8,max=100"`
}

type UpdateUser struct {
	Login    *string `json:"login" binding:"omitnil,min=3,max=30"`
	Password *string `json:"password" binding:"omitnil,min=8,max=100"`
}

const UserIDKey = "user_id"
const UserIsModeratorKey = "user_is_moderator"
