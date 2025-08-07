package logic

import (
	"errors"
	"fmt"
	"github.com/fusioncatltd/fusioncat/common"
	"github.com/fusioncatltd/fusioncat/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"strings"
)

const (
	STATUS_ACTIVE = "active"
)

// UserObject represents a user in the system.
// It performs operations on a single user object, such as serialization,
// authentication, and other user-specific actions.
type UserObject struct {
	Model *db.UsersDBModel
}

type UserDBSerializerStruct struct {
	ID     string `json:"id"`
	Handle string `json:"handle"`
	Status string `json:"status"`
}

func (user *UserObject) Serialize() *UserDBSerializerStruct {
	return &UserDBSerializerStruct{
		ID:     user.Model.ID.String(),
		Handle: user.Model.Handle,
		Status: user.Model.Status,
	}
}

func (user *UserObject) GetID() uuid.UUID {
	return user.Model.ID
}

// VerifyPassword checks if the provided password matches the stored password hash.
func (user *UserObject) VerifyPassword(password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(user.Model.PasswordHash), []byte(password))
	return err == nil
}

// UserObjectsManager manages user objects in the system. It accumulates functions
// which perform operations over multiple user objects, such as creating new users,
// retrieving users by ID or email, etc.
type UserObjectsManager struct {
}

// FindByID retrieves a user from the database by their ID.
func (usersManager *UserObjectsManager) FindByID(id uuid.UUID) (*UserObject, error) {
	userDbRecord := db.UsersDBModel{}
	dbResult := db.GetDB().First(&userDbRecord, id)

	if errors.Is(dbResult.Error, gorm.ErrRecordNotFound) {
		return nil, common.FusioncatErrRecordNotFound
	}

	u := &UserObject{}
	u.Model = &userDbRecord
	return u, nil
}

// RegisterNewUserWithEmailAndPassword creates a new user in the database with the provided email and password.
func (usersManager *UserObjectsManager) RegisterNewUserWithEmailAndPassword(email string, password string) (
	*UserObject, error) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	var userObject *UserObject
	err := db.GetDB().Transaction(func(tx *gorm.DB) error {
		newUser := &db.UsersDBModel{
			Email:        strings.ToLower(email),
			PasswordHash: string(hashedPassword),
			Status:       STATUS_ACTIVE,
			Handle:       generateNewDefaultHandle(),
		}

		if err := tx.Create(newUser).Error; err != nil {
			if pgErr, ok := err.(*pgconn.PgError); ok && pgErr.Code == "23505" {
				return common.FusioncatErrUniqueConstraintViolations
			}
			return err
		}

		userObject = &UserObject{Model: newUser}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return userObject, nil
}

// FindByEmail retrieves a user from the database by their email.
func (usersManager *UserObjectsManager) FindByEmail(email string) (*UserObject, error) {
	userDbRecord := db.UsersDBModel{}
	dbResult := db.GetDB().Where("email = ?", strings.ToLower(email)).First(&userDbRecord)

	if errors.Is(dbResult.Error, gorm.ErrRecordNotFound) {
		return nil, common.FusioncatErrRecordNotFound
	}

	u := &UserObject{}
	u.Model = &userDbRecord
	return u, nil
}

// generateNewSequenceID generates a new sequence ID from the database.
func generateNewSequenceID() uint {
	connection := db.GetDB()
	sequenceName := "handle_sequence"

	// Query the next value of the sequence
	var nextVal uint
	result := connection.Raw(fmt.Sprintf("SELECT nextval('%s')", sequenceName)).Scan(&nextVal)
	if result.Error != nil {
		panic(result.Error)
	}

	return nextVal
}

func generateNewDefaultHandle() string {
	return fmt.Sprintf("%s%d", "user", generateNewSequenceID())
}
