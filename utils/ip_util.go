package utils

import (
	"fmt"
	"strconv"
	"strings"
)

// A IntIP ip convert bean
type IntIP struct {
	IP    string
	Intip int
}

func main() {
	x := &IntIP{IP: "192.168.1.99"}
	fmt.Println(x)
	x.ToIntIP()
	fmt.Println(x.Intip)
}

// String to String
func (intIP *IntIP) String() string {
	return intIP.IP
}

// ToIntIP convert IntIP
func (intIP *IntIP) ToIntIP() (int, error) {
	Intip, err := ConvertToIntIP(intIP.IP)
	if err != nil {
		return 0, err
	}
	intIP.Intip = Intip
	return Intip, nil
}

// ToString to string method
func (intIP *IntIP) ToString() (string, error) {
	i4 := intIP.Intip & 255
	i3 := intIP.Intip >> 8 & 255
	i2 := intIP.Intip >> 16 & 255
	i1 := intIP.Intip >> 24 & 255
	if i1 > 255 || i2 > 255 || i3 > 255 || i4 > 255 {
		return "", fmt.Errorf("Isn't a IntIP Type.")
	}
	ipstring := fmt.Sprintf("%d.%d.%d.%d", i4, i3, i2, i1)
	intIP.IP = ipstring
	return ipstring, nil
}

// ConvertToIntIP ConvertToIntIP method
func ConvertToIntIP(ip string) (int, error) {
	ips := strings.Split(ip, ".")
	E := fmt.Errorf("Not A IP.")
	if len(ips) != 4 {
		return 0, E
	}
	var intIP int
	for k, v := range ips {
		i, err := strconv.Atoi(v)
		if err != nil || i > 255 {
			return 0, E
		}
		intIP = intIP | i<<uint(8*(3-k))
	}
	return intIP, nil
}
