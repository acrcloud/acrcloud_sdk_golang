package main

import (
    "fmt"
    "os"
    "time"
    "io/ioutil"
    "./acrcloud"
)

func main() {
    b, _ := ioutil.ReadFile(os.Args[1])
    fmt.Println(len(b))

    access_key := "xxx"
    access_secret := "xxx"
    host := "xxx"

    configs := map[string]string {
        "access_key": access_key,
        "access_secret": access_secret,
        "host": host,
        "recognize_type": acrcloud.ACR_OPT_REC_HUMMING,
    }

    start := time.Now().UnixNano() / 1e6
    fmt.Println(time.Now())
    var recHandler = acrcloud.NewRecognizer(configs)

    result, _ := recHandler.RecognizeByFileBuffer(b, 0, 20)
    end := time.Now().UnixNano() / 1e6
    fmt.Println(end - start)
    fmt.Println(time.Now())

    fmt.Println(result)
}
