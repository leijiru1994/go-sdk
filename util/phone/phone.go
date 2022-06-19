package phone

import "regexp"

const (
	Pattern4ChinaMobile = "^((13[0-9])|(14[5,7])|(15[0-3,5-9])|(17[0,3,5-8])|(18[0-9])|166|198|199|(147))\\d{8}$"
)

var (
	chinaMobileRegexp *regexp.Regexp
)

func init() {
	chinaMobileRegexp = regexp.MustCompile(Pattern4ChinaMobile)
}

func IsChinaMobile(phone string) (is bool) {
	is = false
	if len(phone) != 11 {
		return
	}

	is = chinaMobileRegexp.MatchString(phone)
	return
}
