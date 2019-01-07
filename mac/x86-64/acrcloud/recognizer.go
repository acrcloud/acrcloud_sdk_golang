package acrcloud

/*
#cgo CFLAGS: -I.
#cgo LDFLAGS: -L. -lacrcloud_extr_tool -lpthread
#include <stdio.h>
#include <stdlib.h>
#include "dll_acr_extr_tool.h"
*/
import "C"

import (
    "fmt"
    "bytes"
    "unsafe"
    "strconv"
    "encoding/base64"
    "crypto/hmac"
    "crypto/sha1"
    "io/ioutil"
    "net/http"
    "mime/multipart"
    "time"
)

const ACR_OPT_REC_AUDIO string = "audio"
const ACR_OPT_REC_HUMMING string = "humming"
const ACR_OPT_REC_BOTH string = "both"

type Recognizer struct {
    Host string
    AccessKey string
    AccessSecret string
    RecType string
    TimeoutS int
}

func NewRecognizer(configs map[string]string) *Recognizer {
    var result = new(Recognizer)
    result.Host = configs["host"]
    result.AccessKey = configs["access_key"]
    result.AccessSecret = configs["access_secret"]
    result.RecType = configs["recognize_type"]
    result.TimeoutS = 10

    C.acr_init()
    return result
}

func (self *Recognizer) Post(url string, fieldParams map[string]string, fileParams map[string][]byte, timeoutS int) (string, error) {
    postDataBuffer := bytes.Buffer{}
    mpWriter := multipart.NewWriter(&postDataBuffer)

    for key, val := range fieldParams {
        _ = mpWriter.WriteField(key, val)
    }

    for key, val := range fileParams {
        fw, err := mpWriter.CreateFormFile(key, key)
        if err != nil {
            mpWriter.Close()
            return "", fmt.Errorf("Create Form File Error: %v", err)
        }
        fw.Write(val)
    }


    mpWriter.Close()

    hClient := &http.Client {
        Timeout: time.Duration(10 * time.Second),
    }

    req, err := http.NewRequest("POST", url, &postDataBuffer)
    if err != nil {
        return "", fmt.Errorf("NewRequest Error: %v", err)
    }
    req.Header.Set("Content-Type", mpWriter.FormDataContentType())
    response, err := hClient.Do(req)
    if err != nil {
        return "", fmt.Errorf("Http Client Do Error: %v", err)
    }
    defer response.Body.Close()

    if response.StatusCode != 200 {
        return "", fmt.Errorf("Http Response Status Code Is Not 200: %d", response.StatusCode)
    }

    body, err := ioutil.ReadAll(response.Body)
    if err != nil {
        return "", fmt.Errorf("Read From Http Response Error: %v", err)
    }

    return string(body), nil
}

func (self *Recognizer) GetSign(str string, key string) string {
    hmacHandler := hmac.New(sha1.New, []byte(key))
    hmacHandler.Write([]byte(str))
    return base64.StdEncoding.EncodeToString(hmacHandler.Sum(nil))
}

func (self *Recognizer) CreateHummingFingerprintByBuffer(pcmData []byte, startSeconds int, lenSeconds int) ([]byte, error) {
    if (pcmData == nil || len(pcmData) == 0) {
        return nil, fmt.Errorf("Parameter pcmData is nil or len(pcmData) == 0")
    }

    var fp *C.char
    fpLenC := C.create_humming_fingerprint_by_filebuffer((*C.char)(unsafe.Pointer(&pcmData[0])), C.int(len(pcmData)), C.int(startSeconds), C.int(lenSeconds), &fp)
    fpLen := int(fpLenC)
    if fpLen <= 0 {
        return nil, fmt.Errorf("Can not Create Humming Fingerprint")
    }

    fpBytes := C.GoBytes(unsafe.Pointer(fp), C.int(fpLen))
    C.acr_free(fp)

    return fpBytes, nil
}

func (self *Recognizer) RecognizeByFileBuffer(data []byte, startSeconds int, lenSeconds int) (string, error) {
    qurl := "http://" + self.Host + "/v1/identify"
    http_method := "POST"
    http_uri := "/v1/identify"
    data_type := "fingerprint"
    signature_version := "1"
    timestamp := strconv.FormatInt(time.Now().Unix(), 10)

    string_to_sign := http_method+"\n"+http_uri+"\n"+self.AccessKey+"\n"+data_type+"\n"+signature_version+"\n"+ timestamp
    sign := self.GetSign(string_to_sign, self.AccessSecret)

    humFp,err := self.CreateHummingFingerprintByBuffer(data, startSeconds, lenSeconds)
    if err != nil {
        return "", err
    }

    field_params := map[string]string {
        "access_key": self.AccessKey,
        "sample_hum_bytes": strconv.Itoa(len(humFp)),
        //"sample_bytes": strconv.Itoa(len(data)),
        "timestamp": timestamp,
        "signature": sign,
        "data_type": data_type,
        "signature_version": signature_version,
    }

    file_params := map[string][]byte {
        "sample_hum": humFp,
    }

    result,err := self.Post(qurl, field_params, file_params, self.TimeoutS)
    return result,err
}
