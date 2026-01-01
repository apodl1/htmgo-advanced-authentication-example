package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/maddalax/htmgo/framework/h"
	"github.com/maddalax/htmgo/framework/service"
	"advancedauth/internal/db"
)

type CreateUserRequest struct {
	Email    string
	Password string
}

type LoginUserRequest struct {
	Email    string
	Password string
	Remember bool
}

type CreatedUser struct {
	Id    string
	Email string
}

func Create(ctx *h.RequestContext, request CreateUserRequest) (int64, error) {
	if len(request.Password) < 6 {
		return 0, errors.New("password must be at least 6 characters long")
	}

	queries := service.Get[db.Queries](ctx.ServiceLocator())

	hashedPassword, err := HashPassword(request.Password)

	if err != nil {
		return 0, errors.New("something went wrong")
	}

	id, err := queries.CreateUser(ctx.Request.Context(), db.CreateUserParams{
		Email:    request.Email,
		Password: hashedPassword,
		Metadata: string("{}"),
	})

	if err != nil {

		if err.Error() == "UNIQUE constraint failed: user.email" {
			return 0, errors.New("email already exists")
		}

		return 0, err
	}

	return id, nil
}

func Login(ctx *h.RequestContext, request LoginUserRequest) (int64, error) {

	queries := service.Get[db.Queries](ctx.ServiceLocator())

	user, err := queries.GetUserByEmail(ctx.Request.Context(), request.Email)

	if err != nil {
		fmt.Printf("error: %s\n", err.Error())
		return 0, errors.New("email or password is incorrect")
	}

	if !PasswordMatches(request.Password, user.Password) {
		return 0, errors.New("email or password is incorrect")
	}

	_, err = CreateSession(ctx, user.ID)

	if err != nil {
		return 0, errors.New("something went wrong")
	}

	if request.Remember {
		_, err = CreateRememberToken(ctx, user.ID)

		if err != nil {
			return 0, err
		}

	}

	return user.ID, nil
}

func ParseMeta(meta any) map[string]interface{} {
	if meta == nil {
		return map[string]interface{}{}
	}

	var data []byte

	switch v := meta.(type) {
	case []byte:
		data = v
	case string:
		data = []byte(v)
	case json.RawMessage:
		data = v
	case map[string]interface{}:
		return v
	default:
		return map[string]interface{}{}
	}

	var dest map[string]interface{}
	err := json.Unmarshal(data, &dest)
	if err != nil {
		return map[string]interface{}{}
	}
	return dest
}


func GetMetaKey(meta map[string]interface{}, key string) string {
	if val, ok := meta[key]; ok {
		return val.(string)
	}
	return ""
}

func SetMeta(ctx *h.RequestContext, userId int64, meta map[string]interface{}) error {
	queries := service.Get[db.Queries](ctx.ServiceLocator())
	serialized, _ := json.Marshal(meta)
	fmt.Printf("serialized: %s\n", string(serialized))
	err := queries.UpdateUserMetadata(context.Background(), db.UpdateUserMetadataParams{
		JsonPatch: serialized,
		ID:        userId,
	})
	if err != nil {
		return err
	}
	return nil
}
