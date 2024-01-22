package util

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestPaassword(t *testing.T) {
	randomPass := RandomString(6)

	hashedPassword, err := HashedPassword(randomPass)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword)
	fmt.Printf("hashedPassword:%v,randomPass%v", hashedPassword, randomPass)

	err = CheckPassword(randomPass, hashedPassword)
	fmt.Printf("error%s", err)

	require.NoError(t, err)

	wrongPassword := RandomString(6)
	err = CheckPassword(wrongPassword, hashedPassword)
	require.EqualError(t, err, bcrypt.ErrMismatchedHashAndPassword.Error())

	hashedPassword2, err := HashedPassword(randomPass)
	require.NoError(t, err)
	require.NotEmpty(t, hashedPassword2)
	require.NotEqual(t, hashedPassword, hashedPassword2)
}
