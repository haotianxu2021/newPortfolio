// db/sqlc/user_test.go

package db

import (
	"context"
	"database/sql"
	"testing"

	"github.com/haotianxu2021/newPortfolio/util"
	"github.com/stretchr/testify/require"
)

func createRandomUser(t *testing.T) User {
	arg := CreateUserParams{
		Username:     "user_" + util.RandomString(5),
		Email:        util.RandomString(5) + "@example.com",
		PasswordHash: util.RandomString(10),
		FirstName: sql.NullString{
			String: "Test",
			Valid:  true,
		},
		LastName: sql.NullString{
			String: "User",
			Valid:  true,
		},
		Bio: sql.NullString{
			String: "Test bio",
			Valid:  true,
		},
	}

	user, err := testQueries.CreateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user)

	require.Equal(t, arg.Username, user.Username)
	require.Equal(t, arg.Email, user.Email)
	require.Equal(t, arg.PasswordHash, user.PasswordHash)
	require.Equal(t, arg.FirstName, user.FirstName)
	require.Equal(t, arg.LastName, user.LastName)
	require.Equal(t, arg.Bio, user.Bio)

	require.NotZero(t, user.ID)
	require.NotZero(t, user.CreatedAt)

	return user
}

func TestUpdateUser(t *testing.T) {
	user1 := createRandomUser(t)

	arg := UpdateUserParams{
		ID:       user1.ID,
		Username: "updated_" + util.RandomString(5),
		Email:    "updated_" + util.RandomString(5) + "@example.com",
		FirstName: sql.NullString{
			String: "UpdatedFirst",
			Valid:  true,
		},
		LastName: sql.NullString{
			String: "UpdatedLast",
			Valid:  true,
		},
		Bio: sql.NullString{
			String: "Updated bio",
			Valid:  true,
		},
	}

	user2, err := testQueries.UpdateUser(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, user2)

	require.Equal(t, user1.ID, user2.ID)
	require.Equal(t, arg.Username, user2.Username)
	require.Equal(t, arg.Email, user2.Email)
	require.Equal(t, arg.FirstName, user2.FirstName)
	require.Equal(t, arg.LastName, user2.LastName)
	require.Equal(t, arg.Bio, user2.Bio)
}

func TestUpdateUserPassword(t *testing.T) {
	user1 := createRandomUser(t)
	newPassword := util.RandomString(10)

	result, err := testQueries.UpdateUserPassword(context.Background(), UpdateUserPasswordParams{
		ID:           user1.ID,
		PasswordHash: newPassword,
	})

	require.NoError(t, err)
	require.NotEmpty(t, result)
	require.Equal(t, user1.ID, result.ID)
	require.Equal(t, user1.Email, result.Email)
	require.Equal(t, user1.Username, result.Username)
}
