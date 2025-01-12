package model

import (
	"context"
	"mime/multipart"

	"github.com/google/uuid"
	"github.com/imkishore16/go-cloudStorage/internal/model/entities"
)

// UserService defines methods the handler layer expects
// any service it interacts with to implement
type UserService interface {
	ClearProfileImage(ctx context.Context, uid uuid.UUID) error
	Get(ctx context.Context, uid uuid.UUID) (*entities.User, error)
	Signup(ctx context.Context, u *entities.User) error
	Signin(ctx context.Context, u *entities.User) error
	UpdateDetails(ctx context.Context, u *entities.User) error
	SetProfileImage(ctx context.Context, uid uuid.UUID, imageFileHeader *multipart.FileHeader) (*entities.User, error)
}

// UserRepository defines methods the service layer expects
// any repository it interacts with to implement
type UserRepository interface {
	FindByID(ctx context.Context, uid uuid.UUID) (*entities.User, error)
	FindByEmail(ctx context.Context, email string) (*entities.User, error)
	Create(ctx context.Context, u *entities.User) error
	Update(ctx context.Context, u *entities.User) error
	UpdateImage(ctx context.Context, uid uuid.UUID, imageURL string) (*entities.User, error)
}

// ImageRepository defines methods it expects a repository
// it interacts with to implement
type ImageRepository interface {
	DeleteProfile(ctx context.Context, objName string) error
	UpdateProfile(ctx context.Context, objName string, imageFile multipart.File) (string, error)
}
