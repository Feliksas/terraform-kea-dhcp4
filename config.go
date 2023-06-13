package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"sync"
)

//SERVICE define which IP version we are use (currently only dhcp4 supported)
var SERVICE []string = []string{"dhcp4"}

// REST API response codes
const (
	KEA_SUCCESS int = iota
	KEA_ERROR
	KEA_UNSUPPORTED
	KEA_EMPTY
)

type configReq struct {
	Command string   `json:"command"`
	Service []string `json:"service"`
}
type configSet struct {
	Command   string    `json:"command"`
	Service   []string  `json:"service"`
	Arguments Arguments `json:"arguments"`
}

//NestedElem elements in root
type NestedElem struct {
	Arguments Arguments `json:"arguments"`
	Result    int       `json:"result"`
}

//Arguments structure
type Arguments struct {
	Dhcp4 Dhcp4 `json:"Dhcp4"`
}

//Response structure
type Response struct {
	Result int    `json:"result"`
	Text   string `json:"text"`
}

//Dhcp4 structure
type Dhcp4 struct {
	Authoritative              bool                    `json:"authoritative"`
	BootFileName               string                  `json:"boot-file-name"`
	CalculateTeeTimes          bool                    `json:"calculate-tee-times"`
	ClientClasses              []ClientClasses         `json:"client-classes,omitempty"`
	ControlSocket              ControlSocket           `json:"control-socket"`
	DDNSGeneratedPrefix        string                  `json:"ddns-generated-prefix,omitempty"`
	DDNSOverrideClientUpdates  bool                    `json:"ddns-override-client-update,omitempty"`
	DDNSOverrideNoUpdate       bool                    `json:"ddns-override-no-update,omitempty"`
	DDNSQualifyingSuffix       string                  `json:"ddns-qualifying-suffix,omitempty"`
	DDNSReplaceClientName      string                  `json:"ddns-replace-client-name,omitempty"`
	DDNSSendUpdates            bool                    `json:"ddns-send-updates,omitempty"`
	DeclineProbationPeriod     int                     `json:"decline-probation-period"`
	DHCPDDNS                   DHCPDDNS                `json:"dhcp-ddns"`
	DHCPQueueControl           DHCPQueueControl        `json:"dhcp-queue-control"`
	DHCP4o6Port                int                     `json:"dhcp4o6-port"`
	EchoClientId               bool                    `json:"echo-client-id"`
	ExpiredLeasesProcessing    ExpiredLeasesProcessing `json:"expired-leases-processing"`
	HooksLibraries             []HooksLibraries        `json:"hooks-libraries"`
	HostReservationIdentifiers []string                `json:"host-reservation-identifiers"`
	InterfacesConfig           InterfacesConfig        `json:"interfaces-config"`
	LeaseDatabase              LeaseDatabase           `json:"lease-database"`
	Loggers                    []Loggers               `json:"loggers,omitempty"`
	MatchClientId              bool                    `json:"match-client-id"`
	NextServer                 string                  `json:"next-server"`
	OptionData                 []OptionData            `json:"option-data"`
	OptionDef                  interface{}             `json:"option-def,omitempty"` // Not implemented
	RebindTimer                int                     `json:"rebind-timer"`
	RenewTimer                 int                     `json:"renew-timer"`
	ReservationMode            string                  `json:"reservation-mode,omitempty"`
	SanityChecks               SanityChecks            `json:"sanity-checks"`
	ServerHostname             string                  `json:"server-hostname"`
	ServerTag                  string                  `json:"server-tag"`
	SharedNetworks             interface{}             `json:"shared-networks,omitempty"` // Not implemented
	Subnet4                    []Subnet4               `json:"subnet4"`
	T1Percent                  float64                 `json:"t1-percent"`
	T2Percent                  float64                 `json:"t2-percent"`
	ValidLifetime              int                     `json:"valid-lifetime"`
}

//ClientClasses structure
type ClientClasses struct {
	Name       string       `json:"name"`
	OptionData []OptionData `json:"option-data"`
	Test       string       `json:"test"`
}

//ControlSocket structure
type ControlSocket struct {
	SocketName string `json:"socket-name"`
	SocketType string `json:"socket-type"`
}

//DHCPDDNS structure
type DHCPDDNS struct {
	EnableUpdates        bool   `json:"enable-updates"`
	GeneratedPrefix      string `json:"generated-prefix"`
	MaxQueueSize         int    `json:"max-queue-size"`
	NCRFormat            string `json:"ncr-format"`
	NCRProtocol          string `json:"ncr-protocol"`
	OverrideClientUpdate bool   `json:"override-client-update"`
	OverrideNoUpdate     bool   `json:"override-no-update"`
	QualifyingSuffix     string `json:"qualifying-suffix"`
	ReplaceClientName    string `json:"replace-client-name"`
	SenderIP             string `json:"sender-ip"`
	SenderPort           int    `json:"sender-port"`
	ServerIP             string `json:"server-ip"`
	ServerPort           int    `json:"server-port"`
}

//DHCPQueueControl structure
type DHCPQueueControl struct {
	Capacity    int    `json:"capacity"`
	EnableQueue bool   `json:"enable-queue"`
	QueueType   string `json:"queue-type"`
}

//ExpiredLeasesProcessing structure
type ExpiredLeasesProcessing struct {
	FlushReclaimedTimerWaitTime int `json:"flush-reclaimed-timer-wait-time"`
	HoldReclaimedTime           int `json:"hold-reclaimed-time"`
	MaxReclaimLeases            int `json:"max-reclaim-leases"`
	MaxReclaimTime              int `json:"max-reclaim-time"`
	ReclaimTimerWaitTime        int `json:"reclaim-timer-wait-time"`
	UnwarnedReclaimCycles       int `json:"unwarned-reclaim-cycles"`
}

//HooksLibraries structure
type HooksLibraries struct {
	Library string `json:"library"`
}

//InterfacesConfig structure
type InterfacesConfig struct {
	Interfaces []string `json:"interfaces"`
	ReDetect   bool     `json:"re-detect"`
}

//LeaseDatabase structure
type LeaseDatabase struct {
	LFCInterval int    `json:"lfc-interval"`
	Name        string `json:"name"`
	Persist     bool   `json:"persist"`
	Type        string `json:"type"`
}

//Loggers structure
type Loggers struct {
	DebugLevel    int             `json:"debuglevel"`
	Name          string          `json:"name"`
	OutputOptions []OutputOptions `json:"output_options"`
	Severity      string          `json:"severity"`
}

//OptionData structure
type OptionData struct {
	AlwaysSend bool   `json:"always-send"`
	Code       int    `json:"code,omitempty"`
	CSVFormat  bool   `json:"csv-format"`
	Data       string `json:"data"`
	Name       string `json:"name,omitempty"`
	Space      string `json:"space"`
}

//OutputOptions structure
type OutputOptions struct {
	Output string `json:"output"`
}

//Pools structure
type Pools struct {
	OptionData []OptionData `json:"option-data"`
	Pool       string       `json:"pool"`
}

//SanityChecks structure
type SanityChecks struct {
	LeaseChecks string `json:"lease-checks"`
}

//Subnet4 structure
type Subnet4 struct {
	FourOverSixInterface   string         `json:"4o6-interface"`
	FourOverSixInterfaceId string         `json:"4o6-interface-id"`
	FourOverSixSubnet      string         `json:"4o6-subnet"`
	Authoritative          bool           `json:"authoritative"`
	CalculateTeeTimes      bool           `json:"calculate-tee-times"`
	Id                     int            `json:"id"`
	MatchClientId          bool           `json:"match-client-id"`
	NextServer             string         `json:"next-server"`
	OptionData             []OptionData   `json:"option-data"`
	Pools                  []Pools        `json:"pools"`
	RebindTimer            int            `json:"rebind-timer"`
	Relay                  interface{}    `json:"relay"`
	RenewTimer             int            `json:"renew-timer"`
	ReservationMode        string         `json:"reservation-mode,omitempty"`
	Reservations           []Reservations `json:"reservations"`
	Subnet                 string         `json:"subnet"`
	T1Percent              float64        `json:"t1-percent"`
	T2Percent              float64        `json:"t2-percent"`
	ValidLifetime          int            `json:"valid-lifetime"`
}

// Relay structure
type Relay struct {
	IPAddresses interface{} `json:"ip-addresses"` // Not implemented
}

//Reservations structure
type Reservations struct {
	BootFileName   string       `json:"boot-file-name"`
	ClientClasses  []string     `json:"client-classes"`
	Hostname       string       `json:"hostname"`
	HWAddress      string       `json:"hw-address"`
	IPAddress      string       `json:"ip-address"`
	NextServer     string       `json:"next-server"`
	OptionData     []OptionData `json:"option-data"`
	ServerHostname string       `json:"server-hostname"`
}

// Config struct for provider
type Config struct {
	Server     string
	Username   string
	Password   string
	Configfile string
}

// Client for connections
type Client struct {
	Config        *Config
	currentConfig []NestedElem
	httpClient    *http.Client
	lock          sync.Mutex
}

func check(e error) {
	if e != nil {
		log.Fatal(e)
		panic(e)
	}
}

//Client In go (c *Config) means this is method assosieted to struct Config
func (c *Config) Client() (*Client, error) {

	log.Println("[INFO] Configuring kea-api client")
	htclient := &http.Client{}
	var data []byte
	services := []string{"dhcp4"}
	jsonStr := configReq{Command: "config-get", Service: services}
	b, err := json.Marshal(jsonStr)
	check(err)
	req, err := http.NewRequest("POST", c.Server, bytes.NewBuffer(b))
	check(err)
	req.SetBasicAuth(c.Username, c.Password)
	req.Header.Set("Content-Type", "application/json")
	resp, err := htclient.Do(req)
	if err != nil {
		log.Fatalf("The HTTP request failed with error %s\n", err)
		return nil, err
	} else if resp.StatusCode >= 400 {
		data, _ = ioutil.ReadAll(resp.Body)
		errtext := fmt.Sprintf("The HTTP request failed with code %d, response was %s\n", resp.StatusCode, data)
		log.Print(errtext)
		return nil, fmt.Errorf(errtext)
	} else {
		data, _ = ioutil.ReadAll(resp.Body)
		var resp []Response
		err := json.Unmarshal(data, &resp)
		if err != nil {
			log.Printf("[Error] Could not unmarshal API response: %s\n", err)
			return nil, err
		}
		if resp[0].Result != KEA_SUCCESS && resp[0].Result != KEA_EMPTY {
			log.Printf("[Error] The HTTP request failed with error %s\n", resp[0].Text)
			return nil, fmt.Errorf(resp[0].Text)
		}
		log.Printf("[INFO] %s\n", string(data))
	}
	log.Println("[INFO] Terminating GET_CONF application...")
	var m []NestedElem
	err = json.Unmarshal(data, &m)
	check(err)

	client := &Client{
		Config:        c,
		currentConfig: m,
		httpClient:    htclient,
	}

	return client, nil
}

//ReadLease method to check if lease exists in kea-dhcpd4
func (c *Client) ReadLease(r Reservations) bool {
	for _, reservation := range c.currentConfig[0].Arguments.Dhcp4.Subnet4[0].Reservations {
		if reservation.Hostname == r.Hostname {
			log.Printf("[DEBUG] read function found the host\n")
			return true
		}
	}
	log.Printf("[DEBUG] read function cannot find the host\n")
	return false
}

//NewLease method to create new lease fot kea-dhcpd4 works!
func (c *Client) NewLease(r Reservations) error {
	var data []byte
	c.currentConfig[0].Arguments.Dhcp4.Subnet4[0].Reservations = append(c.currentConfig[0].Arguments.Dhcp4.Subnet4[0].Reservations, r)
	jsonSet := configSet{Command: "config-set", Service: SERVICE, Arguments: c.currentConfig[0].Arguments}
	enc, err := json.Marshal(jsonSet)
	check(err)
	log.Printf("[DEBUG] Generated JSON: %s\n", enc)
	req, err := http.NewRequest("POST", c.Config.Server, bytes.NewBuffer(enc))
	check(err)
	req.SetBasicAuth(c.Config.Username, c.Config.Password)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("[Error] The HTTP request failed with error %s\n", err)
		return err
	} else if resp.StatusCode >= 400 {
		data, _ = ioutil.ReadAll(resp.Body)
		errtext := fmt.Sprintf("The HTTP request failed with code %d, response was %s\n", resp.StatusCode, data)
		log.Print(errtext)
		return fmt.Errorf(errtext)
	} else {
		data, _ = ioutil.ReadAll(resp.Body)
		var resp []Response
		err := json.Unmarshal(data, &resp)
		if err != nil {
			log.Printf("[Error] Could not unmarshal API response: %s\n", err)
			return err
		}
		if resp[0].Result != KEA_SUCCESS && resp[0].Result != KEA_EMPTY {
			log.Printf("[Error] The HTTP request failed with error %s\n", resp[0].Text)
			return fmt.Errorf(resp[0].Text)
		}
		c.SaveConfig()
		log.Printf("[INFO] New lease shoud be added.. %s\n", string(data))
	}
	return nil
}

// UpdateLease resource works!
func (c *Client) UpdateLease(r Reservations) error {
	var data []byte
	for index, reservation := range c.currentConfig[0].Arguments.Dhcp4.Subnet4[0].Reservations {
		if reservation.Hostname == r.Hostname {
			c.currentConfig[0].Arguments.Dhcp4.Subnet4[0].Reservations[index] = r
		}
	}
	jsonSet := configSet{Command: "config-set", Service: SERVICE, Arguments: c.currentConfig[0].Arguments}
	enc, err := json.Marshal(jsonSet)
	check(err)
	log.Printf("[DEBUG] Generated JSON: %s\n", enc)
	req, err := http.NewRequest("POST", c.Config.Server, bytes.NewBuffer(enc))
	check(err)
	req.SetBasicAuth(c.Config.Username, c.Config.Password)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("[Error] The HTTP request failed with error %s\n", err)
		return err
	} else if resp.StatusCode >= 400 {
		data, _ = ioutil.ReadAll(resp.Body)
		errtext := fmt.Sprintf("The HTTP request failed with code %d, response was %s\n", resp.StatusCode, data)
		log.Print(errtext)
		return fmt.Errorf(errtext)
	} else {
		data, _ = ioutil.ReadAll(resp.Body)
		var resp []Response
		err := json.Unmarshal(data, &resp)
		if err != nil {
			log.Printf("[Error] Could not unmarshal API response: %s\n", err)
			return err
		}
		if resp[0].Result != KEA_SUCCESS && resp[0].Result != KEA_EMPTY {
			log.Printf("[Error] The HTTP request failed with error %s\n", resp[0].Text)
			return fmt.Errorf(resp[0].Text)
		}
		c.SaveConfig()
		log.Printf("[INFO] The lease shoud be updated.. %s\n", string(data))
	}
	return nil

}

// DeleteLease resource
func (c *Client) DeleteLease(r Reservations) error {
	var data []byte
	filteredReservations := make([]Reservations, 0)
	for _, reservation := range c.currentConfig[0].Arguments.Dhcp4.Subnet4[0].Reservations {
		if reservation.Hostname != r.Hostname {
			filteredReservations = append(filteredReservations, reservation)
		}
	}

	c.currentConfig[0].Arguments.Dhcp4.Subnet4[0].Reservations = filteredReservations
	jsonSet := configSet{Command: "config-set", Service: SERVICE, Arguments: c.currentConfig[0].Arguments}
	enc, err := json.Marshal(jsonSet)
	check(err)
	log.Printf("[DEBUG] Generated JSON: %s\n", enc)
	req, err := http.NewRequest("POST", c.Config.Server, bytes.NewBuffer(enc))
	check(err)
	req.SetBasicAuth(c.Config.Username, c.Config.Password)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("[Error] The HTTP request failed with error %s\n", err)
		return err
	} else if resp.StatusCode >= 400 {
		data, _ = ioutil.ReadAll(resp.Body)
		errtext := fmt.Sprintf("The HTTP request failed with code %d, response was %s\n", resp.StatusCode, data)
		log.Print(errtext)
		return fmt.Errorf(errtext)
	} else {
		data, _ = ioutil.ReadAll(resp.Body)
		var resp []Response
		err := json.Unmarshal(data, &resp)
		if err != nil {
			log.Printf("[Error] Could not unmarshal API response: %s\n", err)
			return err
		}
		if resp[0].Result != KEA_SUCCESS && resp[0].Result != KEA_EMPTY {
			log.Printf("[Error] The HTTP request failed with error %s\n", resp[0].Text)
			return fmt.Errorf(resp[0].Text)
		}
		c.SaveConfig()
		log.Printf("[INFO] The lease should be deleted if existed.. %s\n", string(data))
	}
	return nil

}

//SaveConfig method to save a config file for kea-dhcpd4
func (c *Client) SaveConfig() error {
	var data []byte
	jsonWrite := "{ \"command\": \"config-write\", \"service\": [ \"dhcp4\" ], \"arguments\":{\"filename\":\"" + c.Config.Configfile + "\"} }"
	buff := new(bytes.Buffer)
	json.NewEncoder(buff).Encode(jsonWrite)
	req, err := http.NewRequest("POST", c.Config.Server, strings.NewReader(jsonWrite))
	check(err)
	req.SetBasicAuth(c.Config.Username, c.Config.Password)
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("[Error] The HTTP request failed with error %s\n", err)
		return err
	} else if resp.StatusCode >= 400 {
		data, _ = ioutil.ReadAll(resp.Body)
		errtext := fmt.Sprintf("The HTTP request failed with code %d, response was %s\n", resp.StatusCode, data)
		log.Print(errtext)
		return fmt.Errorf(errtext)
	} else {
		data, _ = ioutil.ReadAll(resp.Body)
		var resp []Response
		err := json.Unmarshal(data, &resp)
		if err != nil {
			log.Printf("[Error] Could not unmarshal API response: %s\n", err)
			return err
		}
		if resp[0].Result != KEA_SUCCESS && resp[0].Result != KEA_EMPTY {
			log.Printf("[Error] The HTTP request failed with error %s\n", resp[0].Text)
			return fmt.Errorf(resp[0].Text)
		}
		log.Printf("[INFO] Configuration file should be saved.. %s\n", string(data))
	}
	return nil
}
