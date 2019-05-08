package acrcloud

/*
#cgo CFLAGS: -I.
#cgo LDFLAGS: -L. -lacrcloud_extr_tool -lpthread -lm
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
        //fmt.Println(val)
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

func (self *Recognizer) CreateHummingFingerprint(pcmData []byte) ([]byte, error) {
    if (pcmData == nil || len(pcmData) == 0) {
        return nil, fmt.Errorf("Parameter pcmData is nil or len(pcmData) == 0")
    }

    var fp *C.char
    fpLenC := C.create_humming_fingerprint((*C.char)(unsafe.Pointer(&pcmData[0])), C.int(len(pcmData)), &fp)
    fpLen := int(fpLenC)
    if fpLen <= 0 {
        return nil, fmt.Errorf("Can not Create Humming Fingerprint")
    }

    fpBytes := C.GoBytes(unsafe.Pointer(fp), C.int(fpLen))
    C.acr_free(fp)

    return fpBytes, nil
}

func (self *Recognizer) CreateAudioFingerprint(pcmData []byte) ([]byte, error) {
    if (pcmData == nil || len(pcmData) == 0) {
        return nil, fmt.Errorf("Parameter pcmData is nil or len(pcmData) == 0")
    }

    var fp *C.char
    fpLenC := C.create_fingerprint((*C.char)(unsafe.Pointer(&pcmData[0])), C.int(len(pcmData)), 0, &fp)
    fpLen := int(fpLenC)
    if fpLen <= 0 {
        return nil, fmt.Errorf("Can not Create Audio Fingerprint")
    }

    fpBytes := C.GoBytes(unsafe.Pointer(fp), C.int(fpLen))
    C.acr_free(fp)

    return fpBytes, nil
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

func (self *Recognizer) CreateAudioFingerprintByBuffer(pcmData []byte, startSeconds int, lenSeconds int) ([]byte, error) {
    if (pcmData == nil || len(pcmData) == 0) {
        return nil, fmt.Errorf("Parameter pcmData is nil or len(pcmData) == 0")
    }

    var fp *C.char
    fpLenC := C.create_fingerprint_by_filebuffer((*C.char)(unsafe.Pointer(&pcmData[0])), C.int(len(pcmData)), C.int(startSeconds), C.int(lenSeconds), 0, &fp)
    fpLen := int(fpLenC)
    if fpLen <= 0 {
        return nil, fmt.Errorf("Can not Create Audio Fingerprint")
    }

    fpBytes := C.GoBytes(unsafe.Pointer(fp), C.int(fpLen))
    C.acr_free(fp)

    return fpBytes, nil
}

func (self *Recognizer) DoRecognize(audioFp []byte, humFp []byte, userParams map[string]string) (string, error) {
    qurl := "http://" + self.Host + "/v1/identify"
    http_method := "POST"
    http_uri := "/v1/identify"
    data_type := "fingerprint"
    signature_version := "1"
    timestamp := strconv.FormatInt(time.Now().Unix(), 10)

    string_to_sign := http_method+"\n"+http_uri+"\n"+self.AccessKey+"\n"+data_type+"\n"+signature_version+"\n"+ timestamp
    sign := self.GetSign(string_to_sign, self.AccessSecret)

    if audioFp == nil && humFp == nil {
        return "", fmt.Errorf("Can not Create Fingerprint")
    }

    field_params := map[string]string {
        "access_key": self.AccessKey,
        "timestamp": timestamp,
        "signature": sign,
        "data_type": data_type,
        "signature_version": signature_version,
    }

    if userParams != nil {
        for key, val := range userParams {
            field_params[key] = val
        }
    }

    file_params := map[string][]byte {}
    if audioFp != nil && len(audioFp) != 0 {
        file_params["sample"] = audioFp
        field_params["sample_bytes"] = strconv.Itoa(len(audioFp))
    }
    if humFp != nil && len(humFp) != 0 {
        file_params["sample_hum"] = humFp
        field_params["sample_hum_bytes"] = strconv.Itoa(len(humFp))
    }

    result,err := self.Post(qurl, field_params, file_params, self.TimeoutS)
    return result,err
}


/*
 *    This function support most of audio / video files.
 *    
 *    Audio: mp3, wav, m4a, flac, aac, amr, ape, ogg ...
 *    Video: mp4, mkv, wmv, flv, ts, avi ...
 *
 *    @param data: file_path query buffer
 *    @param startSeconds: skip (start_seconds) seconds from from the beginning of (data)
 *    @param lenSeconds: use rec_length seconds data to recongize
 *    @param userParams: some User-defined fields.
 *    @return result metainfos
*/
func (self *Recognizer) RecognizeByFileBuffer(data []byte, startSeconds int, lenSeconds int, userParams map[string]string) (string, error) {
    var humFp []byte
    var audioFp []byte
    if self.RecType == ACR_OPT_REC_HUMMING || self.RecType == ACR_OPT_REC_BOTH {
        humFp,_ = self.CreateHummingFingerprintByBuffer(data, startSeconds, lenSeconds)
    }
    if self.RecType == ACR_OPT_REC_AUDIO || self.RecType == ACR_OPT_REC_BOTH {
        audioFp,_ = self.CreateAudioFingerprintByBuffer(data, startSeconds, lenSeconds)
    }

    result,err := self.DoRecognize(audioFp, humFp, userParams)
    return result,err
}

// Only support Microsoft PCM, 16 bit, mono 8000 Hz
func (self *Recognizer) Recognize(data []byte, userParams map[string]string) (string, error) {
    var humFp []byte
    var audioFp []byte
    if self.RecType == ACR_OPT_REC_HUMMING || self.RecType == ACR_OPT_REC_BOTH {
        humFp,_ = self.CreateHummingFingerprint(data)
    }
    if self.RecType == ACR_OPT_REC_AUDIO || self.RecType == ACR_OPT_REC_BOTH {
        audioFp,_ = self.CreateAudioFingerprint(data)
    }

    result,err := self.DoRecognize(audioFp, humFp, userParams)
    return result,err
}
