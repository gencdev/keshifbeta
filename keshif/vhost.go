package keshif

import (
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
)

const UNKNOWN = 0
const EMPTY = 10
const COMMENT = 20
const ADDRESS = 30

// HostFileLine
type HostFileLine struct {
	OriginalLineNum int
	LineType        int
	Address         string
	Parts           []string
	Hostnames       []string
	Raw             string
	Trimed          string
	Comment         string
}

// HostFileLines
type HostFileLines []HostFileLine

type HostsConfig struct {
	ReadFilePath  string
	WriteFilePath string
}

type Hosts struct {
	sync.Mutex
	*HostsConfig
	hostFileLines HostFileLines
}

func ParseHosts(path string) ([]HostFileLine, error) {
	input, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	inputNormalized := strings.Replace(string(input), "\r\n", "\n", -1)

	dataLines := strings.Split(inputNormalized, "\n")

	hostFileLines := make([]HostFileLine, len(dataLines))

	// trim leading an trailing whitespace
	for i, l := range dataLines {
		curLine := &hostFileLines[i]
		curLine.OriginalLineNum = i
		curLine.Raw = l

		// trim line
		curLine.Trimed = strings.TrimSpace(l)

		// check for comment
		if strings.HasPrefix(curLine.Trimed, "#") {
			curLine.LineType = COMMENT
			continue
		}

		if curLine.Trimed == "" {
			curLine.LineType = EMPTY
			continue
		}

		curLineSplit := strings.SplitN(curLine.Trimed, "#", 2)
		if len(curLineSplit) > 1 {
			curLine.Comment = curLineSplit[1]
		}
		curLine.Trimed = curLineSplit[0]

		curLine.Parts = strings.Fields(curLine.Trimed)

		if len(curLine.Parts) > 1 {
			curLine.LineType = ADDRESS
			curLine.Address = strings.ToLower(curLine.Parts[0])
			// lower case all
			for _, p := range curLine.Parts[1:] {
				curLine.Hostnames = append(curLine.Hostnames, strings.ToLower(p))
			}

			continue
		}

		// if we can't figure out what this line is
		// at this point mark it as unknown
		curLine.LineType = UNKNOWN

	}

	return hostFileLines, nil
}

func NewHosts(hc *HostsConfig) (*Hosts, error) {
	h := &Hosts{HostsConfig: hc}
	h.Lock()
	defer h.Unlock()

	defaultHostsFile := "/etc/hosts"

	if h.ReadFilePath == "" {
		h.ReadFilePath = defaultHostsFile
	}

	if h.WriteFilePath == "" {
		h.WriteFilePath = h.ReadFilePath
	}

	hfl, err := ParseHosts(h.ReadFilePath)
	if err != nil {
		return nil, err
	}

	h.hostFileLines = hfl

	return h, nil
}

func NewHostsDefault() (*Hosts, error) {
	return NewHosts(&HostsConfig{})
}

func (h *Hosts) HostAddressLookup(host string) (bool, string, int) {
	h.Lock()
	defer h.Unlock()

	for i, hfl := range h.hostFileLines {
		for _, hn := range hfl.Hostnames {
			//fmt.Println(hn, host)
			if hn == strings.ToLower(host) {
				return true, hfl.Address, i
			}
		}
	}
	return false, "", 0
}

func removeStringElement(slice []string, s int) []string {
	return append(slice[:s], slice[s+1:]...)
}

// removeHFLElement removed an element of a HostFileLine slice
func removeHFLElement(slice []HostFileLine, s int) []HostFileLine {
	return append(slice[:s], slice[s+1:]...)
}

func (h *Hosts) Save() error {
	return h.SaveAs(h.WriteFilePath)
}

// lineFormatter
func lineFormatter(hfl HostFileLine) string {

	if hfl.LineType < ADDRESS {
		return hfl.Raw
	}

	if len(hfl.Comment) > 0 {
		return fmt.Sprintf("%-16s %s #%s", hfl.Address, strings.Join(hfl.Hostnames, " "), hfl.Comment)
	}
	return fmt.Sprintf("%-16s %s", hfl.Address, strings.Join(hfl.Hostnames, " "))
}

func (h *Hosts) RenderHostsFile() string {
	h.Lock()
	defer h.Unlock()

	hf := ""

	for _, hfl := range h.hostFileLines {
		hf = hf + fmt.Sprintln(lineFormatter(hfl))
	}

	return hf
}

// SaveAs saves rendered hosts file to the filename specified
func (h *Hosts) SaveAs(fileName string) error {
	hfData := []byte(h.RenderHostsFile())

	h.Lock()
	defer h.Unlock()

	err := ioutil.WriteFile(fileName, hfData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func (h *Hosts) AddHost(addressRaw string, hostRaw string) {
	host := strings.TrimSpace(strings.ToLower(hostRaw))
	address := strings.TrimSpace(strings.ToLower(addressRaw))

	// does the host already exist
	if ok, exAdd, hflIdx := h.HostAddressLookup(host); ok {
		// if the address is the same we are done
		if address == exAdd {
			return
		}

		// if the hostname is at a different address, go and remove it from the address
		for hidx, hst := range h.hostFileLines[hflIdx].Hostnames {
			if hst == host {
				h.Lock()
				h.hostFileLines[hflIdx].Hostnames = removeStringElement(h.hostFileLines[hflIdx].Hostnames, hidx)
				h.Unlock()

				// remove the address line if empty
				if len(h.hostFileLines[hflIdx].Hostnames) < 1 {
					h.Lock()
					h.hostFileLines = removeHFLElement(h.hostFileLines, hflIdx)
					h.Unlock()
				}

				break // unless we should continue because it could have duplicates
			}
		}
	}

	// if the address exists add it to the address line
	for i, hfl := range h.hostFileLines {
		if hfl.Address == address {
			h.Lock()
			h.hostFileLines[i].Hostnames = append(h.hostFileLines[i].Hostnames, host)
			h.Unlock()
			return
		}
	}
	// the address and host do not already exist so go ahead and create them
	hfl := HostFileLine{
		LineType:  ADDRESS,
		Address:   address,
		Hostnames: []string{host},
	}

	h.Lock()
	h.hostFileLines = append(h.hostFileLines, hfl)
	h.Unlock()
}

func (h *Hosts) RemoveHosts(hosts []string) {
	for _, host := range hosts {
		if h.RemoveFirstHost(host) {
			h.RemoveHost(host)
		}
	}
}

// RemoveHost removes all hostname entries of provided host
func (h *Hosts) RemoveHost(host string) {
	if h.RemoveFirstHost(host) {
		h.RemoveHost(host)
	}
}

// RemoveHost the first hostname entry found and returns true if successful
func (h *Hosts) RemoveFirstHost(host string) bool {
	h.Lock()
	defer h.Unlock()

	for hflIdx := range h.hostFileLines {
		for hidx, hst := range h.hostFileLines[hflIdx].Hostnames {
			if hst == host {
				h.hostFileLines[hflIdx].Hostnames = removeStringElement(h.hostFileLines[hflIdx].Hostnames, hidx)

				// remove the address line if empty
				if len(h.hostFileLines[hflIdx].Hostnames) < 1 {
					h.hostFileLines = removeHFLElement(h.hostFileLines, hflIdx)
				}
				return true
			}
		}
	}

	return false
}

func ClearHosts(routes map[string]Route) {
	host, err := NewHostsDefault()

	if err != nil {
		panic(err)
	}

	hosts := []string{}

	for _, route := range routes {
		hosts = append(hosts, route.Vhost)
	}

	host.RemoveHosts(hosts)
	err = host.Save()

	if err != nil {
		panic(err)
	}
}

func AddHost(routes map[string]Route) {
	hosts, err := NewHostsDefault()

	if err != nil {
		panic(err)
	}

	fmt.Println("----------------------------------------------------------------")
	for k, v := range routes {
		hosts.AddHost(v.Ip, v.Vhost)
		fmt.Printf("name: %s \nvhost: %s \nip: %s \nport: %s\n", k, v.Vhost, v.Ip, v.Port)
		fmt.Println("----------------------------------------------------------------")
	}

	err = hosts.Save()

	if err != nil {
		panic(err)
	}
}
