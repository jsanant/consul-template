// go:build ignore

package dependency

import (
	"bytes"
	"encoding/pem"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/consul-template/renderer"
	"github.com/hashicorp/vault/api"
)

func Test_VaultPKI_goodFor(t *testing.T) {
	_, cert, err := getCert([]byte(testCert))
	if err != nil {
		t.Error(err)
	}
	dur, ok := goodFor(cert)
	if ok != false {
		t.Error("should be false")
	}
	// duration should be negative as cert has already expired
	// but still tests cert time parsing (it'd be 0 if there was an issue)
	if dur >= 0 {
		t.Error("cert shouldn't be 0 or positive (old cert)")
	}
}

func Test_VaultPKI_getCert(t *testing.T) {
	for _, testcert := range []string{testCert, testCertSecond, testCertGarbo} {
		pemBlk, cert, err := getCert([]byte(testcert))
		if err != nil {
			t.Error(err)
		}
		got := strings.TrimRight(string(pem.EncodeToMemory(pemBlk)), "\n")
		want := strings.TrimRight(strings.TrimSpace(testCert), "\n")
		if got != want {
			t.Errorf("certs didn't match:\ngot: %v\nwant: %v", got, want)
		}
		if err := cert.VerifyHostname("foo.example.com"); err != nil {
			t.Error(err)
		}
	}
}

func Test_VaultPKI_fetchPEM(t *testing.T) {
	setupVaultPKI(t)

	clients := testClients
	data := map[string]interface{}{
		"common_name": "foo.example.com",
		"ttl":         "2h",
		"ip_sans":     "127.0.0.1,192.168.2.2",
	}
	d, err := NewVaultPKIQuery("pki/issue/example-dot-com", "/dev/null", data)
	if err != nil {
		t.Error(err)
	}
	encPEM, err := d.fetchPEM(clients)
	if err != nil {
		t.Error(err)
	}
	if !bytes.Contains(encPEM, []byte("CERTIFICATE")) {
		t.Errorf("certificate not fetched, got: %s", string(encPEM))
	}
	// test path error
	d, err = NewVaultPKIQuery("pki/issue/does-not-exist", "/dev/null", data)
	if err != nil {
		t.Error(err)
	}
	_, err = d.fetchPEM(clients)
	var respErr *api.ResponseError
	if !errors.As(err, &respErr) {
		t.Error(err)
	}
}

func Test_VaultPKI_fetch(t *testing.T) {
	setupVaultPKI(t)

	f, err := os.CreateTemp("", "")
	if err != nil {
		t.Fatal(err)
	}
	os.Remove(f.Name())
	defer os.Remove(f.Name())

	clients := testClients
	/// above is prep work
	data := map[string]interface{}{
		"common_name": "foo.example.com",
		"ttl":         "2h",
		"ip_sans":     "127.0.0.1,192.168.2.2",
	}
	d, err := NewVaultPKIQuery("pki/issue/example-dot-com", f.Name(), data)
	if err != nil {
		t.Fatal(err)
	}
	act1, _, err := d.Fetch(clients, nil)
	if err != nil {
		t.Fatal(err)
	}

	cert1, ok := act1.(string)
	if !ok || !strings.Contains(cert1, "BEGIN") {
		t.Fatalf("expected a cert but found: %s", cert1)
	}

	// Fake template rendering file to disk
	renderer.Render(&renderer.RenderInput{
		Contents: []byte(cert1),
		Path:     f.Name(),
	})

	// re-fetch, should be the same cert pulled from the file
	// if re-fetched from Vault it will be different
	<-d.sleepCh // drain sleepCh so we don't wait
	act2, _, err := d.Fetch(clients, nil)
	if err != nil {
		t.Fatal(err)
	}

	cert2, ok := act2.(string)
	if !ok || !strings.Contains(cert2, "BEGIN") {
		t.Fatalf("expected a cert but found: %s", cert2)
	}

	if cert1 != cert2 {
		t.Errorf("certs don't match and should. cert1: %s, cert2: %s", cert1, cert2)
	}
}

func setupVaultPKI(t *testing.T) {
	clients := testClients

	err := clients.Vault().Sys().Mount("pki", &api.MountInput{
		Type: "pki",
	})
	switch {
	case err == nil:
	case strings.Contains(err.Error(), "path is already in use"):
		// for idempotency
		return
	default:
		t.Fatal(err)
	}

	vc := clients.Vault()
	_, err = vc.Logical().Write("pki/root/generate/internal",
		map[string]interface{}{
			"common_name": "example.com",
			"ttl":         "48h",
		})
	if err != nil {
		t.Fatal(err)
	}
	_, err = vc.Logical().Write("pki/roles/example-dot-com",
		map[string]interface{}{
			"allowed_domains":  "example.com",
			"allow_subdomains": "true",
			"ttl":              "24h",
		})
	if err != nil {
		t.Fatal(err)
	}
}

const testCert = `
-----BEGIN CERTIFICATE-----
MIIDWTCCAkGgAwIBAgIUUARA+vQExU8zjdsX/YXMMu1K5FkwDQYJKoZIhvcNAQEL
BQAwFjEUMBIGA1UEAxMLZXhhbXBsZS5jb20wHhcNMjIwMzAxMjIzMzAzWhcNMjIw
MzA0MjIzMzMzWjAaMRgwFgYDVQQDEw9mb28uZXhhbXBsZS5jb20wggEiMA0GCSqG
SIb3DQEBAQUAA4IBDwAwggEKAoIBAQDD3sktiNGo/CSvtL84+GIcsuDzp1VFjG++
8P682ZPiqPGjrgwe3P8ypyhQv6I8ZGOyu7helMqBN/S1mrhmHWUONy/4o95QWDsJ
CGw4H44dRil5hKC6K8BUrf79XGAGIQJr3T6I5CCwxukfYhU/+xNE3dq5AgLrIIB2
BtzZA6m1T5CmgAzSzI1byTjaRpxOJjucI37iKzkx7AkYS5hGfVsFmJgGi/UXhvzK
uwnHHIq9rLItx7p261dJV8mxRDFaf4x+4bZh2kYkEaG8REOfyHSCJ78RniWbF/DN
Jtgh8bT2/938/ecBtWcTN+psICD62DJii6988FD2qS+Yd8Eu8M5rAgMBAAGjgZow
gZcwDgYDVR0PAQH/BAQDAgOoMB0GA1UdJQQWMBQGCCsGAQUFBwMBBggrBgEFBQcD
AjAdBgNVHQ4EFgQUfmm32UJb3xJNxfA7ZB0Q5RXsQIkwHwYDVR0jBBgwFoAUDoYJ
CtobWJrR1xmTsYJd9buj2jwwJgYDVR0RBB8wHYIPZm9vLmV4YW1wbGUuY29thwR/
AAABhwTAqAEpMA0GCSqGSIb3DQEBCwUAA4IBAQBzB+RM2PSZPmDG3xJssS1litV8
TOlGtBAOUi827W68kx1lprp35c9Jyy7l4AAu3Q1+az3iDQBfYBazq89GOZeXRvml
x9PVCjnXP2E7mH9owA6cE+Z1cLN/5h914xUZCb4t9Ahu04vpB3/bnoucXdM5GJsZ
EJylY99VsC/bZKPCheZQnC/LtFBC31WEGYb8rnB7gQxmH99H91+JxnJzYhT1a6lw
arHERAKScrZMTrYPLt2YqYoeyO//aCuT9YW6YdIa9jPQhzjeMKXywXLetE+Ip18G
eB01bl42Y5WwHl0IrjfbEevzoW0+uhlUlZ6keZHr7bLn/xuRCUkVfj3PRlMl
-----END CERTIFICATE-----
`

const testCertSecond = `
-----BEGIN RSA PRIVATE KEY-----
MIIEpQIBAAKCAQEAw97JLYjRqPwkr7S/OPhiHLLg86dVRYxvvvD+vNmT4qjxo64M
Htz/MqcoUL+iPGRjsru4XpTKgTf0tZq4Zh1lDjcv+KPeUFg7CQhsOB+OHUYpeYSg
uivAVK3+/VxgBiECa90+iOQgsMbpH2IVP/sTRN3auQIC6yCAdgbc2QOptU+QpoAM
0syNW8k42kacTiY7nCN+4is5MewJGEuYRn1bBZiYBov1F4b8yrsJxxyKvayyLce6
dutXSVfJsUQxWn+MfuG2YdpGJBGhvERDn8h0gie/EZ4lmxfwzSbYIfG09v/d/P3n
AbVnEzfqbCAg+tgyYouvfPBQ9qkvmHfBLvDOawIDAQABAoIBAQC4QZT469NnZ0LP
s3WLj0UkgDXTn988nL7mXWkVmIxg1dLyyiEGy5iaOttXEt74duu+0I7BErFpW40t
ZY4AKbjN5aaP/P9+j3GBrtW2+iBDc6RCdzyHxe6Y+lF8X/DI8zaG58sTFZ+XDJdy
+V7KIFPhHd7K2ZSLQbj2zr/kumhkcP1powom44jt7dtjtsR7w2LVA0RkR1HunCgm
GlhwBJKbmyRRz3Uan2OvzuYSC0hoQXHBYIT2GCA2e5pY8/aU6MHgNwIFw3kDxP/B
3J9ePHZaadg8HGZ2vFnOrU5ird6NeReDC5o3t7GDcSH+H7qtNRavxz3D4hJyVlrH
JYJCrjFxAoGBANDK3KW5KZINVSTg2e1HgOHWqq+fwLOTwksT8bMiRgKUfWTdvgBN
4MbayijYl2kYK/9OEYoL1M5dbjfOPgTPtAlIiumQawwWLRUOQA01tFTgINJfV3kN
X4yOTzLtoZYHkEb0z1yhS5XNKeUiDXvlTfQXgv9jKczyN6K5zwcw+fjzAoGBAPAn
+WiDI1zEAknO0fj3NTIjn4d0KcKYoKW8ntTCO1iGLTWJ9qRPQDnvalD2wLdyS744
D7BORAxQR+z64N7mqEFfL8/oj7MTPKLw/kHOFOsbAa+LSVd2QGwSmRbNQessk9DM
tvd88eUgMMD0zvhC4ZCRxk4mKJo7JN8X+CsuhDKpAoGBAMC99HBb/P8hpa70jtjX
ACf69fhIPijIRzztfVsDUaPCFfuOI36+ZbjMcoDAaQ2QTdVR6SkJgPq8DyofDut8
HdPQDsRMGDXBJv7f98r5/622dTYe424RJVpoaL431cnc05hdGCuHjnIMQheOlun/
pTWmmrxNe2IBW9CxPGeEE853AoGAPCfHMYantPTkHdjQf6xshsKlkyhlzXitxNYa
cvC0LNhvOpn0TfQMAncWCnHElC7tChjA1UjFgtAZNCMjcLIWM0nEkC+QzypiZe43
wgP8+WcqZO5e0KmuOWPvNOb1PBNOc17T9eo2LU6C59JqhYU7OxtIsQqd4QQvmDJI
14gvVQECgYEAkOKJXlimzn09MD0XgftqoWy50A+cygGuwiDE/VVWyB8LR74l1OZh
NZzlExhxuW3KlafLh+aiHmT+aZ+SVjIv0sDH741qerrQYgaLHWlIeo1ALte6sFGB
P3l4OcR/dmUiJV3rELS8g7sofHP/s/SvXLzG8PJFqMH5N4FqKV4Yih0=
-----END RSA PRIVATE KEY-----
` + testCert

const testCertGarbo = `
aa983w4;/amndsfm908q26035vc;ng902338(%@%!@QY!&DVLMNSALX>PT(RQ!QO*%@
` + testCert + `
!Q)(*@^YUO!Q#MN%$#WP(G^&+_!%)!+^%$Y	:!#QLKENFVJ)	!#*&%YHTM
`
