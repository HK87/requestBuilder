package nifcloud

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"

	"github.com/nifcloud/nifcloud-sdk-go/nifcloud"
	"github.com/nifcloud/nifcloud-sdk-go/service/computing"
)

const (
	signatureVersion = "2"
	signatureMethod  = "HmacSHA256"
	timeFormat       = "2006-01-02T15:04:05Z"
)

type signer struct {
	// Values that must be populated from the request
	Request     *http.Request
	Time        time.Time
	Credentials aws.CredentialsProvider
	Debug       aws.LogLevel
	Logger      aws.Logger

	Query        url.Values
	stringToSign string
	signature    string
}

func CreateRequest(region string, request string) string {
	// Create config with credentials and region.
	cfg := nifcloud.NewConfig(
		os.Getenv("NIFCLOUD_ACCESS_KEY_ID"),
		os.Getenv("NIFCLOUD_SECRET_ACCESS_KEY"),
		region,
	)
	svc := computing.New(cfg)
	values := url.Values{}
	var mapData map[string]string
	if err := json.Unmarshal([]byte(request), &mapData); err != nil {
		fmt.Println(err)
	}

	for k, v := range mapData {
		values.Set(k, v)
	}

	req := svc.NewRequest(
		&aws.Operation{
			HTTPMethod: "GET",
			HTTPPath:   "/api/",
		},
		&values,
		nil,
	)

	req.HTTPRequest.URL.Query()
	v2 := signer{
		Request:     req.HTTPRequest,
		Time:        req.Time,
		Credentials: req.Config.Credentials,
		Debug:       req.Config.LogLevel,
		Logger:      req.Config.Logger,
		Query:       values,
	}

	req.Error = v2.Sign(req.Context())
	// Create the Computing client with Config value.
	req.HTTPRequest.URL.RawQuery = v2.Query.Encode()
	return req.HTTPRequest.URL.String()
}

func (v2 *signer) Sign(ctx context.Context) error {
	credValue, err := v2.Credentials.Retrieve(ctx)
	if err != nil {
		return err
	}

	v2.Query.Set("AccessKeyId", credValue.AccessKeyID)
	v2.Query.Set("SignatureVersion", signatureVersion)
	v2.Query.Set("SignatureMethod", signatureMethod)
	v2.Query.Set("Timestamp", v2.Time.UTC().Format(timeFormat))
	if credValue.SessionToken != "" {
		v2.Query.Set("SecurityToken", credValue.SessionToken)
	}

	v2.Query.Del("Signature")

	method := v2.Request.Method
	host := v2.Request.URL.Host
	path := v2.Request.URL.Path
	if path == "" {
		path = "/"
	}

	queryKeys := make([]string, 0, len(v2.Query))
	for key := range v2.Query {
		queryKeys = append(queryKeys, key)
	}
	sort.Strings(queryKeys)

	queryKeysAndValues := make([]string, len(queryKeys))
	for i, key := range queryKeys {
		k := strings.Replace(url.QueryEscape(key), "+", "%20", -1)
		v := strings.Replace(url.QueryEscape(v2.Query.Get(key)), "+", "%20", -1)
		queryKeysAndValues[i] = k + "=" + v
	}

	query := strings.Join(queryKeysAndValues, "&")

	v2.stringToSign = strings.Join([]string{
		method,
		host,
		path,
		query,
	}, "\n")

	hash := hmac.New(sha256.New, []byte(credValue.SecretAccessKey))
	hash.Write([]byte(v2.stringToSign))
	v2.signature = base64.StdEncoding.EncodeToString(hash.Sum(nil))
	v2.Query.Set("Signature", v2.signature)

	return nil
}
