package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
        "flag"
)

func main() {
        usrNamePtr := flag.String("name","anaymous","user name")
        usrPasswdPtr := flag.String("passwd","anaymous","user password")
        flag.Parse()

        v := url.Values{}
        v.Set("principal", *usrNamePtr)
        v.Set("password", *usrPasswdPtr)

	body := ioutil.NopCloser(strings.NewReader(v.Encode())) //endode v:[body struce]
        fmt.Println(v)
 
	client := &http.Client{}
	reqest, err := http.NewRequest("POST", "http://localhost/login", body)
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
	}

	reqest.Header.Set("Content-Type", "application/x-www-form-urlencoded;param=value") //setting post head

	resp, err := client.Do(reqest)
	defer resp.Body.Close() //close resp.Body

	fmt.Println("login status: ", resp.StatusCode) //print status code

	//content_post, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//    fmt.Println("Fatal error ", err.Error())
	//}

	//fmt.Println(string(content_post))    //print reply

	response, err := http.Get("http://localhost/api/logout")
	if err != nil {
		fmt.Println("Fatal error ", err.Error())
	}

	defer response.Body.Close()

	fmt.Println("logout status: ", resp.StatusCode) //print status code
	//content_get, err := ioutil.ReadAll(response.Body)
	//fmt.Println(string(content_get))

}
