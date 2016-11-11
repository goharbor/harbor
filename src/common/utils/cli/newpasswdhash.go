// copyright: none included feel to relicense as you like.

package main

import ("fmt"
        "crypto/sha1"
        "bufio"
        "strings"
        "os"

        "golang.org/x/crypto/pbkdf2"
        "golang.org/x/crypto/sha3" 
)

func main() {

    reader := bufio.NewReader(os.Stdin)
    fmt.Print("Password: ")
    passwd, _ := reader.ReadString('\n')
    passwd = strings.TrimSpace(passwd)
    fmt.Print("Hash: ")
    salt, _ := reader.ReadString('\n')
    salt = strings.TrimSpace(salt)

    hash := ""
    hash = fmt.Sprintf("%x",pbkdf2.Key([]byte(passwd), []byte(salt), 4096, 16, sha1.New))
    // rip off from https://godoc.org/golang.org/x/crypto/sha3#ex-package--Sum
    buf := []byte(passwd)
    h := make([]byte, 64)
    sha3.ShakeSum256(h, buf)
    hashshakes256 := ""
    hashshakes256 = fmt.Sprintf("%x",h)
    fmt.Println("Your new password pbkdf2/sha1 hash is: ")
    fmt.Println(hash)
    fmt.Println("Your new unsalted password shakesum256 is: ")
    fmt.Println(hashshakes256)
}
