// Package client provides a typed Go client for the Atlantic.Net Cloud API.
// All requests are HMAC-SHA256 signed per the Atlantic.Net authentication spec.
package client

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	DefaultBaseURL = "https://cloudapi.atlantic.net/"
	APIVersion     = "2010-12-30"
	defaultTimeout = 60 * time.Second
)

// APIError represents a structured error returned by the Atlantic.Net API.
type APIError struct {
	Code    string
	Message string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("Atlantic.Net API error %s: %s", e.Code, e.Message)
}

// Client is a signed HTTP client for the Atlantic.Net Cloud API.
type Client struct {
	AccessKey  string
	PrivateKey string
	BaseURL    string
	HTTPClient *http.Client
}

// New creates a new API client with the given credentials.
func New(accessKey, privateKey string) *Client {
	return &Client{
		AccessKey:  accessKey,
		PrivateKey: privateKey,
		BaseURL:    DefaultBaseURL,
		HTTPClient: &http.Client{Timeout: defaultTimeout},
	}
}

// sign computes the HMAC-SHA256 signature required by the Atlantic.Net API.
func (c *Client) sign(timestamp int64, rndguid string) string {
	msg := strconv.FormatInt(timestamp, 10) + rndguid
	mac := hmac.New(sha256.New, []byte(c.PrivateKey))
	mac.Write([]byte(msg))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

// do performs a signed GET request to the API and decodes the JSON response.
func (c *Client) do(action string, extra map[string]string) (map[string]interface{}, error) {
	ts := time.Now().Unix()
	guid := uuid.New().String()
	sig := c.sign(ts, guid)

	params := url.Values{}
	params.Set("Action", action)
	params.Set("Format", "json")
	params.Set("Version", APIVersion)
	params.Set("ACSAccessKeyId", c.AccessKey)
	params.Set("Timestamp", strconv.FormatInt(ts, 10))
	params.Set("Rndguid", guid)
	params.Set("Signature", sig)
	for k, v := range extra {
		params.Set(k, v)
	}

	reqURL := c.BaseURL + "?" + params.Encode()
	resp, err := c.HTTPClient.Get(reqURL)
	if err != nil {
		return nil, fmt.Errorf("HTTP request to Atlantic.Net API failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read API response body: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to parse API response as JSON: %w (body: %s)", err, string(body))
	}

	// Surface API-level errors
	if errObj, ok := result["error"].(map[string]interface{}); ok {
		return nil, &APIError{
			Code:    fmt.Sprintf("%v", errObj["code"]),
			Message: fmt.Sprintf("%v", errObj["message"]),
		}
	}

	return result, nil
}

// str safely extracts a string from an interface{} map.
func str(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		return fmt.Sprintf("%v", v)
	}
	return ""
}

// ─── Locations ──────────────────────────────────────────────────────────────

// Location represents an Atlantic.Net datacenter.
type Location struct {
	Code        string
	Name        string
	Description string
	IsActive    string
}

// ListLocations returns all available datacenter locations.
func (c *Client) ListLocations() ([]Location, error) {
	resp, err := c.do("list-locations", nil)
	if err != nil {
		return nil, err
	}
	lresp, ok := resp["list-locationsresponse"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response structure from list-locations")
	}
	ks, _ := lresp["KeysSet"].(map[string]interface{})
	var locs []Location
	for _, v := range ks {
		item, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		locs = append(locs, Location{
			Code:        str(item, "location_code"),
			Name:        str(item, "location_name"),
			Description: str(item, "description"),
			IsActive:    str(item, "is_active"),
		})
	}
	return locs, nil
}

// ─── Plans ───────────────────────────────────────────────────────────────────

// Plan represents an Atlantic.Net server plan.
type Plan struct {
	Name      string
	RAM       string
	Disk      string
	CPU       string
	RatePerHr string
}

// ListPlans returns all available server plans.
func (c *Client) ListPlans() ([]Plan, error) {
	resp, err := c.do("describe-plan", nil)
	if err != nil {
		return nil, err
	}
	presp, ok := resp["describe-planresponse"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response structure from describe-plan")
	}
	plans, _ := presp["plans"].(map[string]interface{})
	var out []Plan
	for _, v := range plans {
		item, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		out = append(out, Plan{
			Name:      str(item, "plan_name"),
			RAM:       str(item, "ram"),
			Disk:      str(item, "disk"),
			CPU:       str(item, "cpu"),
			RatePerHr: str(item, "rate_per_hr"),
		})
	}
	return out, nil
}

// ─── Cloud Servers ───────────────────────────────────────────────────────────

// Instance represents an Atlantic.Net Cloud Server.
type Instance struct {
	ID          string
	Name        string
	IPAddress   string
	PlanName    string
	Image       string
	Status      string
	CPU         string
	RAM         string
	Disk        string
	CreatedDate string
	RatePerHr   string
}

// RunInstanceInput holds parameters for creating a new server.
type RunInstanceInput struct {
	ServerName   string
	ImageID      string
	PlanName     string
	VMLocation   string
	SSHKeyID     string
	EnableBackup bool
	Term         string
}

// RunInstance provisions a new Cloud Server and waits for it to reach RUNNING.
func (c *Client) RunInstance(in RunInstanceInput) (*Instance, error) {
	backup := "N"
	if in.EnableBackup {
		backup = "Y"
	}
	term := in.Term
	if term == "" {
		term = "on-demand"
	}
	params := map[string]string{
		"servername":   in.ServerName,
		"imageid":      in.ImageID,
		"planname":     in.PlanName,
		"vm_location":  in.VMLocation,
		"server_qty":   "1",
		"enablebackup": backup,
		"term":         term,
	}
	if in.SSHKeyID != "" {
		params["key_id"] = in.SSHKeyID
	}

	resp, err := c.do("run-instance", params)
	if err != nil {
		return nil, fmt.Errorf("run-instance failed: %w", err)
	}

	rresp, ok := resp["run-instanceresponse"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response from run-instance")
	}
	set, _ := rresp["instancesSet"].(map[string]interface{})
	item, _ := set["item"].(map[string]interface{})
	if item == nil {
		return nil, fmt.Errorf("no instance item in run-instance response")
	}

	instanceID := str(item, "instanceid")
	return c.waitForInstance(instanceID, "RUNNING", 15*time.Minute)
}

// GetInstance returns details for a single Cloud Server.
func (c *Client) GetInstance(id string) (*Instance, error) {
	resp, err := c.do("describe-instance", map[string]string{"instanceid": id})
	if err != nil {
		return nil, fmt.Errorf("describe-instance failed: %w", err)
	}
	dresp, ok := resp["describe-instanceresponse"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response from describe-instance")
	}
	set, _ := dresp["instanceSet"].(map[string]interface{})
	item, _ := set["item"].(map[string]interface{})
	if item == nil {
		return nil, fmt.Errorf("instance %s not found", id)
	}
	return parseInstance(item), nil
}

// ResizeInstance upgrades a Cloud Server to a larger plan (API requires larger disk).
func (c *Client) ResizeInstance(id, planName string) (*Instance, error) {
	_, err := c.do("resize-instance", map[string]string{
		"instanceid": id,
		"planname":   planName,
	})
	if err != nil {
		return nil, fmt.Errorf("resize-instance failed: %w", err)
	}
	return c.waitForInstance(id, "RUNNING", 15*time.Minute)
}

// TerminateInstance deletes a Cloud Server.
func (c *Client) TerminateInstance(id string) error {
	_, err := c.do("terminate-instance", map[string]string{"instanceid": id})
	if err != nil {
		return fmt.Errorf("terminate-instance failed: %w", err)
	}
	return nil
}

// waitForInstance polls describe-instance until the desired status is reached.
func (c *Client) waitForInstance(id, want string, timeout time.Duration) (*Instance, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		inst, err := c.GetInstance(id)
		if err != nil {
			return nil, err
		}
		switch inst.Status {
		case want:
			return inst, nil
		case "FAILED", "REMOVED":
			return nil, fmt.Errorf("instance %s entered terminal status %q while waiting for %q", id, inst.Status, want)
		}
		time.Sleep(15 * time.Second)
	}
	return nil, fmt.Errorf("timed out after %s waiting for instance %s to reach %q", timeout, id, want)
}

func parseInstance(item map[string]interface{}) *Instance {
	return &Instance{
		ID:          str(item, "InstanceId"),
		Name:        str(item, "vm_description"),
		IPAddress:   str(item, "vm_ip_address"),
		PlanName:    str(item, "vm_plan_name"),
		Image:       str(item, "vm_image"),
		Status:      str(item, "vm_status"),
		CPU:         str(item, "vm_cpu_req"),
		RAM:         str(item, "vm_ram_req"),
		Disk:        str(item, "vm_disk_req"),
		CreatedDate: str(item, "vm_created_date"),
		RatePerHr:   str(item, "rate_per_hr"),
	}
}

// ─── SSH Keys ────────────────────────────────────────────────────────────────

// SSHKey represents an Atlantic.Net SSH key.
type SSHKey struct {
	ID          string
	Name        string
	PublicKey   string
	Fingerprint string
}

// ListSSHKeys returns all SSH keys on the account.
func (c *Client) ListSSHKeys() ([]SSHKey, error) {
	resp, err := c.do("list-sshkeys", nil)
	if err != nil {
		return nil, fmt.Errorf("list-sshkeys failed: %w", err)
	}
	kresp, ok := resp["list-sshkeysresponse"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response from list-sshkeys")
	}
	ks, _ := kresp["KeysSet"].(map[string]interface{})
	var keys []SSHKey
	for _, v := range ks {
		item, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		keys = append(keys, SSHKey{
			ID:          str(item, "key_id"),
			Name:        str(item, "key_name"),
			PublicKey:   str(item, "key"),
			Fingerprint: str(item, "key_fingerprint"),
		})
	}
	return keys, nil
}

// GetSSHKey returns a single SSH key by ID.
func (c *Client) GetSSHKey(id string) (*SSHKey, error) {
	keys, err := c.ListSSHKeys()
	if err != nil {
		return nil, err
	}
	for _, k := range keys {
		if k.ID == id {
			return &k, nil
		}
	}
	return nil, fmt.Errorf("SSH key %q not found", id)
}

// AddSSHKey uploads a new SSH key to the account.
func (c *Client) AddSSHKey(name, publicKey string) (*SSHKey, error) {
	resp, err := c.do("add-sshkey", map[string]string{
		"key_name": name,
		"key":      publicKey,
	})
	if err != nil {
		return nil, fmt.Errorf("add-sshkey failed: %w", err)
	}
	aresp, ok := resp["add-sshkeyresponse"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response from add-sshkey")
	}
	return &SSHKey{
		ID:        str(aresp, "key_id"),
		Name:      name,
		PublicKey: publicKey,
	}, nil
}

// DeleteSSHKey removes an SSH key from the account.
func (c *Client) DeleteSSHKey(id string) error {
	_, err := c.do("delete-sshkey", map[string]string{"key_id": id})
	if err != nil {
		return fmt.Errorf("delete-sshkey failed: %w", err)
	}
	return nil
}

// ─── DNS Zones ───────────────────────────────────────────────────────────────

// DNSZone represents an Atlantic.Net DNS zone.
type DNSZone struct {
	ID   string
	Name string
}

// ListDNSZones returns all DNS zones on the account.
func (c *Client) ListDNSZones() ([]DNSZone, error) {
	resp, err := c.do("DNS-list-zones", nil)
	if err != nil {
		return nil, fmt.Errorf("list-dns-zones failed: %w", err)
	}
	zresp, ok := resp["DNS-list-zonesresponse"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response from list-dns-zones")
	}
	zones, _ := zresp["zones"].(map[string]interface{})
	var out []DNSZone
	for _, v := range zones {
		item, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		out = append(out, DNSZone{
			ID:   str(item, "zone_id"),
			Name: str(item, "zone_name"),
		})
	}
	return out, nil
}

// GetDNSZone retrieves a DNS zone by ID.
func (c *Client) GetDNSZone(id string) (*DNSZone, error) {
	zones, err := c.ListDNSZones()
	if err != nil {
		return nil, err
	}
	for _, z := range zones {
		if z.ID == id {
			return &z, nil
		}
	}
	return nil, fmt.Errorf("DNS zone %q not found", id)
}

// CreateDNSZone creates a new DNS zone.
func (c *Client) CreateDNSZone(name string) (*DNSZone, error) {
	resp, err := c.do("DNS-create-zone", map[string]string{"zone_name": name})
	if err != nil {
		return nil, fmt.Errorf("create-dns-zone failed: %w", err)
	}
	cresp, ok := resp["DNS-create-zoneresponse"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response from create-dns-zone")
	}
	return &DNSZone{
		ID:   str(cresp, "zone_id"),
		Name: name,
	}, nil
}

// DeleteDNSZone removes a DNS zone and all its records.
func (c *Client) DeleteDNSZone(id string) error {
	_, err := c.do("DNS-delete-zone", map[string]string{"zone_id": id})
	if err != nil {
		return fmt.Errorf("delete-dns-zone failed: %w", err)
	}
	return nil
}

// ─── DNS Records ─────────────────────────────────────────────────────────────

// DNSRecord represents a record within an Atlantic.Net DNS zone.
type DNSRecord struct {
	ID       string
	ZoneID   string
	Type     string
	Host     string
	Data     string
	TTL      string
	Priority string
}

// ListDNSRecords returns all records within a zone.
func (c *Client) ListDNSRecords(zoneID string) ([]DNSRecord, error) {
	resp, err := c.do("DNS-list-zone-records", map[string]string{"zone_id": zoneID})
	if err != nil {
		return nil, fmt.Errorf("list-dns-zone-records failed: %w", err)
	}
	rresp, ok := resp["DNS-list-zone-recordsresponse"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response from list-dns-zone-records")
	}
	records, _ := rresp["records"].(map[string]interface{})
	var out []DNSRecord
	for _, v := range records {
		item, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		out = append(out, DNSRecord{
			ID:       str(item, "record_id"),
			ZoneID:   zoneID,
			Type:     str(item, "type"),
			Host:     str(item, "host"),
			Data:     str(item, "data"),
			TTL:      str(item, "ttl"),
			Priority: str(item, "priority"),
		})
	}
	return out, nil
}

// GetDNSRecord returns a single DNS record by ID within a zone.
func (c *Client) GetDNSRecord(zoneID, recordID string) (*DNSRecord, error) {
	records, err := c.ListDNSRecords(zoneID)
	if err != nil {
		return nil, err
	}
	for _, r := range records {
		if r.ID == recordID {
			return &r, nil
		}
	}
	return nil, fmt.Errorf("DNS record %q not found in zone %q", recordID, zoneID)
}

// CreateDNSRecordInput holds parameters for a new DNS record.
type CreateDNSRecordInput struct {
	ZoneID   string
	Type     string
	Host     string
	Data     string
	TTL      string
	Priority string
}

// CreateDNSRecord adds a record to a DNS zone.
func (c *Client) CreateDNSRecord(in CreateDNSRecordInput) (*DNSRecord, error) {
	params := map[string]string{
		"zone_id": in.ZoneID,
		"type":    in.Type,
		"host":    in.Host,
		"data":    in.Data,
		"ttl":     in.TTL,
	}
	if in.Priority != "" {
		params["priority"] = in.Priority
	}
	resp, err := c.do("DNS-create-zone-record", params)
	if err != nil {
		return nil, fmt.Errorf("create-dns-zone-record failed: %w", err)
	}
	cresp, ok := resp["DNS-create-zone-recordresponse"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response from create-dns-zone-record")
	}
	return &DNSRecord{
		ID:       str(cresp, "record_id"),
		ZoneID:   in.ZoneID,
		Type:     in.Type,
		Host:     in.Host,
		Data:     in.Data,
		TTL:      in.TTL,
		Priority: in.Priority,
	}, nil
}

// UpdateDNSRecord modifies an existing DNS record.
func (c *Client) UpdateDNSRecord(in CreateDNSRecordInput, recordID string) (*DNSRecord, error) {
	params := map[string]string{
		"zone_id":   in.ZoneID,
		"record_id": recordID,
		"type":      in.Type,
		"host":      in.Host,
		"data":      in.Data,
		"ttl":       in.TTL,
	}
	if in.Priority != "" {
		params["priority"] = in.Priority
	}
	_, err := c.do("DNS-update-zone-record", params)
	if err != nil {
		return nil, fmt.Errorf("update-dns-zone-record failed: %w", err)
	}
	return c.GetDNSRecord(in.ZoneID, recordID)
}

// DeleteDNSRecord removes a DNS record from a zone.
func (c *Client) DeleteDNSRecord(zoneID, recordID string) error {
	_, err := c.do("DNS-delete-zone-record", map[string]string{
		"zone_id":   zoneID,
		"record_id": recordID,
	})
	if err != nil {
		return fmt.Errorf("delete-dns-zone-record failed: %w", err)
	}
	return nil
}

// ─── Block Storage (SBS) ─────────────────────────────────────────────────────
// Note: Atlantic.Net's SBS API uses separate endpoints.
// The actions below are the documented SBS actions from support documentation.

// BlockVolume represents an Atlantic.Net Scalable Block Storage volume.
type BlockVolume struct {
	ID         string
	Name       string
	Size       string
	Location   string
	Status     string
	InstanceID string
}

// ListBlockVolumes returns all SBS volumes on the account.
func (c *Client) ListBlockVolumes() ([]BlockVolume, error) {
	resp, err := c.do("list-volumes", nil)
	if err != nil {
		return nil, fmt.Errorf("list-volumes failed: %w", err)
	}
	vresp, ok := resp["list-volumesresponse"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response from list-volumes")
	}
	vols, _ := vresp["volumes"].(map[string]interface{})
	var out []BlockVolume
	for _, v := range vols {
		item, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		out = append(out, BlockVolume{
			ID:         str(item, "volume_id"),
			Name:       str(item, "volume_name"),
			Size:       str(item, "size"),
			Location:   str(item, "location"),
			Status:     str(item, "status"),
			InstanceID: str(item, "instance_id"),
		})
	}
	return out, nil
}

// GetBlockVolume returns a single SBS volume by ID.
func (c *Client) GetBlockVolume(id string) (*BlockVolume, error) {
	vols, err := c.ListBlockVolumes()
	if err != nil {
		return nil, err
	}
	for _, v := range vols {
		if v.ID == id {
			return &v, nil
		}
	}
	return nil, fmt.Errorf("block volume %q not found", id)
}

// CreateBlockVolumeInput holds parameters for a new SBS volume.
type CreateBlockVolumeInput struct {
	Name     string
	Size     string // in GB, minimum 50, increments of 50
	Location string
}

// CreateBlockVolume creates a new SBS block storage volume.
func (c *Client) CreateBlockVolume(in CreateBlockVolumeInput) (*BlockVolume, error) {
	resp, err := c.do("create-volume", map[string]string{
		"volume_name": in.Name,
		"size":        in.Size,
		"location":    in.Location,
	})
	if err != nil {
		return nil, fmt.Errorf("create-volume failed: %w", err)
	}
	cresp, ok := resp["create-volumeresponse"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("unexpected response from create-volume")
	}
	volID := str(cresp, "volume_id")
	return c.waitForVolume(volID, "available", 10*time.Minute)
}

// AttachBlockVolume attaches an SBS volume to a Cloud Server.
func (c *Client) AttachBlockVolume(volumeID, instanceID string) error {
	_, err := c.do("attach-volume", map[string]string{
		"volume_id":   volumeID,
		"instance_id": instanceID,
	})
	if err != nil {
		return fmt.Errorf("attach-volume failed: %w", err)
	}
	return nil
}

// DetachBlockVolume detaches an SBS volume from its Cloud Server.
func (c *Client) DetachBlockVolume(volumeID string) error {
	_, err := c.do("detach-volume", map[string]string{
		"volume_id": volumeID,
	})
	if err != nil {
		return fmt.Errorf("detach-volume failed: %w", err)
	}
	return nil
}

// DeleteBlockVolume deletes an SBS volume (must be detached first).
func (c *Client) DeleteBlockVolume(id string) error {
	_, err := c.do("delete-volume", map[string]string{"volume_id": id})
	if err != nil {
		return fmt.Errorf("delete-volume failed: %w", err)
	}
	return nil
}

// waitForVolume polls until the volume reaches the desired status.
func (c *Client) waitForVolume(id, want string, timeout time.Duration) (*BlockVolume, error) {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		vol, err := c.GetBlockVolume(id)
		if err != nil {
			return nil, err
		}
		if strings.EqualFold(vol.Status, want) {
			return vol, nil
		}
		if strings.EqualFold(vol.Status, "error") {
			return nil, fmt.Errorf("volume %s entered error status", id)
		}
		time.Sleep(10 * time.Second)
	}
	return nil, fmt.Errorf("timed out waiting for volume %s to reach %q", id, want)
}
