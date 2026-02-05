package main_test

import (
	"testing"

	main "github.com/metal-stack/gardener-extension-csi-driver-synology/test"
)

func TestSynologyClient_CreateUser(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		username    string
		password    string
		description string
		email       string
		wantErr     bool
	}{
		{
			name:        "create a user",
			baseURL:     "http://127.0.0.1:5000",
			username:    "test",
			password:    "Secret1234",
			description: "Test User",
			email:       "test@email.com",
			wantErr:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, err := main.NewSynologyClient(tt.baseURL)
			if err != nil {
				t.Fatalf("could not construct receiver type: %v", err)
			}
			err = c.Login("stefan", "Start1234")
			if err != nil {
				t.Fatalf("could not login: %v", err)
			}

			gotErr := c.CreateUser(tt.username, tt.password, tt.description, tt.email)
			if gotErr != nil {
				if !tt.wantErr {
					t.Errorf("CreateUser() failed: %v", gotErr)
				}
				return
			}
			if tt.wantErr {
				t.Fatal("CreateUser() succeeded unexpectedly")
			}
		})
	}
}
