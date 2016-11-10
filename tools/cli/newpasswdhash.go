// copyright: none included feel to relicense as you like.

package main

import ("fmt"
        "crypto/sha1"
        "bufio"
        "strings"
        "os"

        "golang.org/x/crypto/pbkdf2"
)

func main() {

    reader := bufio.NewReader(os.Stdin)
    fmt.Print("Password: ")
    passwd, _ := reader.ReadString('\n')
    passwd = strings.TrimSpace(passwd)
    fmt.Print("Hash: ")
    salt, _ := reader.ReadString('\n')
    salt = strings.TrimSpace(salt)

    fmt.Println(passwd)
    fmt.Println(salt)
    hash := ""
    hash = fmt.Sprintf("%x",pbkdf2.Key([]byte(passwd), []byte(salt), 4096, 16, sha1.New))
    fmt.Println("Your new password hash is: ")
    fmt.Println(hash)
}
