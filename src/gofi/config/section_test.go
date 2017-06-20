package config

import (
	"strings"
	"testing"
)

func TestInvalidParse(t *testing.T) {
	if _, err := Parse([]byte("\ninput\n\n")); err != ErrInvalid {
		t.Error("Expected ErrInvalid")
	}
}

var basicInput = `
aaa.1.br.devname=br0
aaa.1.devname=ath0
aaa.1.driver=madwifi
aaa.1.eapol_version=2
aaa.1.ssid=5465454564
aaa.1.status=enabled
aaa.1.verbose=2
aaa.1.wpa.1.pairwise=CCMP
aaa.1.wpa.group_rekey=0
aaa.1.wpa.key.1.mgmt=WPA-PSK
aaa.1.wpa.psk=54645654546
aaa.1.wpa=3
aaa.2.br.devname=br0
aaa.2.devname=ath1
aaa.2.driver=madwifi
aaa.2.ssid=vport
aaa.2.status=disabled
   aaa.status=enabled
`

func TestGenerateSort(t *testing.T) {
	comparisons := [][2]string{
		[2]string{"aaa.1.br.devname", "aaa.2.br.devname"},
		[2]string{"aaa.1.devname", "aaa.1.driver"},
		[2]string{"aaa.1.verbose", "aaa.1.wpa.1.pairwise"},
		[2]string{"aaa.1.br.devname", "aaa.2.br.devname"},
		[2]string{"aaa.2.status", "aaa.status"},
	}

	obj, err := Parse([]byte(basicInput))
	if err != nil {
		t.Fatal(err)
	}
	out, err := obj.Serialize()
	if err != nil {
		t.Fatal(err)
	}

	t.Log(out)

	for _, testCase := range comparisons {
		if strings.Index(out, testCase[0]) >= strings.Index(out, testCase[1]) {
			t.Error("Expected", testCase[0], "before", testCase[1], "in output")
		}
	}
}

func TestGenerate(t *testing.T) {
	obj, err := Parse([]byte(basicInput))
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	obj.generate("", &out)
	outStr := strings.Join(out, "\n")

	for _, line := range strings.Split(basicInput, "\n") {
		if !strings.Contains(outStr, strings.TrimSpace(line)) {
			t.Error("Output is missing", line)
		}
	}
}

func TestParse(t *testing.T) {

	obj, err := Parse([]byte(basicInput))
	if err != nil {
		t.Fatal(err)
	}
	//pretty.Print(obj)
	if len(obj.NamedSubs) != 1 {
		t.Fatal("Expected 1 child section, got ", len(obj.NamedSubs))
	}

	a := obj.NamedSubs["aaa"]
	if len(a.NamedSubs) != 3 || !in("1", a) || !in("2", a) || !in("status", a) {
		t.Error("Expected 0,1,status")
	}
	if a.NamedSubs["status"].Value != "enabled" {
		t.Error("Expected aaa.status=enabled")
	}
	if a.NamedSubs["1"].NamedSubs["wpa"].NamedSubs["key"].NamedSubs["1"].NamedSubs["mgmt"].Value != "WPA-PSK" {
		t.Error("Expected aaa.1.wpa.key.1.mgmt=WPA-PSK")
	}
}

func in(a string, m *Section) bool {
	_, ok := m.NamedSubs[a]
	return ok
}
