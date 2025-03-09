package secret

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/udhos/boilerplate/awsconfig"
)

func TestHttp(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprintln(w, `{"uri": "mongodb://localhost:27017/?retryWrites=false"}`)
	}))
	defer ts.Close()

	u, errURL := url.Parse(ts.URL)
	if errURL != nil {
		t.Errorf("url: %v", errURL)
		return
	}

	host := u.Hostname()
	port := u.Port()

	name := fmt.Sprintf("#http::GET,http,%s,%s,/,text/plain,eyJwYXJhbWV0ZXIiOiJtb25nb2RiIn0=,Bearer secret:uri", host, port)

	roleArn := os.Getenv("ROLE_ARN")

	log.Printf("ROLE_ARN='%s'", roleArn)

	awsConfOptions := awsconfig.Options{
		RoleArn:         roleArn,
		RoleSessionName: "test",
	}

	secretOptions := Options{
		AwsConfigSource: &AwsConfigSource{AwsConfigOptions: awsConfOptions},
	}
	secret := New(secretOptions)
	value := secret.Retrieve(name)

	const expected = "mongodb://localhost:27017/?retryWrites=false"

	if value != expected {
		t.Errorf("secret error: expected=%s got=%s", expected, value)
	}
}
