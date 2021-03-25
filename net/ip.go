package net

import "regexp"

func IsLeagueIp(ip string) bool {
	ipReg := `^((0|[1-9]\d?|1\d\d|2[0-4]\d|25[0-5])\.){3}(0|[1-9]\d?|1\d\d|2[0-4]\d|25[0-5])$`
	match, _ := regexp.MatchString(ipReg, ip)
	return match
}
