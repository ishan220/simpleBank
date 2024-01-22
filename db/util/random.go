package util

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
)

const alphabet = "dewwbrwgeiugbweiqwirueiworiotiigjadsfgffljghjhghzcxnvmcxbvbn"
const digit = "0123456789"

func init() {
	var randNo = rand.Intn(64)
	fmt.Println("Inside init method of util", randNo)

}

func RandomInt(min int64, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)
	for i := 0; i < n; i++ {
		sb.WriteByte(alphabet[rand.Intn(k)])
	}
	return sb.String()
}
func RandomOwner() string {
	return RandomString(6)
}

func RandomMoney() int64 {
	return RandomInt(0, 1000)
}

func RandomCurrency() string {
	currencies := []string{"EUR", "USD", "CAD"}
	n := len(currencies)
	return currencies[rand.Intn(n)]
}

func RandomAccountID() int64 {
	var AccountID strings.Builder
	k := len(digit)
	for i := 0; i < 10; i++ {
		AccountID.WriteByte(digit[rand.Intn(k)])
	}
	genAccount, err := strconv.ParseInt((AccountID.String()), 10, 64)
	if err != nil {
		log.Fatal("Unable to Generate Random Account ID ")
	}
	return genAccount
}

func RandomEmail() string {
	return fmt.Sprintf("%s@email.com", RandomString(6))
}
