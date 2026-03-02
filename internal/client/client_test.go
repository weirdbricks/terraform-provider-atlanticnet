package client_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/terraform-provider-atlanticnet/internal/client"
)

// newMockServer creates a test HTTP server that returns the given JSON payload.
func newMockServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *client.Client) {
	t.Helper()
	srv := httptest.NewServer(handler)
	c := client.New("test-access-key", "test-private-key")
	c.BaseURL = srv.URL + "/"
	return srv, c
}

func respond(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

// ─── Authentication ──────────────────────────────────────────────────────────

func TestSignatureIsIncludedInRequest(t *testing.T) {
	srv, c := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		sig := r.URL.Query().Get("Signature")
		if sig == "" {
			t.Error("expected Signature parameter, got none")
		}
		guid := r.URL.Query().Get("Rndguid")
		if guid == "" {
			t.Error("expected Rndguid parameter, got none")
		}
		respond(w, map[string]interface{}{
			"list-locationsresponse": map[string]interface{}{
				"KeysSet": map[string]interface{}{},
			},
		})
	})
	defer srv.Close()
	_, _ = c.ListLocations()
}

func TestAPIErrorIsSurfaced(t *testing.T) {
	srv, c := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respond(w, map[string]interface{}{
			"error": map[string]interface{}{
				"code":    "E0002",
				"message": "API key/Signature is invalid",
			},
		})
	})
	defer srv.Close()

	_, err := c.ListLocations()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*client.APIError)
	if !ok {
		t.Fatalf("expected *client.APIError, got %T", err)
	}
	if apiErr.Code != "E0002" {
		t.Errorf("expected error code E0002, got %q", apiErr.Code)
	}
}

// ─── Locations ───────────────────────────────────────────────────────────────

func TestListLocations(t *testing.T) {
	srv, c := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("Action") != "list-locations" {
			t.Errorf("unexpected action: %s", r.URL.Query().Get("Action"))
		}
		respond(w, map[string]interface{}{
			"list-locationsresponse": map[string]interface{}{
				"KeysSet": map[string]interface{}{
					"1item": map[string]interface{}{
						"location_code": "USEAST2",
						"location_name": "USA-East-2",
						"description":   "USA-East-2 (New York, NY)",
						"is_active":     "Y",
					},
					"2item": map[string]interface{}{
						"location_code": "EUWEST1",
						"location_name": "EU-West-1",
						"description":   "EU-West-1 (London, UK)",
						"is_active":     "Y",
					},
				},
			},
		})
	})
	defer srv.Close()

	locs, err := c.ListLocations()
	if err != nil {
		t.Fatalf("ListLocations() error: %v", err)
	}
	if len(locs) != 2 {
		t.Errorf("expected 2 locations, got %d", len(locs))
	}
}

func TestListLocationsEmpty(t *testing.T) {
	srv, c := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respond(w, map[string]interface{}{
			"list-locationsresponse": map[string]interface{}{
				"KeysSet": map[string]interface{}{},
			},
		})
	})
	defer srv.Close()

	locs, err := c.ListLocations()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(locs) != 0 {
		t.Errorf("expected 0 locations, got %d", len(locs))
	}
}

// ─── SSH Keys ────────────────────────────────────────────────────────────────

func TestAddSSHKey(t *testing.T) {
	srv, c := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("Action") != "add-sshkey" {
			t.Errorf("unexpected action: %s", r.URL.Query().Get("Action"))
		}
		respond(w, map[string]interface{}{
			"add-sshkeyresponse": map[string]interface{}{
				"key_id": "abc123",
			},
		})
	})
	defer srv.Close()

	key, err := c.AddSSHKey("my-key", "ssh-rsa AAAAB3Nz...")
	if err != nil {
		t.Fatalf("AddSSHKey() error: %v", err)
	}
	if key.ID != "abc123" {
		t.Errorf("expected key_id abc123, got %q", key.ID)
	}
	if key.Name != "my-key" {
		t.Errorf("expected name my-key, got %q", key.Name)
	}
}

func TestListSSHKeys(t *testing.T) {
	srv, c := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respond(w, map[string]interface{}{
			"list-sshkeysresponse": map[string]interface{}{
				"KeysSet": map[string]interface{}{
					"1item": map[string]interface{}{
						"key_id":          "abc123",
						"key_name":        "my-key",
						"key":             "ssh-rsa AAAA...",
						"key_fingerprint": "aa:bb:cc",
					},
				},
			},
		})
	})
	defer srv.Close()

	keys, err := c.ListSSHKeys()
	if err != nil {
		t.Fatalf("ListSSHKeys() error: %v", err)
	}
	if len(keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(keys))
	}
	if keys[0].Fingerprint != "aa:bb:cc" {
		t.Errorf("unexpected fingerprint: %s", keys[0].Fingerprint)
	}
}

func TestGetSSHKeyNotFound(t *testing.T) {
	srv, c := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respond(w, map[string]interface{}{
			"list-sshkeysresponse": map[string]interface{}{
				"KeysSet": map[string]interface{}{},
			},
		})
	})
	defer srv.Close()

	_, err := c.GetSSHKey("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing key, got nil")
	}
}

func TestDeleteSSHKey(t *testing.T) {
	called := false
	srv, c := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("Action") == "delete-sshkey" {
			called = true
			if r.URL.Query().Get("key_id") != "abc123" {
				t.Errorf("unexpected key_id: %s", r.URL.Query().Get("key_id"))
			}
		}
		respond(w, map[string]interface{}{"delete-sshkeyresponse": map[string]interface{}{"return": "true"}})
	})
	defer srv.Close()

	if err := c.DeleteSSHKey("abc123"); err != nil {
		t.Fatalf("DeleteSSHKey() error: %v", err)
	}
	if !called {
		t.Error("delete-sshkey was never called")
	}
}

// ─── DNS Zones ───────────────────────────────────────────────────────────────

func TestCreateDNSZone(t *testing.T) {
	srv, c := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respond(w, map[string]interface{}{
			"dns-create-zoneresponse": map[string]interface{}{
				"zone_id": "zone-001",
			},
		})
	})
	defer srv.Close()

	zone, err := c.CreateDNSZone("example.com")
	if err != nil {
		t.Fatalf("CreateDNSZone() error: %v", err)
	}
	if zone.ID != "zone-001" {
		t.Errorf("expected zone_id zone-001, got %q", zone.ID)
	}
	if zone.Name != "example.com" {
		t.Errorf("expected zone name example.com, got %q", zone.Name)
	}
}

func TestListDNSZones(t *testing.T) {
	srv, c := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respond(w, map[string]interface{}{
			"dns-list-zonesresponse": map[string]interface{}{
				"zones": map[string]interface{}{
					"1item": map[string]interface{}{
						"zone_id":   "zone-001",
						"zone_name": "example.com",
					},
				},
			},
		})
	})
	defer srv.Close()

	zones, err := c.ListDNSZones()
	if err != nil {
		t.Fatalf("ListDNSZones() error: %v", err)
	}
	if len(zones) != 1 {
		t.Fatalf("expected 1 zone, got %d", len(zones))
	}
}

// ─── DNS Records ─────────────────────────────────────────────────────────────

func TestCreateDNSRecord(t *testing.T) {
	srv, c := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respond(w, map[string]interface{}{
			"dns-create-zone-recordresponse": map[string]interface{}{
				"record_id": "rec-001",
			},
		})
	})
	defer srv.Close()

	rec, err := c.CreateDNSRecord(client.CreateDNSRecordInput{
		ZoneID: "zone-001",
		Type:   "A",
		Host:   "www",
		Data:   "1.2.3.4",
		TTL:    "3600",
	})
	if err != nil {
		t.Fatalf("CreateDNSRecord() error: %v", err)
	}
	if rec.ID != "rec-001" {
		t.Errorf("expected record_id rec-001, got %q", rec.ID)
	}
}

func TestGetDNSRecordNotFound(t *testing.T) {
	srv, c := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respond(w, map[string]interface{}{
			"dns-list-zone-recordsresponse": map[string]interface{}{
				"records": map[string]interface{}{},
			},
		})
	})
	defer srv.Close()

	_, err := c.GetDNSRecord("zone-001", "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing record, got nil")
	}
}

// ─── Instances ───────────────────────────────────────────────────────────────

func TestGetInstance(t *testing.T) {
	srv, c := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		respond(w, map[string]interface{}{
			"describe-instanceresponse": map[string]interface{}{
				"instanceSet": map[string]interface{}{
					"item": map[string]interface{}{
						"InstanceId":     "12345",
						"vm_description": "web-01",
						"vm_ip_address":  "1.2.3.4",
						"vm_plan_name":   "G2.4GB",
						"vm_image":       "ubuntu-22.04_64bit",
						"vm_status":      "RUNNING",
						"vm_cpu_req":     "2",
						"vm_ram_req":     "4096",
						"vm_disk_req":    "100",
						"vm_created_date": "1440018294",
						"rate_per_hr":    "0.0547",
					},
				},
			},
		})
	})
	defer srv.Close()

	inst, err := c.GetInstance("12345")
	if err != nil {
		t.Fatalf("GetInstance() error: %v", err)
	}
	if inst.Status != "RUNNING" {
		t.Errorf("expected status RUNNING, got %q", inst.Status)
	}
	if inst.IPAddress != "1.2.3.4" {
		t.Errorf("unexpected IP address: %s", inst.IPAddress)
	}
}

func TestTerminateInstance(t *testing.T) {
	called := false
	srv, c := newMockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("Action") == "terminate-instance" {
			called = true
		}
		respond(w, map[string]interface{}{
			"terminate-instanceresponse": map[string]interface{}{},
		})
	})
	defer srv.Close()

	if err := c.TerminateInstance("12345"); err != nil {
		t.Fatalf("TerminateInstance() error: %v", err)
	}
	if !called {
		t.Error("terminate-instance was never called")
	}
}
