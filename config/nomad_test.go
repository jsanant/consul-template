package config

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/hashicorp/go-retryablehttp"
)

func TestNomadConfig_Copy(t *testing.T) {

	cases := []struct {
		name string
		a    *NomadConfig
	}{
		{
			"nil",
			nil,
		},
		{
			"empty",
			&NomadConfig{},
		},
		{
			"full",
			&NomadConfig{
				Address:      String("address"),
				Namespace:    String("foo"),
				Token:        String("token"),
				AuthUsername: String("admin"),
				AuthPassword: String("admin"),
				HttpClient:   retryablehttp.NewClient().StandardClient(),
			},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d_%s", i, tc.name), func(t *testing.T) {
			r := tc.a.Copy()
			if !reflect.DeepEqual(tc.a, r) {
				t.Errorf("\nexp: %#v\nact: %#v", tc.a, r)
			}
		})
	}
}

func TestNomadConfig_Merge(t *testing.T) {

	cases := []struct {
		name string
		a    *NomadConfig
		b    *NomadConfig
		r    *NomadConfig
	}{
		{
			"nil_a",
			nil,
			&NomadConfig{},
			&NomadConfig{},
		},
		{
			"nil_b",
			&NomadConfig{},
			nil,
			&NomadConfig{},
		},
		{
			"nil_both",
			nil,
			nil,
			nil,
		},
		{
			"empty",
			&NomadConfig{},
			&NomadConfig{},
			&NomadConfig{},
		},
		{
			"address_overrides",
			&NomadConfig{Address: String("same")},
			&NomadConfig{Address: String("different")},
			&NomadConfig{Address: String("different")},
		},
		{
			"address_empty_one",
			&NomadConfig{Address: String("same")},
			&NomadConfig{},
			&NomadConfig{Address: String("same")},
		},
		{
			"address_empty_two",
			&NomadConfig{},
			&NomadConfig{Address: String("same")},
			&NomadConfig{Address: String("same")},
		},
		{
			"address_same",
			&NomadConfig{Address: String("same")},
			&NomadConfig{Address: String("same")},
			&NomadConfig{Address: String("same")},
		},
		{
			"namespace_overrides",
			&NomadConfig{Namespace: String("foo")},
			&NomadConfig{Namespace: String("bar")},
			&NomadConfig{Namespace: String("bar")},
		},
		{
			"namespace_empty_one",
			&NomadConfig{Namespace: String("foo")},
			&NomadConfig{},
			&NomadConfig{Namespace: String("foo")},
		},
		{
			"namespace_empty_two",
			&NomadConfig{},
			&NomadConfig{Namespace: String("bar")},
			&NomadConfig{Namespace: String("bar")},
		},
		{
			"namespace_same",
			&NomadConfig{Namespace: String("foo")},
			&NomadConfig{Namespace: String("foo")},
			&NomadConfig{Namespace: String("foo")},
		},
		{
			"token_overrides",
			&NomadConfig{Token: String("same")},
			&NomadConfig{Token: String("different")},
			&NomadConfig{Token: String("different")},
		},
		{
			"token_empty_one",
			&NomadConfig{Token: String("same")},
			&NomadConfig{},
			&NomadConfig{Token: String("same")},
		},
		{
			"token_empty_two",
			&NomadConfig{},
			&NomadConfig{Token: String("same")},
			&NomadConfig{Token: String("same")},
		},
		{
			"token_same",
			&NomadConfig{Token: String("same")},
			&NomadConfig{Token: String("same")},
			&NomadConfig{Token: String("same")},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d_%s", i, tc.name), func(t *testing.T) {
			r := tc.a.Merge(tc.b)
			if !reflect.DeepEqual(tc.r, r) {
				t.Errorf("\nexp: %#v\nact: %#v", tc.r, r)
			}
		})
	}
}

func TestNomadConfig_Finalize(t *testing.T) {

	cases := []struct {
		name string
		i    *NomadConfig
		r    *NomadConfig
	}{
		{
			"empty",
			&NomadConfig{},
			&NomadConfig{
				Address:      String(""),
				Namespace:    String(""),
				Token:        String(""),
				AuthUsername: String(""),
				AuthPassword: String(""),
				HttpClient:   http.DefaultClient,
			},
		},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("%d_%s", i, tc.name), func(t *testing.T) {
			tc.i.Finalize()
			if !reflect.DeepEqual(tc.r, tc.i) {
				t.Errorf("\nexp: %#v\nact: %#v", tc.r, tc.i)
			}
		})
	}
}
