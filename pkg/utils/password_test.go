package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "password123",
			wantErr:  false,
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false, // bcrypt can hash empty strings
		},
		{
			name:     "long password",
			password: "this-is-a-very-long-password-that-should-still-work-fine",
			wantErr:  false,
		},
		{
			name:     "password with special characters",
			password: "P@ssw0rd!@#$%^&*()",
			wantErr:  false,
		},
		{
			name:     "unicode password",
			password: "пароль密码🔒",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("HashPassword() expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("HashPassword() unexpected error: %v", err)
				return
			}
			
			// Verify hash is not empty
			if hash == "" {
				t.Error("HashPassword() returned empty hash")
			}
			
			// Verify hash is different from original password
			if hash == tt.password {
				t.Error("HashPassword() returned same value as input password")
			}
			
			// Verify hash length is reasonable (bcrypt hashes are typically 60 chars)
			if len(hash) < 50 {
				t.Errorf("HashPassword() returned hash too short: %d chars", len(hash))
			}
		})
	}
}

func TestCheckPasswordHash(t *testing.T) {
	// Test with a known password
	password := "testpassword123"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to create hash for testing: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
	}{
		{
			name:     "correct password",
			password: password,
			hash:     hash,
			want:     true,
		},
		{
			name:     "wrong password",
			password: "wrongpassword",
			hash:     hash,
			want:     false,
		},
		{
			name:     "empty password",
			password: "",
			hash:     hash,
			want:     false,
		},
		{
			name:     "empty hash",
			password: password,
			hash:     "",
			want:     false,
		},
		{
			name:     "invalid hash format",
			password: password,
			hash:     "not-a-valid-bcrypt-hash",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CheckPasswordHash(tt.password, tt.hash)
			if got != tt.want {
				t.Errorf("CheckPasswordHash() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPasswordHashConsistency(t *testing.T) {
	// Test that the same password produces different hashes (salt)
	password := "testpassword"
	
	hash1, err1 := HashPassword(password)
	if err1 != nil {
		t.Fatalf("First hash failed: %v", err1)
	}
	
	hash2, err2 := HashPassword(password)
	if err2 != nil {
		t.Fatalf("Second hash failed: %v", err2)
	}
	
	// Hashes should be different due to random salt
	if hash1 == hash2 {
		t.Error("Same password produced identical hashes - salt not working")
	}
	
	// But both should verify correctly
	if !CheckPasswordHash(password, hash1) {
		t.Error("First hash does not verify correctly")
	}
	
	if !CheckPasswordHash(password, hash2) {
		t.Error("Second hash does not verify correctly")
	}
}

func TestPasswordHashRoundTrip(t *testing.T) {
	// Test complete round trip: hash -> verify
	passwords := []string{
		"simple",
		"password123",
		"P@ssw0rd!",
		"very-long-password-with-many-characters-and-symbols!@#$%^&*()",
		"пароль密码🔒",
	}
	
	for _, password := range passwords {
		t.Run("password_"+password, func(t *testing.T) {
			hash, err := HashPassword(password)
			if err != nil {
				t.Errorf("HashPassword() failed: %v", err)
				return
			}
			
			if !CheckPasswordHash(password, hash) {
				t.Errorf("CheckPasswordHash() failed for password: %s", password)
			}
		})
	}
}

func BenchmarkHashPassword(b *testing.B) {
	password := "benchmarkpassword123"
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := HashPassword(password)
		if err != nil {
			b.Fatalf("HashPassword() failed: %v", err)
		}
	}
}

func BenchmarkCheckPasswordHash(b *testing.B) {
	password := "benchmarkpassword123"
	hash, err := HashPassword(password)
	if err != nil {
		b.Fatalf("Failed to create hash for benchmark: %v", err)
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		CheckPasswordHash(password, hash)
	}
}

// Response utility tests
func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"message": "test"}
	
	WriteJSON(w, http.StatusOK, data)
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("Expected Content-Type to be application/json")
	}
}

func TestWriteSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	
	WriteSuccess(w, "Success message", map[string]string{"id": "123"})
	
	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	
	WriteError(w, http.StatusBadRequest, "Error message")
	
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestWriteCreated(t *testing.T) {
	w := httptest.NewRecorder()
	
	WriteCreated(w, "Created message", map[string]string{"id": "123"})
	
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
	}
}
