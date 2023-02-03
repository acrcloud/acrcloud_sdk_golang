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

    access_key := "XXXXXX"
    access_secret := "XXXXXX"
    host := "XXXXXX"

    configs := map[string]string {
        "access_key": access_key,
        "access_secret": access_secret,
        "host": host,
        "recognize_type": acrcloud.ACR_OPT_REC_AUDIO,
    }

    start := time.Now().UnixNano() / 1e6
    fmt.Println(time.Now())
    var recHandler = acrcloud.NewRecognizer(configs)

    userParams := map[string]string {
        "title": "title",
        "artist": "artist",
        "lyrics": "lyrics",
    }

    //result := recHandler.RecognizeByFileBuffer(b, 0, 20, nil)
    result := recHandler.RecognizeByFileBuffer(b, 0, 12, userParams)
    end := time.Now().UnixNano() / 1e6
    fmt.Println(end - start)
    fmt.Println(time.Now())

    fmt.Println(result)
}
