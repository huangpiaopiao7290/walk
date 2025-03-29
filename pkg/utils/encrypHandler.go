// Auth: pp
// Created: 2025/03/27
// Description: encryption handler in data

package utils

// import (
// 	"time"
// 	"crypto/aes"
// 	"crypto/cipher"
// 	"crypto/rand"
// 	"crypto/bcrypt"
// 	"encoding/base64"
// 	"errors"
// 	"fmt"
// 	"io"
// 	"log"
// 	"os"

// 	jwt "github.com/golang-jwt/jwt/v5"
// )

// type Claims struct {

	
// }


// generate token
// @param userID uint: user id
// @param ttl int: token ttl
// @return string: token
// func (e *Encrypter) GenerateToken(userID uint, ttl int) (string, error) {
// 	expirationTime := time.Now().Add(time.Second * time.Duration(ttl))		// token expiration time in one hour


// }


// func Encrypt(salt string, data string) string {
// 	dk, _ := NewDK(salt)
// 	key := dk.Key()
// 	block, err := aes.NewCipher(key)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	iv := make([]byte, aes.BlockSize)
// 	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
// 		log.Fatal(err)
// 	}
// }