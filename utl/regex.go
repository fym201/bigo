// Copyright 2013 com authors
//
// Licensed under the Apache License, Version 2.0 (the "License"): you may
// not use this file except in compliance with the License. You may obtain
// a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package utl

import "regexp"

const (
	regex_email_pattern        = `(?i)[A-Z0-9._%+-]+@(?:[A-Z0-9-]+\.)+[A-Z]{2,6}`
	regex_strict_email_pattern = `(?i)[A-Z0-9!#$%&'*+/=?^_{|}~-]+` +
		`(?:\.[A-Z0-9!#$%&'*+/=?^_{|}~-]+)*` +
		`@(?:[A-Z0-9](?:[A-Z0-9-]*[A-Z0-9])?\.)+` +
		`[A-Z0-9](?:[A-Z0-9-]*[A-Z0-9])?`
	regex_url_pattern    = `(ftp|http|https):\/\/(\w+:{0,1}\w*@)?(\S+)(:[0-9]+)?(\/|\/([\w#!:.?+=&%@!\-\/]))?`
	regex_mobile_pattern = `^1(3|5|7|8)\d{9}$`
)

var (
	regex_email        *regexp.Regexp
	regex_strict_email *regexp.Regexp
	regex_url          *regexp.Regexp
	regex_mobile       *regexp.Regexp
)

func init() {
	regex_email = regexp.MustCompile(regex_email_pattern)
	regex_strict_email = regexp.MustCompile(regex_strict_email_pattern)
	regex_url = regexp.MustCompile(regex_url_pattern)
	regex_mobile = regexp.MustCompile(regex_mobile_pattern)
}

// validate string is an email address, if not return false
// basically validation can match 99% cases
func IsEmail(email string) bool {
	return regex_email.MatchString(email)
}

// validate string is an email address, if not return false
// this validation omits RFC 2822
func IsEmailRFC(email string) bool {
	return regex_strict_email.MatchString(email)
}

// validate string is a url link, if not return false
// simple validation can match 99% cases
func IsUrl(url string) bool {
	return regex_url.MatchString(url)
}

// validate string is a mobile phone number, if not return false
func IsMobile(number string) bool {
	return regex_mobile.MatchString(number)
}
